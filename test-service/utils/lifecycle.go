package utils

import (
	"context"
	"time"

	"github.com/saurabh/entgo-microservices/test-service/utils/database"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// GRPCServer interface for gRPC server
type GRPCServer interface {
	Stop()
}

// Lifecycle manages startup and shutdown of long-lived resources
type Lifecycle struct {
	DB         *database.DB
	Redis      *database.RedisClient
	Srv        *ServerConfig
	GRPCServer GRPCServer
}

// NewLifecycle creates a new lifecycle manager
func NewLifecycle(db *database.DB, redis *database.RedisClient, srv *ServerConfig) *Lifecycle {
	return &Lifecycle{DB: db, Redis: redis, Srv: srv}
}

// NewLifecycleWithGRPC creates a new lifecycle manager with gRPC support
func NewLifecycleWithGRPC(db *database.DB, redis *database.RedisClient, srv *ServerConfig, grpcSrv GRPCServer) *Lifecycle {
	return &Lifecycle{DB: db, Redis: redis, Srv: srv, GRPCServer: grpcSrv}
}

// Shutdown will attempt to gracefully stop the HTTP server, gRPC server and close DB/Redis
func (l *Lifecycle) Shutdown(ctx context.Context) error {
	logger.Info("Lifecycle: shutting down resources")

	// First shutdown HTTP server
	if l.Srv != nil {
		if err := l.Srv.GracefulShutdown(ctx); err != nil {
			logger.WithError(err).Error("Lifecycle: server graceful shutdown failed")
		}
	}

	// Shutdown gRPC server
	if l.GRPCServer != nil {
		logger.Info("Lifecycle: stopping gRPC server")
		l.GRPCServer.Stop()
	}

	// Close DB
	if l.DB != nil {
		if err := l.DB.Close(); err != nil {
			logger.WithError(err).Error("Lifecycle: failed to close DB")
		}
	}

	// Close Redis
	if l.Redis != nil {
		if err := l.Redis.Close(); err != nil {
			logger.WithError(err).Error("Lifecycle: failed to close Redis")
		}
	}

	// Wait briefly to let deferred logs flush
	time.Sleep(200 * time.Millisecond)

	logger.Info("Lifecycle: shutdown complete")
	return nil
}
