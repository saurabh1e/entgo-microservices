package grpc

import (
	"context"
	"fmt"

	permissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/permission/v1"
	rolev1 "github.com/saurabh/entgo-microservices/pkg/proto/role/v1"
	rolepermissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/rolepermission/v1"
	userv1 "github.com/saurabh/entgo-microservices/pkg/proto/user/v1"
)

// AuthServiceClient provides access to all models in the Auth microservice
type AuthServiceClient struct {
	gatewayClient *GatewayClient
}

// NewAuthServiceClient creates a new Auth service client
func NewAuthServiceClient(gatewayClient *GatewayClient) *AuthServiceClient {
	return &AuthServiceClient{
		gatewayClient: gatewayClient,
	}
}

// ============================================
// Permission operations
// ============================================

// GetPermissionByID fetches a permission by ID
func (c *AuthServiceClient) GetPermissionByID(ctx context.Context, id int32) (*permissionv1.Permission, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := permissionv1.NewPermissionServiceClient(conn)
	resp, err := client.GetPermissionByID(ctx, &permissionv1.GetPermissionByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by id: %w", err)
	}

	return resp.Permission, nil
}

// GetPermissionsByIDs fetches multiple permissions by their IDs
func (c *AuthServiceClient) GetPermissionsByIDs(ctx context.Context, ids []int32) ([]*permissionv1.Permission, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := permissionv1.NewPermissionServiceClient(conn)
	resp, err := client.GetPermissionsByIDs(ctx, &permissionv1.GetPermissionsByIDsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by ids: %w", err)
	}

	return resp.Permissions, nil
}

// ============================================
// Role operations
// ============================================

// GetRoleByID fetches a role by ID
func (c *AuthServiceClient) GetRoleByID(ctx context.Context, id int32) (*rolev1.Role, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := rolev1.NewRoleServiceClient(conn)
	resp, err := client.GetRoleByID(ctx, &rolev1.GetRoleByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	return resp.Role, nil
}

// GetRolesByIDs fetches multiple roles by their IDs
func (c *AuthServiceClient) GetRolesByIDs(ctx context.Context, ids []int32) ([]*rolev1.Role, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := rolev1.NewRoleServiceClient(conn)
	resp, err := client.GetRolesByIDs(ctx, &rolev1.GetRolesByIDsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by ids: %w", err)
	}

	return resp.Roles, nil
}

// ============================================
// RolePermission operations
// ============================================

// GetRolePermissionByID fetches a rolepermission by ID
func (c *AuthServiceClient) GetRolePermissionByID(ctx context.Context, id int32) (*rolepermissionv1.RolePermission, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := rolepermissionv1.NewRolePermissionServiceClient(conn)
	resp, err := client.GetRolePermissionByID(ctx, &rolepermissionv1.GetRolePermissionByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rolepermission by id: %w", err)
	}

	return resp.Rolepermission, nil
}

// GetRolePermissionsByIDs fetches multiple rolepermissions by their IDs
func (c *AuthServiceClient) GetRolePermissionsByIDs(ctx context.Context, ids []int32) ([]*rolepermissionv1.RolePermission, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := rolepermissionv1.NewRolePermissionServiceClient(conn)
	resp, err := client.GetRolePermissionsByIDs(ctx, &rolepermissionv1.GetRolePermissionsByIDsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rolepermissions by ids: %w", err)
	}

	return resp.Rolepermissions, nil
}

// ============================================
// User operations
// ============================================

// GetUserByID fetches a user by ID
func (c *AuthServiceClient) GetUserByID(ctx context.Context, id int32) (*userv1.User, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := userv1.NewUserServiceClient(conn)
	resp, err := client.GetUserByID(ctx, &userv1.GetUserByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return resp.User, nil
}

// GetUsersByIDs fetches multiple users by their IDs
func (c *AuthServiceClient) GetUsersByIDs(ctx context.Context, ids []int32) ([]*userv1.User, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := userv1.NewUserServiceClient(conn)
	resp, err := client.GetUsersByIDs(ctx, &userv1.GetUsersByIDsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get users by ids: %w", err)
	}

	return resp.Users, nil
}
