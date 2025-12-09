#!/bin/bash

# Microservice Generator - Main Script
# This script orchestrates the creation of a new microservice using modular components

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULES_DIR="$SCRIPT_DIR/modules"

# Source all module files
source "$MODULES_DIR/common.sh"
source "$MODULES_DIR/templates/go_files.sh"
source "$MODULES_DIR/templates/docker_files.sh"
source "$MODULES_DIR/generators/schema_generator.sh"
source "$MODULES_DIR/generators/file_copier.sh"
source "$MODULES_DIR/generators/grpc_generator.sh"
source "$MODULES_DIR/integrations/gateway.sh"

# Configuration
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BASE_MODULE="github.com/saurabh/entgo-microservices"
AUTH_DIR="$PROJECT_ROOT/auth"
GATEWAY_DIR="$PROJECT_ROOT/gateway"

# Parse arguments
SERVICE_NAME="$1"

if [ -z "$SERVICE_NAME" ]; then
    echo "Usage: $0 <service-name>"
    echo ""
    echo "Example:"
    echo "  $0 user-service"
    exit 1
fi

# Validate service name
validate_service_name "$SERVICE_NAME" || exit 1

# Setup variables
SERVICE_DIR="$PROJECT_ROOT/$SERVICE_NAME"
SERVICE_NAME_UPPER=$(to_env_var_name "$SERVICE_NAME")
SERVICE_PORT=$((8081 + $(ls -d "$PROJECT_ROOT"/*/ 2>/dev/null | wc -l | tr -d ' ')))
IMPORT_PREFIX="${BASE_MODULE}/${SERVICE_NAME}"

# Validate directory doesn't exist
validate_directory_not_exists "$SERVICE_DIR" || exit 1

log_info "Creating service: $SERVICE_NAME"
log_info "Service directory: $SERVICE_DIR"
log_info "Using base module: $BASE_MODULE"
echo ""

# ============================================================================
# STEP 1: Create Directory Structure
# ============================================================================
log_step "Step 1: Creating service directory structure"
mkdir -p "$SERVICE_DIR"
mkdir -p "$SERVICE_DIR/ent/schema"
mkdir -p "$SERVICE_DIR/graph"
mkdir -p "$SERVICE_DIR/graph/model"
mkdir -p "$SERVICE_DIR/graph/schemas"
mkdir -p "$SERVICE_DIR/logs"
mkdir -p "$SERVICE_DIR/tmp"
mkdir -p "$SERVICE_DIR/grpc"
mkdir -p "$SERVICE_DIR/bin"
log_success "Created directory structure"

# ============================================================================
# STEP 2: Generate Bootstrap Files
# ============================================================================
log_step "Step 2: Creating bootstrap files"

# Create graph/model/bootstrap.go
generate_model_bootstrap "$SERVICE_DIR"

# Setup base_mixin and dummy schema
setup_base_mixin "$SERVICE_DIR" "$AUTH_DIR"
generate_dummy_schema "$SERVICE_DIR"

log_success "Created bootstrap files"

# ============================================================================
# STEP 3: Copy Files from Auth Service
# ============================================================================
log_step "Step 3: Copying files from auth service"

copy_from_auth "$AUTH_DIR" "$SERVICE_DIR" "$SERVICE_NAME" "$BASE_MODULE"
copy_config_files "$AUTH_DIR" "$SERVICE_DIR" "$SERVICE_NAME"
copy_graph_files "$AUTH_DIR" "$SERVICE_DIR" "$SERVICE_NAME" "$BASE_MODULE"
copy_main_go "$AUTH_DIR/main.go" "$SERVICE_DIR/main.go" "$SERVICE_NAME" "$BASE_MODULE"

log_success "Copied files from auth service"

# ============================================================================
# STEP 4: Generate Configuration Files
# ============================================================================
log_step "Step 4: Generating configuration files"

# Generate .env file (simplified - copy from auth and update)
if [ -f "$AUTH_DIR/.env.example" ]; then
    cp "$AUTH_DIR/.env.example" "$SERVICE_DIR/.env.example"
    SERVICE_NAME_UPPER=$(to_env_var_name "$SERVICE_NAME")
    sed -i.bak "s/auth-service/${SERVICE_NAME}-service/g" "$SERVICE_DIR/.env.example"
    sed -i.bak "s/auth_/${SERVICE_NAME}_/g" "$SERVICE_DIR/.env.example"
    sed -i.bak "s/AUTH_/${SERVICE_NAME_UPPER}_/g" "$SERVICE_DIR/.env.example"
    sed -i.bak "s/entgo_auth_postgres/entgo_${SERVICE_NAME}_postgres/g" "$SERVICE_DIR/.env.example"
    sed -i.bak "s/DB_NAME=auth/DB_NAME=${SERVICE_NAME}/g" "$SERVICE_DIR/.env.example"
    sed -i.bak "s/SERVER_PORT=8081/SERVER_PORT=${SERVICE_PORT}/g" "$SERVICE_DIR/.env.example"
    sed -i.bak "s/GRPC_PORT=9081/GRPC_PORT=9082/g" "$SERVICE_DIR/.env.example"
    rm "$SERVICE_DIR/.env.example.bak"
    cp "$SERVICE_DIR/.env.example" "$SERVICE_DIR/.env"
fi

# Generate dev.sh
if [ -f "$AUTH_DIR/dev.sh" ]; then
    cp "$AUTH_DIR/dev.sh" "$SERVICE_DIR/"
    sed -i.bak "s/auth/${SERVICE_NAME}/g" "$SERVICE_DIR/dev.sh"
    rm "$SERVICE_DIR/dev.sh.bak"
    chmod +x "$SERVICE_DIR/dev.sh"
fi

log_success "Generated configuration files"

# ============================================================================
# STEP 5: Generate Docker Files
# ============================================================================
log_step "Step 5: Generating Docker files"

generate_dockerfile "$SERVICE_DIR" "$SERVICE_NAME" "$AUTH_DIR" "$BASE_MODULE"
generate_docker_compose "$SERVICE_DIR" "$SERVICE_NAME"
generate_docker_compose_dev "$SERVICE_DIR" "$SERVICE_NAME"

log_success "Generated Docker files"

# ============================================================================
# STEP 6: Generate Go Module Files
# ============================================================================
log_step "Step 6: Generating Go module files"

generate_go_mod "$SERVICE_DIR" "$BASE_MODULE" "$SERVICE_NAME"
generate_gqlgen_yml "$SERVICE_DIR" "$BASE_MODULE" "$SERVICE_NAME"
generate_generate_go "$SERVICE_DIR"


log_success "Generated Go module files"

# ============================================================================
# STEP 7: Setup gRPC Files
# ============================================================================
log_step "Step 7: Setting up gRPC files"

setup_grpc_files "$SERVICE_DIR" "$AUTH_DIR" "$SERVICE_NAME" "$BASE_MODULE"

log_success "gRPC files setup completed"

# ============================================================================
# STEP 8: Create Makefile and README
# ============================================================================
log_step "Step 8: Creating Makefile and README"

# Copy Makefile from auth (it should already have the correct structure)
if [ -f "$AUTH_DIR/Makefile" ]; then
    cp "$AUTH_DIR/Makefile" "$SERVICE_DIR/"
fi

# Create README
cat > "$SERVICE_DIR/README.md" << EOF
# $SERVICE_NAME

Microservice for ${SERVICE_NAME}.

## Getting Started

\`\`\`bash
# Generate code
go generate generate.go

# Build
go build -o bin/${SERVICE_NAME} main.go

# Run in development mode
./dev.sh dev
\`\`\`

## Ports

- HTTP/GraphQL: ${SERVICE_PORT}
- gRPC: 9082

## Database

- PostgreSQL: entgo_${SERVICE_NAME}_postgres
- Database name: ${SERVICE_NAME}
EOF

log_success "Created Makefile and README"

# ============================================================================
# STEP 9: Add to go.work
# ============================================================================
log_step "Step 9: Adding service to go.work"

if command -v go &> /dev/null; then
    cd "$PROJECT_ROOT"
    if [ -f "go.work" ]; then
        if ! grep -q "use ./${SERVICE_NAME}" go.work; then
            go work use "./${SERVICE_NAME}"
            log_success "Added ${SERVICE_NAME} to go.work"
        fi
    else
        go work init "./${SERVICE_NAME}"
        log_success "Created go.work and added ${SERVICE_NAME}"
    fi
fi

# ============================================================================
# STEP 10: Gateway Integration
# ============================================================================
log_step "Step 10: Integrating with Gateway"

integrate_with_gateway "$SERVICE_NAME" "$SERVICE_PORT" "$GATEWAY_DIR" "$SERVICE_NAME_UPPER"

# ============================================================================
# COMPLETION
# ============================================================================
echo ""
echo "=========================================="
log_success "Microservice '$SERVICE_NAME' created successfully!"
echo "=========================================="
echo ""
echo -e "${CYAN}Next steps:${NC}"
echo "  1. cd $SERVICE_NAME"
echo "  2. go generate generate.go  # Generate all code"
echo "  3. go build -o bin/${SERVICE_NAME} main.go  # Build"
echo "  4. ./dev.sh dev  # Run in development mode"
echo ""
echo -e "${CYAN}Service details:${NC}"
echo "  - HTTP/GraphQL Port: ${SERVICE_PORT}"
echo "  - gRPC Port: 9082"
echo "  - Database: ${SERVICE_NAME}"
echo "  - Container: entgo_${SERVICE_NAME}_dev"
echo ""
echo -e "${GREEN}✓ Gateway Integration:${NC}"
echo "  - Service registered in gateway/services.conf"
echo "  - Gateway will auto-discover this service"
echo ""
echo -e "${YELLOW}Note:${NC} The service is ready to run but won't start HTTP server until"
echo "      you define GraphQL schemas. gRPC server will start immediately."
echo "      Remember to restart the gateway to pick up the new service!"
echo ""

