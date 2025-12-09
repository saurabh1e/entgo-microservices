package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// GetModuleName reads the module name from go.mod file
func GetModuleName() (string, error) {
	file, err := os.Open("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading go.mod: %w", err)
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// GetSchemaFiles returns list of schema files in the given directory
func GetSchemaFiles(schemaDir string) ([]string, error) {
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema directory: %w", err)
	}

	var schemaFiles []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Skip special files
		if strings.Contains(entry.Name(), "enum/") ||
			entry.Name() == "base_mixin.go" {
			continue
		}

		schemaFiles = append(schemaFiles, filepath.Join(schemaDir, entry.Name()))
	}

	return schemaFiles, nil
}

// ExtractEntityName extracts the entity name from a schema file
func ExtractEntityName(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Look for type EntityName struct pattern
	re := regexp.MustCompile(`type\s+(\w+)\s+struct\s*{`)
	matches := re.FindSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("entity struct not found in %s", filePath)
	}

	return string(matches[1]), nil
}

// ParseAnnotations extracts all annotation data from a schema file
func ParseAnnotations(filePath string) (*AnnotationData, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	data := &AnnotationData{}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse boolean flags
		if strings.Contains(line, "@generate-resolver:") {
			data.GenerateResolver = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@generate-mutation:") {
			data.GenerateMutation = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@generate-service:") {
			data.GenerateService = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@generate-grpc:") {
			data.GenerateGRPC = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@generate-hooks:") {
			data.GenerateHooks = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@generate-privacy:") {
			data.GeneratePrivacy = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@tenant-isolated:") {
			data.TenantIsolated = strings.Contains(line, "true")
		}
		if strings.Contains(line, "@user-owned:") {
			data.UserOwned = strings.Contains(line, "true")
		}

		// Parse string values
		if strings.Contains(line, "@role-level:") {
			parts := strings.Split(line, "@role-level:")
			if len(parts) > 1 {
				data.RoleLevel = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "@permission-level:") {
			parts := strings.Split(line, "@permission-level:")
			if len(parts) > 1 {
				data.PermissionLevel = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "@filter-by:") {
			parts := strings.Split(line, "@filter-by:")
			if len(parts) > 1 {
				data.FilterByEdge = strings.TrimSpace(parts[1])
			}
		}
	}

	return data, nil
}

// CheckAnnotation checks if a specific annotation exists in a file
func CheckAnnotation(filePath, annotation string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}

	return strings.Contains(string(content), annotation), nil
}

// ToLowerSnakeCase converts a string to lower_snake_case
func ToLowerSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// ToEntPackageName converts entity name to ent package name (all lowercase, no separators)
// Example: "InlineItem" -> "inlineitem", "BrandName" -> "brandname"
func ToEntPackageName(s string) string {
	return strings.ToLower(s)
}

// ParseRolesList extracts roles from role-level annotation
func ParseRolesList(roleLevel string) []string {
	if roleLevel == "" {
		return []string{"admin"}
	}

	// Split by comma and trim spaces
	roles := strings.Split(roleLevel, ",")
	result := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			result = append(result, role)
		}
	}

	if len(result) == 0 {
		return []string{"admin"}
	}

	return result
}
