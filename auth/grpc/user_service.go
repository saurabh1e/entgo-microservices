package grpc

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/user"

	"entgo.io/ent/privacy"
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

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	entity, err := s.db.User.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get user")
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &userv1.GetUserByIDResponse{
		User: convertEntUserToProto(entity),
	}, nil
}

func (s *UserService) GetUsersByIDs(ctx context.Context, req *userv1.GetUsersByIDsRequest) (*userv1.GetUsersByIDsResponse, error) {
	logger.WithField("user_ids", req.Ids).Debug("GetUsersByIDs called")

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	entities, err := s.db.User.Query().
		Where(user.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get users")
		return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	protoUsers := make([]*userv1.User, len(entities))
	for i, e := range entities {
		protoUsers[i] = convertEntUserToProto(e)
	}

	return &userv1.GetUsersByIDsResponse{
		Users: protoUsers,
	}, nil
}

func convertEntUserToProto(e *ent.User) *userv1.User {
	if e == nil {
		return nil
	}

	protoUser := &userv1.User{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// Map additional fields
	protoUser.Email = e.Email
	protoUser.Username = e.Username
	protoUser.PasswordHash = e.PasswordHash
	protoUser.Name = e.Name
	protoUser.Phone = e.Phone
	protoUser.Address = e.Address
	protoUser.UserType = e.UserType
	protoUser.UserCode = e.UserCode
	protoUser.CompanyName = e.CompanyName
	protoUser.CustomerType = e.CustomerType
	protoUser.PaymentTerms = int32(e.PaymentTerms)
	protoUser.IsActive = e.IsActive
	protoUser.EmailVerified = e.EmailVerified
	if e.EmailVerifiedAt != nil && !e.EmailVerifiedAt.IsZero() {
		protoUser.EmailVerifiedAt = timestamppb.New(*e.EmailVerifiedAt)
	}
	if e.LastLogin != nil && !e.LastLogin.IsZero() {
		protoUser.LastLogin = timestamppb.New(*e.LastLogin)
	}

	return protoUser
}
