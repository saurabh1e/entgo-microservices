package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/permission"

	"entgo.io/ent/privacy"
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

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	entity, err := s.db.Permission.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "permission not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get permission")
		return nil, status.Errorf(codes.Internal, "failed to get permission: %v", err)
	}

	return &permissionv1.GetPermissionByIDResponse{
		Permission: convertEntPermissionToProto(entity),
	}, nil
}

func (s *PermissionService) GetPermissionsByIDs(ctx context.Context, req *permissionv1.GetPermissionsByIDsRequest) (*permissionv1.GetPermissionsByIDsResponse, error) {
	logger.WithField("permission_ids", req.Ids).Debug("GetPermissionsByIDs called")

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	entities, err := s.db.Permission.Query().
		Where(permission.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get permissions")
		return nil, status.Errorf(codes.Internal, "failed to get permissions: %v", err)
	}

	protoPermissions := make([]*permissionv1.Permission, len(entities))
	for i, e := range entities {
		protoPermissions[i] = convertEntPermissionToProto(e)
	}

	return &permissionv1.GetPermissionsByIDsResponse{
		Permissions: protoPermissions,
	}, nil
}

func convertEntPermissionToProto(e *ent.Permission) *permissionv1.Permission {
	if e == nil {
		return nil
	}

	protoPermission := &permissionv1.Permission{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// Map additional fields
	protoPermission.Name = e.Name
	protoPermission.DisplayName = e.DisplayName
	protoPermission.Description = e.Description
	protoPermission.Resource = e.Resource
	protoPermission.IsActive = e.IsActive

	return protoPermission
}
