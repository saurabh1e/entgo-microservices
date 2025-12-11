package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	GraphQL  GraphQLConfig
	Logging  LoggingConfig
}

type AppConfig struct {
	Name           string
	Version        string
	Environment    string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
	PingTimeout     int // in seconds
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type GraphQLConfig struct {
	PlaygroundEnabled bool
	MaxComplexity     int
	MaxDepth          int
}

type LoggingConfig struct {
	Level      string
	LogDir     string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "MyApp"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENVIRONMENT", "development"),
			AllowedOrigins: func() []string {
				v := getEnv("ALLOWED_ORIGINS", "*")
				parts := strings.Split(v, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}
				return parts
			}(),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "discovery_app"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 5),  // 5 minutes default
			ConnMaxIdleTime: getEnvInt("DB_CONN_MAX_IDLE_TIME", 5), // 5 minutes default
			PingTimeout:     getEnvInt("DB_PING_TIMEOUT", 5),       // 5 seconds default
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key"),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout: getEnvInt("SERVER_WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvInt("SERVER_IDLE_TIMEOUT", 60),
		},
		GraphQL: GraphQLConfig{
			PlaygroundEnabled: getEnvBool("GRAPHQL_PLAYGROUND", true),
			MaxComplexity:     getEnvInt("GRAPHQL_MAX_COMPLEXITY", 100),
			MaxDepth:          getEnvInt("GRAPHQL_MAX_DEPTH", 10),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "debug"),
			LogDir:     getEnv("LOG_DIR", "logs"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 5),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 30),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
	}

	return cfg, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func (d DatabaseConfig) PostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	// Validate App config
	if c.App.Name == "" {
		errors = append(errors, "APP_NAME is required")
	}
	if c.App.Environment == "" {
		errors = append(errors, "APP_ENVIRONMENT is required")
	} else {
		validEnvs := []string{"development", "staging", "production"}
		valid := false
		for _, env := range validEnvs {
			if c.App.Environment == env {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, fmt.Sprintf("APP_ENVIRONMENT must be one of: %s", strings.Join(validEnvs, ", ")))
		}
	}

	// Validate Database config
	if c.Database.Host == "" {
		errors = append(errors, "DB_HOST is required")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errors = append(errors, "DB_PORT must be between 1 and 65535")
	}
	if c.Database.User == "" {
		errors = append(errors, "DB_USER is required")
	}
	if c.Database.Name == "" {
		errors = append(errors, "DB_NAME is required")
	}
	if c.Database.MaxOpenConns <= 0 {
		errors = append(errors, "DB_MAX_OPEN_CONNS must be greater than 0")
	}
	if c.Database.MaxIdleConns <= 0 {
		errors = append(errors, "DB_MAX_IDLE_CONNS must be greater than 0")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		errors = append(errors, "DB_MAX_IDLE_CONNS cannot exceed DB_MAX_OPEN_CONNS")
	}

	// Validate Redis config
	if c.Redis.Host == "" {
		errors = append(errors, "REDIS_HOST is required")
	}
	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		errors = append(errors, "REDIS_PORT must be between 1 and 65535")
	}
	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		errors = append(errors, "REDIS_DB must be between 0 and 15")
	}

	// Validate JWT config
	if c.JWT.Secret == "" {
		errors = append(errors, "JWT_SECRET is required and cannot be empty")
	}
	if len(c.JWT.Secret) < 32 {
		errors = append(errors, "JWT_SECRET must be at least 32 characters for security")
	}
	if c.JWT.ExpiryHours <= 0 {
		errors = append(errors, "JWT_EXPIRY_HOURS must be greater than 0")
	}

	// Validate Server config
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, "SERVER_PORT must be between 1 and 65535")
	}
	if c.Server.ReadTimeout <= 0 {
		errors = append(errors, "SERVER_READ_TIMEOUT must be greater than 0")
	}
	if c.Server.WriteTimeout <= 0 {
		errors = append(errors, "SERVER_WRITE_TIMEOUT must be greater than 0")
	}
	if c.Server.IdleTimeout <= 0 {
		errors = append(errors, "SERVER_IDLE_TIMEOUT must be greater than 0")
	}

	// Validate GraphQL config
	if c.GraphQL.MaxComplexity <= 0 {
		errors = append(errors, "GRAPHQL_MAX_COMPLEXITY must be greater than 0")
	}
	if c.GraphQL.MaxDepth <= 0 {
		errors = append(errors, "GRAPHQL_MAX_DEPTH must be greater than 0")
	}

	// Validate Logging config
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	validLevel := false
	for _, level := range validLogLevels {
		if strings.ToLower(c.Logging.Level) == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		errors = append(errors, fmt.Sprintf("LOG_LEVEL must be one of: %s", strings.Join(validLogLevels, ", ")))
	}
	if c.Logging.MaxSize <= 0 {
		errors = append(errors, "LOG_MAX_SIZE must be greater than 0")
	}
	if c.Logging.MaxBackups < 0 {
		errors = append(errors, "LOG_MAX_BACKUPS cannot be negative")
	}
	if c.Logging.MaxAge < 0 {
		errors = append(errors, "LOG_MAX_AGE cannot be negative")
	}

	// Return all validation errors
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
