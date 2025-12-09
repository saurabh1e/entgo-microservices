package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/rolepermission"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	rolepermissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/rolepermission/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RolePermissionService struct {
	rolepermissionv1.UnimplementedRolePermissionServiceServer
	db *ent.Client
}

func NewRolePermissionService(db *ent.Client) *RolePermissionService {
	return &RolePermissionService{db: db}
}

func (s *RolePermissionService) GetRolePermissionByID(ctx context.Context, req *rolepermissionv1.GetRolePermissionByIDRequest) (*rolepermissionv1.GetRolePermissionByIDResponse, error) {
	logger.WithField("rolepermission_id", req.Id).Debug("GetRolePermissionByID called")

	entity, err := s.db.RolePermission.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "rolepermission not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get rolepermission")
		return nil, status.Errorf(codes.Internal, "failed to get rolepermission: %v", err)
	}

	return &rolepermissionv1.GetRolePermissionByIDResponse{
		Rolepermission: convertEntRolePermissionToProto(entity),
	}, nil
}

func (s *RolePermissionService) GetRolePermissionsByIDs(ctx context.Context, req *rolepermissionv1.GetRolePermissionsByIDsRequest) (*rolepermissionv1.GetRolePermissionsByIDsResponse, error) {
	logger.WithField("rolepermission_ids", req.Ids).Debug("GetRolePermissionsByIDs called")

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	entities, err := s.db.RolePermission.Query().
		Where(rolepermission.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get rolepermissions")
		return nil, status.Errorf(codes.Internal, "failed to get rolepermissions: %v", err)
	}

	protoRolePermissions := make([]*rolepermissionv1.RolePermission, len(entities))
	for i, e := range entities {
		protoRolePermissions[i] = convertEntRolePermissionToProto(e)
	}

	return &rolepermissionv1.GetRolePermissionsByIDsResponse{
		Rolepermissions: protoRolePermissions,
	}, nil
}

func convertEntRolePermissionToProto(e *ent.RolePermission) *rolepermissionv1.RolePermission {
	protoRolePermission := &rolepermissionv1.RolePermission{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// TODO: Map additional fields from ent entity to proto message
	// This needs to be manually filled in based on your entity fields

	return protoRolePermission
}
