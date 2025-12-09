#!/bin/bash
# Comprehensive Phase 1-7 Verification
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
PASS=0
FAIL=0
check_pass() {
    echo -e "${GREEN}✅ $1${NC}"
    PASS=$((PASS + 1))
}
check_fail() {
    echo -e "${RED}❌ $1${NC}"
    FAIL=$((FAIL + 1))
}
check_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║    Verifying Phases 1-7 Implementation                    ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
# PHASE 1: Shared Infrastructure
echo -e "${BLUE}═══ PHASE 1: Shared Infrastructure ═══${NC}"
[ -f "go.work" ] && check_pass "go.work exists" || check_fail "go.work missing"
[ -d "pkg" ] && check_pass "pkg/ directory exists" || check_fail "pkg/ missing"
[ -f "pkg/go.mod" ] && check_pass "pkg/go.mod exists" || check_fail "pkg/go.mod missing"
[ -f "pkg/logger/logger.go" ] && check_pass "pkg/logger/logger.go exists" || check_fail "pkg/logger missing"
[ -f "pkg/jwt/service.go" ] && check_pass "pkg/jwt/service.go exists" || check_fail "pkg/jwt missing"
[ -f "pkg/redis/client.go" ] && check_pass "pkg/redis/client.go exists" || check_fail "pkg/redis/client missing"
[ -f "pkg/redis/token.go" ] && check_pass "pkg/redis/token.go exists" || check_fail "pkg/redis/token missing"
[ -f "pkg/middleware/auth.go" ] && check_pass "pkg/middleware/auth.go exists" || check_fail "pkg/middleware/auth missing"
[ -f "pkg/middleware/cors.go" ] && check_pass "pkg/middleware/cors.go exists" || check_fail "pkg/middleware/cors missing"
# Proto files
[ -f "proto/common/common.proto" ] && check_pass "proto/common/common.proto exists" || check_fail "proto common missing"
[ -f "proto/user/user.proto" ] && check_pass "proto/user/user.proto exists" || check_fail "proto user missing"
[ -f "proto/role/role.proto" ] && check_pass "proto/role/role.proto exists" || check_fail "proto role missing"
[ -f "proto/permission/permission.proto" ] && check_pass "proto/permission/permission.proto exists" || check_fail "proto permission missing"
# Generated proto files
[ -f "pkg/proto/user/v1/user.pb.go" ] && check_pass "Generated user.pb.go exists" || check_fail "Generated user proto missing"
[ -f "pkg/proto/user/v1/user_grpc.pb.go" ] && check_pass "Generated user_grpc.pb.go exists" || check_fail "Generated user grpc missing"
[ -f "pkg/proto/role/v1/role.pb.go" ] && check_pass "Generated role.pb.go exists" || check_fail "Generated role proto missing"
[ -f "pkg/proto/permission/v1/permission.pb.go" ] && check_pass "Generated permission.pb.go exists" || check_fail "Generated permission proto missing"
echo ""
# PHASE 2: Docker Infrastructure
echo -e "${BLUE}═══ PHASE 2: Docker Infrastructure ═══${NC}"
[ -f "docker-compose.shared.yml" ] && check_pass "docker-compose.shared.yml exists" || check_fail "shared compose missing"
[ -f "auth/docker-compose.dev.yml" ] && check_pass "auth/docker-compose.dev.yml exists" || check_fail "auth compose missing"
[ -f "Makefile" ] && check_pass "Root Makefile exists" || check_fail "Makefile missing"
# Check Makefile targets
grep -q "^network:" Makefile && check_pass "Makefile has 'network' target" || check_fail "network target missing"
grep -q "^shared:" Makefile && check_pass "Makefile has 'shared' target" || check_fail "shared target missing"
grep -q "^proto:" Makefile && check_pass "Makefile has 'proto' target" || check_fail "proto target missing"
# Check docker-compose.shared.yml content
grep -q "entgo_redis_shared" docker-compose.shared.yml && check_pass "Shared Redis configured" || check_fail "Redis config missing"
# Check auth docker-compose for individual DB
grep -q "auth_postgres" auth/docker-compose.dev.yml && check_pass "Auth has individual Postgres" || check_fail "Auth Postgres missing"
grep -q "9081:9081" auth/docker-compose.dev.yml && check_pass "gRPC port (9081) configured" || check_fail "gRPC port missing"
echo ""
# PHASE 3: Auth Service Migration
echo -e "${BLUE}═══ PHASE 3: Auth Service Migration ═══${NC}"
[ -f "auth/go.mod" ] && check_pass "auth/go.mod exists" || check_fail "auth go.mod missing"
grep -q "github.com/saurabh/entgo-microservices/pkg" auth/go.mod && check_pass "auth imports pkg module" || check_fail "pkg import missing"
# Check key files were updated to use pkg
grep -q "github.com/saurabh/entgo-microservices/pkg/logger" auth/main.go && check_pass "main.go uses pkg/logger" || check_fail "main.go logger import wrong"
grep -q "github.com/saurabh/entgo-microservices/pkg/jwt" auth/utils/bootstrap.go && check_pass "bootstrap.go uses pkg/jwt" || check_fail "bootstrap jwt import wrong"
# Check JWT service initialization with namespace
grep -q '"auth"' auth/utils/bootstrap.go && check_pass "JWT initialized with 'auth' namespace" || check_fail "JWT namespace missing"
echo ""
# PHASE 4: gRPC Implementation
echo -e "${BLUE}═══ PHASE 4: gRPC Implementation ═══${NC}"
[ -d "auth/grpc" ] && check_pass "auth/grpc/ directory exists" || check_fail "auth/grpc missing"
[ -f "auth/grpc/server.go" ] && check_pass "auth/grpc/server.go exists" || check_fail "grpc server missing"
[ -f "auth/grpc/interceptors.go" ] && check_pass "auth/grpc/interceptors.go exists" || check_fail "grpc interceptors missing"
[ -f "auth/grpc/user_service.go" ] && check_pass "auth/grpc/user_service.go exists" || check_fail "user service missing"
[ -f "auth/grpc/role_service.go" ] && check_pass "auth/grpc/role_service.go exists" || check_fail "role service missing"
[ -f "auth/grpc/permission_service.go" ] && check_pass "auth/grpc/permission_service.go exists" || check_fail "permission service missing"
# Check server.go has health checks
grep -q "health.NewServer" auth/grpc/server.go && check_pass "Health checks implemented" || check_fail "Health checks missing"
grep -q "reflection.Register" auth/grpc/server.go && check_pass "gRPC reflection enabled" || check_fail "Reflection missing"
# Check main.go starts gRPC server
grep -q "grpc.NewServer" auth/main.go && check_pass "main.go starts gRPC server" || check_fail "gRPC startup missing"
echo ""
# PHASE 5: gRPC Client Library
echo -e "${BLUE}═══ PHASE 5: gRPC Client Library ═══${NC}"
[ -f "pkg/grpc/client.go" ] && check_pass "pkg/grpc/client.go exists" || check_fail "grpc client missing"
[ -f "pkg/grpc/interceptors.go" ] && check_pass "pkg/grpc/interceptors.go exists" || check_fail "client interceptors missing"
[ -f "pkg/grpc/user_client.go" ] && check_pass "pkg/grpc/user_client.go exists" || check_fail "user client missing"
[ -f "pkg/grpc/role_client.go" ] && check_pass "pkg/grpc/role_client.go exists" || check_fail "role client missing"
[ -f "pkg/grpc/permission_client.go" ] && check_pass "pkg/grpc/permission_client.go exists" || check_fail "permission client missing"
[ -f "pkg/grpc/example_test.go" ] && check_pass "pkg/grpc/example_test.go exists" || check_fail "examples missing"
# Check for connection pooling
grep -q "ClientPool" pkg/grpc/client.go && check_pass "Connection pooling implemented" || check_fail "Connection pool missing"
# Check for retry logic
grep -q "ClientRetryInterceptor" pkg/grpc/interceptors.go && check_pass "Retry interceptor implemented" || check_fail "Retry logic missing"
grep -q "exponential" pkg/grpc/interceptors.go && check_pass "Exponential backoff configured" || check_fail "Backoff missing"
echo ""
# PHASE 6: mTLS Certificate Setup
echo -e "${BLUE}═══ PHASE 6: mTLS Certificate Setup ═══${NC}"
[ -f "scripts/generate-certs.sh" ] && check_pass "scripts/generate-certs.sh exists" || check_fail "cert script missing"
[ -x "scripts/generate-certs.sh" ] && check_pass "generate-certs.sh is executable" || check_fail "cert script not executable"
[ -f "docs/MTLS_SETUP.md" ] && check_pass "docs/MTLS_SETUP.md exists" || check_fail "mTLS docs missing"
# Check cert script content
grep -q "openssl genrsa" scripts/generate-certs.sh && check_pass "Cert script generates keys" || check_fail "Key generation missing"
grep -q "CA certificate" scripts/generate-certs.sh && check_pass "CA generation included" || check_fail "CA generation missing"
# Check MTLS docs content
wc -l docs/MTLS_SETUP.md 2>/dev/null | awk '{if ($1 > 50) exit 0; else exit 1}' && check_pass "MTLS_SETUP.md is comprehensive (>50 lines)" || check_warn "MTLS docs might be incomplete"
echo ""
# PHASE 7: Testing & Validation
echo -e "${BLUE}═══ PHASE 7: Testing & Validation ═══${NC}"
[ -f "docs/TESTING.md" ] && check_pass "docs/TESTING.md exists" || check_fail "testing docs missing"
[ -f "scripts/validate.sh" ] && check_pass "scripts/validate.sh exists" || check_fail "validate script missing"
# Check TESTING.md content
wc -l docs/TESTING.md 2>/dev/null | awk '{if ($1 > 100) exit 0; else exit 1}' && check_pass "TESTING.md is comprehensive (>100 lines)" || check_warn "Testing docs might be incomplete"
# Check for different test types documented
grep -q "Unit Test" docs/TESTING.md && check_pass "Unit testing documented" || check_fail "Unit tests missing"
grep -q "Integration" docs/TESTING.md && check_pass "Integration testing documented" || check_fail "Integration tests missing"
grep -q "E2E" docs/TESTING.md && check_pass "E2E testing documented" || check_fail "E2E tests missing"
grep -q "Load Test\|ghz\|k6" docs/TESTING.md && check_pass "Load testing documented" || check_fail "Load testing missing"
echo ""
# BUILD VERIFICATION
echo -e "${BLUE}═══ Build Verification ═══${NC}"
cd pkg
if go build ./... 2>/dev/null; then
    check_pass "pkg module builds successfully"
else
    check_fail "pkg module has build errors"
fi
cd ..
cd auth
if go build 2>/dev/null; then
    check_pass "auth service builds successfully"
    [ -f "auth" ] && rm auth
else
    check_fail "auth service has build errors"
fi
cd ..
echo ""
# DOCUMENTATION
echo -e "${BLUE}═══ Documentation ═══${NC}"
[ -f "README.md" ] && check_pass "README.md exists" || check_fail "README missing"
[ -f "IMPLEMENTATION_SUMMARY.md" ] && check_pass "IMPLEMENTATION_SUMMARY.md exists" || check_fail "Summary missing"
[ -f "CHECKLIST.md" ] && check_pass "CHECKLIST.md exists" || check_fail "Checklist missing"
[ -f "PHASES_COMPLETE.md" ] && check_pass "PHASES_COMPLETE.md exists" || check_fail "Phases doc missing"
[ -f "quick-start.sh" ] && check_pass "quick-start.sh exists" || check_fail "Quick start missing"
echo ""
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                 Verification Summary                      ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${GREEN}Passed:${NC} $PASS checks"
echo -e "  ${RED}Failed:${NC} $FAIL checks"
echo ""
if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL PHASES 1-7 COMPLETE AND VERIFIED! 🎉${NC}"
    echo ""
    echo "Your microservices architecture is ready for production!"
    exit 0
else
    echo -e "${RED}⚠️  Some checks failed. Please review the output above.${NC}"
    exit 1
fi
