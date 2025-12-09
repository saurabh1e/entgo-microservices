package grpc

import (
	"context"
	"fmt"

	rolev1 "github.com/saurabh/entgo-microservices/pkg/proto/role/v1"
)

// RoleClient provides helper methods for Role gRPC operations
type RoleClient struct {
	pool        *ClientPool
	serviceAddr string
}

// NewRoleClient creates a new Role service client
func NewRoleClient(pool *ClientPool, serviceAddr string) *RoleClient {
	return &RoleClient{
		pool:        pool,
		serviceAddr: serviceAddr,
	}
}

// GetRoleByID fetches a role by ID
func (c *RoleClient) GetRoleByID(ctx context.Context, roleID int32) (*rolev1.Role, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return nil, err
	}

	client := rolev1.NewRoleServiceClient(conn)
	resp, err := client.GetRoleByID(ctx, &rolev1.GetRoleByIDRequest{
		Id: roleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	return resp.Role, nil
}

// GetRolesByIDs fetches multiple roles by their IDs
func (c *RoleClient) GetRolesByIDs(ctx context.Context, roleIDs []int32) ([]*rolev1.Role, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return nil, err
	}

	client := rolev1.NewRoleServiceClient(conn)
	resp, err := client.GetRolesByIDs(ctx, &rolev1.GetRolesByIDsRequest{
		Ids: roleIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by ids: %w", err)
	}

	return resp.Roles, nil
}
