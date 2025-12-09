package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Service represents a GraphQL service configuration
type Service struct {
	Name string // Service name
	URL  string // GraphQL endpoint URL
}

// SchemaType represents a GraphQL type definition
type SchemaType struct {
	Kind          string      `json:"kind"` // Type kind (OBJECT, SCALAR, etc.)
	Name          string      `json:"name"` // Type name
	Description   string      `json:"description,omitempty"`
	Fields        interface{} `json:"fields,omitempty"`        // Fields for object types
	InputFields   interface{} `json:"inputFields,omitempty"`   // Fields for input object types
	Interfaces    interface{} `json:"interfaces,omitempty"`    // Implemented interfaces
	EnumValues    interface{} `json:"enumValues,omitempty"`    // Values for enum types
	PossibleTypes interface{} `json:"possibleTypes,omitempty"` // Possible types for interfaces/unions
}

// SchemaResponse represents a GraphQL introspection response
type SchemaResponse struct {
	Data struct {
		Schema struct {
			QueryType        map[string]string `json:"queryType"`        // Root Query type
			MutationType     map[string]string `json:"mutationType"`     // Root Mutation type
			SubscriptionType map[string]string `json:"subscriptionType"` // Root Subscription type
			Types            []SchemaType      `json:"types"`            // All schema types
			Directives       []interface{}     `json:"directives"`       // Schema directives
		} `json:"__schema"`
	} `json:"data"`
	Errors []map[string]interface{} `json:"errors,omitempty"` // Any errors from introspection
}

// Manager handles schema collection and merging from multiple services
type Manager struct {
	Services     []Service                  // Configured services
	SchemaCache  map[string]*SchemaResponse // Cache of service schemas
	MergedSchema *SchemaResponse            // Combined schema for introspection
	Routes       map[string]string          // Map of operations to service URLs
	RouteLock    sync.RWMutex               // Lock for thread safety
}

// NewManager creates a new schema manager
func NewManager(services []Service) *Manager {
	return &Manager{
		Services:     services,
		SchemaCache:  make(map[string]*SchemaResponse),
		Routes:       make(map[string]string),
		MergedSchema: &SchemaResponse{},
	}
}

// Initialize collects schemas from all services and merges them
func (m *Manager) Initialize() {
	log.Println("Initializing schema manager")

	// Initialize empty schema
	m.MergedSchema.Data.Schema.Types = []SchemaType{}
	m.MergedSchema.Data.Schema.Directives = []interface{}{}
	m.MergedSchema.Data.Schema.QueryType = map[string]string{"name": "Query"}

	// Collect schemas from all services
	successCount := 0
	for _, service := range m.Services {
		if m.CollectSchema(service.Name, service.URL) {
			successCount++
		}
	}

	// Merge schemas if we have any successful collections
	if successCount > 0 {
		m.MergeSchemas()
		log.Printf("Successfully initialized schema manager with %d services", successCount)
	} else {
		log.Println("Warning: Could not collect schema from any service")
	}
}

// CollectSchema fetches schema from a service and updates routing map
func (m *Manager) CollectSchema(name, url string) bool {
	log.Printf("Collecting schema from %s at %s", name, url)

	// Enhanced introspection query to handle deeper type nesting
	query := `{
		"query": "query IntrospectionQuery { __schema { queryType { name } mutationType { name } subscriptionType { name } types { ...FullType } directives { name description locations args { ...InputValue } } } } fragment FullType on __Type { kind name description fields(includeDeprecated: true) { name description args { ...InputValue } type { ...TypeRef } isDeprecated deprecationReason } inputFields { ...InputValue } interfaces { ...TypeRef } enumValues(includeDeprecated: true) { name description isDeprecated deprecationReason } possibleTypes { ...TypeRef } } fragment InputValue on __InputValue { name description type { ...TypeRef } defaultValue } fragment TypeRef on __Type { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } } } }"
	}`

	// Send request with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(query)))
	if err != nil {
		log.Printf("Failed to connect to %s: %v", name, err)
		return false
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Service %s returned status code %d", name, resp.StatusCode)
		return false
	}

	// Read and parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response from %s: %v", name, err)
		return false
	}

	// Log raw response for debugging
	if len(respBody) < 1000 {
		log.Printf("Raw response from %s: %s", name, string(respBody))
	} else {
		log.Printf("Raw response from %s: %s... (truncated)", name, string(respBody[:1000]))
	}

	var schemaResp SchemaResponse
	if err := json.Unmarshal(respBody, &schemaResp); err != nil {
		log.Printf("Failed to parse schema from %s: %v", name, err)
		return false
	}

	// Validate schema data
	if len(schemaResp.Data.Schema.Types) == 0 {
		log.Printf("Service %s returned empty schema", name)
		return false
	}

	// Cache schema and extract routes
	log.Printf("Successfully received schema from %s with %d types", name, len(schemaResp.Data.Schema.Types))
	m.SchemaCache[name] = &schemaResp
	m.UpdateRoutes(url, &schemaResp)

	return true
}

// UpdateRoutes maps GraphQL operations to their service URLs
func (m *Manager) UpdateRoutes(url string, schema *SchemaResponse) {
	m.RouteLock.Lock()
	defer m.RouteLock.Unlock()

	// Find Query, Mutation, and Subscription types
	for _, typeObj := range schema.Data.Schema.Types {
		if typeObj.Name == "Query" || typeObj.Name == "Mutation" || typeObj.Name == "Subscription" {
			// Extract fields as operation names
			if fields, ok := typeObj.Fields.([]interface{}); ok {
				for _, field := range fields {
					if fieldObj, ok := field.(map[string]interface{}); ok {
						if fieldName, ok := fieldObj["name"].(string); ok {
							m.Routes[fieldName] = url
						}
					}
				}
			}
		}
	}

	log.Printf("Mapped %d operations to service %s", len(m.Routes), url)
}

// MergeSchemas combines all service schemas into a single schema
func (m *Manager) MergeSchemas() {
	log.Println("Merging schemas from all services")

	// Reset merged schema
	m.MergedSchema = &SchemaResponse{}
	m.MergedSchema.Data.Schema.Types = []SchemaType{}
	m.MergedSchema.Data.Schema.Directives = []interface{}{}

	// Track Query and Mutation fields
	queryFields := []interface{}{}
	mutationFields := []interface{}{}

	// Track unique types and added directives
	typeMap := make(map[string]bool)
	directiveNames := make(map[string]bool) // To track directive names

	// Track query and mutation counts per service
	serviceStats := make(map[string]struct {
		QueryCount    int
		MutationCount int
	})

	// Process each service's schema
	for name, schema := range m.SchemaCache {
		log.Printf("Processing schema from %s", name)

		// Initialize stats for this service
		serviceStats[name] = struct {
			QueryCount    int
			MutationCount int
		}{}

		// Set root operation types if not already set
		if m.MergedSchema.Data.Schema.QueryType == nil && schema.Data.Schema.QueryType != nil {
			m.MergedSchema.Data.Schema.QueryType = schema.Data.Schema.QueryType
		}

		if m.MergedSchema.Data.Schema.MutationType == nil && schema.Data.Schema.MutationType != nil {
			m.MergedSchema.Data.Schema.MutationType = schema.Data.Schema.MutationType
		}

		if m.MergedSchema.Data.Schema.SubscriptionType == nil && schema.Data.Schema.SubscriptionType != nil {
			m.MergedSchema.Data.Schema.SubscriptionType = schema.Data.Schema.SubscriptionType
		}

		// Process each type in the schema
		for _, typeObj := range schema.Data.Schema.Types {
			// Collect operation fields separately
			if typeObj.Name == "Query" {
				if fields, ok := typeObj.Fields.([]interface{}); ok {
					// Update service stats
					stats := serviceStats[name]
					stats.QueryCount = len(fields)
					serviceStats[name] = stats

					// Add fields to the collection
					queryFields = append(queryFields, fields...)
					log.Printf("Service %s provides %d queries", name, len(fields))
				}
				continue
			}

			if typeObj.Name == "Mutation" {
				if fields, ok := typeObj.Fields.([]interface{}); ok {
					// Update service stats
					stats := serviceStats[name]
					stats.MutationCount = len(fields)
					serviceStats[name] = stats

					// Add fields to the collection
					mutationFields = append(mutationFields, fields...)
					log.Printf("Service %s provides %d mutations", name, len(fields))
				}
				continue
			}

			// Skip other root operation types
			if typeObj.Name == "Subscription" {
				continue
			}

			// Add other types (avoiding duplicates)
			if !typeMap[typeObj.Name] {
				m.MergedSchema.Data.Schema.Types = append(m.MergedSchema.Data.Schema.Types, typeObj)
				typeMap[typeObj.Name] = true
			}
		}

		// Add directives (avoiding duplicates)
		for _, directive := range schema.Data.Schema.Directives {
			if directiveObj, ok := directive.(map[string]interface{}); ok {
				if name, ok := directiveObj["name"].(string); ok {
					// Only add if we haven't seen this directive name before
					if !directiveNames[name] {
						m.MergedSchema.Data.Schema.Directives = append(m.MergedSchema.Data.Schema.Directives, directive)
						directiveNames[name] = true
					}
				}
			}
		}
	}

	// Print summary of operations per service
	fmt.Println("\n📊 Operations Per Service:")
	for name, stats := range serviceStats {
		fmt.Printf("  • %s: %d queries, %d mutations\n", name, stats.QueryCount, stats.MutationCount)
	}
	fmt.Printf("  • Total: %d queries, %d mutations\n", len(queryFields), len(mutationFields))

	// Add Query type with all fields
	if len(queryFields) > 0 || m.MergedSchema.Data.Schema.QueryType != nil {
		m.MergedSchema.Data.Schema.Types = append(m.MergedSchema.Data.Schema.Types, SchemaType{
			Kind:       "OBJECT",
			Name:       "Query",
			Fields:     queryFields,
			Interfaces: []interface{}{}, // Required for GraphQL introspection
		})
	}

	// Add Mutation type with all fields
	if len(mutationFields) > 0 || m.MergedSchema.Data.Schema.MutationType != nil {
		m.MergedSchema.Data.Schema.Types = append(m.MergedSchema.Data.Schema.Types, SchemaType{
			Kind:       "OBJECT",
			Name:       "Mutation",
			Fields:     mutationFields,
			Interfaces: []interface{}{}, // Required for GraphQL introspection
		})
	}

	// Ensure we have at least an empty Query type
	if len(m.MergedSchema.Data.Schema.Types) == 0 {
		m.MergedSchema.Data.Schema.Types = append(m.MergedSchema.Data.Schema.Types, SchemaType{
			Kind:   "OBJECT",
			Name:   "Query",
			Fields: []interface{}{},
		})
		m.MergedSchema.Data.Schema.QueryType = map[string]string{"name": "Query"}
	}

	log.Printf("Merged schema created with %d types", len(m.MergedSchema.Data.Schema.Types))
}

// GetRouteForOperation finds the service URL for a GraphQL operation
func (m *Manager) GetRouteForOperation(rootField string) string {
	m.RouteLock.RLock()
	defer m.RouteLock.RUnlock()

	return m.Routes[rootField] // Returns "" if not found
}

// GetMergedSchema returns the merged schema for introspection queries
func (m *Manager) GetMergedSchema() interface{} {
	// Return a valid response even if schema is empty
	if m.MergedSchema == nil || len(m.MergedSchema.Data.Schema.Types) == 0 {
		return map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType": map[string]string{"name": "Query"},
					"types": []interface{}{
						map[string]interface{}{
							"kind":   "OBJECT",
							"name":   "Query",
							"fields": []interface{}{},
						},
					},
					"directives": []interface{}{},
				},
			},
		}
	}

	// Process all types to ensure proper structure for GraphQL clients
	for i, typeObj := range m.MergedSchema.Data.Schema.Types {
		// Fix Query and Mutation operation types
		if typeObj.Name == "Query" || typeObj.Name == "Mutation" {
			// Validate and enhance field structure
			if fields, ok := typeObj.Fields.([]interface{}); ok {
				// Ensure each field has the required properties
				for j, field := range fields {
					if fieldObj, ok := field.(map[string]interface{}); ok {
						// Ensure field has required properties for Apollo Sandbox
						if _, ok := fieldObj["name"]; !ok {
							log.Printf("Warning: Field without name found in %s", typeObj.Name)
							continue
						}

						// Ensure type information is present
						if _, ok := fieldObj["type"]; !ok {
							// Add default type if missing
							fieldObj["type"] = map[string]interface{}{
								"kind": "SCALAR",
								"name": "String",
							}
						}

						// Ensure args exists (even if empty)
						if _, ok := fieldObj["args"]; !ok {
							fieldObj["args"] = []interface{}{}
						}

						// Update the field in our array
						fields[j] = fieldObj
					}
				}

				// Update the type in our schema
				m.MergedSchema.Data.Schema.Types[i].Fields = fields
			}
		}

		// Fix INPUT_OBJECT types - add empty inputFields if missing
		// This fixes the "Introspection result missing inputFields" error
		if typeObj.Kind == "INPUT_OBJECT" {
			if typeObj.InputFields == nil {
				m.MergedSchema.Data.Schema.Types[i].InputFields = []interface{}{}
			}
		}

		// Fix OBJECT types - ensure fields is always present
		if typeObj.Kind == "OBJECT" && typeObj.Fields == nil {
			m.MergedSchema.Data.Schema.Types[i].Fields = []interface{}{}
		}

		// Fix INTERFACE types - ensure fields is always present
		if typeObj.Kind == "INTERFACE" && typeObj.Fields == nil {
			m.MergedSchema.Data.Schema.Types[i].Fields = []interface{}{}
		}

		// Fix ENUM types - ensure enumValues is always present
		if typeObj.Kind == "ENUM" && typeObj.EnumValues == nil {
			m.MergedSchema.Data.Schema.Types[i].EnumValues = []interface{}{}
		}

		// Fix OBJECT types - ensure interfaces property is always present
		// This fixes the "Introspection result missing interfaces" error
		if typeObj.Kind == "OBJECT" && typeObj.Interfaces == nil {
			m.MergedSchema.Data.Schema.Types[i].Interfaces = []interface{}{}
		}

		// Fix INTERFACE types - ensure possibleTypes property is always present
		// This fixes the "Introspection result missing interfaces" error for INTERFACE types like Node
		if typeObj.Kind == "INTERFACE" && typeObj.PossibleTypes == nil {
			m.MergedSchema.Data.Schema.Types[i].PossibleTypes = []interface{}{}

			// Log the fix for specific interfaces we're targeting
			if typeObj.Name == "Node" {
				log.Printf("Fixed 'Node' interface by adding possibleTypes property")
			}
		}
	}

	// Fix directives to ensure they all have 'args' property
	for i, directive := range m.MergedSchema.Data.Schema.Directives {
		if directiveObj, ok := directive.(map[string]interface{}); ok {
			// Ensure args exists for each directive
			if _, ok := directiveObj["args"]; !ok {
				directiveObj["args"] = []interface{}{}
				m.MergedSchema.Data.Schema.Directives[i] = directiveObj
			}

			// Log the fixed directive
			if name, ok := directiveObj["name"].(string); ok && name == "authenticated" {
				log.Printf("Fixed 'authenticated' directive: %+v", directiveObj)
			}
		}
	}

	// Debug output - critical operations
	fmt.Println("\n🔍 Critical Types in Schema:")
	for _, typeObj := range m.MergedSchema.Data.Schema.Types {
		if typeObj.Name == "Query" || typeObj.Name == "Mutation" {
			fieldCount := 0
			if fields, ok := typeObj.Fields.([]interface{}); ok {
				fieldCount = len(fields)

				// Show a sample of fields
				fmt.Printf("  • %s has %d operations\n", typeObj.Name, fieldCount)
				if fieldCount > 0 {
					maxShow := 3
					if fieldCount < maxShow {
						maxShow = fieldCount
					}

					fmt.Println("    Sample operations:")
					for i := 0; i < maxShow; i++ {
						if fieldObj, ok := fields[i].(map[string]interface{}); ok {
							if name, ok := fieldObj["name"].(string); ok {
								fmt.Printf("      - %s\n", name)
							}
						}
					}
				}
			}
		}
	}

	// Return the merged schema in proper format for GraphQL clients
	return map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType":        m.MergedSchema.Data.Schema.QueryType,
				"mutationType":     m.MergedSchema.Data.Schema.MutationType,
				"subscriptionType": m.MergedSchema.Data.Schema.SubscriptionType,
				"types":            m.MergedSchema.Data.Schema.Types,
				"directives":       m.MergedSchema.Data.Schema.Directives,
			},
		},
	}
}

// Debug returns a string with information about the schema manager state
func (m *Manager) Debug() string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Services Configured: %d\n", len(m.Services)))
	result.WriteString(fmt.Sprintf("Services Connected: %d\n", len(m.SchemaCache)))
	result.WriteString(fmt.Sprintf("Operations Mapped: %d\n", len(m.Routes)))

	// Show service connection status
	result.WriteString("\nService Status:\n")
	for _, svc := range m.Services {
		if _, ok := m.SchemaCache[svc.Name]; ok {
			result.WriteString(fmt.Sprintf("  ✓ %s (%s): Connected\n", svc.Name, svc.URL))
		} else {
			result.WriteString(fmt.Sprintf("  ✗ %s (%s): Failed\n", svc.Name, svc.URL))
		}
	}

	// Show schema information
	if m.MergedSchema != nil && m.MergedSchema.Data.Schema.Types != nil {
		typeCount := len(m.MergedSchema.Data.Schema.Types)
		result.WriteString(fmt.Sprintf("\nMerged Schema:\n"))
		result.WriteString(fmt.Sprintf("  Total Types: %d\n", typeCount))

		// Count types by kind
		kindCounts := make(map[string]int)
		for _, t := range m.MergedSchema.Data.Schema.Types {
			kindCounts[t.Kind]++
		}

		for kind, count := range kindCounts {
			result.WriteString(fmt.Sprintf("  %s Types: %d\n", kind, count))
		}

		// Show root types
		if m.MergedSchema.Data.Schema.QueryType != nil {
			result.WriteString(fmt.Sprintf("  Query Type: %s\n", m.MergedSchema.Data.Schema.QueryType["name"]))
		}
		if m.MergedSchema.Data.Schema.MutationType != nil {
			result.WriteString(fmt.Sprintf("  Mutation Type: %s\n", m.MergedSchema.Data.Schema.MutationType["name"]))
		}
		if m.MergedSchema.Data.Schema.SubscriptionType != nil {
			result.WriteString(fmt.Sprintf("  Subscription Type: %s\n", m.MergedSchema.Data.Schema.SubscriptionType["name"]))
		}
	}

	return result.String()
}

// GetRoutesMap returns a copy of the routes mapping
func (m *Manager) GetRoutesMap() map[string]string {
	m.RouteLock.RLock()
	defer m.RouteLock.RUnlock()

	// Create a copy of the routes map to prevent concurrent access issues
	routesCopy := make(map[string]string, len(m.Routes))
	for k, v := range m.Routes {
		routesCopy[k] = v
	}

	return routesCopy
}
