package grpc

import (
	"context"
	"fmt"

	dummyv1 "github.com/saurabh/entgo-microservices/pkg/proto/dummy/v1"
)

// MicroserviceServiceClient provides access to all models in the Microservice microservice
type MicroserviceServiceClient struct {
	gatewayClient *GatewayClient
}

// NewMicroserviceServiceClient creates a new Microservice service client
func NewMicroserviceServiceClient(gatewayClient *GatewayClient) *MicroserviceServiceClient {
	return &MicroserviceServiceClient{
		gatewayClient: gatewayClient,
	}
}

// ============================================
// Dummy operations
// ============================================

// GetDummyByID fetches a dummy by ID
func (c *MicroserviceServiceClient) GetDummyByID(ctx context.Context, id int32) (*dummyv1.Dummy, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := dummyv1.NewDummyServiceClient(conn)
	resp, err := client.GetDummyByID(ctx, &dummyv1.GetDummyByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dummy by id: %w", err)
	}

	return resp.Dummy, nil
}

// GetDummysByIDs fetches multiple dummys by their IDs
func (c *MicroserviceServiceClient) GetDummysByIDs(ctx context.Context, ids []int32) ([]*dummyv1.Dummy, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := dummyv1.NewDummyServiceClient(conn)
	resp, err := client.GetDummysByIDs(ctx, &dummyv1.GetDummysByIDsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get dummys by ids: %w", err)
	}

	return resp.Dummys, nil
}
