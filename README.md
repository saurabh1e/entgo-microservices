# Entgo Microservices Architecture

A microservices architecture built with Go, featuring shared infrastructure, gRPC inter-service communication, and individual service databases.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Gateway                              │
│                    (HTTP/GraphQL Entry)                      │
└──────────────────────┬──────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       │               │               │
       ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  Auth Service│ │Future Service│ │Future Service│
│   HTTP:8081  │ │              │ │              │
│   gRPC:9081  │ │              │ │              │
│   DB: auth   │ │  DB: orders  │ │ DB: products │
└──────┬───────┘ └──────────────┘ └──────────────┘
       │
       └──────────────┬──────────────────────────────┐
                      ▼                              ▼
              ┌───────────────┐            ┌──────────────┐
              │ Shared Redis  │            │  Postgres    │
              │  (All Services)│            │ (Per Service)│
              └───────────────┘            └──────────────┘
```

## 📦 Components

### Shared Package (`pkg/`)
Reusable components shared across all microservices:

- **`pkg/logger`**: Structured logging with file rotation
- **`pkg/jwt`**: JWT token generation and validation with Redis
- **`pkg/redis`**: Redis client factory and token management with service namespacing
- **`pkg/middleware`**: CORS and authentication middleware
- **`pkg/proto`**: Generated gRPC protocol buffers (user, role, permission)
- **`pkg/grpc`**: gRPC client utilities (future)

### Auth Service
Full-featured authentication service with:
- GraphQL API (port 8081)
- gRPC API (port 9081)
- User management (Ent ORM)
- JWT authentication with Redis whitelist/blacklist
- Role-based access control
- Individual PostgreSQL database

### Gateway Service
HTTP gateway for routing requests to microservices

## 🚀 Key Features

### ✅ Shared Infrastructure
- **Single Redis instance** shared across all services
- **Service-specific key namespacing**: `{service}:{resource}:{id}`
  - Example: `auth:whitelist:token123`, `orders:cache:order456`
- **Individual databases** per service for data isolation

### ✅ gRPC Inter-Service Communication
- Type-safe protocol buffers
- Services exposed via gRPC:
  - `UserService`: GetUserByID, GetUsersByIDs, ValidateUser
  - `RoleService`: GetRoleByID, GetRolesByIDs
  - `PermissionService`: GetPermissionByID, GetPermissionsByIDs
- Health checks and reflection enabled
- Logging and recovery interceptors

### ✅ Go Workspace
- Simplified local development with `go.work`
- Shared `pkg` module across all services
- No need for replace directives

### ✅ Docker Infrastructure
- Shared Redis container
- Individual PostgreSQL per service
- Hot reload in development (Air)
- Health checks for all services

## 🛠️ Development Setup

### Prerequisites
- Go 1.25.5+
- Docker & Docker Compose
- Protocol Buffers compiler (`protoc`)

### Quick Start

1. **Clone and navigate to project:**
```bash
cd /Users/saurabh/GolandProjects/entgo-microservices
```

2. **Set up environment files (REQUIRED):**

Each microservice needs its own .env file:

```bash
# Auth service
cd auth
cp .env.example .env
cd ..

# Future services (when added)
# cd orders && cp .env.example .env && cd ..
```

The .env.example files contain all necessary configuration with sensible defaults for development. You can use them as-is for local development.

**Important Notes:**
- ✅ `.env.example` files are tracked in git (templates)
- ❌ `.env` files are NOT tracked in git (contain your config)
- 🔒 Never commit `.env` files (they may contain secrets in production)

3. **Install protoc plugins:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

4. **Start everything:**

**Option A - Automated (Recommended):**
```bash
./scripts/quick-start.sh
```
The script will:
- Check/create .env files automatically
- Start Docker network
- Start shared Redis
- Start auth service
- Verify everything is running

**Option B - Manual:**
```bash
make network  # Create Docker network
make shared   # Start shared Redis  
make up       # Start all services
```

### Environment File Structure

Each service has:
- `<service>/.env.example` - Template with all variables (tracked in git)
- `<service>/.env` - Your actual config (NOT in git)

The quick-start script will automatically create .env files from templates if they don't exist.

### Available Make Targets

```bash
make network  # Create Docker network
make shared   # Start shared infrastructure (Redis)
make dev      # Start all services (foreground)
make up       # Start all services (background)
make down     # Stop all services
make logs     # Follow service logs
make ps       # Show running containers
make proto    # Generate proto files
make clean    # Clean up everything
```

## 📁 Project Structure

```
.
├── go.work                          # Go workspace definition
├── Makefile                         # Root orchestration
├── docker-compose.shared.yml        # Shared infrastructure (Redis)
├── .gitignore                       # Git ignore rules
│
├── docs/                            # 📚 Documentation
│   ├── architecture/                # Architecture decisions & plans
│   │   ├── IMPLEMENTATION_SUMMARY.md
│   │   ├── PHASES_COMPLETE.md
│   │   ├── VERIFICATION_REPORT.md
│   │   └── plan-microservicesRefactoring.prompt.md
│   └── guides/                      # How-to guides
│       ├── MTLS_SETUP.md           # mTLS configuration guide
│       └── TESTING.md              # Testing strategies
│
├── scripts/                         # 🔧 Utility scripts
│   ├── quick-start.sh              # One-command startup
│   ├── generate-certs.sh           # mTLS certificate generation
│   └── validate.sh                 # Verify all phases
│
├── pkg/                             # 📦 Shared package
│   ├── go.mod
│   ├── logger/                      # Logging utilities
│   ├── jwt/                         # JWT service
│   ├── redis/                       # Redis utilities with namespacing
│   ├── middleware/                  # HTTP middleware
│   ├── grpc/                        # gRPC client library
│   │   ├── client.go               # Connection pooling
│   │   ├── interceptors.go         # Retry & logging
│   │   ├── user_client.go          # User service client
│   │   ├── role_client.go          # Role service client
│   │   └── permission_client.go    # Permission service client
│   └── proto/                       # Generated gRPC code
│       ├── user/v1/
│       ├── role/v1/
│       └── permission/v1/
│
├── proto/                           # Protocol buffer definitions
│   ├── Makefile                    # Proto generation
│   ├── common/common.proto
│   ├── user/user.proto
│   ├── role/role.proto
│   └── permission/permission.proto
│
├── auth/                            # 🔐 Auth microservice
│   ├── go.mod
│   ├── main.go
│   ├── .env.example                # Environment template
│   ├── .env                        # Environment config (not in git)
│   ├── docker-compose.dev.yml     # Auth service + Postgres
│   ├── grpc/                       # gRPC server
│   │   ├── server.go
│   │   ├── user_service.go
│   │   ├── role_service.go
│   │   └── permission_service.go
│   ├── graph/                      # GraphQL
│   ├── internal/ent/               # Database ORM
│   └── utils/                      # Service utilities
│
└── gateway/                         # 🌐 Gateway service
    ├── go.mod
    └── main.go
```

## 🔑 Environment Variables

### Auth Service
```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8081
GRPC_PORT=9081

# Database (Individual per service)
DB_HOST=entgo_auth_postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth

# Redis (Shared)
REDIS_HOST=entgo_redis_shared
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY_HOURS=24

# Logging
LOG_LEVEL=debug
LOG_DIR=./logs
```

## 🔐 Redis Key Namespacing

All services use prefixed keys to avoid conflicts:

```
{service_name}:{resource_type}:{identifier}

Examples:
- auth:whitelist:abc123       # JWT whitelist
- auth:blacklist:xyz789       # JWT blacklist
- auth:user_cache:42          # User cache
- auth:rate_limit:user:123    # Rate limiting
- orders:cache:order:456      # Future: Orders cache
- products:inventory:101      # Future: Products inventory
```

## 🌐 API Endpoints

### Auth Service

#### HTTP/GraphQL (Port 8081)
```
POST   /graphql           # GraphQL API
GET    /playground        # GraphQL Playground
```

#### gRPC (Port 9081)
```
UserService/GetUserByID
UserService/GetUsersByIDs
UserService/ValidateUser
RoleService/GetRoleByID
RoleService/GetRolesByIDs
PermissionService/GetPermissionByID
PermissionService/GetPermissionsByIDs
```

#### Test gRPC with grpcurl
```bash
# List services
grpcurl -plaintext localhost:9081 list

# Get user by ID
grpcurl -plaintext -d '{"id": 1}' \
  localhost:9081 user.v1.UserService/GetUserByID

# Health check
grpcurl -plaintext localhost:9081 \
  grpc.health.v1.Health/Check
```

## 📝 Adding a New Microservice

1. **Create service directory:**
```bash
mkdir orders && cd orders
go mod init orders
```

2. **Update go.work:**
```go
use (
    ./pkg
    ./auth
    ./gateway
    ./orders  // Add new service
)
```

3. **Add to go.mod:**
```go
require github.com/saurabh/entgo-microservices/pkg v0.0.0
replace github.com/saurabh/entgo-microservices/pkg => ../pkg
```

4. **Create proto definitions:**
```bash
# proto/order/order.proto
```

5. **Generate protos:**
```bash
cd proto && make proto
```

6. **Create Docker Compose:**
```yaml
# orders/docker-compose.dev.yml
services:
  orders:
    # ... service config
    depends_on:
      - orders_postgres
      - redis  # Reference shared redis
  
  orders_postgres:
    # Individual database
```

7. **Update root Makefile:**
```makefile
COMPOSE_DEV := -f docker-compose.shared.yml \
               -f auth/docker-compose.dev.yml \
               -f gateway/docker-compose.dev.yml \
               -f orders/docker-compose.dev.yml
```

8. **Use shared packages:**
```go
import (
    "github.com/saurabh/entgo-microservices/pkg/jwt"
    "github.com/saurabh/entgo-microservices/pkg/logger"
    "github.com/saurabh/entgo-microservices/pkg/middleware"
)

// Initialize with service name for Redis namespacing
jwtService := jwt.NewService(secret, hours, redisClient, "orders")
```

## 🔍 Implementation Status

### ✅ Phase 1: Shared Infrastructure - COMPLETE
- [x] Created pkg module with go.mod
- [x] Extracted logger to pkg/logger
- [x] Extracted JWT service to pkg/jwt with Redis namespacing
- [x] Extracted Redis utilities to pkg/redis
- [x] Extracted middleware to pkg/middleware
- [x] Created proto definitions
- [x] Generated proto code
- [x] Created go.work file

### ✅ Phase 2: Docker Infrastructure - COMPLETE
- [x] Created docker-compose.shared.yml (Redis only)
- [x] Updated auth/docker-compose.dev.yml with individual Postgres
- [x] Added gRPC port (9081) to auth service
- [x] Updated root Makefile
- [x] Created Docker network

### ✅ Phase 3: Auth Service Migration - COMPLETE
- [x] Updated auth/go.mod to use pkg
- [x] Replaced all imports to use pkg modules
- [x] Updated JWT service initialization with "auth" namespace
- [x] Updated middleware imports
- [x] Updated all resolver imports
- [x] Service builds successfully

### ✅ Phase 4: gRPC Implementation - COMPLETE
- [x] Created auth/grpc directory
- [x] Implemented gRPC server with health checks
- [x] Implemented UserService
- [x] Implemented RoleService
- [x] Implemented PermissionService
- [x] Added logging and recovery interceptors
- [x] Updated main.go to start gRPC server
- [x] Updated lifecycle manager for gRPC shutdown

### 🔄 Phase 5-7: Ready for Implementation
- [ ] Phase 5: gRPC Client Library (in pkg/grpc)
- [ ] Phase 6: mTLS Certificate Setup
- [ ] Phase 7: Testing & Validation

## 🔒 Security Features

- JWT token whitelist/blacklist with Redis
- Service-specific key namespacing
- Rate limiting with Redis
- CORS protection
- Password hashing with bcrypt
- Privacy rules with Ent
- Health checks for all services

## 📊 Monitoring

- Structured JSON logging
- Log rotation with lumberjack
- Separate error log files
- gRPC request logging
- Health check endpoints

## 🚧 Future Enhancements

1. **Phase 5: gRPC Client Library**
   - Connection pooling
   - Retry logic
   - Circuit breakers
   - Helper methods for cross-service calls

2. **Phase 6: mTLS**
   - Certificate generation scripts
   - Service-to-service authentication
   - Certificate rotation

3. **Phase 7: Observability**
   - Prometheus metrics
   - Distributed tracing (OpenTelemetry)
   - Request correlation IDs
   - Dashboard integration

4. **Additional Features**
   - API Gateway enhancements
   - Service mesh (Istio/Linkerd)
   - Event-driven architecture (Kafka/NATS)
   - Kubernetes deployment manifests

## 📚 Resources

- [Go Workspaces](https://go.dev/blog/get-familiar-with-workspaces)
- [gRPC Go](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://protobuf.dev/)
- [Ent ORM](https://entgo.io/)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)

## 🤝 Contributing

When adding new services:
1. Follow the naming convention: `{service}:{resource}:{id}` for Redis keys
2. Use shared pkg modules
3. Implement both HTTP and gRPC interfaces
4. Add health checks
5. Include proper logging
6. Write integration tests

## 📝 License

[Your License Here]

---

**Built with ❤️ using Go, gRPC, GraphQL, and Docker**

