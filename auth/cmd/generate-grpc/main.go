package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/saurabh/entgo-microservices/auth/cmd/common"
)

// ProtoField represents a field in a protobuf message
type ProtoField struct {
	Name       string // proto field name (snake_case)
	Type       string // proto type
	Number     int    // field number
	IsOptional bool   // whether field is optional
	Comment    string // field comment
}

// GRPCEntityInfo holds information for gRPC generation
type GRPCEntityInfo struct {
	Name           string       // e.g., "User"
	NameLower      string       // e.g., "user"
	NameCamel      string       // e.g., "user"
	ProtoFieldName string       // e.g., "User" or "Rolepermission" - how proto capitalizes the field
	Package        string       // e.g., "user.v1"
	GoPackage      string       // e.g., "github.com/saurabh/entgo-microservices/pkg/proto/user/v1;userv1"
	ModuleName     string       // e.g., "auth"
	Fields         []ProtoField // proto fields
	HasTimestamps  bool         // whether to import google/protobuf/timestamp.proto
}

const protoTemplate = `syntax = "proto3";

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
  {{if .IsOptional}}optional {{end}}{{.Type}} {{.Name}} = {{.Number}};{{if .Comment}} // {{.Comment}}{{end}}
{{- end}}
}

message Get{{.Name}}ByIDRequest {
  int32 id = 1;
}

message Get{{.Name}}ByIDResponse {
  {{.Name}} {{.NameLower}} = 1;
}

message Get{{.Name}}sByIDsRequest {
  repeated int32 ids = 1;
}

message Get{{.Name}}sByIDsResponse {
  repeated {{.Name}} {{.NameLower}}s = 1;
}
`

const serviceTemplate = `package grpc

import (
	"context"

	"{{.ModuleName}}/internal/ent"
	"{{.ModuleName}}/internal/ent/{{.NameLower}}"

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

	entity, err := s.db.{{.Name}}.Get(ctx, int(req.Id))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "{{.NameLower}} not found: %v", err)
		}
		logger.WithError(err).Error("Failed to get {{.NameLower}}")
		return nil, status.Errorf(codes.Internal, "failed to get {{.NameLower}}: %v", err)
	}

	return &{{.NameLower}}v1.Get{{.Name}}ByIDResponse{
		{{.ProtoFieldName}}: convertEnt{{.Name}}ToProto(entity),
	}, nil
}

func (s *{{.Name}}Service) Get{{.Name}}sByIDs(ctx context.Context, req *{{.NameLower}}v1.Get{{.Name}}sByIDsRequest) (*{{.NameLower}}v1.Get{{.Name}}sByIDsResponse, error) {
	logger.WithField("{{.NameLower}}_ids", req.Ids).Debug("Get{{.Name}}sByIDs called")

	ids := make([]int, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int(id)
	}

	entities, err := s.db.{{.Name}}.Query().
		Where({{.NameLower}}.IDIn(ids...)).
		All(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get {{.NameLower}}s")
		return nil, status.Errorf(codes.Internal, "failed to get {{.NameLower}}s: %v", err)
	}

	proto{{.Name}}s := make([]*{{.NameLower}}v1.{{.Name}}, len(entities))
	for i, e := range entities {
		proto{{.Name}}s[i] = convertEnt{{.Name}}ToProto(e)
	}

	return &{{.NameLower}}v1.Get{{.Name}}sByIDsResponse{
		{{.ProtoFieldName}}s: proto{{.Name}}s,
	}, nil
}

func convertEnt{{.Name}}ToProto(e *ent.{{.Name}}) *{{.NameLower}}v1.{{.Name}} {
	proto{{.Name}} := &{{.NameLower}}v1.{{.Name}}{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// TODO: Map additional fields from ent entity to proto message
	// This needs to be manually filled in based on your entity fields

	return proto{{.Name}}
}
`

func main() {
	log.Println("Starting gRPC generation...")

	// Get module name
	moduleName, err := common.GetModuleName()
	if err != nil {
		log.Fatalf("Error reading module name: %v", err)
	}

	// Get schema files
	schemaFiles, err := common.GetSchemaFiles("ent/schema")
	if err != nil {
		log.Fatalf("Error reading schema files: %v", err)
	}

	// Process each schema file
	for _, schemaFile := range schemaFiles {
		// Check for @generate-grpc: true annotation
		hasAnnotation, err := common.CheckAnnotation(schemaFile, "@generate-grpc: true")
		if err != nil {
			log.Printf("Error checking annotation in %s: %v", schemaFile, err)
			continue
		}

		if !hasAnnotation {
			continue
		}

		// Extract entity name
		entityName, err := common.ExtractEntityName(schemaFile)
		if err != nil {
			log.Printf("Error extracting entity name from %s: %v", schemaFile, err)
			continue
		}

		log.Printf("Processing entity: %s", entityName)

		// Extract fields from schema
		fields, err := extractFieldsFromSchema(schemaFile)
		if err != nil {
			log.Printf("Error extracting fields from %s: %v", schemaFile, err)
			continue
		}

		// Build entity info
		entityInfo := buildEntityInfo(entityName, moduleName, fields)

		// Generate proto file
		if err := generateProtoFile(entityInfo); err != nil {
			log.Printf("Error generating proto file for %s: %v", entityName, err)
			continue
		}

		// Compile proto file
		if err := compileProtoFile(entityInfo); err != nil {
			log.Printf("Error compiling proto file for %s: %v", entityName, err)
			continue
		}

		// Generate service file
		if err := generateServiceFile(entityInfo); err != nil {
			log.Printf("Error generating service file for %s: %v", entityName, err)
			continue
		}

		log.Printf("Successfully generated gRPC code for %s", entityName)
	}

	log.Println("gRPC generation completed!")
}

func extractFieldsFromSchema(filePath string) ([]ProtoField, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var fields []ProtoField
	fieldNum := 1

	// Always add base fields (id, created_at, updated_at)
	fields = append(fields,
		ProtoField{Name: "id", Type: "int32", Number: fieldNum, IsOptional: false},
	)
	fieldNum++

	// Extract field definitions using regex
	// Matches patterns like: field.String("name"), field.Int("age"), field.Bool("active")
	fieldRegex := regexp.MustCompile(`field\.(String|Text|Int|Int64|Bool|Float|Float32|Float64|Time|Enum)\("([^"]+)"\)`)
	matches := fieldRegex.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		fieldType := match[1]
		fieldName := match[2]

		// Skip special fields
		if fieldName == "tenant_id" || strings.HasSuffix(fieldName, "_id") {
			continue
		}

		// Map ent field types to proto types
		protoType := mapEntTypeToProtoType(fieldType)
		if protoType == "" {
			continue // Skip unsupported types
		}

		// Check if field is optional by looking for .Optional() after the field definition
		// Simple heuristic: check if Optional appears on the same or next line
		isOptional := false
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.Contains(line, fmt.Sprintf(`field.%s("%s")`, fieldType, fieldName)) {
				// Check current line and next few lines for Optional()
				for j := i; j < i+5 && j < len(lines); j++ {
					if strings.Contains(lines[j], ".Optional()") || strings.Contains(lines[j], ".Nillable()") {
						isOptional = true
						break
					}
				}
				break
			}
		}

		fields = append(fields, ProtoField{
			Name:       toSnakeCase(fieldName),
			Type:       protoType,
			Number:     fieldNum,
			IsOptional: isOptional,
		})
		fieldNum++
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

func mapEntTypeToProtoType(entType string) string {
	switch entType {
	case "String", "Text":
		return "string"
	case "Int":
		return "int32"
	case "Int64":
		return "int64"
	case "Bool":
		return "bool"
	case "Float", "Float64":
		return "double"
	case "Float32":
		return "float"
	case "Time":
		return "google.protobuf.Timestamp"
	case "Enum":
		return "string" // Enums are mapped to strings for simplicity
	default:
		return "" // Unsupported type
	}
}

func buildEntityInfo(entityName, moduleName string, fields []ProtoField) *GRPCEntityInfo {
	nameLower := strings.ToLower(entityName)
	hasTimestamps := false

	for _, field := range fields {
		if strings.Contains(field.Type, "google.protobuf.Timestamp") {
			hasTimestamps = true
			break
		}
	}

	// ProtoFieldName: capitalize first letter of nameLower (how protoc generates Go field names)
	protoFieldName := strings.ToUpper(nameLower[0:1]) + nameLower[1:]

	return &GRPCEntityInfo{
		Name:           entityName,
		NameLower:      nameLower,
		NameCamel:      nameLower,
		ProtoFieldName: protoFieldName,
		Package:        fmt.Sprintf("%s.v1", nameLower),
		GoPackage:      fmt.Sprintf("github.com/saurabh/entgo-microservices/pkg/%s/v1;%sv1", nameLower, nameLower),
		ModuleName:     moduleName,
		Fields:         fields,
		HasTimestamps:  hasTimestamps,
	}
}

func generateProtoFile(info *GRPCEntityInfo) error {
	// Create proto directory
	protoDir := filepath.Join("../proto", info.NameLower)
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		return fmt.Errorf("failed to create proto directory: %w", err)
	}

	// Create proto file
	protoFile := filepath.Join(protoDir, fmt.Sprintf("%s.proto", info.NameLower))
	f, err := os.Create(protoFile)
	if err != nil {
		return fmt.Errorf("failed to create proto file: %w", err)
	}
	defer f.Close()

	// Execute template
	tmpl, err := template.New("proto").Parse(protoTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse proto template: %w", err)
	}

	if err := tmpl.Execute(f, info); err != nil {
		return fmt.Errorf("failed to execute proto template: %w", err)
	}

	log.Printf("Generated proto file: %s", protoFile)
	return nil
}

func compileProtoFile(info *GRPCEntityInfo) error {
	protoFile := filepath.Join("../proto", info.NameLower, fmt.Sprintf("%s.proto", info.NameLower))
	pkgProtoDir := "../pkg/proto"

	// Create output directory
	outputDir := filepath.Join(pkgProtoDir, info.NameLower, "v1")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Run protoc
	cmd := exec.Command("protoc",
		"--proto_path=../proto",
		"--go_out="+pkgProtoDir,
		"--go_opt=module=github.com/saurabh/entgo-microservices/pkg",
		"--go-grpc_out="+pkgProtoDir,
		"--go-grpc_opt=module=github.com/saurabh/entgo-microservices/pkg",
		protoFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("protoc failed: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Compiled proto file for %s", info.Name)
	return nil
}

func generateServiceFile(info *GRPCEntityInfo) error {
	// Create service file
	serviceFile := filepath.Join("grpc", fmt.Sprintf("%s_service.go", info.NameLower))
	f, err := os.Create(serviceFile)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer f.Close()

	// Execute template
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	if err := tmpl.Execute(f, info); err != nil {
		return fmt.Errorf("failed to execute service template: %w", err)
	}

	log.Printf("Generated service file: %s", serviceFile)
	return nil
}

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
