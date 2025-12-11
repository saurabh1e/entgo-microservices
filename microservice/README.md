# microservice

Microservice for microservice.

## Getting Started

```bash
# Generate code
go generate generate.go

# Build
go build -o bin/microservice main.go

# Run in development mode
./dev.sh dev
```

## Ports

- HTTP/GraphQL: 8082
- gRPC: 9082

## Database

- PostgreSQL: entgo_microservice_postgres
- Database name: microservice
