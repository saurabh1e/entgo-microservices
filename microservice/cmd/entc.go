package main

import (
	"fmt"
	"log"
	"os"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Failed to generate: %v", err)
	}
	log.Println("✅ Successfully generated Ent + GraphQL schema")
}

func run() error {
	// Create the entgql extension with desired features
	ex, err := entgql.NewExtension(
		entgql.WithWhereInputs(true),
		entgql.WithRelaySpec(true), // Enables Relay-style pagination
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaHook(),
		entgql.WithConfigPath("gqlgen.yml"),         // gqlgen config file
		entgql.WithSchemaPath("graph/ent.graphqls"), // Where to write ent-inferred graphql schema
	)
	if err != nil {
		return fmt.Errorf("creating entgql extension: %w", err)
	}

	// Check if this is first run (internal/ent/internal doesn't exist)
	firstRun := true
	if _, err := os.Stat("internal/ent/internal/schema.go"); err == nil {
		firstRun = false
	}

	// Build feature list
	features := []string{
		"privacy",
		"sql/upsert",
		"entql",
		"sql/lock",
		"sql/modifier",
		"versioned",
		"intercept",
		"namedges",
		"bidiedges",
		"mixin",
		"entgql",
	}

	// Only enable snapshot after first run
	if !firstRun {
		features = append(features, "schema/snapshot")
	} else {
		log.Println("ℹ️  First run detected - snapshot feature disabled")
		// Ensure the internal directory exists
		os.MkdirAll("internal/ent/internal", 0755)
	}

	opts := []entc.Option{
		entc.Extensions(ex),
		entc.FeatureNames(features...),
	}

	// Run Ent code generation with custom config
	if err := entc.Generate("./ent/schema", &gen.Config{
		Target:  "internal/ent",
		Package: "github.com/saurabh/entgo-microservices/microservice/internal/ent",
	}, opts...); err != nil {
		return fmt.Errorf("running ent codegen: %w", err)
	}

	return nil
}
