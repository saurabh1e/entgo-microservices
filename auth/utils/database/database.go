package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh/entgo-microservices/auth/config"
	"github.com/saurabh/entgo-microservices/auth/internal/ent"

	"github.com/saurabh/entgo-microservices/pkg/logger"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB encapsulates the database connection and client instances.
// It provides a unified interface for database operations.
type DB struct {
	Client *ent.Client
	SqlDB  *sql.DB
	drv    *entsql.Driver // Added to properly manage driver lifecycle
	closed bool
}

// NewPostgresConnection creates a new PostgreSQL connection with connection pooling.
// It configures the connection pool, tests the connection, and initializes the Ent client.
// Returns a DB instance or an error if the connection fails.
func NewPostgresConnection(cfg *config.Config) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	logger.WithFields(map[string]interface{}{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"database": cfg.Database.Name,
	}).Debug("Creating new PostgreSQL connection")

	// Create standard sql.DB for Ent with lib/pq driver
	sqlDB, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		logger.WithError(err).Error("Failed to open database connection")
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool for sql.DB
	logger.WithFields(map[string]interface{}{
		"max_open_conns": cfg.Database.MaxOpenConns,
		"max_idle_conns": cfg.Database.MaxIdleConns,
	}).Debug("Configuring database connection pool")

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Database.ConnMaxIdleTime) * time.Minute)

	// Test the connection with ping
	logger.Debug("Testing database connection with ping")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Database.PingTimeout)*time.Second)
	defer cancel()

	// Simple ping without retry wrapper
	if err = sqlDB.PingContext(ctx); err != nil {
		logger.WithError(err).Error("Database ping failed")
		if closeErr := sqlDB.Close(); closeErr != nil {
			logger.WithError(closeErr).Warn("Failed to close database connection after ping failure")
		}
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	logger.Info("Database connection ping successful")

	// Create Ent driver and client with the sql.DB connection
	logger.Debug("Creating Ent client with PostgreSQL driver")
	drv := entsql.OpenDB(dialect.Postgres, sqlDB)
	entClient := ent.NewClient(ent.Driver(drv))

	logger.Info("PostgreSQL connection and Ent client initialized successfully")

	return &DB{
		Client: entClient,
		SqlDB:  sqlDB,
		drv:    drv,
		closed: false,
	}, nil
}

// Close gracefully closes all database connections.
// It ensures that all resources are properly released.
// Returns an error if any of the close operations fail.
func (db *DB) Close() error {
	if db.closed {
		logger.Debug("Database connections already closed")
		return nil
	}

	logger.Debug("Closing database connections")
	db.closed = true

	var errs []error

	// Close Ent client first
	if db.Client != nil {
		logger.Debug("Closing Ent client")
		if err := db.Client.Close(); err != nil {
			logger.WithError(err).Error("Failed to close Ent client")
			errs = append(errs, fmt.Errorf("ent client close: %w", err))
		}
	}

	// Close SQL database connection
	if db.SqlDB != nil {
		logger.Debug("Closing SQL database connection")
		if err := db.SqlDB.Close(); err != nil {
			logger.WithError(err).Error("Failed to close SQL database connection")
			errs = append(errs, fmt.Errorf("sql db close: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	logger.Info("Database connections closed successfully")
	return nil
}

// Migrate runs database schema migrations.
// It creates or updates the database schema to match the Ent schema definitions.
// Returns an error if the migration fails.
func (db *DB) Migrate(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	logger.Info("Starting database schema migration")
	if err := db.Client.Schema.Create(ctx); err != nil {
		logger.WithError(err).Error("Database migration failed")
		return fmt.Errorf("schema migration failed: %w", err)
	}
	logger.Info("Database schema migration completed successfully")
	return nil
}

// InitializeDatabase initializes the database connection and runs migrations.
// It creates a connection pool, tests connectivity, and applies schema migrations.
// Returns a DB instance or an error if initialization fails.
func InitializeDatabase(cfg *config.Config) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	logger.WithFields(map[string]interface{}{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"user":     cfg.Database.User,
		"database": cfg.Database.Name,
		"ssl_mode": cfg.Database.SSLMode,
	}).Info("Initializing database connection")

	// Establish database connection
	db, err := NewPostgresConnection(cfg)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to database")
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("✅ Database connection established successfully")

	// Run database migrations with proper error handling
	logger.Info("🔧 Running database migrations...")
	ctx := context.Background()

	// Try to create schema, but handle the case where it might already exist
	if err := runMigrations(ctx, db.Client); err != nil {
		logger.WithError(err).Error("Failed to run migrations")
		// Close the connection if migration fails
		if closeErr := db.Close(); closeErr != nil {
			logger.WithError(closeErr).Warn("Failed to close database connection after migration failure")
		}
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("✅ Database migrations completed successfully")

	return db, nil
}

// runMigrations handles database schema migrations with proper error handling.
// It creates the database schema and handles cases where the schema might already exist.
// Returns an error if the migration fails for reasons other than existing schema.
func runMigrations(ctx context.Context, client *ent.Client) error {
	if client == nil {
		return fmt.Errorf("ent client cannot be nil")
	}

	logger.Info("Starting database schema migration...")

	if err := client.Schema.Create(ctx); err != nil {
		// Check if the error is about existing schema objects
		if isSchemaExistsError(err) {
			logger.Warn("Database schema already exists, skipping migration")
			return nil
		}

		logger.WithError(err).Error("Failed to create database schema")
		return fmt.Errorf("failed to create schema: %w", err)
	}

	logger.Info("Schema creation completed successfully")
	return nil
}

// isSchemaExistsError checks if the error indicates the schema already exists.
// It matches common PostgreSQL error patterns for duplicate objects.
// Returns true if the error indicates an existing schema, false otherwise.
func isSchemaExistsError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// PostgreSQL-specific "already exists" error patterns
	return strings.Contains(errStr, "already exists") ||
		(strings.Contains(errStr, "relation") && strings.Contains(errStr, "already exists")) ||
		strings.Contains(errStr, "duplicate key value") ||
		strings.Contains(errStr, "duplicate object")
}
