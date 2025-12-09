package main

import (
	"fmt"
	"log"

	"github.com/saurabh/entgo-microservices/gateway/router"
	"github.com/saurabh/entgo-microservices/gateway/schema"
	"github.com/saurabh/entgo-microservices/gateway/utils"

	"gopkg.in/natefinch/lumberjack.v2"
)

// schemaManager holds the schema data to be accessed by the router
var schemaManager *schema.Manager

// Setup initializes the gateway components
func Setup() *router.Router {
	// Configure logging
	setupLogging()

	// Load service configurations dynamically from environment
	serviceConfigs := utils.LoadServicesFromEnv()

	// Convert to schema.Service format
	services := make([]schema.Service, len(serviceConfigs))
	for i, cfg := range serviceConfigs {
		services[i] = schema.Service{
			Name: cfg.Name,
			URL:  cfg.URL,
		}
	}

	// Print service configuration for debugging
	fmt.Println("🔌 Configured Services:")
	for _, svc := range services {
		fmt.Printf("  • %s: %s\n", svc.Name, svc.URL)
	}

	// Initialize schema manager and collect schemas
	schemaManager = schema.NewManager(services)
	schemaManager.Initialize()

	// Show schema collection results
	fmt.Println("\n📊 Schema Collection Results:")
	fmt.Println(schemaManager.Debug())

	// Create router using adapter pattern
	return router.NewRouter(newSchemaAdapter(schemaManager))
}

// schemaAdapter adapts the schema manager to the router's interface
type schemaAdapter struct {
	manager *schema.Manager
}

// newSchemaAdapter creates a new schema adapter
func newSchemaAdapter(manager *schema.Manager) *schemaAdapter {
	return &schemaAdapter{manager: manager}
}

// GetRouteForOperation delegates to the schema manager
func (s *schemaAdapter) GetRouteForOperation(rootField string) string {
	return s.manager.GetRouteForOperation(rootField)
}

// GetMergedSchema delegates to the schema manager
func (s *schemaAdapter) GetMergedSchema() interface{} {
	return s.manager.GetMergedSchema()
}

// GetRoutes returns the routes map from the schema manager
// This is needed for REST API forwarding
func (s *schemaAdapter) GetRoutes() map[string]string {
	// Get a copy of the routes map
	return s.manager.GetRoutesMap()
}

// setupLogging configures logging for the gateway
func setupLogging() {
	// Configure file logging with rotation
	log.SetOutput(&lumberjack.Logger{
		Filename:   "logs/gateway.log",
		MaxSize:    10,   // megabytes
		MaxBackups: 5,    // number of backups
		MaxAge:     30,   // days
		Compress:   true, // compress old logs
	})

	// Add timestamp and source file info to logs
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Gateway logging initialized")
}
