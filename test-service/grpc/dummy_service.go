package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/test-service/internal/ent"
	"github.com/saurabh/entgo-microservices/test-service/internal/ent/dummy"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	dummyv1 "github.com/saurabh/entgo-microservices/pkg/proto/dummy/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DummyService struct {
	dummyv1.UnimplementedDummyServiceServer
	db *ent.Client
}

func NewDummyService(db *ent.Client) *DummyService {
	return &DummyService{db: db}
}

func (s *DummyService) GetDummyByID(ctx context.Context, req *dummyv1.GetDummyByIDRequest) (*dummyv1.GetDummyByIDResponse, error) {
	logger.WithField("dummy_id", req.Id).Debug("GetDummyByID called")

	entity, err := s.db.Dummy.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "dummy not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get dummy")
		return nil, status.Errorf(codes.Internal, "failed to get dummy: %v", err)
	}

	return &dummyv1.GetDummyByIDResponse{
		Dummy: convertEntDummyToProto(entity),
	}, nil
}

func (s *DummyService) GetDummysByIDs(ctx context.Context, req *dummyv1.GetDummysByIDsRequest) (*dummyv1.GetDummysByIDsResponse, error) {
	logger.WithField("dummy_ids", req.Ids).Debug("GetDummysByIDs called")

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	entities, err := s.db.Dummy.Query().
		Where(dummy.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get dummys")
		return nil, status.Errorf(codes.Internal, "failed to get dummys: %v", err)
	}

	protoDummys := make([]*dummyv1.Dummy, len(entities))
	for i, e := range entities {
		protoDummys[i] = convertEntDummyToProto(e)
	}

	return &dummyv1.GetDummysByIDsResponse{
		Dummys: protoDummys,
	}, nil
}

func convertEntDummyToProto(e *ent.Dummy) *dummyv1.Dummy {
	protoDummy := &dummyv1.Dummy{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// TODO: Map additional fields from ent entity to proto message
	// This needs to be manually filled in based on your entity fields

	return protoDummy
}
