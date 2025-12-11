package grpc

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"time"

	"github.com/saurabh/entgo-microservices/gateway/utils"
	pkggrpc "github.com/saurabh/entgo-microservices/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ServiceRegistry manages gRPC connections to backend microservices
type ServiceRegistry struct {
	connections map[string]*grpc.ClientConn
	mu          sync.RWMutex
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		connections: make(map[string]*grpc.ClientConn),
	}
}

// LoadFromConfig loads service configurations and establishes connections
func (r *ServiceRegistry) LoadFromConfig() error {
	// Load service configurations from environment
	serviceConfigs := utils.LoadServicesFromEnv()

	fmt.Println("🔌 Loading gRPC service connections:")

	for _, svc := range serviceConfigs {
		// Convert GraphQL URL to gRPC URL
		// Expected format: {SERVICE}_GRPC_URL env var or derive from service name
		grpcURL := r.getGRPCURL(svc.Name)

		if grpcURL == "" {
			fmt.Printf("  ⚠️  %s: No gRPC URL configured, skipping\n", svc.Name)
			continue
		}

		// Establish connection
		if err := r.Connect(svc.Name, grpcURL); err != nil {
			fmt.Printf("  ❌ %s: Failed to connect to %s: %v\n", svc.Name, grpcURL, err)
			continue
		}

		fmt.Printf("  ✅ %s: Connected to %s\n", svc.Name, grpcURL)
	}

	return nil
}

// getGRPCURL determines the gRPC URL for a service
func (r *ServiceRegistry) getGRPCURL(serviceName string) string {
	// Try environment variable first: {SERVICE}_GRPC_URL
	envVarName := strings.ToUpper(serviceName) + "_GRPC_URL"
	if url := utils.GetEnv(envVarName, ""); url != "" {
		return url
	}

	// Try to derive from docker container name pattern
	// Default pattern: entgo_{service}_dev:{grpc_port}
	switch serviceName {
	case "auth":
		return utils.GetEnv("AUTH_GRPC_URL", "entgo_auth_dev:9081")
	case "route":
		return utils.GetEnv("ROUTE_GRPC_URL", "entgo_route_dev:9082")
	case "attendance":
		return utils.GetEnv("ATTENDANCE_GRPC_URL", "entgo_attendance_dev:9083")
	case "yard":
		return utils.GetEnv("YARD_GRPC_URL", "entgo_yard_dev:9084")
	default:
		// Return empty string if unknown service
		return ""
	}
}

// Connect establishes a gRPC connection to a service
func (r *ServiceRegistry) Connect(serviceName, address string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already connected
	if _, exists := r.connections[serviceName]; exists {
		return nil
	}

	// Create connection with keepalive and interceptors
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to %s at %s: %w", serviceName, address, err)
	}

	// Use context to avoid unused warning
	_ = ctx

	r.connections[serviceName] = conn
	return nil
}

// GetConnection returns a connection for a service by name
func (r *ServiceRegistry) GetConnection(serviceName string) (*grpc.ClientConn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, exists := r.connections[serviceName]
	if !exists {
		return nil, fmt.Errorf("no connection found for service: %s", serviceName)
	}

	return conn, nil
}

// GetConnectionByMethod extracts service name from gRPC method path and returns connection
// Method format: /package.ServiceName/MethodName
func (r *ServiceRegistry) GetConnectionByMethod(method string) (*grpc.ClientConn, string, error) {
	// Parse method to extract service name
	// Format: /user.v1.UserService/GetUserByID -> extract "user" or map to service name
	parts := strings.Split(strings.TrimPrefix(method, "/"), "/")
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("invalid method format: %s", method)
	}

	fullService := parts[0] // e.g., "user.v1.UserService"

	// Extract service name from proto package
	serviceName := r.extractServiceName(fullService)

	conn, err := r.GetConnection(serviceName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get connection for method %s: %w", method, err)
	}

	return conn, serviceName, nil
}

// extractServiceName maps proto service to microservice name using service mapper
func (r *ServiceRegistry) extractServiceName(fullService string) string {
	// Parse proto package format: package.version.Service
	// Example: user.v1.UserService -> extract "user.v1" -> lookup "auth"

	// Extract proto package (everything before the last dot)
	lastDot := strings.LastIndex(fullService, ".")
	if lastDot == -1 {
		return "unknown"
	}

	protoPackage := fullService[:lastDot] // e.g., "user.v1"

	// Use service mapper for dynamic lookup
	mapper := pkggrpc.GetGlobalMapper()
	microservice, err := mapper.GetMicroserviceByProtoPackage(protoPackage)
	if err != nil {
		// Fallback: try to extract first part before dot
		parts := strings.Split(fullService, ".")
		if len(parts) > 0 {
			return strings.ToLower(parts[0])
		}
		return "unknown"
	}

	return microservice
}

// Close closes all connections
func (r *ServiceRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, conn := range r.connections {
		if err := conn.Close(); err != nil {
			fmt.Printf("⚠️  Failed to close connection to %s: %v\n", name, err)
		}
	}

	r.connections = make(map[string]*grpc.ClientConn)
	return nil
}

// GetAllServices returns list of all registered service names
func (r *ServiceRegistry) GetAllServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]string, 0, len(r.connections))
	for name := range r.connections {
		services = append(services, name)
	}
	return services
}
