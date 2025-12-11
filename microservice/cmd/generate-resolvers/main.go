package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// EntityInfo holds information about an entity for template generation
type EntityInfo struct {
	Name             string // e.g., "APIKey"
	NameLower        string // e.g., "api_key"
	NameCamel        string // e.g., "apiKey"
	NamePlural       string // e.g., "Categories", "Commissions"
	ModuleName       string // e.g., "github.com/zyne-labs/syphoon_main"
	GenerateResolver bool   // true if @generate-resolver: true
	GenerateMutation bool   // true if @generate-mutation: true
}

// queryOnlyGraphQLTemplate defines the template for generating only queries in GraphQL schema
const queryOnlyGraphQLTemplate = `extend type Query {
	{{.Name}}ByID(id: Int!): {{.Name}} @auth
	{{.NamePlural}}(
		first: Int
		after: Cursor
		last: Int
		before: Cursor
		orderBy: {{.Name}}Order
		where: {{.Name}}WhereInput
	): {{.Name}}Connection! @auth
}
`

// mutationOnlyGraphQLTemplate defines the template for generating only mutations in GraphQL schema
const mutationOnlyGraphQLTemplate = `extend type Mutation {
	create{{.Name}}(input: Create{{.Name}}Input!): {{.Name}}! @auth
	createBulk{{.Name}}(input: [Create{{.Name}}Input!]!): [{{.Name}}!]! @auth
	update{{.Name}}(id: Int!, input: Update{{.Name}}Input!): {{.Name}}! @auth
	delete{{.Name}}(id: Int!): Boolean! @auth
}
`

// combinedGraphQLTemplate defines the template for generating both queries and mutations in one GraphQL schema file
const combinedGraphQLTemplate = `extend type Query {
	{{.Name}}ByID(id: Int!): {{.Name}} @auth
	{{.NamePlural}}(
		first: Int
		after: Cursor
		last: Int
		before: Cursor
		orderBy: {{.Name}}Order
		where: {{.Name}}WhereInput
	): {{.Name}}Connection! @auth
}

extend type Mutation {
	create{{.Name}}(input: Create{{.Name}}Input!): {{.Name}}! @auth
	createBulk{{.Name}}(input: [Create{{.Name}}Input!]!): [{{.Name}}!]! @auth
	update{{.Name}}(id: Int!, input: Update{{.Name}}Input!): {{.Name}}! @auth
	delete{{.Name}}(id: Int!): Boolean! @auth
}
`

// queryOnlyTemplate defines the template for generating only query resolvers using direct client
const queryOnlyTemplate = `package graph

import (
	"context"
	"{{.ModuleName}}/internal/ent"

	"entgo.io/contrib/entgql"
)

// {{.Name}}ByID is the resolver for the {{.Name}}ByID field.
func (r *queryResolver) {{.Name}}ByID(ctx context.Context, id int) (*ent.{{.Name}}, error) {
	return r.Resolver.client.{{.Name}}.Get(ctx, id)
}

// {{.NamePlural}} is the resolver for the {{.NamePlural}} field.
func (r *queryResolver) {{.NamePlural}}(ctx context.Context, first *int, after *entgql.Cursor[int], last *int, before *entgql.Cursor[int], orderBy *ent.{{.Name}}Order, where *ent.{{.Name}}WhereInput) (*ent.{{.Name}}Connection, error) {
	return r.Resolver.client.{{.Name}}.Query().Paginate(ctx, after, first, before, last, ent.With{{.Name}}Filter(where.Filter), ent.With{{.Name}}Order(orderBy))
}
`

// mutationOnlyTemplate defines the template for generating only mutation resolvers using direct client
const mutationOnlyTemplate = `package graph

import (
	"context"
	"{{.ModuleName}}/internal/ent"
)

// Create{{.Name}} is the resolver for the create{{.Name}} mutation.
func (r *mutationResolver) Create{{.Name}}(ctx context.Context, input ent.Create{{.Name}}Input) (*ent.{{.Name}}, error) {
	return r.Resolver.client.{{.Name}}.Create().SetInput(input).Save(ctx)
}

// CreateBulk{{.Name}} is the resolver for the createBulk{{.Name}} mutation.
func (r *mutationResolver) CreateBulk{{.Name}}(ctx context.Context, input []ent.Create{{.Name}}Input) ([]*ent.{{.Name}}, error) {
	builders := make([]*ent.{{.Name}}Create, len(input))
	for i, inp := range input {
		builders[i] = r.Resolver.client.{{.Name}}.Create().SetInput(inp)
	}
	return r.Resolver.client.{{.Name}}.CreateBulk(builders...).Save(ctx)
}

// Update{{.Name}} is the resolver for the update{{.Name}} mutation.
func (r *mutationResolver) Update{{.Name}}(ctx context.Context, id int, input ent.Update{{.Name}}Input) (*ent.{{.Name}}, error) {
	return r.Resolver.client.{{.Name}}.UpdateOneID(id).SetInput(input).Save(ctx)
}

// Delete{{.Name}} is the resolver for the delete{{.Name}} mutation.
func (r *mutationResolver) Delete{{.Name}}(ctx context.Context, id int) (bool, error) {
	err := r.Resolver.client.{{.Name}}.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
`

// combinedTemplate defines the template for generating both queries and mutations resolvers using direct client
const combinedTemplate = `package graph

import (
	"context"
	"{{.ModuleName}}/internal/ent"

	"entgo.io/contrib/entgql"
)

// {{.Name}}ByID is the resolver for the {{.Name}}ByID field.
func (r *queryResolver) {{.Name}}ByID(ctx context.Context, id int) (*ent.{{.Name}}, error) {
	return r.Resolver.client.{{.Name}}.Get(ctx, id)
}

// {{.NamePlural}} is the resolver for the {{.NamePlural}} field.
func (r *queryResolver) {{.NamePlural}}(ctx context.Context, first *int, after *entgql.Cursor[int], last *int, before *entgql.Cursor[int], orderBy *ent.{{.Name}}Order, where *ent.{{.Name}}WhereInput) (*ent.{{.Name}}Connection, error) {
	return r.Resolver.client.{{.Name}}.Query().Paginate(ctx, after, first, before, last, ent.With{{.Name}}Filter(where.Filter), ent.With{{.Name}}Order(orderBy))
}

// Create{{.Name}} is the resolver for the create{{.Name}} mutation.
func (r *mutationResolver) Create{{.Name}}(ctx context.Context, input ent.Create{{.Name}}Input) (*ent.{{.Name}}, error) {
	return r.Resolver.client.{{.Name}}.Create().SetInput(input).Save(ctx)
}

// CreateBulk{{.Name}} is the resolver for the createBulk{{.Name}} mutation.
func (r *mutationResolver) CreateBulk{{.Name}}(ctx context.Context, input []*ent.Create{{.Name}}Input) ([]*ent.{{.Name}}, error) {
	builders := make([]*ent.{{.Name}}Create, len(input))
	for i, inp := range input {
		builders[i] = r.Resolver.client.{{.Name}}.Create().SetInput(*inp)
	}
	return r.Resolver.client.{{.Name}}.CreateBulk(builders...).Save(ctx)
}

// Update{{.Name}} is the resolver for the update{{.Name}} mutation.
func (r *mutationResolver) Update{{.Name}}(ctx context.Context, id int, input ent.Update{{.Name}}Input) (*ent.{{.Name}}, error) {
	return r.Resolver.client.{{.Name}}.UpdateOneID(id).SetInput(input).Save(ctx)
}

// Delete{{.Name}} is the resolver for the delete{{.Name}} mutation.
func (r *mutationResolver) Delete{{.Name}}(ctx context.Context, id int) (bool, error) {
	err := r.Resolver.client.{{.Name}}.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}
`

func main() {
	// Get the module name from go.mod
	moduleName, err := getModuleName()
	if err != nil {
		log.Fatalf("Error reading module name: %v", err)
	}

	// Get entity names from schema files that have @generate-resolver: true or @generate-mutation: true
	entities, err := getEntitiesWithGeneration()
	if err != nil {
		log.Fatalf("Error reading entities: %v", err)
	}

	for _, entity := range entities {
		entity.ModuleName = moduleName
		log.Printf("Processing entity: %s (Resolver: %t, Mutation: %t)\n",
			entity.Name, entity.GenerateResolver, entity.GenerateMutation)

		// Generate resolver files based on flags
		err = generateResolvers(entity)
		if err != nil {
			log.Printf("Error generating resolvers for %s: %v", entity.Name, err)
			continue
		}

		// Generate GraphQL schema files based on flags
		err = generateGraphQLSchemas(entity)
		if err != nil {
			log.Printf("Error generating GraphQL schemas for %s: %v", entity.Name, err)
			continue
		}
	}

	log.Println("Resolver generation completed!")
}

func getModuleName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("module name not found in go.mod")
}

func getEntitiesWithGeneration() ([]EntityInfo, error) {
	var entities []EntityInfo

	err := filepath.Walk("ent/schema", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "enum/") || strings.Contains(path, "base_mixin.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if file has @generate-resolver: true
		generateResolver := strings.Contains(string(content), "@generate-resolver: true")

		// Check if file has @generate-mutation: true
		generateMutation := strings.Contains(string(content), "@generate-mutation: true")

		// Skip if neither annotation is present
		if !generateResolver && !generateMutation {
			return nil
		}

		// Parse the Go file to extract the actual struct name
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, content, parser.ParseComments)
		if err != nil {
			return err
		}

		// Find the struct that embeds ent.Schema
		var entityName string
		ast.Inspect(node, func(n ast.Node) bool {
			if typeSpec, ok := n.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					// Check if this struct embeds ent.Schema
					for _, field := range structType.Fields.List {
						if len(field.Names) == 0 { // Anonymous field (embedding)
							if ident, ok := field.Type.(*ast.SelectorExpr); ok {
								if x, ok := ident.X.(*ast.Ident); ok && x.Name == "ent" && ident.Sel.Name == "Schema" {
									entityName = typeSpec.Name.Name
									return false // Stop searching
								}
							}
						}
					}
				}
			}
			return true
		})

		if entityName == "" {
			// Fallback to filename-based approach if struct parsing fails
			fileName := strings.TrimSuffix(filepath.Base(path), ".go")
			entityName = toPascalCase(fileName)
		}

		entity := EntityInfo{
			Name:             entityName,
			NameLower:        strings.ToLower(entityName),
			NameCamel:        toCamelCase(entityName),
			NamePlural:       pluralize(entityName),
			GenerateResolver: generateResolver,
			GenerateMutation: generateMutation,
		}

		entities = append(entities, entity)
		return nil
	})

	return entities, err
}

func generateGraphQLSchemas(entity EntityInfo) error {
	// Generate a single schema file based on what flags are set
	fileName := fmt.Sprintf("graph/schemas/%s.graphqls", entity.NameLower)

	// Determine which template and expected fields to use based on flags
	var templateToUse string
	var expectedFields []string

	if entity.GenerateResolver && entity.GenerateMutation {
		// Both flags set - generate everything
		templateToUse = combinedGraphQLTemplate
		expectedFields = []string{
			entity.Name + "ByID",
			entity.NamePlural,
			"create" + entity.Name,
			"createBulk" + entity.Name,
			"update" + entity.Name,
			"delete" + entity.Name,
		}
		log.Printf("  GraphQL: Generating both queries and mutations\n")
	} else if entity.GenerateResolver {
		// Only queries
		templateToUse = queryOnlyGraphQLTemplate
		expectedFields = []string{
			entity.Name + "ByID",
			entity.NamePlural,
		}
		log.Printf("  GraphQL: Generating queries only\n")
	} else if entity.GenerateMutation {
		// Only mutations
		templateToUse = mutationOnlyGraphQLTemplate
		expectedFields = []string{
			"create" + entity.Name,
			"createBulk" + entity.Name,
			"update" + entity.Name,
			"delete" + entity.Name,
		}
		log.Printf("  GraphQL: Generating mutations only\n")
	} else {
		// Neither flag set - skip
		log.Printf("  GraphQL: No generation flags set, skipping\n")
		return nil
	}

	// Check if all expected fields exist
	if !graphqlFieldsExist(fileName, expectedFields) {
		log.Printf("  GraphQL: Creating/updating schema at %s\n", fileName)
		err := generateOrAppendGraphQLFile(fileName, templateToUse, entity)
		if err != nil {
			return err
		}
	} else {
		log.Printf("  GraphQL: Schema already exists with all fields at %s\n", fileName)
	}

	return nil
}

func graphqlFieldsExist(fileName string, expectedFields []string) bool {
	content, err := os.ReadFile(fileName)
	if err != nil {
		// File doesn't exist, so fields don't exist
		return false
	}

	contentStr := string(content)
	for _, field := range expectedFields {
		if !strings.Contains(contentStr, field+"(") {
			return false
		}
	}
	return true
}

func generateOrAppendGraphQLFile(fileName, templateStr string, entity EntityInfo) error {
	// If file doesn't exist, create it
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return generateFileFromTemplate(fileName, templateStr, entity)
	}

	// File exists, check what fields are missing
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Parse template to get the fields we want to add
	tmpl, err := template.New("graphql").Parse(templateStr)
	if err != nil {
		return err
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, entity)
	if err != nil {
		return err
	}

	newContent := buf.String()

	// Extract fields from the new content that don't already exist
	var fieldsToAdd []string
	lines := strings.Split(newContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "(") && strings.Contains(line, ":") &&
			!strings.HasPrefix(line, "extend type") && !strings.HasPrefix(line, "}") && line != "" {
			// This looks like a field definition
			fieldName := strings.Split(line, "(")[0]
			fieldName = strings.TrimSpace(fieldName)

			if !strings.Contains(contentStr, fieldName+"(") {
				fieldsToAdd = append(fieldsToAdd, "\t"+line)
			}
		}
	}

	// If there are fields to add, append them
	if len(fieldsToAdd) > 0 {
		// Find the closing brace and insert before it
		lines := strings.Split(contentStr, "\n")
		var newLines []string

		for i, line := range lines {
			if strings.TrimSpace(line) == "}" && i == len(lines)-1 {
				// This is the last closing brace, add fields before it
				newLines = append(newLines, fieldsToAdd...)
				newLines = append(newLines, line)
			} else {
				newLines = append(newLines, line)
			}
		}

		// Write the updated content
		return os.WriteFile(fileName, []byte(strings.Join(newLines, "\n")), 0644)
	}

	return nil
}

func getExistingFunctions(fileName string) ([]string, error) {
	var functions []string

	content, err := os.ReadFile(fileName)
	if err != nil {
		return functions, err
	}

	// Try to parse with Go AST first
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, content, parser.ParseComments)
	if err == nil {
		// Successfully parsed with AST
		ast.Inspect(node, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil {
				functions = append(functions, fn.Name.Name)
			}
			return true
		})
		return functions, nil
	}

	// Fallback to regex if AST parsing fails
	funcRegex := regexp.MustCompile(`func\s+\([^)]+\)\s+(\w+)\s*\(`)
	matches := funcRegex.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	return functions, nil
}

func generateFileFromTemplate(fileName, templateStr string, entity EntityInfo) error {
	tmpl, err := template.New("file").Parse(templateStr)
	if err != nil {
		return err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, entity)
}

func appendMissingFunctions(fileName, templateStr string, entity EntityInfo, missingFuncs []string) error {
	// Parse template to extract individual functions
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return err
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, entity)
	if err != nil {
		return err
	}

	content := buf.String()

	// Extract only the missing functions from the template
	var functionsToAdd []string
	lines := strings.Split(content, "\n")

	var currentFunc strings.Builder
	var currentFuncName string
	inFunction := false
	braceCount := 0

	for _, line := range lines {
		// Check if this is a function declaration
		funcRegex := regexp.MustCompile(`func\s+\([^)]+\)\s+(\w+)\s*\(`)
		if matches := funcRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Save previous function if it was one we needed
			if inFunction && contains(missingFuncs, currentFuncName) {
				functionsToAdd = append(functionsToAdd, currentFunc.String())
			}

			// Start new function
			currentFuncName = matches[1]
			currentFunc.Reset()
			inFunction = true
			braceCount = 0
		}

		if inFunction {
			currentFunc.WriteString(line + "\n")

			// Count braces to determine when function ends
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			// Function ends when braces are balanced and we've seen at least one opening brace
			if braceCount == 0 && strings.Contains(currentFunc.String(), "{") {
				if contains(missingFuncs, currentFuncName) {
					functionsToAdd = append(functionsToAdd, currentFunc.String())
				}
				inFunction = false
				currentFunc.Reset()
			}
		}
	}

	// Handle last function if template doesn't end cleanly
	if inFunction && contains(missingFuncs, currentFuncName) {
		functionsToAdd = append(functionsToAdd, currentFunc.String())
	}

	// Append missing functions to file
	if len(functionsToAdd) > 0 {
		file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		for _, function := range functionsToAdd {
			_, err = file.WriteString("\n" + function)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateResolvers(entity EntityInfo) error {
	fileName := fmt.Sprintf("graph/%s.resolvers.go", entity.NameLower)

	// Check if file exists and get existing functions
	existingFuncs := []string{}
	if _, err := os.Stat(fileName); err == nil {
		existingFuncs, err = getExistingFunctions(fileName)
		if err != nil {
			log.Printf("Warning: Could not parse existing functions in %s: %v", fileName, err)
		}
	}

	// Determine which functions to generate based on flags
	var expectedFuncs []string
	var templateToUse string

	if entity.GenerateResolver && entity.GenerateMutation {
		// Both flags set - generate everything
		expectedFuncs = []string{
			entity.Name + "ByID",
			entity.NamePlural,
			"Create" + entity.Name,
			"CreateBulk" + entity.Name,
			"Update" + entity.Name,
			"Delete" + entity.Name,
		}
		templateToUse = combinedTemplate
		log.Printf("  Resolvers: Generating both queries and mutations\n")
	} else if entity.GenerateResolver {
		// Only queries
		expectedFuncs = []string{
			entity.Name + "ByID",
			entity.NamePlural,
		}
		templateToUse = queryOnlyTemplate
		log.Printf("  Resolvers: Generating queries only\n")
	} else if entity.GenerateMutation {
		// Only mutations
		expectedFuncs = []string{
			"Create" + entity.Name,
			"CreateBulk" + entity.Name,
			"Update" + entity.Name,
			"Delete" + entity.Name,
		}
		templateToUse = mutationOnlyTemplate
		log.Printf("  Resolvers: Generating mutations only\n")
	} else {
		// Neither flag set - skip
		log.Printf("  Resolvers: No generation flags set, skipping\n")
		return nil
	}

	missingFuncs := []string{}
	for _, funcName := range expectedFuncs {
		if !contains(existingFuncs, funcName) {
			missingFuncs = append(missingFuncs, funcName)
		}
	}

	if len(missingFuncs) == 0 {
		log.Printf("  Resolvers: All functions exist, skipping\n")
		return nil
	}

	// If file doesn't exist, create it with appropriate template
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		log.Printf("  Resolvers: Creating new file with %d functions\n", len(expectedFuncs))
		return generateFileFromTemplate(fileName, templateToUse, entity)
	}

	// If file exists, append missing functions
	log.Printf("  Resolvers: Appending %d missing functions\n", len(missingFuncs))
	return appendMissingFunctions(fileName, templateToUse, entity, missingFuncs)
}

// Utility functions
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(pascal[:1]) + pascal[1:]
	}
	return pascal
}

func pluralize(s string) string {
	// Special cases for common programming terms
	specialCases := map[string]string{
		"APIKey":             "APIKeys",
		"ApiKey":             "ApiKeys",
		"API":                "APIs",
		"Api":                "Apis",
		"Key":                "Keys",
		"ModuleGroupPricing": "ModuleGroupPricing", // Pricing doesn't get pluralized
		"Pricing":            "Pricing",            // Pricing is uncountable
	}

	if plural, exists := specialCases[s]; exists {
		return plural
	}

	// Check for compound words ending with "Pricing"
	if strings.HasSuffix(s, "Pricing") {
		return s // Don't pluralize words ending with "Pricing"
	}

	// Standard pluralization rules
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		// Only apply y->ies rule if the character before 'y' is a consonant
		beforeY := strings.ToLower(string(s[len(s)-2]))
		if beforeY != "a" && beforeY != "e" && beforeY != "i" && beforeY != "o" && beforeY != "u" {
			return s[:len(s)-1] + "ies"
		}
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
