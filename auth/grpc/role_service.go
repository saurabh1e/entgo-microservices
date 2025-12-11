package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/role"

	"entgo.io/ent/privacy"
	"github.com/saurabh/entgo-microservices/pkg/logger"
	rolev1 "github.com/saurabh/entgo-microservices/pkg/proto/role/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RoleService struct {
	rolev1.UnimplementedRoleServiceServer
	db *ent.Client
}

func NewRoleService(db *ent.Client) *RoleService {
	return &RoleService{db: db}
}

func (s *RoleService) GetRoleByID(ctx context.Context, req *rolev1.GetRoleByIDRequest) (*rolev1.GetRoleByIDResponse, error) {
	logger.WithField("role_id", req.Id).Debug("GetRoleByID called")

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	entity, err := s.db.Role.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get role")
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}

	return &rolev1.GetRoleByIDResponse{
		Role: convertEntRoleToProto(entity),
	}, nil
}

func (s *RoleService) GetRolesByIDs(ctx context.Context, req *rolev1.GetRolesByIDsRequest) (*rolev1.GetRolesByIDsResponse, error) {
	logger.WithField("role_ids", req.Ids).Debug("GetRolesByIDs called")

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	entities, err := s.db.Role.Query().
		Where(role.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get roles")
		return nil, status.Errorf(codes.Internal, "failed to get roles: %v", err)
	}

	protoRoles := make([]*rolev1.Role, len(entities))
	for i, e := range entities {
		protoRoles[i] = convertEntRoleToProto(e)
	}

	return &rolev1.GetRolesByIDsResponse{
		Roles: protoRoles,
	}, nil
}

func convertEntRoleToProto(e *ent.Role) *rolev1.Role {
	if e == nil {
		return nil
	}

	protoRole := &rolev1.Role{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// Map additional fields
	protoRole.Name = e.Name
	protoRole.DisplayName = e.DisplayName
	protoRole.Description = e.Description
	protoRole.IsActive = e.IsActive
	protoRole.Priority = int32(e.Priority)

	return protoRole
}
