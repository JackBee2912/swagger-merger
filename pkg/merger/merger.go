
package merger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// SwaggerVersion represents the detected version of a swagger file
type SwaggerVersion struct {
	Version string
	IsYAML  bool
}

// Config holds configuration for the merger
type Config struct {
	InputPaths []string
	OutputPath string
	Servers    []Server
}

// Server represents an API server configuration
type Server struct {
	URL         string
	Description string
}

// DefaultServers returns default server configurations
func DefaultServers() []Server {
	return []Server{
		{URL: "https://api-dev.domain.com", Description: "Development Environment"},
		{URL: "https://api-test.domain.com", Description: "Test Environment"},
		{URL: "https://api-stg.domain.com", Description: "Staging Environment"},
		{URL: "https://api.domain.com", Description: "Production Environment"},
	}
}

// Merger handles swagger file merging operations
type Merger struct {
	config Config
}

// New creates a new Merger instance
func New(config Config) *Merger {
	if config.Servers == nil {
		config.Servers = DefaultServers()
	}
	return &Merger{config: config}
}

// detectSwaggerVersion detects if a file is Swagger 2.0 or OpenAPI 3.0
func (m *Merger) detectSwaggerVersion(data []byte) (*SwaggerVersion, error) {
	var obj map[string]interface{}

	// Try YAML first
	if err := yaml.Unmarshal(data, &obj); err == nil {
		if version, exists := obj["swagger"]; exists {
			return &SwaggerVersion{Version: fmt.Sprintf("%v", version), IsYAML: true}, nil
		}
		if version, exists := obj["openapi"]; exists {
			return &SwaggerVersion{Version: fmt.Sprintf("%v", version), IsYAML: true}, nil
		}
	}

	// Try JSON
	if err := json.Unmarshal(data, &obj); err == nil {
		if version, exists := obj["swagger"]; exists {
			return &SwaggerVersion{Version: fmt.Sprintf("%v", version), IsYAML: false}, nil
		}
		if version, exists := obj["openapi"]; exists {
			return &SwaggerVersion{Version: fmt.Sprintf("%v", version), IsYAML: false}, nil
		}
	}

	return nil, fmt.Errorf("unable to detect swagger/openapi version")
}

// convertToOpenAPI3 converts a swagger file to OpenAPI 3.0
func (m *Merger) convertToOpenAPI3(data []byte, version *SwaggerVersion) (*openapi3.T, error) {
	if strings.HasPrefix(version.Version, "3.") {
		// Already OpenAPI 3.0, just parse it
		loader := openapi3.NewLoader()
		return loader.LoadFromData(data)
	}

	// Convert from Swagger 2.0 to OpenAPI 3.0
	var jsonObj map[string]interface{}

	if version.IsYAML {
		// Convert YAML -> JSON
		if err := yaml.Unmarshal(data, &jsonObj); err != nil {
			return nil, fmt.Errorf("failed to parse YAML to map: %v", err)
		}
		jsonBytes, err := json.Marshal(jsonObj)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal YAML to JSON: %v", err)
		}
		data = jsonBytes
	}

	// Parse swagger2 (JSON)
	var swagger2Doc openapi2.T
	if err := swagger2Doc.UnmarshalJSON(data); err != nil {
		return nil, fmt.Errorf("failed to parse Swagger2 JSON: %v", err)
	}

	// Convert to OpenAPI 3.0
	openapi3Doc, err := openapi2conv.ToV3(&swagger2Doc)
	if err != nil {
		return nil, fmt.Errorf("convert to openapi 3.0 failed: %v", err)
	}

	return openapi3Doc, nil
}

// mergeOpenAPI3 merges multiple OpenAPI 3.0 documents
func (m *Merger) mergeOpenAPI3(docs []*openapi3.T) (*openapi3.T, error) {
	if len(docs) == 0 {
		return nil, fmt.Errorf("no documents to merge")
	}

	merged := docs[0]

	for i := 1; i < len(docs); i++ {
		doc := docs[i]

		// Merge paths
		if doc.Paths != nil {
			if merged.Paths == nil {
				merged.Paths = &openapi3.Paths{}
			}
			for path, item := range doc.Paths.Map() {
				merged.Paths.Set(path, item)
			}
		}

		// Initialize components if nil
		if merged.Components.Schemas == nil {
			merged.Components.Schemas = openapi3.Schemas{}
		}
		if merged.Components.Responses == nil {
			merged.Components.Responses = openapi3.ResponseBodies{}
		}
		if merged.Components.Parameters == nil {
			merged.Components.Parameters = openapi3.ParametersMap{}
		}
		if merged.Components.RequestBodies == nil {
			merged.Components.RequestBodies = openapi3.RequestBodies{}
		}
		if merged.Components.Headers == nil {
			merged.Components.Headers = openapi3.Headers{}
		}

		// Merge components
		if doc.Components.Schemas != nil {
			for k, v := range doc.Components.Schemas {
				merged.Components.Schemas[k] = v
			}
		}
		if doc.Components.Responses != nil {
			for k, v := range doc.Components.Responses {
				merged.Components.Responses[k] = v
			}
		}
		if doc.Components.Parameters != nil {
			for k, v := range doc.Components.Parameters {
				merged.Components.Parameters[k] = v
			}
		}
		if doc.Components.RequestBodies != nil {
			for k, v := range doc.Components.RequestBodies {
				merged.Components.RequestBodies[k] = v
			}
		}
		if doc.Components.Headers != nil {
			for k, v := range doc.Components.Headers {
				merged.Components.Headers[k] = v
			}
		}

		// Merge tags
		if doc.Tags != nil {
			merged.Tags = append(merged.Tags, doc.Tags...)
		}
	}

	return merged, nil
}

// readDataFromPath reads data from either a local file or URL
func (m *Merger) readDataFromPath(path string) ([]byte, error) {
	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Make HTTP request
		resp, err := client.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URL %s: %v", path, err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP request failed with status %d for URL %s", resp.StatusCode, path)
		}

		// Read response body
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body from %s: %v", path, err)
		}

		return data, nil
	}

	// Read local file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	return data, nil
}

// processSwaggerFile processes a single swagger file
func (m *Merger) processSwaggerFile(filePath string) (*openapi3.T, error) {
	// Read data from file or URL
	data, err := m.readDataFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filePath, err)
	}

	// Detect version
	version, err := m.detectSwaggerVersion(data)
	if err != nil {
		return nil, fmt.Errorf("failed to detect version for %s: %v", filePath, err)
	}

	// Convert to OpenAPI 3.0
	doc, err := m.convertToOpenAPI3(data, version)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s: %v", filePath, err)
	}

	// Set common properties
	doc.OpenAPI = "3.0.1"

	// Always override servers with configured servers
	servers := make(openapi3.Servers, len(m.config.Servers))
	for i, server := range m.config.Servers {
		servers[i] = &openapi3.Server{
			URL:         server.URL,
			Description: server.Description,
		}
	}
	doc.Servers = servers

	return doc, nil
}

// Merge merges all swagger files and writes the result to output file
func (m *Merger) Merge() error {
	if len(m.config.InputPaths) == 0 {
		return fmt.Errorf("no input paths provided")
	}

	if m.config.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	// Process each file
	var docs []*openapi3.T
	for _, filePath := range m.config.InputPaths {
		doc, err := m.processSwaggerFile(filePath)
		if err != nil {
			return fmt.Errorf("error processing %s: %v", filePath, err)
		}
		docs = append(docs, doc)
	}

	// Merge all documents
	merged, err := m.mergeOpenAPI3(docs)
	if err != nil {
		return fmt.Errorf("error merging documents: %v", err)
	}

	// Write output
	out, err := yaml.Marshal(merged)
	if err != nil {
		return fmt.Errorf("error marshaling to YAML: %v", err)
	}

	if err := os.WriteFile(m.config.OutputPath, out, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

// MergeFromDirectory merges all swagger files found in a directory
func (m *Merger) MergeFromDirectory(inputDir, pattern string) error {
	var swaggerFiles []string

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				swaggerFiles = append(swaggerFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error finding files: %v", err)
	}

	if len(swaggerFiles) == 0 {
		return fmt.Errorf("no swagger files found in %s with pattern %s", inputDir, pattern)
	}

	// Update config with found files
	m.config.InputPaths = swaggerFiles

	return m.Merge()
}

// GetStats returns statistics about the merged document
func (m *Merger) GetStats() (map[string]int, error) {
	if len(m.config.InputPaths) == 0 {
		return nil, fmt.Errorf("no input paths provided")
	}

	// Process each file
	var docs []*openapi3.T
	for _, filePath := range m.config.InputPaths {
		doc, err := m.processSwaggerFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error processing %s: %v", filePath, err)
		}
		docs = append(docs, doc)
	}

	// Merge all documents
	merged, err := m.mergeOpenAPI3(docs)
	if err != nil {
		return nil, fmt.Errorf("error merging documents: %v", err)
	}

	stats := map[string]int{
		"total_files":   len(m.config.InputPaths),
		"total_paths":   len(merged.Paths.Map()),
		"total_schemas": len(merged.Components.Schemas),
		"total_tags":    len(merged.Tags),
	}

	return stats, nil
}
