#!/bin/bash

# Generate Dockerfile
generate_dockerfile() {
    local service_dir=$1
    local service_name=$2
    local auth_dir=$3
    local base_module=$4

    if [ -f "$auth_dir/Dockerfile" ]; then
        cp "$auth_dir/Dockerfile" "$service_dir/"
        # Replace port numbers
        sed -i.bak "s/8081/8082/g" "$service_dir/Dockerfile"
        # Replace auth references with the new service name
        sed -i.bak "s|COPY auth/go.mod auth/go.sum|COPY ${service_name}/go.mod ${service_name}/go.sum|g" "$service_dir/Dockerfile"
        # Fix the COPY . . commands to copy only the service directory
        sed -i.bak "s|^COPY \. \.$|COPY ${service_name} .|g" "$service_dir/Dockerfile"
        # Replace any auth directory references with service name
        sed -i.bak "s|COPY auth |COPY ${service_name} |g" "$service_dir/Dockerfile"
        rm "$service_dir/Dockerfile.bak"
    fi
}

# Generate docker-compose.yml
generate_docker_compose() {
    local service_dir=$1
    local service_name=$2

    cat > "$service_dir/docker-compose.yml" << EOF
services:
  ${service_name}:
    build:
      context: ..
      dockerfile: ${service_name}/Dockerfile
    container_name: entgo_${service_name}
    restart: on-failure
    env_file:
      - .env
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    ports:
      - "8082:8082"
      - "9082:9082"
    networks:
      - entgo_network

  postgres:
    image: postgres:15-alpine
    container_name: entgo_${service_name}_postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: ${service_name}
    volumes:
      - ${service_name}_postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - entgo_network

  redis:
    image: redis:7-alpine
    container_name: entgo_${service_name}_redis
    command: redis-server --appendonly yes
    volumes:
      - ${service_name}_redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - entgo_network

volumes:
  ${service_name}_postgres_data:
  ${service_name}_redis_data:

networks:
  entgo_network:
    external: true
EOF
}

# Generate docker-compose.dev.yml
generate_docker_compose_dev() {
    local service_dir=$1
    local service_name=$2

    cat > "$service_dir/docker-compose.dev.yml" << EOF

services:
  ${service_name}:
    build:
      context: ..
      dockerfile: ${service_name}/Dockerfile
      target: dev
    container_name: entgo_${service_name}_dev
    restart: on-failure
    env_file:
      - .env  # Path relative to this compose file (${service_name} directory)
    environment:
      # Override specific vars for Docker environment
      DB_HOST: entgo_${service_name}_postgres
      REDIS_HOST: entgo_redis_shared

      # Go build cache
      GOMODCACHE: /go/pkg/mod
      GOCACHE: /root/.cache/go-build
      CGO_ENABLED: 0
    ports:
      - "8082:8082"  # HTTP/GraphQL
      - "9082:9082"  # gRPC
    volumes:
      - ./:/src:cached
      - ../pkg:/pkg:cached
      - gomodcache:/go/pkg/mod
      - gocache:/root/.cache/go-build
    command: ["air", "-c", ".air.toml"]
    depends_on:
      ${service_name}_postgres:
        condition: service_healthy
    networks:
      - entgo_network

  ${service_name}_postgres:
    image: postgres:15-alpine
    container_name: entgo_${service_name}_postgres
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: ${service_name}
    volumes:
      - ${service_name}_postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - entgo_network

volumes:
  ${service_name}_postgres_data:
  gomodcache:
  gocache:

networks:
  entgo_network:
    external: true
EOF
}

