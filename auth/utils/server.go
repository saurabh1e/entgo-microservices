package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/saurabh/entgo-microservices/auth/config"
	"github.com/saurabh/entgo-microservices/auth/graph"
	"github.com/saurabh/entgo-microservices/auth/utils/database"

	pkggraphql "github.com/saurabh/entgo-microservices/pkg/graphql"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
	"github.com/saurabh/entgo-microservices/pkg/logger"
	pkgmiddleware "github.com/saurabh/entgo-microservices/pkg/middleware"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Router     *gin.Engine
	GraphQLSrv *handler.Server
	Config     *config.Config
	httpServer *http.Server
}

// InitializeServer sets up the HTTP server with all routes and middleware
func InitializeServer(cfg *config.Config, db *database.DB, jwtService *jwt.Service, redis *database.RedisClient) *ServerConfig {
	logger.Info("Initializing HTTP server")

	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(pkgmiddleware.CORS())

	// Initialize GraphQL resolver with JWT service and Redis client
	// Gateway client will be initialized on-demand when needed (since gateway starts after microservices)
	resolver := graph.NewResolver(db.Client, jwtService, redis.Client, nil)

	// Create GraphQL server with directive configuration
	graphqlSrv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
		Directives: graph.DirectiveRoot{
			Auth:          pkggraphql.AuthDirective,
			HasRole:       pkggraphql.HasRoleDirective,
			HasPermission: pkggraphql.HasPermissionDirective,
		},
	}))

	// Configure GraphQL server transports
	graphqlSrv.AddTransport(transport.Options{})
	graphqlSrv.AddTransport(transport.GET{})
	graphqlSrv.AddTransport(transport.POST{})
	graphqlSrv.AddTransport(transport.MultipartForm{})
	graphqlSrv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// If AllowedOrigins contains '*', allow all origins
				if len(cfg.App.AllowedOrigins) == 0 {
					return true
				}
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				for _, ao := range cfg.App.AllowedOrigins {
					if ao == "*" {
						return true
					}
					if strings.EqualFold(strings.TrimSpace(ao), origin) {
						return true
					}
				}
				return false
			},
		},
	})

	// Add GraphQL extensions
	graphqlSrv.Use(extension.Introspection{})

	// Initialize JWT auth middleware from pkg (just reads from Redis)
	jwtAuthMiddleware := pkgmiddleware.NewJWTAuthMiddleware(jwtService, redis.Client, "auth")

	// GraphQL routes with authentication middleware (use /graphql)
	router.POST("/graphql", gin.WrapH(jwtAuthMiddleware.Middleware(graphqlSrv)))
	router.GET("/graphql", gin.WrapH(jwtAuthMiddleware.Middleware(graphqlSrv)))

	// GraphQL playground (only in development)
	if cfg.App.Environment != "production" {
		router.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))
		router.GET("/playground", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))
	}

	// Health check endpoints (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth",
			"version": cfg.App.Version,
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	return &ServerConfig{
		Router:     router,
		GraphQLSrv: graphqlSrv,
		Config:     cfg,
	}
}

// StartServer starts the HTTP server
func (s *ServerConfig) StartServer() error {
	addr := fmt.Sprintf(":%d", s.Config.Server.Port)

	logger.WithFields(map[string]interface{}{
		"address":     addr,
		"environment": s.Config.App.Environment,
	}).Info("🚀 Starting HTTP server")

	// Create HTTP server with timeouts and store it on the struct for shutdown
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.Router,
		ReadTimeout:  time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}

	logger.Infof("🌟 Server is running on http://localhost%s", addr)
	if s.Config.App.Environment != "production" && s.Config.GraphQL.PlaygroundEnabled {
		logger.Infof("🎮 GraphQL Playground: http://localhost%s/playground", addr)
	}
	logger.Infof("🔍 GraphQL Endpoint: http://localhost%s/graphql", addr)
	logger.Infof("❤️  Health Check: http://localhost%s/health", addr)

	return s.httpServer.ListenAndServe()
}

// GracefulShutdown handles graceful server shutdown
func (s *ServerConfig) GracefulShutdown(ctx context.Context) error {
	logger.Info("Initiating graceful server shutdown")

	// Shutdown HTTP server if running
	if s.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			logger.WithError(err).Error("HTTP server shutdown failed")
			return err
		}
	}

	logger.Info("✅ Server shutdown completed")
	return nil
}
