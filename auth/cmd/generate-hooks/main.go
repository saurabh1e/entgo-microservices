package main

import (
	"bufio"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// EntityInfo holds information about an entity for template generation
type EntityInfo struct {
	Name       string // e.g., "Category"
	NameLower  string // e.g., "category"
	NameCamel  string // e.g., "category"
	ModuleName string // e.g., "github.com/zyne-labs/syphoon_auth"
}

// hooksTemplate defines the template for generating hooks files
const hooksTemplate = `package hooks

import (
	"context"
	"{{.ModuleName}}/internal/ent"
	"{{.ModuleName}}/internal/ent/hook"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func {{.Name}}CreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.{{.Name}}Mutation); ok {
				// Hook executing for create

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-create logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func {{.Name}}BulkUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.{{.Name}}Mutation); ok {
				// Hook executing for bulk update

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-update logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func {{.Name}}SingleUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if {{.NameLower}}Mutation, ok := m.(*ent.{{.Name}}Mutation); ok {
				// Hook executing for single update
				_ = {{.NameLower}}Mutation

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-single-update logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func {{.Name}}BulkDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.{{.Name}}Mutation); ok {
				// Hook executing for bulk delete

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-delete logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func {{.Name}}SingleDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if {{.NameLower}}Mutation, ok := m.(*ent.{{.Name}}Mutation); ok {
				// Hook executing for single delete
				_ = {{.NameLower}}Mutation

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-single-delete logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func {{.Name}}Hooks() []ent.Hook {
	return []ent.Hook{
		// Execute {{.Name}}CreateHook only for Create operations
		hook.On({{.Name}}CreateHook(), ent.OpCreate),
		// Execute {{.Name}}BulkUpdateHook only for bulk Update operations
		hook.On({{.Name}}BulkUpdateHook(), ent.OpUpdate),
		// Execute {{.Name}}SingleUpdateHook only for single UpdateOne operations
		hook.On({{.Name}}SingleUpdateHook(), ent.OpUpdateOne),
		// Execute {{.Name}}BulkDeleteHook only for bulk Delete operations
		hook.On({{.Name}}BulkDeleteHook(), ent.OpDelete),
		// Execute {{.Name}}SingleDeleteHook only for single DeleteOne operations
		hook.On({{.Name}}SingleDeleteHook(), ent.OpDeleteOne),
	}
}
`

// extractEntityName extracts the actual entity struct name from the schema file
func extractEntityName(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Regex to match "type EntityName struct"
	typeRegex := regexp.MustCompile(`^type\s+([A-Z][a-zA-Z0-9]*)\s+struct`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("no entity struct found in %s", filePath)
}

// scanForHooksAnnotation scans schema files for @generate-hooks: true annotation
func scanForHooksAnnotation() ([]EntityInfo, error) {
	var entities []EntityInfo

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	schemaDir := filepath.Join(cwd, "ent", "schema")
	moduleName, err := getModuleName()
	if err != nil {
		return nil, fmt.Errorf("failed to get module name: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(schemaDir, "*.go"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Skip base_mixin.go and enums as they're not entities
		basename := filepath.Base(file)
		if strings.HasSuffix(file, "base_mixin.go") || basename == "enums.go" {
			continue
		}

		hasAnnotation, err := checkForHooksAnnotation(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		if hasAnnotation {
			// Extract the actual entity name from the schema file
			entityName, err := extractEntityName(file)
			if err != nil {
				log.Printf("Failed to extract entity name from %s: %v", file, err)
				continue
			}

			// Extract filename without extension for file naming
			fileBaseName := strings.TrimSuffix(basename, ".go")

			entityInfo := EntityInfo{
				Name:       entityName,                    // Actual struct name like "APIKey"
				NameLower:  strings.ToLower(fileBaseName), // Filename like "api_key"
				NameCamel:  toCamelCase(fileBaseName),     // camelCase like "apiKey"
				ModuleName: moduleName,
			}
			entities = append(entities, entityInfo)
		}
	}

	return entities, nil
}

// checkForHooksAnnotation checks if a file contains @generate-hooks: true
func checkForHooksAnnotation(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "// @generate-hooks: true") {
			return true, nil
		}
	}

	return false, scanner.Err()
}

// getModuleName reads the go.mod file and extracts the module name
func getModuleName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}

	goModPath := filepath.Join(cwd, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// toTitleCase converts snake_case to TitleCase
func toTitleCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// toCamelCase converts snake_case to camelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return result
}

// formatGoFile formats the Go file at the given path
func formatGoFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file content
	var content []byte
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line...)
		content = append(content, '\n')
	}

	// Format the content using go/format
	formattedContent, err := format.Source(content)
	if err != nil {
		return err
	}

	// Write the formatted content back to the file
	err = os.WriteFile(filePath, formattedContent, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("Generating hooks files...")

	// Scan for entities with @generate-hooks: true annotation
	entityList, err := scanForHooksAnnotation()
	if err != nil {
		log.Printf("Error scanning for hooks annotations: %v", err)
		os.Exit(1)
	}

	if len(entityList) == 0 {
		log.Println("No entities found with @generate-hooks: true annotation")
		return
	}

	log.Printf("Found %d entities; generating hooks...", len(entityList))

	// Parse the template
	tmpl, err := template.New("hooks").Parse(hooksTemplate)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		os.Exit(1)
	}

	// Create the schema_hooks directory if it doesn't exist
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
		os.Exit(1)
	}
	hooksDir := filepath.Join(cwd, "ent", "schema_hooks")
	err = os.MkdirAll(hooksDir, 0755)
	if err != nil {
		log.Printf("Error creating directory %s: %v", hooksDir, err)
		os.Exit(1)
	}

	// Generate hooks files for each entity
	for _, entityInfo := range entityList {
		// Generate hooks file
		filename := fmt.Sprintf("%s.go", entityInfo.NameLower)
		filePath := filepath.Join(hooksDir, filename)

		// SAFETY CHECK: Skip if file already exists to preserve custom code
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("⚠️  SKIPPING %s - File already exists. To regenerate, delete the file first.", filename)
			continue
		}

		// Create the hooks file only if it doesn't exist
		file, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file %s: %v", filePath, err)
			continue
		}

		// Execute the template
		err = tmpl.Execute(file, entityInfo)
		file.Close()
		if err != nil {
			log.Printf("Error executing template for %s: %v", filename, err)
			continue
		}

		// Format the generated file
		err = formatGoFile(filePath)
		if err != nil {
			log.Printf("Error formatting file %s: %v", filePath, err)
			continue
		}

		log.Printf("Generated hooks file: %s", filename)
	}

	log.Println("Hooks generation completed.")
}
