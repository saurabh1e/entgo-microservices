package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/privacy"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/user"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	userv1 "github.com/saurabh/entgo-microservices/pkg/proto/user/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	userv1.UnimplementedUserServiceServer
	db *ent.Client
}

func NewUserService(db *ent.Client) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserByID(ctx context.Context, req *userv1.GetUserByIDRequest) (*userv1.GetUserByIDResponse, error) {
	logger.WithField("user_id", req.Id).Debug("GetUserByID called")

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	userEntity, err := s.db.User.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get user")
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &userv1.GetUserByIDResponse{
		User: convertEntUserToProto(userEntity),
	}, nil
}

func (s *UserService) GetUsersByIDs(ctx context.Context, req *userv1.GetUsersByIDsRequest) (*userv1.GetUsersByIDsResponse, error) {
	logger.WithField("user_ids", req.Ids).Debug("GetUsersByIDs called")

	// Bypass privacy rules for gRPC internal service calls
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	users, err := s.db.User.Query().
		Where(user.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get users")
		return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	protoUsers := make([]*userv1.User, len(users))
	for i, u := range users {
		protoUsers[i] = convertEntUserToProto(u)
	}

	return &userv1.GetUsersByIDsResponse{
		Users: protoUsers,
	}, nil
}

func (s *UserService) ValidateUser(ctx context.Context, req *userv1.ValidateUserRequest) (*userv1.ValidateUserResponse, error) {
	logger.WithField("user_id", req.UserId).Debug("ValidateUser called")

	// Bypass privacy rules for gRPC internal service calls
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	userEntity, err := s.db.User.Get(ctx, int(req.UserId))
	if err != nil {
		if ent.IsNotFound(err) {
			return &userv1.ValidateUserResponse{Valid: false}, nil
		}
		logger.WithError(err).Error("Failed to validate user")
		return nil, status.Errorf(codes.Internal, "failed to validate user: %v", err)
	}

	return &userv1.ValidateUserResponse{
		Valid: true,
		User:  convertEntUserToProto(userEntity),
	}, nil
}

func convertEntUserToProto(u *ent.User) *userv1.User {
	protoUser := &userv1.User{
		Id:        int32(u.ID),
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}

	if u.TenantID != nil {
		tenantID := int32(*u.TenantID)
		protoUser.TenantId = &tenantID
	}

	return protoUser
}
