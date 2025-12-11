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

// Step 8: Final GraphQL code generation (use installed binary instead of go run)
//go:generate gqlgen

//go:build ignore

package main
