package grpc

import (
	"context"
	"fmt"

	permissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/permission/v1"
)

// PermissionClient provides helper methods for Permission gRPC operations
type PermissionClient struct {
	pool        *ClientPool
	serviceAddr string
}

// NewPermissionClient creates a new Permission service client
func NewPermissionClient(pool *ClientPool, serviceAddr string) *PermissionClient {
	return &PermissionClient{
		pool:        pool,
		serviceAddr: serviceAddr,
	}
}

// GetPermissionByID fetches a permission by ID
func (c *PermissionClient) GetPermissionByID(ctx context.Context, permissionID int32) (*permissionv1.Permission, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return nil, err
	}

	client := permissionv1.NewPermissionServiceClient(conn)
	resp, err := client.GetPermissionByID(ctx, &permissionv1.GetPermissionByIDRequest{
		Id: permissionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by id: %w", err)
	}

	return resp.Permission, nil
}

// GetPermissionsByIDs fetches multiple permissions by their IDs
func (c *PermissionClient) GetPermissionsByIDs(ctx context.Context, permissionIDs []int32) ([]*permissionv1.Permission, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return nil, err
	}

	client := permissionv1.NewPermissionServiceClient(conn)
	resp, err := client.GetPermissionsByIDs(ctx, &permissionv1.GetPermissionsByIDsRequest{
		Ids: permissionIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by ids: %w", err)
	}

	return resp.Permissions, nil
}
