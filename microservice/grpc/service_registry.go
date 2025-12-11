package grpc

import (
	"github.com/saurabh/entgo-microservices/microservice/internal/ent"
	dummyv1 "github.com/saurabh/entgo-microservices/pkg/proto/dummy/v1"
	"google.golang.org/grpc"
)

// RegisterAllServices registers all generated gRPC services
func RegisterAllServices(s *grpc.Server, db *ent.Client) {
	dummyv1.RegisterDummyServiceServer(s, NewDummyService(db))
}

// GetServiceNames returns list of all registered service names
func GetServiceNames() []string {
	return []string{
		"DummyService",
	}
}
