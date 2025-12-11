package main

import (
	"encoding/json"
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
	GoName     string // Go field name (PascalCase) from Ent
	Type       string // proto type
	EntType    string // original Ent field type
	Number     int    // field number
	IsOptional bool   // whether field is optional
	Comment    string // field comment
}

// GRPCEntityInfo holds information for gRPC generation
type GRPCEntityInfo struct {
	Name            string       // e.g., "User"
	NameLower       string       // e.g., "user"
	NameCamel       string       // e.g., "user"
	ProtoFieldName  string       // e.g., "User" or "Rolepermission" - how proto capitalizes the field
	Package         string       // e.g., "user.v1"
	GoPackage       string       // e.g., "github.com/saurabh/entgo-microservices/pkg/proto/user/v1;userv1"
	ModuleName      string       // e.g., "auth"
	ModuleShortName string       // e.g., "attendance" (extracted from full module path)
	Fields          []ProtoField // proto fields
	HasTimestamps   bool         // whether to import google/protobuf/timestamp.proto
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

	"entgo.io/ent/privacy"
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

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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

	// Bypass privacy policies for internal gRPC communication
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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
	if e == nil {
		return nil
	}

	proto{{.Name}} := &{{.NameLower}}v1.{{.Name}}{
		Id:        int32(e.ID),
		CreatedAt: timestamppb.New(e.CreatedAt),
		UpdatedAt: timestamppb.New(e.UpdatedAt),
	}

	// Map additional fields
{{- range .Fields}}
{{- if and (ne .GoName "") (ne .GoName "ID") (ne .GoName "CreatedAt") (ne .GoName "UpdatedAt")}}
	{{- if and (eq .EntType "Time") .IsOptional}}
	if e.{{.GoName}} != nil && !e.{{.GoName}}.IsZero() {
		proto{{$.Name}}.{{toPascalCase .Name}} = timestamppb.New(*e.{{.GoName}})
	}
	{{- else if eq .EntType "Time"}}
	if !e.{{.GoName}}.IsZero() {
		proto{{$.Name}}.{{toPascalCase .Name}} = timestamppb.New(e.{{.GoName}})
	}
	{{- else if .IsOptional}}
	if e.{{.GoName}} != nil {
		{{- if eq .Type "string"}}
		proto{{$.Name}}.{{toPascalCase .Name}} = e.{{.GoName}}
		{{- else if or (eq .Type "int32") (eq .Type "int64")}}
		val := {{.Type}}(*e.{{.GoName}})
		proto{{$.Name}}.{{toPascalCase .Name}} = &val
		{{- else if eq .Type "bool"}}
		proto{{$.Name}}.{{toPascalCase .Name}} = e.{{.GoName}}
		{{- else if or (eq .Type "float") (eq .Type "double")}}
		val := float64(*e.{{.GoName}})
		proto{{$.Name}}.{{toPascalCase .Name}} = &val
		{{- else}}
		proto{{$.Name}}.{{toPascalCase .Name}} = e.{{.GoName}}
		{{- end}}
	}
	{{- else}}
	{{- if or (eq .Type "int32") (eq .Type "int64")}}
	proto{{$.Name}}.{{toPascalCase .Name}} = {{.Type}}(e.{{.GoName}})
	{{- else if or (eq .Type "float") (eq .Type "double")}}
	proto{{$.Name}}.{{toPascalCase .Name}} = float64(e.{{.GoName}})
	{{- else}}
	proto{{$.Name}}.{{toPascalCase .Name}} = e.{{.GoName}}
	{{- end}}
	{{- end}}
{{- end}}
{{- end}}

	return proto{{.Name}}
}
`

const clientHelperTemplate = `package grpc

import (
	"context"
	"fmt"

	{{.NameLower}}v1 "github.com/saurabh/entgo-microservices/pkg/proto/{{.NameLower}}/v1"
	"google.golang.org/grpc"
)

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

// Get{{.Name}}ByID fetches a {{.NameLower}} by ID
func (c *{{.Name}}Client) Get{{.Name}}ByID(ctx context.Context, id int32) (*{{.NameLower}}v1.{{.Name}}, error) {
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

	return resp.{{.ProtoFieldName}}, nil
}

// Get{{.Name}}sByIDs fetches multiple {{.NameLower}}s by their IDs
func (c *{{.Name}}Client) Get{{.Name}}sByIDs(ctx context.Context, ids []int32) ([]*{{.NameLower}}v1.{{.Name}}, error) {
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

	return resp.{{.ProtoFieldName}}s, nil
}
`

const autoClientsTemplate = `package grpc

import (
	"sync"
)

// AutoClients provides lazy-initialized typed clients for all generated services
type AutoClients struct {
	gatewayClient *GatewayClient
{{- range .Entities}}
	{{.NameLower}}Client     *{{.Name}}Client
	{{.NameLower}}ClientOnce sync.Once
{{- end}}
}

// NewAutoClients creates a new auto clients instance
func NewAutoClients(gatewayClient *GatewayClient) *AutoClients {
	return &AutoClients{
		gatewayClient: gatewayClient,
	}
}

{{- range .Entities}}

// Get{{.Name}}Client returns a lazy-initialized {{.Name}} client
func (a *AutoClients) Get{{.Name}}Client() *{{.Name}}Client {
	a.{{.NameLower}}ClientOnce.Do(func() {
		a.{{.NameLower}}Client = New{{.Name}}Client(a.gatewayClient)
	})
	return a.{{.NameLower}}Client
}
{{- end}}
`

const serverRegistryTemplate = `package grpc

import (
	"google.golang.org/grpc"
	"github.com/saurabh/entgo-microservices/{{.ModuleShortName}}/internal/ent"
{{- range .Entities}}
	{{.NameLower}}v1 "github.com/saurabh/entgo-microservices/pkg/proto/{{.NameLower}}/v1"
{{- end}}
)

// RegisterAllServices registers all generated gRPC services
func RegisterAllServices(s *grpc.Server, db *ent.Client) {
{{- range .Entities}}
	{{.NameLower}}v1.Register{{.Name}}ServiceServer(s, New{{.Name}}Service(db))
{{- end}}
}

// GetServiceNames returns list of all registered service names
func GetServiceNames() []string {
	return []string{
{{- range .Entities}}
		"{{.Name}}Service",
{{- end}}
	}
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

	var entityInfos []*GRPCEntityInfo

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

		entityInfos = append(entityInfos, entityInfo)
		log.Printf("Successfully generated gRPC code for %s", entityName)
	}

	// Generate server registry file
	if len(entityInfos) > 0 {
		if err := generateServerRegistry(entityInfos); err != nil {
			log.Printf("Error generating server registry: %v", err)
		}

		// Generate service metadata
		if err := generateServiceMetadata(entityInfos); err != nil {
			log.Printf("Error generating service metadata: %v", err)
		}

		// Generate consolidated service client (one file per microservice with all models)
		if err := generateConsolidatedServiceClient(entityInfos); err != nil {
			log.Printf("Error generating consolidated service client: %v", err)
		}
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

	// Extract entity name from schema file
	entityName := ""
	entityRegex := regexp.MustCompile(`type (\w+) struct`)
	if matches := entityRegex.FindStringSubmatch(string(content)); len(matches) > 1 {
		entityName = strings.ToLower(matches[1])
	}

	// Read the generated Ent entity file to check which fields are pointers (optional)
	entFilePath := filepath.Join("internal", "ent", entityName+".go")
	entContent, err := os.ReadFile(entFilePath)
	var entFieldTypes map[string]bool // true if field is a pointer
	if err == nil {
		entFieldTypes = extractEntFieldTypes(string(entContent))
	} else {
		log.Printf("Warning: Could not read generated Ent file %s, will assume all fields are required", entFilePath)
		entFieldTypes = make(map[string]bool)
	}

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

		// Check if field is optional by looking at generated Ent code
		goFieldName := toPascalCase(fieldName)
		isOptional := entFieldTypes[goFieldName]

		fields = append(fields, ProtoField{
			Name:       toSnakeCase(fieldName),
			GoName:     goFieldName,
			Type:       protoType,
			EntType:    fieldType,
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

// extractEntFieldTypes reads generated Ent entity code and returns map of field names to whether they're pointers
func extractEntFieldTypes(entContent string) map[string]bool {
	fieldTypes := make(map[string]bool)

	// Find the struct definition
	// Look for field definitions like: FieldName *string `json:"..."` or FieldName string `json:"..."`
	fieldRegex := regexp.MustCompile(`(?m)^\s+(\w+)\s+(\*?)(\w+(?:\.\w+)?)\s+` + "`" + `json:"([^"]+)"`)
	matches := fieldRegex.FindAllStringSubmatch(entContent, -1)

	for _, match := range matches {
		if len(match) < 5 {
			continue
		}

		fieldName := match[1]
		isPointer := match[2] == "*"
		// goType := match[3] // e.g. "string", "int", "float64", "time.Time"

		// Skip internal/system fields
		if fieldName == "selectValues" || fieldName == "ID" || strings.HasSuffix(fieldName, "Edges") {
			continue
		}

		fieldTypes[fieldName] = isPointer
	}

	return fieldTypes
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

	// Extract module short name from full path
	moduleShortName := moduleName
	if parts := strings.Split(moduleName, "/"); len(parts) > 0 {
		moduleShortName = parts[len(parts)-1]
	}

	return &GRPCEntityInfo{
		Name:            entityName,
		NameLower:       nameLower,
		NameCamel:       nameLower,
		ProtoFieldName:  protoFieldName,
		Package:         fmt.Sprintf("%s.v1", nameLower),
		GoPackage:       fmt.Sprintf("github.com/saurabh/entgo-microservices/pkg/proto/%s/v1;%sv1", nameLower, nameLower),
		ModuleName:      moduleName,
		ModuleShortName: moduleShortName,
		Fields:          fields,
		HasTimestamps:   hasTimestamps,
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

func generateServiceFile(info *GRPCEntityInfo) error {
	// Create service file
	serviceFile := filepath.Join("grpc", fmt.Sprintf("%s_service.go", info.NameLower))
	f, err := os.Create(serviceFile)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer f.Close()

	// Execute template with custom functions
	funcMap := template.FuncMap{
		"toPascalCase": toPascalCase,
	}
	tmpl, err := template.New("service").Funcs(funcMap).Parse(serviceTemplate)
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

func toPascalCase(s string) string {
	// Convert snake_case or normal string to PascalCase
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	result := strings.Join(parts, "")

	// Ensure first letter is uppercase
	if len(result) > 0 {
		result = strings.ToUpper(result[:1]) + result[1:]
	}
	return result
}

func generateConsolidatedServiceClient(entityInfos []*GRPCEntityInfo) error {
	if len(entityInfos) == 0 {
		return nil
	}

	// Use first entity for module info
	firstEntity := entityInfos[0]
	moduleName := strings.Title(firstEntity.ModuleShortName)

	// Create client directory
	clientDir := "../pkg/grpc"
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return fmt.Errorf("failed to create client directory: %w", err)
	}

	clientFile := filepath.Join(clientDir, fmt.Sprintf("%s_service_client.go", firstEntity.ModuleShortName))
	f, err := os.Create(clientFile)
	if err != nil {
		return fmt.Errorf("failed to create service client file: %w", err)
	}
	defer f.Close()

	// Write package and imports
	fmt.Fprintf(f, "package grpc\n\n")
	fmt.Fprintf(f, "import (\n")
	fmt.Fprintf(f, "\t\"context\"\n")
	fmt.Fprintf(f, "\t\"fmt\"\n\n")

	// Import all proto packages for this microservice
	for _, entity := range entityInfos {
		fmt.Fprintf(f, "\t%sv1 \"github.com/saurabh/entgo-microservices/pkg/proto/%s/v1\"\n",
			entity.NameLower, entity.NameLower)
	}
	fmt.Fprintf(f, ")\n\n")

	// Write service client struct
	fmt.Fprintf(f, "// %sServiceClient provides access to all models in the %s microservice\n",
		moduleName, moduleName)
	fmt.Fprintf(f, "type %sServiceClient struct {\n", moduleName)
	fmt.Fprintf(f, "\tgatewayClient *GatewayClient\n")
	fmt.Fprintf(f, "}\n\n")

	// Write constructor
	fmt.Fprintf(f, "// New%sServiceClient creates a new %s service client\n", moduleName, moduleName)
	fmt.Fprintf(f, "func New%sServiceClient(gatewayClient *GatewayClient) *%sServiceClient {\n",
		moduleName, moduleName)
	fmt.Fprintf(f, "\treturn &%sServiceClient{\n", moduleName)
	fmt.Fprintf(f, "\t\tgatewayClient: gatewayClient,\n")
	fmt.Fprintf(f, "\t}\n")
	fmt.Fprintf(f, "}\n\n")

	// Generate methods for each entity
	for _, entity := range entityInfos {
		if err := writeEntityMethods(f, entity, moduleName); err != nil {
			return fmt.Errorf("failed to write methods for %s: %w", entity.Name, err)
		}
	}

	log.Printf("Generated consolidated service client file: %s", clientFile)
	return nil
}

func writeEntityMethods(f *os.File, entity *GRPCEntityInfo, moduleName string) error {
	fmt.Fprintf(f, "// ============================================\n")
	fmt.Fprintf(f, "// %s operations\n", entity.Name)
	fmt.Fprintf(f, "// ============================================\n\n")

	// Get by ID method
	fmt.Fprintf(f, "// Get%sByID fetches a %s by ID\n", entity.Name, entity.NameLower)
	fmt.Fprintf(f, "func (c *%sServiceClient) Get%sByID(ctx context.Context, id int32) (*%sv1.%s, error) {\n",
		moduleName, entity.Name, entity.NameLower, entity.Name)
	fmt.Fprintf(f, "\tconn, err := c.gatewayClient.GetGatewayConnection()\n")
	fmt.Fprintf(f, "\tif err != nil {\n")
	fmt.Fprintf(f, "\t\treturn nil, fmt.Errorf(\"failed to get gateway connection: %%w\", err)\n")
	fmt.Fprintf(f, "\t}\n\n")
	fmt.Fprintf(f, "\tclient := %sv1.New%sServiceClient(conn)\n", entity.NameLower, entity.Name)
	fmt.Fprintf(f, "\tresp, err := client.Get%sByID(ctx, &%sv1.Get%sByIDRequest{\n",
		entity.Name, entity.NameLower, entity.Name)
	fmt.Fprintf(f, "\t\tId: id,\n")
	fmt.Fprintf(f, "\t})\n")
	fmt.Fprintf(f, "\tif err != nil {\n")
	fmt.Fprintf(f, "\t\treturn nil, fmt.Errorf(\"failed to get %s by id: %%w\", err)\n", entity.NameLower)
	fmt.Fprintf(f, "\t}\n\n")
	fmt.Fprintf(f, "\treturn resp.%s, nil\n", entity.ProtoFieldName)
	fmt.Fprintf(f, "}\n\n")

	// Get by IDs method
	fmt.Fprintf(f, "// Get%ssByIDs fetches multiple %ss by their IDs\n", entity.Name, entity.NameLower)
	fmt.Fprintf(f, "func (c *%sServiceClient) Get%ssByIDs(ctx context.Context, ids []int32) ([]*%sv1.%s, error) {\n",
		moduleName, entity.Name, entity.NameLower, entity.Name)
	fmt.Fprintf(f, "\tconn, err := c.gatewayClient.GetGatewayConnection()\n")
	fmt.Fprintf(f, "\tif err != nil {\n")
	fmt.Fprintf(f, "\t\treturn nil, fmt.Errorf(\"failed to get gateway connection: %%w\", err)\n")
	fmt.Fprintf(f, "\t}\n\n")
	fmt.Fprintf(f, "\tclient := %sv1.New%sServiceClient(conn)\n", entity.NameLower, entity.Name)
	fmt.Fprintf(f, "\tresp, err := client.Get%ssByIDs(ctx, &%sv1.Get%ssByIDsRequest{\n",
		entity.Name, entity.NameLower, entity.Name)
	fmt.Fprintf(f, "\t\tIds: ids,\n")
	fmt.Fprintf(f, "\t})\n")
	fmt.Fprintf(f, "\tif err != nil {\n")
	fmt.Fprintf(f, "\t\treturn nil, fmt.Errorf(\"failed to get %ss by ids: %%w\", err)\n", entity.NameLower)
	fmt.Fprintf(f, "\t}\n\n")
	fmt.Fprintf(f, "\treturn resp.%ss, nil\n", entity.ProtoFieldName)
	fmt.Fprintf(f, "}\n\n")

	return nil
}

func generateServerRegistry(entityInfos []*GRPCEntityInfo) error {
	if len(entityInfos) == 0 {
		return nil
	}

	// Use first entity info for module information
	firstEntity := entityInfos[0]

	data := struct {
		ModuleShortName string
		Entities        []*GRPCEntityInfo
	}{
		ModuleShortName: firstEntity.ModuleShortName,
		Entities:        entityInfos,
	}

	// Create registry file
	registryFile := filepath.Join("grpc", "service_registry.go")
	f, err := os.Create(registryFile)
	if err != nil {
		return fmt.Errorf("failed to create registry file: %w", err)
	}
	defer f.Close()

	// Execute template
	tmpl, err := template.New("registry").Parse(serverRegistryTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse registry template: %w", err)
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute registry template: %w", err)
	}

	log.Printf("Generated service registry file: %s", registryFile)
	return nil
}

type ServiceMetadata struct {
	Service          string `json:"service"`
	ProtoPackage     string `json:"proto_package"`
	MicroserviceName string `json:"microservice_name"`
	EntityName       string `json:"entity_name"`
}

func generateServiceMetadata(entityInfos []*GRPCEntityInfo) error {
	if len(entityInfos) == 0 {
		return nil
	}

	// Create metadata directory in pkg
	metadataDir := "../pkg/grpc/metadata"
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	firstEntity := entityInfos[0]
	metadataFile := filepath.Join(metadataDir, fmt.Sprintf("%s_services.json", firstEntity.ModuleShortName))

	var metadata []ServiceMetadata
	for _, info := range entityInfos {
		metadata = append(metadata, ServiceMetadata{
			Service:          fmt.Sprintf("%s.%sService", info.Package, info.Name),
			ProtoPackage:     info.Package,
			MicroserviceName: info.ModuleShortName,
			EntityName:       info.Name,
		})
	}

	// Write as JSON
	f, err := os.Create(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	log.Printf("Generated service metadata file: %s", metadataFile)
	return nil
}

func generateAutoClientsFactory(entityInfos []*GRPCEntityInfo) error {
	if len(entityInfos) == 0 {
		return nil
	}

	data := struct {
		Entities []*GRPCEntityInfo
	}{
		Entities: entityInfos,
	}

	// Create auto clients file in pkg/grpc
	clientFile := "../pkg/grpc/auto_clients.go"
	f, err := os.Create(clientFile)
	if err != nil {
		return fmt.Errorf("failed to create auto clients file: %w", err)
	}
	defer f.Close()

	// Execute template
	tmpl, err := template.New("autoclients").Parse(autoClientsTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse auto clients template: %w", err)
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute auto clients template: %w", err)
	}

	log.Printf("Generated auto clients factory file: %s", clientFile)
	return nil
}
