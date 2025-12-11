package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// ProxyServer is a transparent gRPC proxy that forwards requests to backend services
type ProxyServer struct {
	grpcServer *grpc.Server
	listener   net.Listener
	registry   *ServiceRegistry
}

// NewProxyServer creates a new gRPC proxy server
func NewProxyServer(port int) (*ProxyServer, error) {
	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	// Initialize service registry
	registry := NewServiceRegistry()
	if err := registry.LoadFromConfig(); err != nil {
		return nil, fmt.Errorf("failed to load service registry: %w", err)
	}

	// Create proxy server with interceptors
	proxy := &ProxyServer{
		registry: registry,
		listener: listener,
	}

	// Create gRPC server with unknown service handler (transparent proxy)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			proxy.loggingInterceptor,
			proxy.recoveryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			proxy.streamLoggingInterceptor,
			proxy.streamRecoveryInterceptor,
		),
		grpc.UnknownServiceHandler(proxy.transparentHandler()),
	)

	proxy.grpcServer = grpcServer

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Enable reflection for debugging with grpcurl
	reflection.Register(grpcServer)

	log.Printf("🔌 gRPC proxy server initialized on port %d", port)
	log.Printf("📡 Registered services: %v", registry.GetAllServices())

	return proxy, nil
}

// transparentHandler creates a handler that forwards all unknown services
func (p *ProxyServer) transparentHandler() grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		// Get method from stream context
		method, ok := grpc.MethodFromServerStream(stream)
		if !ok {
			return status.Error(codes.Internal, "failed to get method from stream")
		}

		// Get backend connection
		backendConn, serviceName, err := p.registry.GetConnectionByMethod(method)
		if err != nil {
			log.Printf("❌ Failed to route request: %v", err)
			return status.Errorf(codes.Unavailable, "service unavailable: %v", err)
		}

		// Get metadata from incoming request
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			md = metadata.New(nil)
		}

		// Create outgoing context with metadata
		outCtx := metadata.NewOutgoingContext(stream.Context(), md)

		// Create client stream to backend
		clientStream, err := backendConn.NewStream(outCtx, &grpc.StreamDesc{
			StreamName:    method,
			ServerStreams: true,
			ClientStreams: true,
		}, method)
		if err != nil {
			log.Printf("❌ Failed to create stream to backend: %v", err)
			return status.Errorf(codes.Internal, "failed to create backend stream: %v", err)
		}

		// Forward request and response
		return p.forwardStream(stream, clientStream, serviceName, method)
	}
}

// forwardStream bidirectionally forwards data between client and backend
func (p *ProxyServer) forwardStream(serverStream grpc.ServerStream, clientStream grpc.ClientStream, serviceName, method string) error {
	// Channel for errors
	errChan := make(chan error, 2)

	// Forward client -> backend (request)
	go func() {
		for {
			var msg interface{}
			if err := serverStream.RecvMsg(&msg); err != nil {
				if err == io.EOF {
					_ = clientStream.CloseSend()
					errChan <- nil
					return
				}
				errChan <- status.Errorf(codes.Internal, "failed to receive from client: %v", err)
				return
			}

			if err := clientStream.SendMsg(msg); err != nil {
				errChan <- status.Errorf(codes.Internal, "failed to send to backend: %v", err)
				return
			}
		}
	}()

	// Forward backend -> client (response)
	go func() {
		for {
			var msg interface{}
			if err := clientStream.RecvMsg(&msg); err != nil {
				if err == io.EOF {
					errChan <- nil
					return
				}
				errChan <- status.Errorf(codes.Internal, "failed to receive from backend: %v", err)
				return
			}

			if err := serverStream.SendMsg(msg); err != nil {
				errChan <- status.Errorf(codes.Internal, "failed to send to client: %v", err)
				return
			}
		}
	}()

	// Wait for first error or completion
	err := <-errChan

	// Log completion
	if err != nil {
		log.Printf("❌ [%s] %s: %v", serviceName, method, err)
	} else {
		log.Printf("✅ [%s] %s: completed", serviceName, method)
	}

	return err
}

// loggingInterceptor logs unary RPC calls
func (p *ProxyServer) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("📨 gRPC Proxy: %s", info.FullMethod)

	resp, err := handler(ctx, req)

	if err != nil {
		log.Printf("❌ gRPC Proxy Error: %s: %v", info.FullMethod, err)
	}

	return resp, err
}

// streamLoggingInterceptor logs streaming RPC calls
func (p *ProxyServer) streamLoggingInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Printf("📨 gRPC Proxy Stream: %s", info.FullMethod)

	err := handler(srv, stream)

	if err != nil {
		log.Printf("❌ gRPC Proxy Stream Error: %s: %v", info.FullMethod, err)
	}

	return err
}

// recoveryInterceptor recovers from panics in unary calls
func (p *ProxyServer) recoveryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 Panic recovered in gRPC proxy: %v", r)
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}

// streamRecoveryInterceptor recovers from panics in streaming calls
func (p *ProxyServer) streamRecoveryInterceptor(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 Panic recovered in gRPC stream proxy: %v", r)
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()

	return handler(srv, stream)
}

// Start starts the gRPC proxy server
func (p *ProxyServer) Start() error {
	log.Printf("🚀 Starting gRPC proxy server on %s", p.listener.Addr().String())
	return p.grpcServer.Serve(p.listener)
}

// Stop gracefully stops the gRPC proxy server
func (p *ProxyServer) Stop() {
	log.Println("🛑 Stopping gRPC proxy server...")
	p.grpcServer.GracefulStop()
	if err := p.registry.Close(); err != nil {
		log.Printf("⚠️  Error closing registry: %v", err)
	}
}

// Address returns the listening address
func (p *ProxyServer) Address() string {
	return p.listener.Addr().String()
}
