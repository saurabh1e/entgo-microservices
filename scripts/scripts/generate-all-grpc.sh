#!/usr/bin/env bash
#
# generate-all-grpc-centralized.sh
# Centralized script to generate gRPC proto files and services for all microservices
# This script contains all the logic - no individual service scripts needed
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

# List of all microservices
ALL_SERVICES=(
    "attendance"
    "fleet"
    "route"
    "core"
    "notification"
    "reporting"
    "yard"
    "auth"
)

# Function to get service directory
get_service_dir() {
    echo "$PROJECT_ROOT/$1"
}

# Parse command line arguments
SERVICES_TO_GENERATE=()
CLEAN_FIRST=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN_FIRST=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --service|-s)
            SERVICES_TO_GENERATE+=("$2")
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Generate gRPC proto files and services for microservices"
            echo ""
            echo "Options:"
            echo "  --clean              Clean existing proto and gRPC files before generating"
            echo "  --service, -s NAME   Generate only for specific service (can be used multiple times)"
            echo "  --verbose, -v        Show detailed output"
            echo "  --help, -h           Show this help message"
            echo ""
            echo "Available services:"
            for service in "${ALL_SERVICES[@]}"; do
                echo "  - $service"
            done
            echo ""
            echo "Examples:"
            echo "  $0                          # Generate for all services"
            echo "  $0 --clean                  # Clean and regenerate all"
            echo "  $0 -s attendance -s fleet   # Generate only for attendance and fleet"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# If no specific services specified, generate for all
if [ ${#SERVICES_TO_GENERATE[@]} -eq 0 ]; then
    SERVICES_TO_GENERATE=("${ALL_SERVICES[@]}")
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}gRPC Generation Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Clean if requested
if [ "$CLEAN_FIRST" = true ]; then
    echo -e "${YELLOW}🧹 Cleaning existing proto and gRPC files...${NC}"
    if [ -f "$SCRIPT_DIR/clean-all-grpc.sh" ]; then
        "$SCRIPT_DIR/clean-all-grpc.sh"
    else
        echo -e "${YELLOW}   ⚠️  clean-all-grpc.sh not found, skipping cleanup${NC}"
    fi
    echo ""
fi

# Check if generate-grpc-for-service.go exists
GENERATOR_SCRIPT="$PROJECT_ROOT/scripts/generate-grpc-for-service.go"

if [ ! -f "$GENERATOR_SCRIPT" ]; then
    echo -e "${RED}❌ Error: $GENERATOR_SCRIPT not found${NC}"
    echo -e "${YELLOW}Creating the generator script...${NC}"

    # This shouldn't happen in normal operation, but provide helpful error
    echo -e "${RED}Please ensure generate-grpc-for-service.go exists in scripts/ directory${NC}"
    exit 1
fi

# Track results
RESULTS_SERVICES=()
RESULTS_STATUS=()
TOTAL=0
SUCCESS=0
FAILED=0

echo -e "${BLUE}Generating gRPC code for ${#SERVICES_TO_GENERATE[@]} service(s)...${NC}"
echo ""

for service in "${SERVICES_TO_GENERATE[@]}"; do
    TOTAL=$((TOTAL + 1))
    SERVICE_DIR=$(get_service_dir "$service")

    if [ ! -d "$SERVICE_DIR" ]; then
        echo -e "${RED}[$service]${NC} ❌ Directory not found: $SERVICE_DIR"
        RESULTS_SERVICES+=("$service")
        RESULTS_STATUS+=("FAILED: Directory not found")
        FAILED=$((FAILED + 1))
        continue
    fi

    # Check if service has ent schemas
    SCHEMA_DIR="$SERVICE_DIR/ent/schema"
    if [ ! -d "$SCHEMA_DIR" ]; then
        echo -e "${YELLOW}[$service]${NC} ⚠️  Skipping - no ent/schema directory found"
        RESULTS_SERVICES+=("$service")
        RESULTS_STATUS+=("SKIPPED: No schemas")
        continue
    fi

    echo -e "${BLUE}[$service]${NC} 🔧 Generating gRPC code..."

    cd "$SERVICE_DIR"

    # Run the centralized generator for this service
    if [ "$VERBOSE" = true ]; then
        if go run "$GENERATOR_SCRIPT" -service="$service" 2>&1; then
            echo -e "${GREEN}[$service]${NC} ✅ Generation successful"
            RESULTS_SERVICES+=("$service")
            RESULTS_STATUS+=("SUCCESS")
            SUCCESS=$((SUCCESS + 1))
        else
            echo -e "${RED}[$service]${NC} ❌ Generation failed"
            RESULTS_SERVICES+=("$service")
            RESULTS_STATUS+=("FAILED: Generation error")
            FAILED=$((FAILED + 1))
        fi
    else
        OUTPUT=$(go run "$GENERATOR_SCRIPT" -service="$service" 2>&1)
        if [ $? -eq 0 ]; then
            echo "$OUTPUT" | grep -E "Successfully generated|Compiled proto|gRPC generation completed" | tail -5 || true
            echo -e "${GREEN}[$service]${NC} ✅ Generation successful"
            RESULTS_SERVICES+=("$service")
            RESULTS_STATUS+=("SUCCESS")
            SUCCESS=$((SUCCESS + 1))
        else
            echo -e "${RED}[$service]${NC} ❌ Generation failed"
            if [ ! -z "$OUTPUT" ]; then
                echo "$OUTPUT" | tail -10 | sed "s/^/  /"
            fi
            RESULTS_SERVICES+=("$service")
            RESULTS_STATUS+=("FAILED: Generation error")
            FAILED=$((FAILED + 1))
        fi
    fi

    echo ""
done

# Return to project root
cd "$PROJECT_ROOT"

# Generate consolidated service clients if we have metadata
if [ -d "$PROJECT_ROOT/pkg/grpc/metadata" ]; then
    METADATA_COUNT=$(find "$PROJECT_ROOT/pkg/grpc/metadata" -name "*_services.json" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$METADATA_COUNT" -gt 0 ]; then
        echo -e "${BLUE}📦 Generating consolidated service clients...${NC}"
        if [ -f "$SCRIPT_DIR/generate-service-clients.sh" ]; then
            if "$SCRIPT_DIR/generate-service-clients.sh" 2>&1 | tail -5; then
                echo -e "${GREEN}✅ Service clients generated${NC}"
            else
                echo -e "${YELLOW}⚠️  Service client generation had warnings${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  generate-service-clients.sh not found${NC}"
        fi
        echo ""
    fi
fi

# Print summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Generation Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Total services: ${BLUE}$TOTAL${NC}"
echo -e "Successful:     ${GREEN}$SUCCESS${NC}"
echo -e "Failed:         ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${YELLOW}Failed services:${NC}"
    for i in "${!RESULTS_SERVICES[@]}"; do
        service="${RESULTS_SERVICES[$i]}"
        status="${RESULTS_STATUS[$i]}"
        if [[ "$status" == FAILED* ]]; then
            echo -e "  ${RED}✗${NC} $service: $status"
        fi
    done
    echo ""
fi

if [ $SUCCESS -eq $TOTAL ]; then
    echo -e "${GREEN}✅ All services generated successfully!${NC}"
    exit 0
elif [ $FAILED -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Some services failed to generate${NC}"
    echo -e "${YELLOW}Run with --verbose flag for detailed output${NC}"
    exit 1
else
    echo -e "${GREEN}✅ Generation complete${NC}"
    exit 0
fi

