package merger

import (
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestDefaultServers(t *testing.T) {
	servers := DefaultServers()

	if len(servers) != 4 {
		t.Errorf("Expected 4 default servers, got %d", len(servers))
	}

	expectedURLs := []string{
		"https://api-dev.domain.com",
		"https://api-test.domain.com",
		"https://api-stg.domain.com",
		"https://api.domain.com",
	}

	for i, server := range servers {
		if server.URL != expectedURLs[i] {
			t.Errorf("Expected URL %s, got %s", expectedURLs[i], server.URL)
		}
	}
}

func TestNewMerger(t *testing.T) {
	config := Config{
		InputPaths: []string{"https://github.com/4runfit/activity-service/blob/dev/openapi.yaml"},
		OutputPath: "output.yaml",
	}

	merger := New(config)

	if merger.config.OutputPath != "output.yaml" {
		t.Errorf("Expected output path 'output.yaml', got '%s'", merger.config.OutputPath)
	}

	if len(merger.config.Servers) != 4 {
		t.Errorf("Expected 4 default servers, got %d", len(merger.config.Servers))
	}
}

func TestNewMergerWithCustomServers(t *testing.T) {
	customServers := []Server{
		{URL: "https://custom1.com", Description: "Custom 1"},
		{URL: "https://custom2.com", Description: "Custom 2"},
	}

	config := Config{
		InputPaths: []string{"test1.yaml"},
		OutputPath: "output.yaml",
		Servers:    customServers,
	}

	merger := New(config)

	if len(merger.config.Servers) != 2 {
		t.Errorf("Expected 2 custom servers, got %d", len(merger.config.Servers))
	}

	if merger.config.Servers[0].URL != "https://custom1.com" {
		t.Errorf("Expected URL 'https://custom1.com', got '%s'", merger.config.Servers[0].URL)
	}
}

func TestDetectSwaggerVersion(t *testing.T) {
	// Test Swagger 2.0 YAML
	swagger2YAML := []byte(`swagger: "2.0"
info:
  title: Test API
  version: 1.0.0`)

	merger := &Merger{}
	version, err := merger.detectSwaggerVersion(swagger2YAML)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if version.Version != "2.0" {
		t.Errorf("Expected version '2.0', got '%s'", version.Version)
	}

	if !version.IsYAML {
		t.Errorf("Expected YAML format, got JSON")
	}

	// Test OpenAPI 3.0 YAML
	openapi3YAML := []byte(`openapi: "3.0.1"
info:
  title: Test API
  version: 1.0.0`)

	version, err = merger.detectSwaggerVersion(openapi3YAML)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if version.Version != "3.0.1" {
		t.Errorf("Expected version '3.0.1', got '%s'", version.Version)
	}
}

func TestMergeValidation(t *testing.T) {
	// Test empty input paths
	config := Config{
		InputPaths: []string{},
		OutputPath: "output.yaml",
	}

	merger := New(config)
	err := merger.Merge()

	if err == nil {
		t.Error("Expected error for empty input paths")
	}

	// Test empty output path
	config = Config{
		InputPaths: []string{"test.yaml"},
		OutputPath: "",
	}

	merger = New(config)
	err = merger.Merge()

	if err == nil {
		t.Error("Expected error for empty output path")
	}
}

func TestGetStatsValidation(t *testing.T) {
	// Test empty input paths
	config := Config{
		InputPaths: []string{},
		OutputPath: "output.yaml",
	}

	merger := New(config)
	_, err := merger.GetStats()

	if err == nil {
		t.Error("Expected error for empty input paths")
	}
}

// Helper function to create temporary test files
func createTempSwaggerFile(content string) (string, error) {
	tmpfile, err := os.CreateTemp("", "swagger_test_*.yaml")
	if err != nil {
		return "", err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return "", err
	}

	tmpfile.Close()
	return tmpfile.Name(), nil
}

func TestMergeOpenAPI3(t *testing.T) {
	merger := &Merger{}

	// Test empty docs
	_, err := merger.mergeOpenAPI3([]*openapi3.T{})
	if err == nil {
		t.Error("Expected error for empty documents")
	}

	// Test single doc
	doc1 := &openapi3.T{
		OpenAPI: "3.0.1",
		Info: &openapi3.Info{
			Title:   "API 1",
			Version: "1.0.0",
		},
	}

	merged, err := merger.mergeOpenAPI3([]*openapi3.T{doc1})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if merged.Info.Title != "API 1" {
		t.Errorf("Expected title 'API 1', got '%s'", merged.Info.Title)
	}
}
