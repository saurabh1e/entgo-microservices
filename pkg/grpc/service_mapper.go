package grpc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ServiceMapping represents the mapping between proto package and microservice
type ServiceMapping struct {
	Service          string `json:"service"`           // e.g., "user.v1.UserService"
	ProtoPackage     string `json:"proto_package"`     // e.g., "user.v1"
	MicroserviceName string `json:"microservice_name"` // e.g., "auth"
	EntityName       string `json:"entity_name"`       // e.g., "User"
}

// ServiceMapper provides dynamic service-to-microservice mapping
type ServiceMapper struct {
	mappings map[string]*ServiceMapping // key: proto package prefix (e.g., "user.v1")
	mu       sync.RWMutex
}

var (
	globalMapper     *ServiceMapper
	globalMapperOnce sync.Once
)

// GetGlobalMapper returns the global service mapper instance
func GetGlobalMapper() *ServiceMapper {
	globalMapperOnce.Do(func() {
		globalMapper = NewServiceMapper()
		// Load mappings from metadata files
		if err := globalMapper.LoadFromMetadata(); err != nil {
			fmt.Printf("⚠️  Failed to load service mappings: %v\n", err)
		}
	})
	return globalMapper
}

// NewServiceMapper creates a new service mapper
func NewServiceMapper() *ServiceMapper {
	return &ServiceMapper{
		mappings: make(map[string]*ServiceMapping),
	}
}

// LoadFromMetadata loads service mappings from generated metadata files
func (m *ServiceMapper) LoadFromMetadata() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Look for metadata files in pkg/grpc/metadata directory
	metadataDir := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "saurabh", "entgo-microservices", "pkg", "grpc", "metadata")

	// If GOPATH not set or dir doesn't exist, try relative path
	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		// Try to find it relative to this file
		metadataDir = "./metadata"
	}

	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		// If metadata dir doesn't exist yet, use hardcoded fallback
		m.loadFallbackMappings()
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_services.json") {
			continue
		}

		filePath := filepath.Join(metadataDir, entry.Name())
		if err := m.loadMetadataFile(filePath); err != nil {
			fmt.Printf("⚠️  Failed to load metadata from %s: %v\n", entry.Name(), err)
			continue
		}
	}

	// If no mappings loaded, use fallback
	if len(m.mappings) == 0 {
		m.loadFallbackMappings()
	}

	return nil
}

// loadMetadataFile loads service mappings from a single metadata file
func (m *ServiceMapper) loadMetadataFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var mappings []ServiceMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	for _, mapping := range mappings {
		// Store by proto package for efficient lookup
		m.mappings[mapping.ProtoPackage] = &ServiceMapping{
			Service:          mapping.Service,
			ProtoPackage:     mapping.ProtoPackage,
			MicroserviceName: mapping.MicroserviceName,
			EntityName:       mapping.EntityName,
		}
	}

	return nil
}

// loadFallbackMappings provides hardcoded mappings as fallback
func (m *ServiceMapper) loadFallbackMappings() {
	// Auth service entities
	authEntities := []string{"user", "role", "permission", "rolepermission"}
	for _, entity := range authEntities {
		m.mappings[entity+".v1"] = &ServiceMapping{
			Service:          fmt.Sprintf("%s.v1.%sService", entity, strings.Title(entity)),
			ProtoPackage:     entity + ".v1",
			MicroserviceName: "auth",
			EntityName:       strings.Title(entity),
		}
	}

	// Other services (add more as needed)
	fallbackServices := map[string]string{
		"dummy": "attendance",
		// Add more entity -> service mappings here
	}

	for entity, service := range fallbackServices {
		m.mappings[entity+".v1"] = &ServiceMapping{
			Service:          fmt.Sprintf("%s.v1.%sService", entity, strings.Title(entity)),
			ProtoPackage:     entity + ".v1",
			MicroserviceName: service,
			EntityName:       strings.Title(entity),
		}
	}
}

// GetMicroserviceByProtoPackage returns the microservice name for a proto package
func (m *ServiceMapper) GetMicroserviceByProtoPackage(protoPackage string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mapping, exists := m.mappings[protoPackage]
	if !exists {
		return "", fmt.Errorf("no microservice mapping found for proto package: %s", protoPackage)
	}

	return mapping.MicroserviceName, nil
}

// GetMicroserviceByMethod extracts the proto package from a gRPC method and returns the microservice
// Method format: /package.ServiceName/MethodName
func (m *ServiceMapper) GetMicroserviceByMethod(method string) (string, error) {
	// Parse method: /user.v1.UserService/GetUserByID -> extract "user.v1"
	parts := strings.Split(strings.TrimPrefix(method, "/"), "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid method format: %s", method)
	}

	fullService := parts[0] // e.g., "user.v1.UserService"

	// Extract proto package (everything before the last dot)
	lastDot := strings.LastIndex(fullService, ".")
	if lastDot == -1 {
		return "", fmt.Errorf("invalid service format: %s", fullService)
	}

	protoPackage := fullService[:lastDot] // e.g., "user.v1"

	return m.GetMicroserviceByProtoPackage(protoPackage)
}

// GetMapping returns the full mapping for a proto package
func (m *ServiceMapper) GetMapping(protoPackage string) (*ServiceMapping, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mapping, exists := m.mappings[protoPackage]
	if !exists {
		return nil, fmt.Errorf("no mapping found for proto package: %s", protoPackage)
	}

	return mapping, nil
}

// GetAllMappings returns all service mappings
func (m *ServiceMapper) GetAllMappings() map[string]*ServiceMapping {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ServiceMapping, len(m.mappings))
	for k, v := range m.mappings {
		result[k] = v
	}
	return result
}

// AddMapping manually adds a service mapping (useful for testing or dynamic services)
func (m *ServiceMapper) AddMapping(protoPackage, microservice, entityName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mappings[protoPackage] = &ServiceMapping{
		Service:          fmt.Sprintf("%s.%sService", protoPackage, entityName),
		ProtoPackage:     protoPackage,
		MicroserviceName: microservice,
		EntityName:       entityName,
	}
}
