# syntax=docker/dockerfile:1.5

# Base image
FROM golang:1.25.5-alpine AS base
RUN apk add --no-cache git ca-certificates bash
WORKDIR /src

# --------------------
# Dev image (hot-reload)
# --------------------
FROM base AS dev
RUN apk add --no-cache build-base

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Install air for hot-reload
RUN go install github.com/cosmtrek/air@latest || go install github.com/air-verse/air@latest

# Cache go modules
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download

COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# --------------------
# Builder for production
# --------------------
FROM golang:1.25.5-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/main ./

# --------------------
# Production image (final)
# --------------------
FROM alpine:3.18 AS prod
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/tmp/main /app/main
COPY .env /app/.env
EXPOSE 8080
ENTRYPOINT ["/app/main"]
