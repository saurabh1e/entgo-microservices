package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the gateway service
type Config struct {
	Port           string
	RedisHost      string
	RedisPort      string
	RedisPassword  string
	RedisDB        int
	AuthServiceURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	redisHost := GetEnv("REDIS_HOST", "localhost")
	redisPort := GetEnv("REDIS_PORT", "6379")

	return &Config{
		Port:           GetEnv("PORT", "8080"),
		RedisHost:      redisHost,
		RedisPort:      redisPort,
		RedisPassword:  GetEnv("REDIS_PASSWORD", ""),
		RedisDB:        GetEnvInt("REDIS_DB", 0),
		AuthServiceURL: GetEnv("AUTH_SERVICE_URL", "http://localhost:8081/graphql"),
	}
}

// LoadServicesFromEnv dynamically loads service configurations from environment variables
// Reads GATEWAY_SERVICES env var as comma-separated list of service names
// For each service, reads {SERVICE_NAME}_SERVICE_URL environment variable
func LoadServicesFromEnv() []ServiceConfig {
	// Try loading from file first (easier to manage)
	if services, err := LoadServicesFromFile("services.conf"); err == nil && len(services) > 0 {
		return services
	}

	// Fall back to environment variables
	// Get comma-separated list of services (e.g., "auth,microservice-1,user-service")
	servicesStr := GetEnv("GATEWAY_SERVICES", "auth")
	serviceNames := parseServiceNames(servicesStr)

	services := make([]ServiceConfig, 0, len(serviceNames))

	for _, name := range serviceNames {
		// Convert service name to uppercase env var format (e.g., "auth" -> "AUTH_SERVICE_URL")
		envVarName := toEnvVarName(name) + "_SERVICE_URL"

		// Get service URL from environment or use default pattern
		defaultURL := fmt.Sprintf("http://localhost:808%d/graphql", len(services)+1)
		serviceURL := GetEnv(envVarName, defaultURL)

		if serviceURL != "" {
			services = append(services, ServiceConfig{
				Name: name,
				URL:  serviceURL,
			})
		}
	}

	return services
}

// LoadServicesFromFile loads service configurations from a file
// File format: service-name|service-url (one per line)
// Lines starting with # are comments
func LoadServicesFromFile(filename string) ([]ServiceConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	services := make([]ServiceConfig, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line: service-name|service-url
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		url := strings.TrimSpace(parts[1])

		if name != "" && url != "" {
			services = append(services, ServiceConfig{
				Name: name,
				URL:  url,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

// ServiceConfig holds configuration for a microservice
type ServiceConfig struct {
	Name string
	URL  string
}

// parseServiceNames splits comma-separated service names and trims whitespace
func parseServiceNames(servicesStr string) []string {
	if servicesStr == "" {
		return []string{}
	}

	parts := strings.Split(servicesStr, ",")
	names := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}

	return names
}

// toEnvVarName converts service name to environment variable format
// Examples: "auth" -> "AUTH", "microservice-1" -> "MICROSERVICE_1"
func toEnvVarName(serviceName string) string {
	// Replace hyphens with underscores and convert to uppercase
	envName := strings.ReplaceAll(serviceName, "-", "_")
	return strings.ToUpper(envName)
}

// GetRedisAddr formats the Redis address from host and port
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return fallback
}
