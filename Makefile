# Root Makefile to orchestrate microservices

COMPOSE_SHARED := -f docker-compose.shared.yml
COMPOSE_DEV := -f docker-compose.shared.yml -f auth/docker-compose.dev.yml -f gateway/docker-compose.dev.yml
COMPOSE_PROD := -f docker-compose.shared.yml -f auth/docker-compose.yml -f gateway/docker-compose.yml

.PHONY: help network shared dev prod up down build logs ps clean proto generate-grpc generate-clients

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  network - create Docker network"
	@echo "  shared  - start shared infrastructure (Redis, Postgres)"
	@echo "  dev     - start services in development mode"
	@echo "  prod    - start services in production mode"
	@echo "  up      - alias for dev (starts in background)"
	@echo "  down    - stop and remove all containers"
	@echo "  build   - build images for prod"
	@echo "  logs    - follow logs from services"
	@echo "  ps      - show running containers"
	@echo "  proto           - generate proto files"
	@echo "  generate-grpc   - generate gRPC service clients for all microservices"
	@echo "  generate-clients- aggregate metadata and generate service_clients.go"
	@echo "  clean           - remove generated files and containers"

# Create Docker network
network:
	@docker network create entgo_network || true
	@echo "✅ Docker network 'entgo_network' ready"

# Start shared infrastructure only
shared: network
	@echo "🚀 Starting shared infrastructure..."
	docker-compose $(COMPOSE_SHARED) up -d
	@echo "✅ Shared infrastructure started (Redis, Postgres)"

# Start dev in foreground (attached)
dev: network
	@echo "🚀 Starting development environment..."
	docker-compose $(COMPOSE_DEV) up --build

# Start prod in foreground (attached)
prod: network
	@echo "🚀 Starting production environment..."
	docker-compose $(COMPOSE_PROD) up --build

# Start dev in background
up: network
	@echo "🚀 Starting development environment in background..."
	docker-compose $(COMPOSE_DEV) up --build -d
	@echo "✅ Services started successfully"

# Stop and remove containers
down:
	@echo "🛑 Stopping all services..."
	docker-compose $(COMPOSE_DEV) down || true
	docker-compose $(COMPOSE_PROD) down || true
	docker-compose $(COMPOSE_SHARED) down || true
	@echo "✅ All services stopped"

# Build images for prod compose
build:
	@echo "🔨 Building production images..."
	docker-compose $(COMPOSE_PROD) build --no-cache

# Follow logs
logs:
	docker-compose $(COMPOSE_DEV) logs -f

# Show containers
ps:
	@echo "📊 Running containers:"
	docker-compose $(COMPOSE_DEV) ps

# Generate proto files
proto:
	@echo "🔧 Generating proto files..."
	cd proto && make proto
	@echo "✅ Proto generation complete"

# Generate gRPC service clients for all microservices
generate-grpc:
	@echo "🔧 Generating gRPC service clients..."
	@cd auth && go run cmd/generate-grpc/main.go
	@cd attendance && go run cmd/generate-grpc/main.go
	@cd route && go run cmd/generate-grpc/main.go
	@cd yard && go run cmd/generate-grpc/main.go
	@echo "✅ gRPC service clients generated for all microservices"

# Aggregate metadata and generate service_clients.go
generate-clients: generate-grpc
	@echo "🔧 Generating consolidated service_clients.go..."
	@./scripts/generate-service-clients.sh
	@echo "✅ Service clients generation complete"

# Clean up
clean:
	@echo "🧹 Cleaning up..."
	docker-compose $(COMPOSE_DEV) down -v || true
	docker-compose $(COMPOSE_PROD) down -v || true
	docker-compose $(COMPOSE_SHARED) down -v || true
	@echo "✅ Cleanup complete"

