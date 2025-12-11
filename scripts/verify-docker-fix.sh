#!/bin/bash

# Verification script to check if all services have the Docker fix applied

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Docker Fix Verification ===${NC}\n"

SERVICES=("attendance" "auth" "route" "yard")
ALL_GOOD=true

for service in "${SERVICES[@]}"; do
    echo -e "${BLUE}Checking $service...${NC}"

    # Check if docker-entrypoint-dev.sh exists
    if [ -f "$service/docker-entrypoint-dev.sh" ]; then
        echo -e "  ${GREEN}✓${NC} docker-entrypoint-dev.sh exists"
    else
        echo -e "  ${RED}✗${NC} docker-entrypoint-dev.sh missing"
        ALL_GOOD=false
    fi

    # Check if docker-entrypoint-dev.sh is executable
    if [ -x "$service/docker-entrypoint-dev.sh" ]; then
        echo -e "  ${GREEN}✓${NC} docker-entrypoint-dev.sh is executable"
    else
        echo -e "  ${RED}✗${NC} docker-entrypoint-dev.sh is not executable"
        ALL_GOOD=false
    fi

    # Check if generate.go uses gqlgen (not go run github.com/99designs/gqlgen)
    if grep -q "//go:generate gqlgen" "$service/generate.go" 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} generate.go uses gqlgen binary"
    else
        echo -e "  ${RED}✗${NC} generate.go still uses 'go run github.com/99designs/gqlgen'"
        ALL_GOOD=false
    fi

    # Check if Dockerfile has ENTRYPOINT set
    if grep -q "ENTRYPOINT.*docker-entrypoint-dev.sh" "$service/Dockerfile" 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} Dockerfile has ENTRYPOINT set"
    else
        echo -e "  ${RED}✗${NC} Dockerfile missing ENTRYPOINT"
        ALL_GOOD=false
    fi

    echo ""
done

echo -e "${BLUE}=== Summary ===${NC}"
if [ "$ALL_GOOD" = true ]; then
    echo -e "${GREEN}✓ All services have the Docker fix applied correctly!${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo "1. Rebuild all services:"
    echo "   cd <service> && docker-compose -f docker-compose.dev.yml build"
    echo ""
    echo "2. Restart all services:"
    echo "   cd <service> && docker-compose -f docker-compose.dev.yml down && docker-compose -f docker-compose.dev.yml up -d"
    exit 0
else
    echo -e "${RED}✗ Some services are missing the Docker fix${NC}"
    echo ""
    echo "Please review the issues above and apply the fix as documented in docs/DOCKER_FIX.md"
    exit 1
fi

