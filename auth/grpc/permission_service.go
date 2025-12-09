package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/permission"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/privacy"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	permissionv1 "github.com/saurabh/entgo-microservices/pkg/proto/permission/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PermissionService struct {
	permissionv1.UnimplementedPermissionServiceServer
	db *ent.Client
}

func NewPermissionService(db *ent.Client) *PermissionService {
	return &PermissionService{db: db}
}

func (s *PermissionService) GetPermissionByID(ctx context.Context, req *permissionv1.GetPermissionByIDRequest) (*permissionv1.GetPermissionByIDResponse, error) {
	logger.WithField("permission_id", req.Id).Debug("GetPermissionByID called")

	// Bypass privacy rules for gRPC internal service calls
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	permissionEntity, err := s.db.Permission.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "permission not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get permission")
		return nil, status.Errorf(codes.Internal, "failed to get permission: %v", err)
	}

	return &permissionv1.GetPermissionByIDResponse{
		Permission: convertEntPermissionToProto(permissionEntity),
	}, nil
}

func (s *PermissionService) GetPermissionsByIDs(ctx context.Context, req *permissionv1.GetPermissionsByIDsRequest) (*permissionv1.GetPermissionsByIDsResponse, error) {
	logger.WithField("permission_ids", req.Ids).Debug("GetPermissionsByIDs called")

	// Bypass privacy rules for gRPC internal service calls
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	permissions, err := s.db.Permission.Query().
		Where(permission.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get permissions")
		return nil, status.Errorf(codes.Internal, "failed to get permissions: %v", err)
	}

	protoPermissions := make([]*permissionv1.Permission, len(permissions))
	for i, p := range permissions {
		protoPermissions[i] = convertEntPermissionToProto(p)
	}

	return &permissionv1.GetPermissionsByIDsResponse{
		Permissions: protoPermissions,
	}, nil
}

func convertEntPermissionToProto(p *ent.Permission) *permissionv1.Permission {
	return &permissionv1.Permission{
		Id:          int32(p.ID),
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
