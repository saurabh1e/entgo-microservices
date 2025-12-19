#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Array of microservices
SERVICES=("auth" "attendance" "route" "yard" "reporting" "notification" "core" "fleet")

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Building All Microservices in Parallel${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Sync workspace dependencies first
echo -e "${YELLOW}Syncing workspace dependencies...${NC}"
cd "$PROJECT_ROOT"
if go work sync; then
    echo -e "${GREEN}✓ Workspace dependencies synced${NC}"
else
    echo -e "${RED}✗ Workspace sync failed${NC}"
    exit 1
fi
echo ""

# Function to clean and generate for a service
build_service() {
    local service=$1
    local service_dir="$PROJECT_ROOT/$service"

    echo -e "${YELLOW}[$service]${NC} Starting build process..."

    # Check if service directory exists
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}[$service]${NC} Directory not found: $service_dir"
        return 1
    fi

    # Navigate to service directory
    cd "$service_dir"

    # Run dev.sh clean (required)
    if [ -f "./dev.sh" ]; then
        echo -e "${YELLOW}[$service]${NC} Running ./dev.sh clean..."
        if ./dev.sh clean 2>&1 | sed "s/^/[$service] /"; then
            echo -e "${GREEN}[$service]${NC} ✓ Clean completed"
        else
            echo -e "${RED}[$service]${NC} ✗ Clean failed"
            return 1
        fi
    else
        echo -e "${RED}[$service]${NC} ✗ dev.sh not found"
        return 1
    fi

    # Run go generate (required)
    if [ -f "./generate.go" ]; then
        echo -e "${YELLOW}[$service]${NC} Running go generate generate.go..."
        if go generate generate.go 2>&1 | sed "s/^/[$service] /"; then
            echo -e "${GREEN}[$service]${NC} ✓ Code generation completed"
        else
            echo -e "${RED}[$service]${NC} ✗ Code generation failed"
            return 1
        fi
    else
        echo -e "${RED}[$service]${NC} ✗ generate.go not found"
        return 1
    fi

    echo -e "${GREEN}[$service]${NC} ✅ Build process completed successfully"
    return 0
}

# Export the function so it can be used by parallel processes
export -f build_service
export PROJECT_ROOT
export RED GREEN YELLOW BLUE NC

# Run builds in parallel
echo -e "${BLUE}Starting parallel builds...${NC}"
echo ""

# Create a temporary directory for tracking results
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Run all services in parallel
for service in "${SERVICES[@]}"; do
    (
        if build_service "$service"; then
            touch "$TEMP_DIR/$service.success"
        else
            touch "$TEMP_DIR/$service.failed"
        fi
    ) &
done

# Wait for all background jobs to complete
wait

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Build Summary${NC}"
echo -e "${BLUE}========================================${NC}"

# Check results
failed_services=()
successful_services=()

for service in "${SERVICES[@]}"; do
    if [ -f "$TEMP_DIR/$service.success" ]; then
        successful_services+=("$service")
        echo -e "${GREEN}✓${NC} $service"
    elif [ -f "$TEMP_DIR/$service.failed" ]; then
        failed_services+=("$service")
        echo -e "${RED}✗${NC} $service"
    else
        echo -e "${YELLOW}?${NC} $service (unknown status)"
    fi
done

echo ""
echo -e "Successful: ${GREEN}${#successful_services[@]}${NC} / ${#SERVICES[@]}"

if [ ${#failed_services[@]} -gt 0 ]; then
    echo -e "Failed: ${RED}${#failed_services[@]}${NC}"
    echo ""
    echo -e "${RED}The following services failed to build:${NC}"
    for service in "${failed_services[@]}"; do
        echo -e "  - $service"
    done
    exit 1
else
    echo -e "${GREEN}All services built successfully!${NC}"
    exit 0
fi

