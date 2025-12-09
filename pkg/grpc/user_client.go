package grpc

import (
	"context"
	"fmt"

	userv1 "github.com/saurabh/entgo-microservices/pkg/proto/user/v1"
)

// UserClient provides helper methods for User gRPC operations
type UserClient struct {
	pool        *ClientPool
	serviceAddr string
}

// NewUserClient creates a new User service client
func NewUserClient(pool *ClientPool, serviceAddr string) *UserClient {
	return &UserClient{
		pool:        pool,
		serviceAddr: serviceAddr,
	}
}

// GetUserByID fetches a user by ID
func (c *UserClient) GetUserByID(ctx context.Context, userID int32) (*userv1.User, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return nil, err
	}

	client := userv1.NewUserServiceClient(conn)
	resp, err := client.GetUserByID(ctx, &userv1.GetUserByIDRequest{
		Id: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return resp.User, nil
}

// GetUsersByIDs fetches multiple users by their IDs
func (c *UserClient) GetUsersByIDs(ctx context.Context, userIDs []int32) ([]*userv1.User, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return nil, err
	}

	client := userv1.NewUserServiceClient(conn)
	resp, err := client.GetUsersByIDs(ctx, &userv1.GetUsersByIDsRequest{
		Ids: userIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get users by ids: %w", err)
	}

	return resp.Users, nil
}

// ValidateUser validates if a user exists and is active
func (c *UserClient) ValidateUser(ctx context.Context, userID int32) (bool, *userv1.User, error) {
	conn, err := c.pool.GetConnection(ctx, c.serviceAddr)
	if err != nil {
		return false, nil, err
	}

	client := userv1.NewUserServiceClient(conn)
	resp, err := client.ValidateUser(ctx, &userv1.ValidateUserRequest{
		UserId: userID,
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to validate user: %w", err)
	}

	return resp.Valid, resp.User, nil
}
