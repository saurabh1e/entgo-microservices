# test-service

Microservice for test-service.

## Getting Started

```bash
# Generate code
go generate generate.go

# Build
go build -o bin/test-service main.go

# Run in development mode
./dev.sh dev
```

## Ports

- HTTP/GraphQL: 8087
- gRPC: 9082

## Database

- PostgreSQL: entgo_test-service_postgres
- Database name: test-service
