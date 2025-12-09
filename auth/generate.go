//go:generate go get -u entgo.io/ent/cmd/ent

// Step 1: Initial Ent generation WITHOUT privacy/hooks (creates internal/ent types)
//go:generate go run ./cmd/entc.go

// Step 2: Generate privacy and hooks policies (NOW ent types exist)
//go:generate go run ./cmd/generate-privacy/main.go
//go:generate go run ./cmd/generate-hooks/main.go

// Step 3: Add Policy() and Hooks() imports to schema files
//go:generate go run ./cmd/generate-imports/main.go

// Step 4: Regenerate Ent WITH privacy/hooks integrated
//go:generate go run ./cmd/entc.go

// Step 5: Generate gRPC services
//go:generate go run ./cmd/generate-grpc/main.go

// Step 6: Generate GraphQL resolvers
//go:generate go run ./cmd/generate-resolvers/main.go

// Step 7: Final GraphQL code generation
//go:generate go run github.com/99designs/gqlgen

//go:build ignore

package main
