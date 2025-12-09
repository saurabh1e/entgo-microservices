package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// GraphQLRequest represents a client's GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`                   // GraphQL query/mutation string
	Variables     map[string]interface{} `json:"variables"`               // Query variables
	OperationName string                 `json:"operationName,omitempty"` // Optional operation name
}

// SchemaManager interface defines methods needed from the schema manager
type SchemaManager interface {
	// GetRouteForOperation returns the service URL for a given operation field name
	GetRouteForOperation(rootField string) string

	// GetMergedSchema returns the merged schema for introspection queries
	GetMergedSchema() interface{}
}

// Router handles GraphQL request routing to microservices
type Router struct {
	SchemaManager SchemaManager // Schema manager for routing and introspection
}

// NewRouter creates a new router with schema manager
func NewRouter(manager SchemaManager) *Router {
	return &Router{
		SchemaManager: manager,
	}
}

// HandleRequest processes incoming GraphQL requests
func (r *Router) HandleRequest(w http.ResponseWriter, req *http.Request) {
	// Set comprehensive CORS headers for external tools like Apollo Studio
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Cache-Control", "no-cache")

	// If this is a preflight OPTIONS request, handle it and return
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if content type is application/json
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		log.Printf("Invalid content type: %s", contentType)
		errorJSON, _ := json.Marshal(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Content-Type must be application/json"},
			},
		})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorJSON)
		return
	}

	// Parse the GraphQL request
	var graphQLReq GraphQLRequest
	if err := json.NewDecoder(req.Body).Decode(&graphQLReq); err != nil {
		log.Printf("Error parsing request: %v", err)
		errorJSON, _ := json.Marshal(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Invalid GraphQL request: " + err.Error()},
			},
		})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorJSON)
		return
	}

	// Handle introspection queries (schema information) with special care
	if isIntrospectionQuery(graphQLReq.Query) {
		log.Printf("Handling introspection query: %s", graphQLReq.OperationName)
		fmt.Println("📡 Introspection request received")

		// Get merged schema
		schema := r.SchemaManager.GetMergedSchema()

		// Write response
		if err := json.NewEncoder(w).Encode(schema); err != nil {
			log.Printf("Error encoding schema response: %v", err)
			errorJSON, _ := json.Marshal(map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": "Internal server error: Failed to encode response"},
				},
			})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errorJSON)
		}
		return
	}

	// Extract the root operation field
	rootField := extractRootField(graphQLReq.Query)
	if rootField == "" {
		log.Printf("Could not determine root operation field from query: %s", graphQLReq.Query)
		errorJSON, _ := json.Marshal(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Unable to parse GraphQL query - could not identify operation"},
			},
		})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorJSON)
		return
	}

	// Find which service should handle this operation
	serviceURL := r.SchemaManager.GetRouteForOperation(rootField)
	if serviceURL == "" {
		log.Printf("No service found for operation: %s", rootField)
		errorJSON, _ := json.Marshal(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": fmt.Sprintf("Operation '%s' not supported by any service", rootField)},
			},
		})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorJSON)
		return
	}

	// Forward the request to the appropriate service
	r.forwardRequest(w, req, graphQLReq, serviceURL)
}

// forwardRequest sends the GraphQL request to the target service
func (r *Router) forwardRequest(w http.ResponseWriter, originalReq *http.Request,
	graphQLReq GraphQLRequest, serviceURL string) {

	// Marshal request for forwarding
	requestBody, err := json.Marshal(graphQLReq)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create request to the target service
	req, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating service request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy relevant headers
	req.Header.Set("Content-Type", "application/json")
	copyHeaders(originalReq.Header, req.Header)

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding to service %s: %v", serviceURL, err)
		http.Error(w, "Service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Return the service response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	log.Printf("Forwarded request to %s, status: %d", serviceURL, resp.StatusCode)
}

// copyHeaders copies HTTP headers except Host
func copyHeaders(src, dst http.Header) {
	for key, values := range src {
		if key == "Host" {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// isIntrospectionQuery detects if a query is requesting schema information
func isIntrospectionQuery(query string) bool {
	// Match common introspection patterns
	return regexp.MustCompile(`\b__schema\b`).MatchString(query) ||
		regexp.MustCompile(`\b__type\b`).MatchString(query) ||
		regexp.MustCompile(`\bIntrospectionQuery\b`).MatchString(query)
}

// extractRootField gets the main operation field from a GraphQL query
func extractRootField(query string) string {
	// Match the first field inside the operation
	pattern := regexp.MustCompile(`(?i)^\s*(?:query|mutation|subscription)?\s*(?:\w+\s*)?(?:\([\s\S]*?\))?\s*\{?\s*(\w+)`)
	matches := pattern.FindStringSubmatch(query)

	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}

// ServePlayground serves the GraphQL Playground - an in-browser GraphQL IDE
func ServePlayground(w http.ResponseWriter, r *http.Request) {
	// Set headers for HTML content
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Use a template for the playground HTML
	playgroundTemplate := template.Must(template.New("playground").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>Syphoon GraphQL Playground</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@apollographql/graphql-playground-react@1.7.42/build/static/css/index.css" />
    <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/@apollographql/graphql-playground-react@1.7.42/build/favicon.png" />
    <script src="https://cdn.jsdelivr.net/npm/@apollographql/graphql-playground-react@1.7.42/build/static/js/middleware.js"></script>
</head>
<body>
    <style>
        body {
            background-color: rgb(23, 42, 58);
            font-family: Open Sans, sans-serif;
            height: 100vh;
            margin: 0;
            overflow: hidden;
        }
        #root {
            height: 100vh;
            width: 100%;
        }
    </style>
    <div id="root"></div>
    <script>
        window.addEventListener('load', function (event) {
            const root = document.getElementById('root');
            root.classList.add('playgroundIn');
            
            GraphQLPlayground.init(root, {
                endpoint: '/graphql',
                settings: {
                    'request.credentials': 'same-origin',
                    'schema.polling.enable': true,
                    'schema.polling.interval': 5000
                },
                tabs: [
                    {
                        name: 'Syphoon API',
                        endpoint: '/graphql',
                        query: '# Welcome to Syphoon GraphQL API\n# Try querying your API here\n\n{\n  __schema {\n    queryType {\n      name\n      fields {\n        name\n        description\n      }\n    }\n  }\n}'
                    }
                ]
            });
        });
    </script>
</body>
</html>
`))

	// Execute the template
	if err := playgroundTemplate.Execute(w, nil); err != nil {
		log.Printf("Error rendering playground: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleRESTRequest processes incoming REST API requests and forwards them to appropriate services
func (r *Router) HandleRESTRequest(w http.ResponseWriter, req *http.Request) {
	// Extract service name and path from the URL
	// Expected format: /api/v1/{service_name}/{path}
	path := req.URL.Path

	// Parse the URL to get service name and remaining path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid API path format", http.StatusBadRequest)
		log.Printf("Invalid API path: %s", path)
		return
	}

	// Extract service name (parts[3])
	serviceName := parts[3]

	// Get base URL for this service from the schema manager
	serviceBaseURL := r.getServiceBaseURL(serviceName)
	if serviceBaseURL == "" {
		http.Error(w, fmt.Sprintf("Unknown service: %s", serviceName), http.StatusNotFound)
		log.Printf("Unknown service requested: %s", serviceName)
		return
	}

	// Rebuild the path without the /api/v1/{service_name} prefix
	remainingPath := strings.Join(parts[4:], "/")

	// Extract the base URL without the GraphQL path
	serviceURLParts := strings.Split(serviceBaseURL, "/graphql")
	serviceHostURL := serviceURLParts[0]

	// Build the target URL
	targetURL := fmt.Sprintf("%s/%s", serviceHostURL, remainingPath)

	// Log the forwarding information
	log.Printf("Forwarding REST request to service %s: %s → %s", serviceName, path, targetURL)

	// Forward the request
	r.forwardRESTRequest(w, req, targetURL)
}

// getServiceBaseURL returns the base URL for a service by name
func (r *Router) getServiceBaseURL(serviceName string) string {
	// Simple example field to get a URL for each service
	// This assumes the first operation for each service is mapped correctly
	for field, url := range r.getRoutesMap() {
		if strings.HasPrefix(field, serviceName) ||
			(serviceName == "main" && !strings.HasPrefix(field, "auth") &&
				!strings.HasPrefix(field, "getgrass")) {
			return url
		}
	}

	// Special case mapping if no routes are found
	switch serviceName {
	case "main":
		return "http://localhost:8088/graphql"
	case "auth":
		return "http://localhost:8081/graphql"
	case "getgrass":
		return "http://localhost:8085/graphql"
	default:
		return ""
	}
}

// getRoutesMap retrieves all available routes from the schema manager
func (r *Router) getRoutesMap() map[string]string {
	// This method is needed because SchemaManager interface doesn't expose the routes map directly
	// We'll use reflection to access it - in a real-world scenario, this should be properly exposed

	// Try to extract the routes map using type assertion if the underlying implementation is compatible
	if adapter, ok := r.SchemaManager.(interface{ GetRoutes() map[string]string }); ok {
		return adapter.GetRoutes()
	}

	// Fallback to an empty map if not available
	return map[string]string{}
}

// forwardRESTRequest sends the REST request to the target service
func (r *Router) forwardRESTRequest(w http.ResponseWriter, originalReq *http.Request, targetURL string) {
	// Create a new request with the same method, URL, and body
	proxyReq, err := http.NewRequest(originalReq.Method, targetURL, originalReq.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy all headers from the original request
	for key, values := range originalReq.Header {
		if key != "Host" { // Skip the Host header
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Copy query parameters
	proxyReq.URL.RawQuery = originalReq.URL.RawQuery

	// Send the request to the target service
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Error forwarding request to %s: %v", targetURL, err)
		http.Error(w, "Service Unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy all headers from the response
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}

	log.Printf("REST request forwarded to %s, status: %d", targetURL, resp.StatusCode)
}
