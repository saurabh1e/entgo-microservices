#!/usr/bin/env bash
#
# clean-all-grpc.sh
# Deletes all proto files and gRPC service files to allow clean regeneration
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Cleaning All Proto Files and gRPC Services${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# List of all microservices
MICROSERVICES=(
    "attendance"
    "fleet"
    "route"
    "core"
    "notification"
    "reporting"
    "yard"
    "auth"
)

# Clean proto files
echo -e "${YELLOW}🧹 Cleaning proto files from proto/ directory...${NC}"
if [ -d "$PROJECT_ROOT/proto" ]; then
    # Count files before deletion
    PROTO_COUNT=$(find "$PROJECT_ROOT/proto" -type f -name "*.proto" 2>/dev/null | wc -l | tr -d ' ')

    if [ "$PROTO_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}   Found $PROTO_COUNT proto files${NC}"

        # Delete all .proto files
        find "$PROJECT_ROOT/proto" -type f -name "*.proto" -delete

        # Delete all generated .pb.go files in pkg/proto
        if [ -d "$PROJECT_ROOT/pkg/proto" ]; then
            PB_COUNT=$(find "$PROJECT_ROOT/pkg/proto" -type f -name "*.pb.go" 2>/dev/null | wc -l | tr -d ' ')
            if [ "$PB_COUNT" -gt 0 ]; then
                echo -e "${YELLOW}   Found $PB_COUNT generated .pb.go files${NC}"
                find "$PROJECT_ROOT/pkg/proto" -type f -name "*.pb.go" -delete
            fi
        fi

        echo -e "${GREEN}   ✅ Deleted all proto files${NC}"
    else
        echo -e "${YELLOW}   ℹ️  No proto files found${NC}"
    fi
else
    echo -e "${YELLOW}   ℹ️  proto/ directory not found${NC}"
fi

echo ""

# Clean gRPC service files from each microservice
echo -e "${YELLOW}🧹 Cleaning gRPC service files from microservices...${NC}"
for service in "${MICROSERVICES[@]}"; do
    SERVICE_DIR="$PROJECT_ROOT/$service"

    if [ ! -d "$SERVICE_DIR" ]; then
        echo -e "${YELLOW}   ⚠️  Skipping $service - directory not found${NC}"
        continue
    fi

    GRPC_DIR="$SERVICE_DIR/grpc"

    if [ ! -d "$GRPC_DIR" ]; then
        echo -e "${YELLOW}   ⚠️  Skipping $service - grpc/ directory not found${NC}"
        continue
    fi

    # Count service files
    SERVICE_COUNT=$(find "$GRPC_DIR" -type f -name "*_service.go" 2>/dev/null | wc -l | tr -d ' ')
    REGISTRY_EXISTS=0
    [ -f "$GRPC_DIR/service_registry.go" ] && REGISTRY_EXISTS=1

    TOTAL_FILES=$((SERVICE_COUNT + REGISTRY_EXISTS))

    if [ "$TOTAL_FILES" -gt 0 ]; then
        echo -e "${BLUE}   [$service]${NC} Deleting $TOTAL_FILES gRPC files..."

        # Delete all *_service.go files
        find "$GRPC_DIR" -type f -name "*_service.go" -delete

        # Delete service_registry.go
        rm -f "$GRPC_DIR/service_registry.go"

        echo -e "${GREEN}   [$service]${NC} ✅ Cleaned"
    else
        echo -e "${YELLOW}   [$service]${NC} ℹ️  No gRPC files to clean"
    fi
done

echo ""

# Clean service client files from pkg/grpc
echo -e "${YELLOW}🧹 Cleaning service client files from pkg/grpc...${NC}"
if [ -d "$PROJECT_ROOT/pkg/grpc" ]; then
    # Delete all *_service_client.go files
    CLIENT_COUNT=$(find "$PROJECT_ROOT/pkg/grpc" -type f -name "*_service_client.go" 2>/dev/null | wc -l | tr -d ' ')

    if [ "$CLIENT_COUNT" -gt 0 ]; then
        echo -e "${YELLOW}   Found $CLIENT_COUNT service client files${NC}"
        find "$PROJECT_ROOT/pkg/grpc" -type f -name "*_service_client.go" -delete
        echo -e "${GREEN}   ✅ Deleted all service client files${NC}"
    else
        echo -e "${YELLOW}   ℹ️  No service client files found${NC}"
    fi

    # Delete metadata directory
    if [ -d "$PROJECT_ROOT/pkg/grpc/metadata" ]; then
        METADATA_COUNT=$(find "$PROJECT_ROOT/pkg/grpc/metadata" -type f -name "*.json" 2>/dev/null | wc -l | tr -d ' ')
        if [ "$METADATA_COUNT" -gt 0 ]; then
            echo -e "${YELLOW}   Found $METADATA_COUNT metadata files${NC}"
            rm -rf "$PROJECT_ROOT/pkg/grpc/metadata"
            echo -e "${GREEN}   ✅ Deleted metadata directory${NC}"
        fi
    fi
else
    echo -e "${YELLOW}   ℹ️  pkg/grpc directory not found${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Cleanup Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Regenerate gRPC code for each service:"
echo "   cd attendance && go run cmd/generate-grpc/main.go"
echo "   cd ../fleet && go run cmd/generate-grpc/main.go"
echo "   cd ../route && go run cmd/generate-grpc/main.go"
echo "   # ... etc for each service"
echo ""
echo "2. Or use the full generation:"
echo "   cd attendance && make gen"
echo "   cd ../fleet && make gen"
echo "   # ... etc for each service"
echo ""
echo "3. Or regenerate all at once using the fix script:"
echo "   ./scripts/fix-all-grpc-generation.sh"
echo ""

