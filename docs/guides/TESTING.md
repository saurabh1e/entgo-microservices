# Testing Guide
Comprehensive testing strategy for the microservices architecture.
## Testing Pyramid
```
           /\
          /E2E\         End-to-End Tests
         /______\
        /        \
       /Integration\    Integration Tests
      /____________\
     /              \
    /   Unit Tests   \  Unit Tests (Most)
   /__________________\
```
## 1. Unit Tests
Test individual components in isolation.
### Example: JWT Service Test
```go
package jwt_test
import (
    "testing"
    "github.com/saurabh/entgo-microservices/pkg/jwt"
)
func TestJWTService_GenerateTokenPair(t *testing.T) {
    jwtService := jwt.NewService("secret", 24, redisClient, "test")
    accessToken, refreshToken, err := jwtService.GenerateTokenPair(ctx, 1, "user", "email")
    if err != nil {
        t.Fatal(err)
    }
    if accessToken == "" {
        t.Error("access token is empty")
    }
}
```
### Run Unit Tests
```bash
cd pkg
go test ./... -v
go test ./... -cover
```
## 2. Integration Tests
Test interactions between components.
### Example: gRPC Service Test
```go
func TestUserService_Integration(t *testing.T) {
    client := enttest.Open(t, "sqlite3", "file:ent?mode=memory")
    defer client.Close()
    service := grpc.NewUserService(client)
    resp, err := service.GetUserByID(ctx, &userv1.GetUserByIDRequest{Id: 1})
    if err != nil {
        t.Fatal(err)
    }
}
```
### Run Integration Tests
```bash
go test ./... -v -tags=integration
```
## 3. End-to-End Tests
Test complete flows across services.
### Run E2E Tests
```bash
# Start services
make up
# Run E2E tests
go test ./tests/e2e/... -v -timeout 5m
```
## 4. Load Testing
### Using ghz (gRPC benchmarking)
```bash
# Install ghz
go install github.com/bojand/ghz/cmd/ghz@latest
# Run benchmark
ghz --insecure \
  --proto ./proto/user/user.proto \
  --call user.v1.UserService/GetUserByID \
  -d '{"id":1}' \
  -c 50 \
  -n 10000 \
  localhost:9081
```
### Using k6 (HTTP/GraphQL load testing)
```javascript
import http from 'k6/http';
import { check } from 'k6';
export const options = {
  vus: 50,
  duration: '30s',
};
export default function () {
  const res = http.post('http://localhost:8081/graphql', JSON.stringify({
    query: '{ me { id username } }'
  }));
  check(res, { 'status is 200': (r) => r.status === 200 });
}
```
## 5. Manual Testing
### Test GraphQL
```bash
# Open playground
open http://localhost:8081/playground
# Example query
query {
  me {
    id
    username
    email
  }
}
```
### Test gRPC
```bash
# List services
grpcurl -plaintext localhost:9081 list
# Call method
grpcurl -plaintext -d '{"id": 1}' \
  localhost:9081 user.v1.UserService/GetUserByID
# Health check
grpcurl -plaintext localhost:9081 \
  grpc.health.v1.Health/Check
```
### Test Redis
```bash
# Connect to Redis
docker exec -it entgo_redis_shared redis-cli
# Check keys
127.0.0.1:6379> KEYS auth:*
127.0.0.1:6379> GET auth:whitelist:token123
```
## CI/CD Integration
### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:7-alpine
        ports: [6379:6379]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      - name: Run tests
        run: go test ./... -v -cover
```
## Best Practices
1. **Isolate tests** - Each test should be independent
2. **Use test databases** - Never test against production
3. **Clean up** - Reset state after each test
4. **Mock external services** - Use testcontainers
5. **Test edge cases** - Not just happy paths
6. **Automate** - Run tests in CI/CD
## Test Coverage Goals
- Unit Tests: > 80%
- Integration Tests: All critical paths
- E2E Tests: All user journeys
