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

# Copy module files and the local pkg module so `replace ../pkg` resolves during go mod download
COPY gateway/go.mod gateway/go.sum ./
COPY pkg /pkg
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download

# Copy the rest of the source for the dev environment (mounted in compose for live edits)
COPY gateway .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# --------------------
# Builder for production
# --------------------
FROM golang:1.25.5-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
# Copy module files and local pkg module for build
COPY gateway/go.mod gateway/go.sum ./
COPY pkg /pkg
RUN go env -w GOPROXY=https://proxy.golang.org && go mod download
COPY gateway .
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
