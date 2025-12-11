package grpc

import (
	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	permissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/permission/v1"
	rolev1 "github.com/saurabh/entgo-microservices/pkg/proto/role/v1"
	rolepermissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/rolepermission/v1"
	userv1 "github.com/saurabh/entgo-microservices/pkg/proto/user/v1"
	"google.golang.org/grpc"
)

// RegisterAllServices registers all generated gRPC services
func RegisterAllServices(s *grpc.Server, db *ent.Client) {
	permissionv1.RegisterPermissionServiceServer(s, NewPermissionService(db))
	rolev1.RegisterRoleServiceServer(s, NewRoleService(db))
	rolepermissionv1.RegisterRolePermissionServiceServer(s, NewRolePermissionService(db))
	userv1.RegisterUserServiceServer(s, NewUserService(db))
}

// GetServiceNames returns list of all registered service names
func GetServiceNames() []string {
	return []string{
		"PermissionService",
		"RoleService",
		"RolePermissionService",
		"UserService",
	}
}
