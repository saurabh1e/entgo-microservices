# Scripts Directory

This directory contains automation scripts for the microservices project.

## Service Client Generation

### 🔧 generate-service-clients.sh

**Purpose:** Aggregates metadata from all microservices and generates the consolidated `service_clients.go` file.

**Usage:**
```bash
./scripts/generate-service-clients.sh
```

**What it does:**
1. Scans `pkg/grpc/metadata/*.json` for all microservices
2. Discovers microservices automatically (auth, attendance, route, yard, etc.)
3. Generates `pkg/grpc/service_clients.go` with:
   - One client per microservice
   - Lazy initialization with sync.Once
   - Clean API: `r.Services().Auth().GetUserByID()`

**When to run:**
- After running `generate-grpc` in any microservice
- After creating a new microservice
- Use `make generate-clients` from project root (RECOMMENDED)

**Output:**
```
✅ Found 4 microservices: auth attendance route yard
📝 Generating /path/to/pkg/grpc/service_clients.go...
✅ Generated service_clients.go
📊 Summary:
   Microservices: 4
   - Auth: 4 models
   - Attendance: 1 models
   - Route: 1 models
   - Yard: 1 models
✨ Service clients generation complete!
```

## Microservice Creation

### 🏗️ create-microservice-v2.sh

**Purpose:** Creates a new microservice with complete structure and auto-generates service clients.

**Usage:**
```bash
./scripts/create-microservice-v2.sh my-service
```

**What it does:**
1. Creates complete microservice structure
2. Generates Ent schemas, GraphQL schemas
3. Sets up Docker and development environment
4. Integrates with gateway
5. **✅ Automatically generates gRPC service client**
6. **✅ Updates consolidated service_clients.go**

**Steps Performed:**
- Step 1-11: Standard microservice creation
- **Step 12: Generate Service Clients** (NEW!)
  - Runs `cmd/generate-grpc/main.go` for new service
  - Runs `generate-service-clients.sh` to update registry
  - New service immediately available: `r.Services().Myservice()`

## Other Scripts

### 📝 build-all.sh
Builds all microservices in the project.

### 🔐 generate-certs.sh
Generates SSL/TLS certificates for secure communication.

### 🔧 install-proto.sh
Installs protocol buffer compiler and plugins.

### 🌱 seed.sh
Seeds the database with initial data.

### ✅ validate.sh
Validates project structure and dependencies.

## Integration with Makefile

### Project Root Makefile

```bash
make generate-grpc       # Generate service clients for all microservices
make generate-clients    # Run generate-grpc + aggregator (RECOMMENDED)
```

### Microservice Makefiles

```bash
make gen                 # Generate all code (includes gRPC)
make generate-grpc       # Generate only gRPC service client
```

## Automated Triggers

The service client generation runs automatically in these scenarios:

1. **`go generate` in microservice** - Includes gRPC generation as Step 5
2. **`make gen` in microservice** - Cleans and regenerates everything
3. **`./scripts/create-microservice-v2.sh`** - Step 12 auto-generates clients
4. **`make generate-clients` from root** - Regenerates everything

## File Structure

```
scripts/
├── README.md                          # This file
├── generate-service-clients.sh        # Aggregator (auto-generates service_clients.go)
├── create-microservice-v2.sh          # Microservice creator (calls aggregator)
├── build-all.sh
├── generate-certs.sh
├── install-proto.sh
├── seed.sh
├── validate.sh
└── modules/                           # Helper modules for create-microservice-v2.sh
    ├── common.sh
    ├── generators/
    ├── integrations/
    └── templates/
```

## Generated Files

```
pkg/grpc/
├── service_clients.go                 # Auto-generated (DO NOT EDIT)
├── auth_service_client.go            # Generated per microservice
├── attendance_service_client.go
├── route_service_client.go
├── yard_service_client.go
└── metadata/
    ├── auth_services.json            # Metadata per microservice
    ├── attendance_services.json
    ├── route_services.json
    └── yard_services.json
```

## Best Practices

1. ✅ **Use `make generate-clients`** - Handles everything correctly
2. ✅ **Commit generated files** - They are part of the source code
3. ✅ **Run before commits** - Ensure service clients are up to date
4. ✅ **Use in CI/CD** - Verify generated files match source
5. ❌ **Don't edit manually** - service_clients.go has "DO NOT EDIT" header

## Quick Reference

| Task | Command |
|------|---------|
| Regenerate all service clients | `make generate-clients` |
| Create new microservice | `./scripts/create-microservice-v2.sh NAME` |
| Generate for one service | `cd SERVICE && make generate-grpc` |
| Run only aggregator | `./scripts/generate-service-clients.sh` |
| Full regeneration | `cd SERVICE && make gen && cd .. && make generate-clients` |

## Troubleshooting

**Problem:** Service client not updated after schema change  
**Solution:** `make generate-clients` from project root

**Problem:** New microservice not in service_clients.go  
**Solution:** `./scripts/generate-service-clients.sh`

**Problem:** Build errors after generation  
**Solution:** `cd pkg && go build ./...` to verify

For detailed documentation, see:
- `/docs/guides/SERVICE_CLIENT_AUTOMATION.md` - Complete automation guide
- `/docs/guides/SERVICE_CLIENT_GENERATION.md` - Architecture and implementation details

