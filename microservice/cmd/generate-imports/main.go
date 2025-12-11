package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/saurabh/entgo-microservices/microservice/cmd/common"
)

// SchemaInfo holds information about a schema
type SchemaInfo struct {
	Name            string
	LowerName       string
	FilePath        string
	HasPrivacy      bool
	HasHooks        bool
	RoleLevel       string
	PermissionLevel string
	ModuleName      string
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Generation failed: %v", err)
	}
	log.Println("✅ Successfully processed all schemas!")
}

func run() error {
	log.Println("🔗 Auto-generating imports and methods for schemas...")

	// Get module name using common utility
	moduleName, err := common.GetModuleName()
	if err != nil {
		return fmt.Errorf("getting module name: %w", err)
	}
	log.Printf("Detected module name: %s", moduleName)

	// Get all schema files
	schemas, err := getSchemaFiles("./ent/schema", moduleName)
	if err != nil {
		return fmt.Errorf("getting schema files: %w", err)
	}

	log.Printf("Found %d schemas to process", len(schemas))

	for _, schema := range schemas {
		if schema.HasPrivacy {
			log.Printf("Adding Policy() method to %s", schema.Name)
			if err := addPolicyMethod(schema); err != nil {
				log.Printf("Error adding Policy method to %s: %v", schema.Name, err)
			} else {
				log.Printf("✅ Added Policy() method to %s", schema.Name)
			}
		}

		if schema.HasHooks {
			log.Printf("Adding Hooks() method to %s", schema.Name)
			if err := addHooksMethod(schema); err != nil {
				log.Printf("Error adding Hooks method to %s: %v", schema.Name, err)
			} else {
				log.Printf("✅ Added Hooks() method to %s", schema.Name)
			}
		}
	}

	return nil
}

func getSchemaFiles(schemaDir, moduleName string) ([]SchemaInfo, error) {
	var schemas []SchemaInfo

	// Use common utility to list schema files
	schemaFiles, err := common.GetSchemaFiles(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("reading schema directory: %w", err)
	}

	for _, filePath := range schemaFiles {
		// Extract schema name using common utility
		schemaName, err := common.ExtractEntityName(filePath)
		if err != nil {
			log.Printf("Warning: Could not extract schema name from %s: %v", filePath, err)
			continue
		}

		if schemaName == "" {
			continue
		}

		// Parse annotations using common utility
		annotations, err := common.ParseAnnotations(filePath)
		if err != nil {
			log.Printf("Warning: Could not parse annotations from %s: %v", filePath, err)
			continue
		}

		schema := SchemaInfo{
			Name:            schemaName,
			LowerName:       strings.ToLower(schemaName),
			FilePath:        filePath,
			HasPrivacy:      annotations.GeneratePrivacy,
			HasHooks:        annotations.GenerateHooks,
			RoleLevel:       annotations.RoleLevel,
			PermissionLevel: annotations.PermissionLevel,
			ModuleName:      moduleName,
		}

		// Log what we found
		privacyStatus := "❌"
		if schema.HasPrivacy {
			privacyStatus = "✅"
		}
		hooksStatus := "❌"
		if schema.HasHooks {
			hooksStatus = "✅"
		}

		log.Printf("🔐 %s: Privacy %s, Hooks %s", schema.Name, privacyStatus, hooksStatus)
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func addPolicyMethod(schema SchemaInfo) error {
	// Check if Policy method already exists
	hasPolicy, err := common.CheckFileHasContent(schema.FilePath, "func ("+schema.Name+") Policy()")
	if err != nil {
		return fmt.Errorf("checking for existing Policy method: %w", err)
	}
	if hasPolicy {
		log.Printf("Policy method already exists in %s, skipping", schema.Name)
		return nil
	}

	content, err := os.ReadFile(schema.FilePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	fileContent := string(content)

	// Check if privacy import already exists
	privacyImportPattern := `privacy\s+"` + regexp.QuoteMeta(schema.ModuleName) + `/ent/schema_privacy"`
	privacyImportRegex := regexp.MustCompile(privacyImportPattern)

	if !privacyImportRegex.MatchString(fileContent) {
		// Add privacy import
		importPattern := regexp.MustCompile(`(import\s*\(\s*\n)`)
		if importPattern.MatchString(fileContent) {
			fileContent = importPattern.ReplaceAllString(fileContent,
				`${1}	privacy "`+schema.ModuleName+`/ent/schema_privacy"`+"\n")
		}
	}

	// Add Policy method at the end of the file
	policyMethod := fmt.Sprintf(`
func (%s) Policy() ent.Policy {
	return privacy.%sPolicy()
}
`, schema.Name, schema.Name)

	fileContent += policyMethod

	return os.WriteFile(schema.FilePath, []byte(fileContent), 0644)
}

func addHooksMethod(schema SchemaInfo) error {
	// Check if Hooks method already exists
	hasHooks, err := common.CheckFileHasContent(schema.FilePath, "func ("+schema.Name+") Hooks()")
	if err != nil {
		return fmt.Errorf("checking for existing Hooks method: %w", err)
	}
	if hasHooks {
		log.Printf("Hooks method already exists in %s, skipping", schema.Name)
		return nil
	}

	content, err := os.ReadFile(schema.FilePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	fileContent := string(content)

	// Check if hooks import already exists
	hooksImportPattern := `hook\s+"` + regexp.QuoteMeta(schema.ModuleName) + `/ent/schema_hooks"`
	hooksImportRegex := regexp.MustCompile(hooksImportPattern)

	if !hooksImportRegex.MatchString(fileContent) {
		// Add hooks import
		importPattern := regexp.MustCompile(`(import\s*\(\s*\n)`)
		if importPattern.MatchString(fileContent) {
			fileContent = importPattern.ReplaceAllString(fileContent,
				`${1}	hook "`+schema.ModuleName+`/ent/schema_hooks"`+"\n")
		}
	}

	// Add Hooks method at the end of the file
	hooksMethod := fmt.Sprintf(`
func (%s) Hooks() []ent.Hook {
	return hook.%sHooks()
}
`, schema.Name, schema.Name)

	fileContent += hooksMethod

	return os.WriteFile(schema.FilePath, []byte(fileContent), 0644)
}
