package grpc

import (
	"fmt"
	"net"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	permissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/permission/v1"
	rolev1 "github.com/saurabh/entgo-microservices/pkg/proto/role/v1"
	userv1 "github.com/saurabh/entgo-microservices/pkg/proto/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	db         *ent.Client
}

func NewServer(db *ent.Client, port int) (*Server, error) {
	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor(),
			RecoveryInterceptor(),
		),
	)

	// Register services
	userv1.RegisterUserServiceServer(grpcServer, NewUserService(db))
	rolev1.RegisterRoleServiceServer(grpcServer, NewRoleService(db))
	permissionv1.RegisterPermissionServiceServer(grpcServer, NewPermissionService(db))

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Enable reflection for grpcurl and other tools
	reflection.Register(grpcServer)

	logger.WithField("port", port).Info("gRPC server initialized")

	return &Server{
		grpcServer: grpcServer,
		listener:   listener,
		db:         db,
	}, nil
}

func (s *Server) Start() error {
	logger.WithField("address", s.listener.Addr().String()).Info("Starting gRPC server")
	return s.grpcServer.Serve(s.listener)
}

func (s *Server) Stop() {
	logger.Info("Stopping gRPC server")
	s.grpcServer.GracefulStop()
}

func (s *Server) Address() string {
	return s.listener.Addr().String()
}
