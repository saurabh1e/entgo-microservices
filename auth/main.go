package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/saurabh/entgo-microservices/auth/config"
	"github.com/saurabh/entgo-microservices/auth/grpc"
	"github.com/saurabh/entgo-microservices/auth/utils"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

import _ "github.com/saurabh/entgo-microservices/auth/internal/ent/runtime"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration before proceeding
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed:\n%v", err)
	}

	// Initialize logger
	logConfig := logger.LogConfig{
		Level:      cfg.Logging.Level,
		LogDir:     cfg.Logging.LogDir,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   cfg.Logging.Compress,
	}
	if err := logger.InitLogger(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("🚀 Starting Auth Service...")
	logger.WithFields(map[string]interface{}{
		"log_level": cfg.Logging.Level,
		"log_dir":   cfg.Logging.LogDir,
	}).Info("Logger initialized successfully")

	// Initialize dependencies (DB, Redis, JWT)
	deps, err := utils.InitializeDependencies(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize dependencies")
	}

	// Initialize and start HTTP server with JWT service
	server := utils.InitializeServer(cfg, deps.DB, deps.JWTService, deps.Redis)

	// Initialize gRPC server
	grpcPort := cfg.Server.Port + 1000 // Default: 9081 if HTTP is 8081
	if os.Getenv("GRPC_PORT") != "" {
		// Override from env if set
	}
	grpcServer, err := grpc.NewServer(deps.DB.Client, grpcPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize gRPC server")
	}

	// Start gRPC server in background
	go func() {
		logger.WithField("port", grpcPort).Info("Starting gRPC server")
		if err := grpcServer.Start(); err != nil {
			logger.WithError(err).Fatal("gRPC server failed")
		}
	}()

	// Create lifecycle manager (include gRPC server shutdown)
	lifecycle := utils.NewLifecycleWithGRPC(deps.DB, deps.Redis, server, grpcServer)

	// Setup graceful shutdown
	setupGracefulShutdown(lifecycle)

	// Start server
	if err := server.StartServer(); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}

func setupGracefulShutdown(lifecycle *utils.Lifecycle) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutdown signal received...")

		ctx := context.Background()
		if err := lifecycle.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Failed during graceful shutdown")
		}

		os.Exit(0)
	}()
}
