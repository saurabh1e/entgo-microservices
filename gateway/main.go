package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/saurabh/entgo-microservices/gateway/router"
	"github.com/saurabh/entgo-microservices/gateway/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Printf("⚠️  Warning: Could not load .env file: %v\n", err)
	}

	// Load configuration
	config := utils.LoadConfig()

	// Initialize dependencies
	if err := initDependencies(); err != nil {
		fmt.Printf("❌ Failed to initialize dependencies: %v\n", err)
		os.Exit(1)
	}
	defer cleanupDependencies()

	// Setup gateway components
	gatewayRouter := Setup()

	// Setup HTTP router
	r := chi.NewRouter()
	setupMiddleware(r)

	// Add explicit OPTIONS handler for GraphQL endpoint
	r.Options("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Register GraphQL handler
	r.Post("/graphql", func(w http.ResponseWriter, r *http.Request) {
		gatewayRouter.HandleRequest(w, r)
	})

	// Add GraphQL Playground route
	r.Get("/playground", router.ServePlayground)

	// Add REST API routes for service proxying
	// This will handle routes like /api/v1/{service_name}/{path}
	r.HandleFunc("/api/v1/*", func(w http.ResponseWriter, r *http.Request) {
		gatewayRouter.HandleRESTRequest(w, r)
	})

	// Print info about available endpoints
	fmt.Println("📊 API Endpoints:")
	fmt.Println("  • GraphQL API: http://localhost:" + config.Port + "/graphql")
	fmt.Println("  • GraphQL Playground: http://localhost:" + config.Port + "/playground")
	fmt.Println("  • REST API: http://localhost:" + config.Port + "/api/v1/{service_name}/{path}")

	// Start server
	fmt.Printf("🚀 GraphQL Gateway running on port %s...\n", config.Port)
	err := http.ListenAndServe(":"+config.Port, r)
	if err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

// setupMiddleware configures the middleware for the HTTP router
func setupMiddleware(r *chi.Mux) {
	// Add custom middleware to handle private network access - must be first
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add Access-Control-Allow-Private-Network header for Apollo Studio and other tools
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
			next.ServeHTTP(w, r)
		})
	})

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},                                                        // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"}, // Comprehensive HTTP methods
		AllowedHeaders:   []string{"*"},                                                        // Allow all headers
		ExposedHeaders:   []string{"*"},                                                        // Expose all headers
		AllowCredentials: true,                                                                 // Allow credentials
		MaxAge:           86400,                                                                // Cache preflight requests for 24 hours
	}))
}

// initDependencies initializes external services and connections
func initDependencies() error {
	// Initialize Redis connection
	if err := utils.InitRedis(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}

// cleanupDependencies performs cleanup of resources
func cleanupDependencies() {
	utils.CloseRedis()
}
