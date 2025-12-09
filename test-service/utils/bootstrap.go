package utils

import (
	"errors"

	"github.com/saurabh/entgo-microservices/test-service/config"
	"github.com/saurabh/entgo-microservices/test-service/utils/database"

	"github.com/saurabh/entgo-microservices/pkg/jwt"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// Deps groups runtime dependencies initialized at startup
type Deps struct {
	DB         *database.DB
	Redis      *database.RedisClient
	JWTService *jwt.Service
}

// InitializeDependencies sets up DB, Redis and JWT service
func InitializeDependencies(cfg *config.Config) (*Deps, error) {
	var (
		db          *database.DB
		redisClient *database.RedisClient
		jwtService  *jwt.Service
		err         error
	)

	// Ensure resources are cleaned up on error
	defer func() {
		if err != nil {
			if redisClient != nil {
				if cerr := redisClient.Close(); cerr != nil {
					logger.WithError(cerr).Error("Failed to close Redis during cleanup")
				}
			}
			if db != nil {
				if cerr := db.Close(); cerr != nil {
					logger.WithError(cerr).Error("Failed to close DB during cleanup")
				}
			}
		}
	}()

	// Initialize database
	db, err = database.InitializeDatabase(cfg)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize database")
		return nil, err
	}

	// Initialize Redis
	redisClient, err = database.NewRedisClient(cfg)
	if err != nil {
		logger.WithError(err).Error("Failed to initialize Redis")
		return nil, err
	}

	// Defensive nil-check
	if redisClient == nil || redisClient.Client == nil {
		logger.Error("Redis client is nil after initialization")
		err = errors.New("redis client is nil")
		return nil, err
	}

	// Initialize JWT service with Redis and "auth" service name for key namespacing
	jwtService = jwt.NewService(cfg.JWT.Secret, cfg.JWT.ExpiryHours, redisClient.Client, "auth")
	logger.Info("JWT service initialized with Redis token management")

	// Success — cancel deferred cleanup by setting err to nil and returning resources
	return &Deps{DB: db, Redis: redisClient, JWTService: jwtService}, nil
}
