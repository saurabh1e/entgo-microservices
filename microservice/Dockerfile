# syntax=docker/dockerfile:1.5

# Base image with common packages
FROM golang:1.25.5-alpine AS base
RUN apk add --no-cache git ca-certificates bash
WORKDIR /src

# --------------------
# Dev image (hot-reload)
# --------------------
FROM base AS dev
# Install build tools needed by some go packages and for `go install` to work reliably
# Also install protoc and protobuf-dev for gRPC code generation
RUN apk add --no-cache build-base protobuf protobuf-dev

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Install air (hot-reload) into GOPATH/bin
# Try the popular cosmtrek/air first, fall back to air-verse if unavailable
RUN go install github.com/cosmtrek/air@latest || go install github.com/air-verse/air@latest

# Install protoc-gen-go and protoc-gen-go-grpc for gRPC code generation
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy module files and the local pkg module so `replace ../pkg` resolves during go mod download
COPY microservice/go.mod microservice/go.sum ./
COPY pkg /pkg
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download

# Copy the rest of the source for the dev environment (mounted in compose for live edits)
COPY microservice .

# Copy and set up the development entrypoint script
COPY auth/docker-entrypoint-dev.sh /usr/local/bin/docker-entrypoint-dev.sh
RUN chmod +x /usr/local/bin/docker-entrypoint-dev.sh

EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/docker-entrypoint-dev.sh"]
CMD ["air", "-c", ".air.toml"]

# --------------------
# Builder for production
# --------------------
FROM golang:1.25.5-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Copy module files and local pkg module for build
COPY microservice/go.mod microservice/go.sum ./
COPY pkg /pkg
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download
COPY microservice .

# Build a static binary for smaller final image
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/main ./

# --------------------
# Production image (final)
# --------------------
FROM alpine:3.18 AS prod
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy binary from builder
COPY --from=builder /src/tmp/main /app/main

# Copy default .env example (do not include secrets in image — prefer docker-compose env_file)
# If you want to avoid copying .env into image altogether, remove the next line
COPY .env /app/.env

EXPOSE 8082
ENTRYPOINT ["/app/main"]
