package grpc

import (
	"context"
	"time"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs all gRPC requests
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			}
		}

		logger.WithFields(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration.String(),
			"code":     code.String(),
		}).Info("gRPC request")

		return resp, err
	}
}

// RecoveryInterceptor recovers from panics in gRPC handlers
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithFields(map[string]interface{}{
					"method": info.FullMethod,
					"panic":  r,
				}).Error("Panic in gRPC handler")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
