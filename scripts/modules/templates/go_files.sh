#!/bin/bash

# Generate go.mod file
generate_go_mod() {
    local service_dir=$1
    local base_module=$2
    local service_name=$3

    cat > "$service_dir/go.mod" << EOF
module ${base_module}/${service_name}

go 1.25.5

require (
	entgo.io/contrib v0.7.0
	entgo.io/ent v0.14.5
	github.com/99designs/gqlgen v0.17.84
	github.com/gin-gonic/gin v1.11.0
	github.com/gorilla/websocket v1.5.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/redis/go-redis/v9 v9.7.1
	github.com/saurabh/entgo-microservices/pkg v0.0.0
	github.com/vektah/gqlparser/v2 v2.5.24
	golang.org/x/crypto v0.33.0
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.4
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217
)

replace github.com/saurabh/entgo-microservices/pkg => ../pkg

exclude google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6
EOF
}

# Generate gqlgen.yml file
generate_gqlgen_yml() {
    local service_dir=$1
    local base_module=$2
    local service_name=$3

    cat > "$service_dir/gqlgen.yml" << EOF
# GraphQL schema files
schema:
  - "graph/*.graphqls"
  - "graph/**/**/*.graphqls"

# Generated server code
exec:
  filename: graph/generated.go
  package: graph
  dir: graph
  layout: single-file

# Generated models
model:
  filename: graph/model/models_gen.go
  package: model

# Resolver implementation structure
resolver:
  layout: follow-schema
  dir: graph
  package: graph
  filename_template: "{name}.resolvers.go"


# Automatically bind types from Ent
autobind:
  - ${base_module}/${service_name}/internal/ent

# Scalar mappings
models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32

  Time:
    model:
      - github.com/99designs/gqlgen/graphql.Time

  JSON:
    model:
      - github.com/99designs/gqlgen/graphql.Map

  Json:
    model:
      - github.com/99designs/gqlgen/graphql.Map

  Map:
    model:
      - github.com/99designs/gqlgen/graphql.Map

  Node:
    model:
      - ${base_module}/${service_name}/internal/ent.Noder

skip_validation: true
EOF
}

# Generate generate.go file
generate_generate_go() {
    local service_dir=$1

    cat > "$service_dir/generate.go" << 'EOF'
//go:generate go get -u entgo.io/ent/cmd/ent

// Step 1: Initial Ent generation WITHOUT privacy/hooks (creates internal/ent types)
//go:generate go run ./cmd/entc.go

// Step 2: Resolve dependencies from generated code
//go:generate go mod tidy

// Step 3: Generate privacy and hooks policies (NOW ent types exist)
//go:generate go run ./cmd/generate-privacy/main.go
//go:generate go run ./cmd/generate-hooks/main.go

// Step 4: Add Policy() and Hooks() imports to schema files
//go:generate go run ./cmd/generate-imports/main.go

// Step 5: Regenerate Ent WITH privacy/hooks integrated
//go:generate go run ./cmd/entc.go

// Step 6: Generate gRPC services
//go:generate go run ./cmd/generate-grpc/main.go

// Step 7: Generate GraphQL resolvers
//go:generate go run ./cmd/generate-resolvers/main.go

// Step 8: Final GraphQL code generation
//go:generate go run github.com/99designs/gqlgen

//go:build ignore

package main
EOF
}

# Generate bootstrap.go for graph/model
generate_model_bootstrap() {
    local service_dir=$1

    cat > "$service_dir/graph/model/bootstrap.go" << 'EOF'
package model

// This file is a placeholder to bootstrap gqlgen.
// It will be replaced during generation.
EOF
}

