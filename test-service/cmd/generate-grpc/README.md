# gRPC Code Generator

This tool automatically generates gRPC services for Ent schema entities.

## Usage

### 1. Add Annotation to Schema

Add the `@generate-grpc: true` annotation to any schema file you want to generate gRPC services for:

```go
// Fields of the RolePermission.
// @generate-mutation: true
// @generate-resolver: true
// @generate-grpc: true
func (RolePermission) Fields() []ent.Field {
    return []ent.Field{
        field.Bool("can_read").Default(false).Comment("Can read the resource"),
        field.Bool("can_create").Default(false).Comment("Can create the resource"),
        // ... more fields
    }
}
```

### 2. Run Generation

From the `auth` directory, run:

```bash
go run ./cmd/generate-grpc/main.go
```

Or use the full generation pipeline:

```bash
go generate
```

### 3. Generated Files

The generator creates three types of files:

1. **Proto Definition**: `proto/{entity}/{entity}.proto`
   - Protocol buffer service and message definitions
   - Service methods: Get{Entity}ByID, Get{Entity}sByIDs

2. **Generated Proto Code**: `pkg/proto/{entity}/v1/*.pb.go`
   - Compiled protobuf Go code
   - gRPC service stubs

3. **Service Implementation**: `auth/grpc/{entity}_service.go`
   - Service struct and methods
   - Converter function (with TODO for field mapping)

## Supported Field Types

The generator currently supports these basic Ent field types:

- `String`, `Text` → `string`
- `Int` → `int32`
- `Int64` → `int64`
- `Bool` → `bool`
- `Float`, `Float64` → `double`
- `Float32` → `float`
- `Time` → `google.protobuf.Timestamp`
- `Enum` → `string` (mapped to string for simplicity)

## Generated Service Methods

Each generated service includes:

- `Get{Entity}ByID(id: int32)` - Get single entity by ID
- `Get{Entity}sByIDs(ids: []int32)` - Get multiple entities by IDs

## Field Mapping

The converter function (`convertEnt{Entity}ToProto`) is generated with a TODO comment.
You need to manually add field mappings for entity-specific fields.

Example:
```go
func convertEntRolePermissionToProto(e *ent.RolePermission) *rolepermissionv1.RolePermission {
    protoRolePermission := &rolepermissionv1.RolePermission{
        Id:        int32(e.ID),
        CreatedAt: timestamppb.New(e.CreatedAt),
        UpdatedAt: timestamppb.New(e.UpdatedAt),
        // Add your fields here:
        CanRead:   e.CanRead,
        CanCreate: e.CanCreate,
        CanUpdate: e.CanUpdate,
        CanDelete: e.CanDelete,
    }
    return protoRolePermission
}
```

## Limitations

- Currently only generates read operations (GetByID, GetByIDs)
- Complex types (JSON, edges) are not supported
- Field names with `_id` suffix are skipped (assumed to be foreign keys)
- Optional/nullable fields are not yet properly marked in proto
- Converter functions need manual field mapping

## Integration with Build Pipeline

The generator is integrated into the `generate.go` pipeline:

```go
// Step 6: Generate gRPC services
//go:generate go run ./cmd/generate-grpc/main.go
```

It runs after Ent code generation and before GraphQL resolver generation.

## Proto Compilation

Proto files are automatically compiled using `protoc`. Ensure you have:

- `protoc` installed (Protocol Buffer Compiler)
- `protoc-gen-go` (Go protobuf plugin)
- `protoc-gen-go-grpc` (Go gRPC plugin)

Install plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Dynamic Proto Compilation

The `proto/Makefile` has been updated to dynamically discover and compile all proto files:

```bash
cd proto
make proto
```

This will compile all `.proto` files in subdirectories.

