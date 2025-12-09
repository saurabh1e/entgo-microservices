# Syphoon GraphQL Gateway

An HTTP gateway that routes GraphQL and REST requests to underlying microservices. Uses chi, CORS, and a simple schema-aware router. Optional Redis for shared state.

## Overview
- GraphQL endpoint at /graphql forwards operations to target services based on root field
- GraphQL Playground at /playground
- REST passthrough at /api/v1/{service}/{path} forwards to service base URL
- CORS enabled (including Access-Control-Allow-Private-Network)

## Requirements
- Go 1.23+
- Redis (optional but recommended)

## Configuration
Environment variables (defaults shown):
- PORT: 8080
- REDIS_HOST: localhost
- REDIS_PORT: 6379
- REDIS_PASSWORD: ""
- REDIS_DB: 0
- AUTH_SERVICE_URL: http://localhost:8081/graphql
- MAIN_SERVICE_URL: http://localhost:8088/graphql
- GETGRASS_SERVICE_URL: http://localhost:8082/graphql

Notes:
- syphoon_main defaults to 8082 in its code, but this gateway uses 8088 as the default target; override MAIN_SERVICE_URL if needed.
- syphoon_getgrass service listens on 8085 by default; override GETGRASS_SERVICE_URL to http://localhost:8085/graphql if you use the default getgrass port.

## Run
```bash
# From syphoon_gateway/
go mod download

go run ./main.go
```

Server starts on http://localhost:8080 by default.

## Endpoints
- POST /graphql — GraphQL router
- GET /playground — in-browser GraphQL IDE pointing to /graphql
- Any /api/v1/{service}/{path} — REST passthrough to that service (service one of: auth, main, getgrass)

## How routing works
- Introspection queries are handled by an internal merged schema provider
- For normal GraphQL queries, the router extracts the root field and picks a service URL
- REST passthrough strips the /api/v1/{service} prefix and forwards headers/query/body

## Troubleshooting
- 400 Invalid GraphQL content-type: ensure Content-Type: application/json
- 502 Service unavailable: verify target service URL envs and that services are up
- CORS errors in tools like Apollo Studio: CORS headers are enabled; ensure you’re hitting /graphql and not a different path
