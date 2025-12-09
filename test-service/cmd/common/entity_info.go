package common

// EntityInfo holds information about an entity for template generation
type EntityInfo struct {
	Name             string // e.g., "Category", "APIKey"
	NameLower        string // e.g., "category", "api_key"
	NameCamel        string // e.g., "category", "apiKey"
	NamePlural       string // e.g., "Categories", "APIKeys"
	EntPackageName   string // e.g., "category", "inlineitem" (lowercase, no separators - for ent imports)
	ModuleName       string // e.g., "auth"
	GenerateResolver bool   // true if @generate-resolver: true
	GenerateMutation bool   // true if @generate-mutation: true
	GenerateService  bool   // true if @generate-service: true
	GenerateGRPC     bool   // true if @generate-grpc: true
	GenerateHooks    bool   // true if @generate-hooks: true
	GeneratePrivacy  bool   // true if @generate-privacy: true
}

// AnnotationData holds parsed annotation data from schema files
type AnnotationData struct {
	RoleLevel        string   // e.g., "admin"
	PermissionLevel  string   // e.g., "brand"
	CreateRoles      []string // Roles for create operation
	UpdateRoles      []string // Roles for update operation
	DeleteRoles      []string // Roles for delete operation
	GenerateResolver bool
	GenerateMutation bool
	GenerateService  bool
	GenerateGRPC     bool
	GenerateHooks    bool
	GeneratePrivacy  bool
	TenantIsolated   bool
	UserOwned        bool
	FilterByEdge     string
}
