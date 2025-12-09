#!/bin/bash

# Verification script for newly created microservices
# Usage: ./verify-service.sh <service-name>

set -e

SERVICE_NAME="$1"

if [ -z "$SERVICE_NAME" ]; then
    echo "Usage: $0 <service-name>"
    exit 1
fi

echo "🔍 Verifying service: $SERVICE_NAME"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

check_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if service directory exists
if [ ! -d "$SERVICE_NAME" ]; then
    check_fail "Service directory '$SERVICE_NAME' not found"
fi

cd "$SERVICE_NAME"

echo "📁 Checking directory structure..."
[ -d "graph/model" ] && check_pass "graph/model directory exists" || check_fail "graph/model directory missing"
[ -f "graph/model/bootstrap.go" ] && check_pass "graph/model/bootstrap.go exists" || check_fail "bootstrap.go missing"
[ -d "internal/ent" ] && check_pass "internal/ent directory exists" || check_warning "internal/ent not generated yet (run 'make gen')"

echo ""
echo "📝 Checking import paths..."

# Check main.go for correct imports
if grep -q "github.com/saurabh/entgo-microservices/$SERVICE_NAME" main.go; then
    check_pass "main.go uses full module paths"
else
    check_fail "main.go has incorrect import paths"
fi

# Check grpc/server.go for correct imports
if [ -f "grpc/server.go" ]; then
    if grep -q "github.com/saurabh/entgo-microservices/$SERVICE_NAME" grpc/server.go; then
        check_pass "grpc/server.go uses full module paths"
    else
        check_fail "grpc/server.go has incorrect import paths"
    fi
fi

# Check gqlgen.yml for correct autobind path
if grep -q "github.com/saurabh/entgo-microservices/$SERVICE_NAME/internal/ent" gqlgen.yml; then
    check_pass "gqlgen.yml has correct autobind path"
else
    check_fail "gqlgen.yml has incorrect autobind path"
fi

echo ""
echo "🌐 Checking gateway integration..."
cd ..

SERVICE_NAME_UPPER=$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_')

if [ -d "gateway" ]; then
    # Check gateway.go
    if grep -q "\"$SERVICE_NAME\"" gateway/gateway.go; then
        check_pass "Service registered in gateway/gateway.go"
    else
        check_warning "Service not found in gateway/gateway.go"
    fi

    # Check config.go
    if grep -q "${SERVICE_NAME_UPPER}ServiceURL" gateway/utils/config.go; then
        check_pass "Service configuration in gateway/utils/config.go"
    else
        check_warning "Service configuration not in gateway/utils/config.go"
    fi

    # Check .env
    if grep -q "${SERVICE_NAME_UPPER}_SERVICE_URL" gateway/.env; then
        check_pass "Service URL in gateway/.env"
    else
        check_warning "Service URL not in gateway/.env"
    fi
else
    check_warning "Gateway directory not found"
fi

echo ""
echo "🏗️  Testing build..."
cd "$SERVICE_NAME"

# Try to build
if go build -o bin/test main.go 2>/dev/null; then
    check_pass "Service builds successfully"
    rm -f bin/test
else
    check_warning "Service build failed (may need 'make gen' first)"
fi

echo ""
echo "✅ Verification complete!"
echo ""
echo "Next steps:"
echo "  1. cd $SERVICE_NAME"
echo "  2. make gen  # Generate all code"
echo "  3. make build  # Build the service"
echo "  4. ./dev.sh dev  # Run in development mode"

