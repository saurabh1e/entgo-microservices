package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	serviceName = flag.String("service", "", "Name of the microservice to generate gRPC for")
	projectRoot = ""
)

// ProtoField represents a field in a protobuf message
type ProtoField struct {
	Name            string // proto field name (snake_case)
	GoName          string // Go field name (PascalCase) from Ent
	Type            string // proto type
	EntType         string // original Ent field type
	Number          int    // field number
	IsOptional      bool   // whether ent field is a pointer
	IsProtoOptional bool   // whether proto field should be optional
	Comment         string // field comment
}

// GRPCEntityInfo holds information for gRPC generation
type GRPCEntityInfo struct {
	Name            string       // e.g., "User"
	NameLower       string       // e.g., "user"
	NameCamel       string       // e.g., "user"
	ProtoFieldName  string       // e.g., "User" or "Rolepermission"
	Package         string       // e.g., "user.v1"
	GoPackage       string       // e.g., "github.com/saurabh/entgo-microservices/pkg/proto/user/v1;userv1"
	ModuleName      string       // e.g., "auth"
	ModuleShortName string       // e.g., "attendance"
	Fields          []ProtoField // proto fields
	HasTimestamps   bool         // whether to import google/protobuf/timestamp.proto
}

// ServiceMetadata holds metadata about generated services
type ServiceMetadata struct {
	ServiceName string   `json:"service"`
	Models      []string `json:"models"`
}

func main() {
	flag.Parse()

	if *serviceName == "" {
		log.Fatal("Service name is required. Use -service=<name>")
	}

	// Get project root (assuming we're in scripts/ directory)
	scriptDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Check if we're in a service directory or scripts directory
	if strings.HasSuffix(scriptDir, *serviceName) {
		projectRoot = filepath.Dir(scriptDir)
	} else if strings.HasSuffix(scriptDir, "scripts") {
		projectRoot = filepath.Dir(scriptDir)
	} else {
		projectRoot = scriptDir
	}

	serviceDir := filepath.Join(projectRoot, *serviceName)

	log.Printf("Starting gRPC generation for service: %s", *serviceName)
	log.Printf("Service directory: %s", serviceDir)

	// Change to service directory
	if err := os.Chdir(serviceDir); err != nil {
		log.Fatalf("Failed to change to service directory: %v", err)
	}

	// Generate gRPC code
	if err := generateGRPCForService(*serviceName, serviceDir); err != nil {
		log.Fatalf("Failed to generate gRPC: %v", err)
	}

	log.Println("gRPC generation completed!")
}

func generateGRPCForService(serviceName, serviceDir string) error {
	schemaDir := filepath.Join(serviceDir, "ent", "schema")

	// Find all schema files
	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("failed to read schema directory: %w", err)
	}

	var entityInfos []*GRPCEntityInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Skip non-entity files
		if strings.Contains(entry.Name(), "mixin") || strings.Contains(entry.Name(), "enum") {
			continue
		}

		entityName := strings.TrimSuffix(entry.Name(), ".go")
		entityName = toPascalCase(entityName)

		schemaFile := filepath.Join(schemaDir, entry.Name())

		// Check if this entity should have gRPC generation
		shouldGenerate, err := shouldGenerateGRPC(schemaFile)
		if err != nil {
			log.Printf("Warning: Failed to check if %s should generate gRPC: %v", entityName, err)
			continue
		}

		if !shouldGenerate {
			continue
		}

		log.Printf("Processing entity: %s", entityName)

		// Read the actual Ent entity name from the generated code
		// Ent may use different casing for abbreviations (e.g., KML, POI instead of Kml, Poi)
		entEntityName := entityName
		entFilePath := filepath.Join(serviceDir, "internal", "ent", strings.ToLower(entityName)+".go")
		if entContent, err := os.ReadFile(entFilePath); err == nil {
			// Extract the actual type name from the generated ent file
			typeRegex := regexp.MustCompile(`type\s+(\w+)\s+struct`)
			if matches := typeRegex.FindStringSubmatch(string(entContent)); len(matches) > 1 {
				entEntityName = matches[1]
				log.Printf("  Using Ent entity name: %s (from generated code)", entEntityName)
			}
		}

		// Extract fields from schema
		fields, err := extractFieldsFromSchema(schemaFile, serviceDir, entEntityName)
		if err != nil {
			log.Printf("Error extracting fields for %s: %v", entityName, err)
			continue
		}

		entityInfo := &GRPCEntityInfo{
			Name:            entEntityName,
			NameLower:       strings.ToLower(entityName),
			NameCamel:       toLowerCamelCase(entityName),
			ProtoFieldName:  entityName,
			Package:         strings.ToLower(entityName) + ".v1",
			GoPackage:       fmt.Sprintf("github.com/saurabh/entgo-microservices/pkg/proto/%s/v1;%sv1", strings.ToLower(entityName), strings.ToLower(entityName)),
			ModuleName:      "github.com/saurabh/entgo-microservices/" + serviceName,
			ModuleShortName: serviceName,
			Fields:          fields,
			HasTimestamps:   true,
		}

		// Generate proto file
		if err := generateProtoFile(entityInfo); err != nil {
			return fmt.Errorf("failed to generate proto file for %s: %w", entityName, err)
		}

		// Compile proto file
		if err := compileProtoFile(entityInfo); err != nil {
			return fmt.Errorf("failed to compile proto file for %s: %w", entityName, err)
		}

		// Generate service file
		if err := generateServiceFile(entityInfo, serviceDir); err != nil {
			return fmt.Errorf("failed to generate service file for %s: %w", entityName, err)
		}

		entityInfos = append(entityInfos, entityInfo)
		log.Printf("Successfully generated gRPC code for %s", entityName)
	}

	if len(entityInfos) > 0 {
		// Generate server registry
		if err := generateServerRegistry(entityInfos, serviceDir); err != nil {
			return fmt.Errorf("failed to generate server registry: %w", err)
		}

		// Generate service metadata
		if err := generateServiceMetadata(entityInfos, serviceName); err != nil {
			return fmt.Errorf("failed to generate service metadata: %w", err)
		}

		// Generate consolidated service client
		if err := generateConsolidatedServiceClient(entityInfos, serviceName); err != nil {
			return fmt.Errorf("failed to generate consolidated service client: %w", err)
		}
	}

	return nil
}

func shouldGenerateGRPC(schemaFile string) (bool, error) {
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return false, err
	}

	// Look for @generate-grpc: true annotation in the Fields() method comment
	// Match the comment block before Fields() method
	fieldsMethodRegex := regexp.MustCompile(`(?s)//.*?@generate-grpc:\s*true.*?func\s+\(\w+\)\s+Fields\(\)`)
	return fieldsMethodRegex.MatchString(string(content)), nil
}

func extractFieldsFromSchema(filePath, serviceDir, entityName string) ([]ProtoField, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var fields []ProtoField
	fieldNum := 1

	// Always add base fields
	fields = append(fields,
		ProtoField{Name: "id", Type: "string", Number: fieldNum, IsOptional: false, Comment: "UUID"},
	)
	fieldNum++

	// Read the generated Ent entity file to check which fields are pointers
	entFilePath := filepath.Join(serviceDir, "internal", "ent", strings.ToLower(entityName)+".go")
	entContent, err := os.ReadFile(entFilePath)
	var entFieldTypes map[string]bool
	if err == nil {
		entFieldTypes = extractEntFieldTypes(string(entContent))
	} else {
		entFieldTypes = make(map[string]bool)
	}

	// Extract field definitions - now check for .Optional() and .Nillable()
	schemaContent := string(content)

	// Simple approach: split by lines and build field definitions
	lines := strings.Split(schemaContent, "\n")
	var currentField strings.Builder
	var inField bool
	var fieldType, fieldName string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line starts a new field
		if strings.HasPrefix(trimmed, "field.") {
			// If we were building a field, process it
			if inField && fieldType != "" && fieldName != "" {
				fullFieldDef := currentField.String()

				// Skip special fields
				if fieldName != "tenant_id" && !strings.HasSuffix(fieldName, "_id") {
					// Skip UUID and JSON fields
					if fieldType != "UUID" && fieldType != "JSON" {
						protoType := mapEntTypeToProtoType(fieldType)
						if protoType != "" {
							goFieldName := toPascalCase(fieldName)
							isSchemaOptional := strings.Contains(fullFieldDef, ".Optional()") || strings.Contains(fullFieldDef, ".Nillable()")
							isEntPointer := entFieldTypes[goFieldName]

							fields = append(fields, ProtoField{
								Name:            toSnakeCase(fieldName),
								GoName:          goFieldName,
								Type:            protoType,
								EntType:         fieldType,
								Number:          fieldNum,
								IsOptional:      isEntPointer,
								IsProtoOptional: isSchemaOptional,
							})
							fieldNum++
						}
					}
				}
			}

			// Start new field
			currentField.Reset()
			currentField.WriteString(trimmed)
			inField = true

			// Extract field type and name from this line
			fieldMatch := regexp.MustCompile(`field\.(String|Text|Int|Int64|Bool|Float|Float32|Float64|Time|Enum|UUID|JSON)\("([^"]+)"`).FindStringSubmatch(trimmed)
			if len(fieldMatch) >= 3 {
				fieldType = fieldMatch[1]
				fieldName = fieldMatch[2]
			}
		} else if inField {
			// Continue building current field
			currentField.WriteString(" ")
			currentField.WriteString(trimmed)

			// Check if field definition ended (ends with comma)
			if strings.HasSuffix(trimmed, ",") {
				inField = false
			}
		}
	}

	// Process last field if any
	if inField && fieldType != "" && fieldName != "" {
		fullFieldDef := currentField.String()
		if fieldName != "tenant_id" && !strings.HasSuffix(fieldName, "_id") {
			if fieldType != "UUID" && fieldType != "JSON" {
				protoType := mapEntTypeToProtoType(fieldType)
				if protoType != "" {
					goFieldName := toPascalCase(fieldName)
					isSchemaOptional := strings.Contains(fullFieldDef, ".Optional()") || strings.Contains(fullFieldDef, ".Nillable()")
					isEntPointer := entFieldTypes[goFieldName]

					fields = append(fields, ProtoField{
						Name:            toSnakeCase(fieldName),
						GoName:          goFieldName,
						Type:            protoType,
						EntType:         fieldType,
						Number:          fieldNum,
						IsOptional:      isEntPointer,
						IsProtoOptional: isSchemaOptional,
					})
					fieldNum++
				}
			}
		}
	}

	// Add timestamps
	fields = append(fields,
		ProtoField{Name: "created_at", Type: "google.protobuf.Timestamp", Number: fieldNum, IsOptional: false},
	)
	fieldNum++
	fields = append(fields,
		ProtoField{Name: "updated_at", Type: "google.protobuf.Timestamp", Number: fieldNum, IsOptional: false},
	)

	return fields, nil
}

func extractEntFieldTypes(entContent string) map[string]bool {
	fieldTypes := make(map[string]bool)
	fieldRegex := regexp.MustCompile(`(?m)^\s+(\w+)\s+(\*?)(\w+(?:\.\w+)?)\s+` + "`" + `json:"([^"]+)"`)
	matches := fieldRegex.FindAllStringSubmatch(entContent, -1)

	for _, match := range matches {
		if len(match) < 5 {
			continue
		}
		fieldName := match[1]
		isPointer := match[2] == "*"
		if fieldName == "selectValues" || fieldName == "ID" || strings.HasSuffix(fieldName, "Edges") {
			continue
		}
		fieldTypes[fieldName] = isPointer
	}

	return fieldTypes
}

func mapEntTypeToProtoType(entType string) string {
	switch entType {
	case "String", "Text", "Enum":
		return "string"
	case "Int":
		return "int32"
	case "Int64":
		return "int64"
	case "Bool":
		return "bool"
	case "Float", "Float32":
		return "float"
	case "Float64":
		return "double"
	case "Time":
		return "google.protobuf.Timestamp"
	default:
		return ""
	}
}

// Continue in next file...

func generateProtoFile(info *GRPCEntityInfo) error {
	protoDir := filepath.Join(projectRoot, "proto", info.NameLower)
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		return fmt.Errorf("failed to create proto directory: %w", err)
	}

	protoFile := filepath.Join(protoDir, info.NameLower+".proto")
	f, err := os.Create(protoFile)
	if err != nil {
		return fmt.Errorf("failed to create proto file: %w", err)
	}
	defer f.Close()

	tmpl := `syntax = "proto3";

package {{.Package}};
option go_package = "{{.GoPackage}}";
{{if .HasTimestamps}}
import "google/protobuf/timestamp.proto";
{{end}}
service {{.Name}}Service {
  rpc Get{{.Name}}ByID(Get{{.Name}}ByIDRequest) returns (Get{{.Name}}ByIDResponse);
  rpc Get{{.Name}}sByIDs(Get{{.Name}}sByIDsRequest) returns (Get{{.Name}}sByIDsResponse);
}

message {{.Name}} {
{{- range .Fields}}
  {{if .IsProtoOptional}}optional {{end}}{{.Type}} {{.Name}} = {{.Number}};{{if .Comment}} // {{.Comment}}{{end}}
{{- end}}
}

message Get{{.Name}}ByIDRequest {
  string id = 1; // UUID
}

message Get{{.Name}}ByIDResponse {
  {{.Name}} {{.NameLower}} = 1;
}

message Get{{.Name}}sByIDsRequest {
  repeated string ids = 1; // UUIDs
}

message Get{{.Name}}sByIDsResponse {
  repeated {{.Name}} {{.NameLower}}s = 1;
}
`

	t, err := template.New("proto").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := t.Execute(f, info); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Generated proto file: %s", protoFile)
	return nil
}

func compileProtoFile(info *GRPCEntityInfo) error {
	protoFile := filepath.Join(projectRoot, "proto", info.NameLower, info.NameLower+".proto")
	pkgProtoDir := filepath.Join(projectRoot, "pkg", "proto")

	cmd := exec.Command("protoc",
		"--proto_path="+filepath.Join(projectRoot, "proto"),
		"--go_out="+pkgProtoDir,
		"--go_opt=module=github.com/saurabh/entgo-microservices/pkg/proto",
		"--go-grpc_out="+pkgProtoDir,
		"--go-grpc_opt=module=github.com/saurabh/entgo-microservices/pkg/proto",
		protoFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("protoc failed: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Compiled proto file for %s", info.Name)
	return nil
}

func generateServiceFile(info *GRPCEntityInfo, serviceDir string) error {
	grpcDir := filepath.Join(serviceDir, "grpc")
	if err := os.MkdirAll(grpcDir, 0755); err != nil {
		return fmt.Errorf("failed to create grpc directory: %w", err)
	}

	serviceFile := filepath.Join(grpcDir, fmt.Sprintf("%s_service.go", info.NameLower))
	f, err := os.Create(serviceFile)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer f.Close()

	funcMap := template.FuncMap{
		"toProtoPascalCase": toProtoPascalCase,
	}

	tmpl := `package grpc

import (
	"context"

	"{{.ModuleName}}/internal/ent"
	"{{.ModuleName}}/internal/ent/{{.NameLower}}"

	"entgo.io/ent/privacy"
	"github.com/google/uuid"
	"github.com/saurabh/entgo-microservices/pkg/logger"
	{{.NameLower}}v1 "github.com/saurabh/entgo-microservices/pkg/proto/{{.NameLower}}/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type {{.Name}}Service struct {
	{{.NameLower}}v1.Unimplemented{{.Name}}ServiceServer
	db *ent.Client
}

func New{{.Name}}Service(db *ent.Client) *{{.Name}}Service {
	return &{{.Name}}Service{db: db}
}

func (s *{{.Name}}Service) Get{{.Name}}ByID(ctx context.Context, req *{{.NameLower}}v1.Get{{.Name}}ByIDRequest) (*{{.NameLower}}v1.Get{{.Name}}ByIDResponse, error) {
	logger.WithField("{{.NameLower}}_id", req.Id).Debug("Get{{.Name}}ByID called")

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID: %v", err)
	}

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	entity, err := s.db.{{.Name}}.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "{{.NameLower}} not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get {{.NameLower}}")
		return nil, status.Errorf(codes.Internal, "failed to get {{.NameLower}}: %v", err)
	}

	return &{{.NameLower}}v1.Get{{.Name}}ByIDResponse{
		{{toProtoPascalCase .NameLower}}: convertEnt{{.Name}}ToProto(entity),
	}, nil
}

func (s *{{.Name}}Service) Get{{.Name}}sByIDs(ctx context.Context, req *{{.NameLower}}v1.Get{{.Name}}sByIDsRequest) (*{{.NameLower}}v1.Get{{.Name}}sByIDsResponse, error) {
	logger.WithField("{{.NameLower}}_ids", req.Ids).Debug("Get{{.Name}}sByIDs called")

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	ids := make([]uuid.UUID, 0, len(req.Ids))
	for _, idStr := range req.Ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid UUID: %v", err)
		}
		ids = append(ids, id)
	}

	entities, err := s.db.{{.Name}}.Query().
		Where({{.NameLower}}.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get {{.NameLower}}s")
		return nil, status.Errorf(codes.Internal, "failed to get {{.NameLower}}s: %v", err)
	}

	protoEntities := make([]*{{.NameLower}}v1.{{.Name}}, len(entities))
	for i, entity := range entities {
		protoEntities[i] = convertEnt{{.Name}}ToProto(entity)
	}

	return &{{.NameLower}}v1.Get{{.Name}}sByIDsResponse{
		{{toProtoPascalCase .NameLower}}s: protoEntities,
	}, nil
}


func convertEnt{{.Name}}ToProto(e *ent.{{.Name}}) *{{.NameLower}}v1.{{.Name}} {
	if e == nil {
		return nil
	}

	proto{{.Name}} := &{{.NameLower}}v1.{{.Name}}{
		Id:        e.ID.String(),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// Map additional fields
{{- range .Fields}}
{{- if and (ne .GoName "") (ne .GoName "ID") (ne .GoName "CreatedAt") (ne .GoName "UpdatedAt")}}
	{{- if and (eq .EntType "Time") .IsOptional}}
	if e.{{.GoName}} != nil && !e.{{.GoName}}.IsZero() {
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = timestamppb.New(*e.{{.GoName}})
	}
	{{- else if eq .EntType "Time"}}
	if !e.{{.GoName}}.IsZero() {
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = timestamppb.New(e.{{.GoName}})
	}
	{{- else if .IsOptional}}
	// Ent field IS a pointer
	if e.{{.GoName}} != nil {
		{{- if eq .Type "string"}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = e.{{.GoName}}
		{{- else if eq .Type "int32"}}
		{{- if .IsProtoOptional}}
		val{{.GoName}} := int32(*e.{{.GoName}})
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
		{{- else}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = int32(*e.{{.GoName}})
		{{- end}}
		{{- else if eq .Type "int64"}}
		{{- if .IsProtoOptional}}
		val{{.GoName}} := int64(*e.{{.GoName}})
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
		{{- else}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = int64(*e.{{.GoName}})
		{{- end}}
		{{- else if eq .Type "bool"}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = e.{{.GoName}}
		{{- else if eq .Type "float"}}
		{{- if .IsProtoOptional}}
		val{{.GoName}} := float32(*e.{{.GoName}})
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
		{{- else}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = float32(*e.{{.GoName}})
		{{- end}}
		{{- else if eq .Type "double"}}
		{{- if .IsProtoOptional}}
		val{{.GoName}} := float64(*e.{{.GoName}})
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
		{{- else}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = float64(*e.{{.GoName}})
		{{- end}}
		{{- else}}
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = e.{{.GoName}}
		{{- end}}
	}
	{{- else if .IsProtoOptional}}
	// Ent field is NOT a pointer but proto field IS optional
	{{- if eq .Type "string"}}
	if e.{{.GoName}} != "" {
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = &e.{{.GoName}}
	}
	{{- else if eq .Type "int32"}}
	val{{.GoName}} := int32(e.{{.GoName}})
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
	{{- else if eq .Type "int64"}}
	val{{.GoName}} := int64(e.{{.GoName}})
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
	{{- else if eq .Type "float"}}
	val{{.GoName}} := float32(e.{{.GoName}})
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
	{{- else if eq .Type "double"}}
	val{{.GoName}} := float64(e.{{.GoName}})
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
	{{- else if eq .EntType "Enum"}}
	if e.{{.GoName}} != "" {
		val{{.GoName}} := string(e.{{.GoName}})
		proto{{$.Name}}.{{toProtoPascalCase .Name}} = &val{{.GoName}}
	}
	{{- else}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = &e.{{.GoName}}
	{{- end}}
	{{- else}}
	// Ent field is NOT a pointer, proto field is NOT optional
	{{- if eq .Type "int32"}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = int32(e.{{.GoName}})
	{{- else if eq .Type "int64"}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = int64(e.{{.GoName}})
	{{- else if eq .Type "float"}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = float32(e.{{.GoName}})
	{{- else if eq .Type "double"}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = float64(e.{{.GoName}})
	{{- else if eq .EntType "Enum"}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = string(e.{{.GoName}})
	{{- else}}
	proto{{$.Name}}.{{toProtoPascalCase .Name}} = e.{{.GoName}}
	{{- end}}
	{{- end}}
{{- end}}
{{- end}}

	return proto{{.Name}}
}
`

	t, err := template.New("service").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	if err := t.Execute(f, info); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	log.Printf("Generated service file: %s", serviceFile)
	return nil
}

func generateServerRegistry(entityInfos []*GRPCEntityInfo, serviceDir string) error {
	grpcDir := filepath.Join(serviceDir, "grpc")
	registryFile := filepath.Join(grpcDir, "service_registry.go")

	f, err := os.Create(registryFile)
	if err != nil {
		return fmt.Errorf("failed to create registry file: %w", err)
	}
	defer f.Close()

	tmpl := `package grpc

import (
	"{{.ModuleName}}/internal/ent"
{{- range .EntityInfos}}
	{{.NameLower}}v1 "github.com/saurabh/entgo-microservices/pkg/proto/{{.NameLower}}/v1"
{{- end}}
	"google.golang.org/grpc"
)

// RegisterServices registers all gRPC services
func RegisterServices(grpcServer *grpc.Server, db *ent.Client) {
{{- range .EntityInfos}}
	{{.NameLower}}v1.Register{{.Name}}ServiceServer(grpcServer, New{{.Name}}Service(db))
{{- end}}
}
`

	t, err := template.New("registry").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse registry template: %w", err)
	}

	data := struct {
		ModuleName  string
		EntityInfos []*GRPCEntityInfo
	}{
		ModuleName:  entityInfos[0].ModuleName,
		EntityInfos: entityInfos,
	}

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute registry template: %w", err)
	}

	log.Printf("Generated service registry file: %s", registryFile)
	return nil
}

func generateServiceMetadata(entityInfos []*GRPCEntityInfo, serviceName string) error {
	metadataDir := filepath.Join(projectRoot, "pkg", "grpc", "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metadataFile := filepath.Join(metadataDir, serviceName+"_services.json")

	models := make([]string, len(entityInfos))
	for i, info := range entityInfos {
		models[i] = info.Name
	}

	metadata := ServiceMetadata{
		ServiceName: serviceName,
		Models:      models,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	log.Printf("Generated service metadata file: %s", metadataFile)
	return nil
}

func generateConsolidatedServiceClient(entityInfos []*GRPCEntityInfo, serviceName string) error {
	clientDir := filepath.Join(projectRoot, "pkg", "grpc")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}

	clientFile := filepath.Join(clientDir, serviceName+"_service_client.go")
	f, err := os.Create(clientFile)
	if err != nil {
		return fmt.Errorf("failed to create client file: %w", err)
	}
	defer f.Close()

	serviceTitleCase := strings.Title(serviceName)

	funcMap := template.FuncMap{
		"toProtoPascalCase": toProtoPascalCase,
	}

	tmpl := `// Code generated by generate-grpc-for-service.go. DO NOT EDIT.
package grpc

import (
	"context"
	"fmt"
	"sync"

{{- range .EntityInfos}}
	{{.NameLower}}v1 "github.com/saurabh/entgo-microservices/pkg/proto/{{.NameLower}}/v1"
{{- end}}
	"google.golang.org/grpc"
)

// {{.ServiceTitleCase}}ServiceClient provides access to all {{.ServiceName}} service models
type {{.ServiceTitleCase}}ServiceClient struct {
	gatewayClient *GatewayClient
{{- range .EntityInfos}}
	{{.NameLower}}Client     *{{.Name}}Client
	{{.NameLower}}ClientOnce sync.Once
{{- end}}
}

// New{{.ServiceTitleCase}}ServiceClient creates a new {{.ServiceName}} service client
func New{{.ServiceTitleCase}}ServiceClient(gatewayClient *GatewayClient) *{{.ServiceTitleCase}}ServiceClient {
	return &{{.ServiceTitleCase}}ServiceClient{
		gatewayClient: gatewayClient,
	}
}

{{- range .EntityInfos}}

// {{.Name}} returns the {{.Name}} client (lazy initialized)
func (c *{{$.ServiceTitleCase}}ServiceClient) {{.Name}}() *{{.Name}}Client {
	c.{{.NameLower}}ClientOnce.Do(func() {
		c.{{.NameLower}}Client = New{{.Name}}Client(c.gatewayClient)
	})
	return c.{{.NameLower}}Client
}

// {{.Name}}Client provides helper methods for {{.Name}} gRPC operations
type {{.Name}}Client struct {
	gatewayClient *GatewayClient
}

// New{{.Name}}Client creates a new {{.Name}} service client
func New{{.Name}}Client(gatewayClient *GatewayClient) *{{.Name}}Client {
	return &{{.Name}}Client{
		gatewayClient: gatewayClient,
	}
}

// New{{.Name}}ClientFromConn creates a new {{.Name}} service client from a direct connection
func New{{.Name}}ClientFromConn(conn *grpc.ClientConn) *{{.Name}}Client {
	return &{{.Name}}Client{
		gatewayClient: &GatewayClient{conn: conn, useGateway: true},
	}
}

// Get{{.Name}}ByID fetches a {{.NameLower}} by ID (UUID)
func (c *{{.Name}}Client) Get{{.Name}}ByID(ctx context.Context, id string) (*{{.NameLower}}v1.{{.Name}}, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := {{.NameLower}}v1.New{{.Name}}ServiceClient(conn)
	resp, err := client.Get{{.Name}}ByID(ctx, &{{.NameLower}}v1.Get{{.Name}}ByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get {{.NameLower}} by id: %w", err)
	}
	return resp.{{toProtoPascalCase .NameLower}}, nil
}

// Get{{.Name}}sByIDs fetches multiple {{.NameLower}}s by their IDs (UUIDs)
func (c *{{.Name}}Client) Get{{.Name}}sByIDs(ctx context.Context, ids []string) ([]*{{.NameLower}}v1.{{.Name}}, error) {
	conn, err := c.gatewayClient.GetGatewayConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway connection: %w", err)
	}

	client := {{.NameLower}}v1.New{{.Name}}ServiceClient(conn)
	resp, err := client.Get{{.Name}}sByIDs(ctx, &{{.NameLower}}v1.Get{{.Name}}sByIDsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get {{.NameLower}}s by ids: %w", err)
	}

	return resp.{{toProtoPascalCase .NameLower}}s, nil
}
{{- end}}
`

	t, err := template.New("client").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse client template: %w", err)
	}

	data := struct {
		ServiceName      string
		ServiceTitleCase string
		EntityInfos      []*GRPCEntityInfo
	}{
		ServiceName:      serviceName,
		ServiceTitleCase: serviceTitleCase,
		EntityInfos:      entityInfos,
	}

	if err := t.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute client template: %w", err)
	}

	log.Printf("Generated consolidated service client file: %s", clientFile)
	return nil
}

// Helper functions

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func toPascalCase(s string) string {
	commonAbbreviations := map[string]bool{
		"ip": true, "id": true, "url": true, "uri": true, "api": true,
		"ui": true, "uuid": true, "http": true, "https": true, "sql": true,
		"db": true, "acl": true, "json": true, "tls": true,
		"xml": true, "csv": true, "pdf": true, "html": true, "css": true,
	}

	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			lowerPart := strings.ToLower(part)
			if commonAbbreviations[lowerPart] {
				parts[i] = strings.ToUpper(lowerPart)
			} else {
				parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
	}
	result := strings.Join(parts, "")

	if len(result) > 0 {
		result = strings.ToUpper(result[:1]) + result[1:]
	}
	return result
}

func toProtoPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	result := strings.Join(parts, "")

	if len(result) > 0 {
		result = strings.ToUpper(result[:1]) + result[1:]
	}
	return result
}

func toLowerCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(pascal[:1]) + pascal[1:]
	}
	return pascal
}
