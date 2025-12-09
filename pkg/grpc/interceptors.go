package grpc

import (
	"context"
	"time"

	"github.com/saurabh/entgo-microservices/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientRetryInterceptor retries failed requests with exponential backoff
func ClientRetryInterceptor(maxRetries int, initialDelay time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var err error
		delay := initialDelay

		for attempt := 0; attempt <= maxRetries; attempt++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			// Check if error is retryable
			if st, ok := status.FromError(err); ok {
				switch st.Code() {
				case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
					// Retryable errors
					if attempt < maxRetries {
						logger.WithFields(map[string]interface{}{
							"method":  method,
							"attempt": attempt + 1,
							"delay":   delay.String(),
						}).Warn("Retrying gRPC call")

						time.Sleep(delay)
						delay *= 2 // Exponential backoff
						continue
					}
				}
			}
			break
		}
		return err
	}
}

// ClientLoggingInterceptor logs gRPC client calls
func ClientLoggingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		fields := map[string]interface{}{
			"method":   method,
			"duration": duration.String(),
		}

		if err != nil {
			if st, ok := status.FromError(err); ok {
				fields["code"] = st.Code().String()
			}
			logger.WithFields(fields).Debug("gRPC client call failed")
		} else {
			logger.WithFields(fields).Debug("gRPC client call succeeded")
		}

		return err
	}
}
