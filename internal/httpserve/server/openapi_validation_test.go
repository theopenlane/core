package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/core/internal/httpserve/server"
)

// TestOpenAPISpecValidation builds the full OpenAPI document the same way spec
// generation does, with all routes registered, and validates it for compliance
func TestOpenAPISpecValidation(t *testing.T) {
	spec, err := server.GenerateOpenAPISpecDocument()
	require.NoError(t, err, "failed to generate OpenAPI spec")

	// Convert to JSON bytes for openapi3 loader (same as the published artifact)
	specBytes, err := json.Marshal(spec)
	require.NoError(t, err, "failed to marshal spec for validation")

	t.Logf("OpenAPI spec generated successfully (%d bytes)", len(specBytes))

	// Load and validate the OpenAPI specification
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(specBytes)
	require.NoError(t, err, "OpenAPI specification MUST load without errors using kin-openapi loader")

	// Validate the specification against OpenAPI 3.0 standards
	ctx := context.Background()
	err = doc.Validate(ctx)
	require.NoError(t, err, "OpenAPI specification MUST pass full kin-openapi validation")

	// Log basic stats
	pathCount := len(doc.Paths.Map())
	t.Logf("OpenAPI spec validated successfully: %d paths, version %s", pathCount, doc.OpenAPI)
}

func TestEchoToOpenAPIConversion(t *testing.T) {
	// Create a spec-mode router and register routes to test actual conversion
	router, err := server.NewSpecRouter(server.LogConfig{})
	require.NoError(t, err)

	// Mock handler to avoid panics
	router.Handler = &handlers.Handler{}

	// Register routes (this will apply our conversion logic)
	err = route.RegisterRoutes(router)
	require.NoError(t, err)

	// Check that our actual registered paths match expectations
	registeredPaths := make(map[string]bool)
	for path := range router.OAS.Paths.Map() {
		registeredPaths[path] = true
	}

	// Verify that known paths with parameters are properly converted
	expectedPaths := []string{
		"/v1/account/roles/organization/{id}",
		"/v1/account/features/{id}",
	}

	for _, expectedPath := range expectedPaths {
		assert.True(t, registeredPaths[expectedPath],
			"Expected path '%s' should be registered in OpenAPI spec", expectedPath)
	}

	// Log some example paths for verification
	parametricPaths := []string{}
	for path := range registeredPaths {
		if strings.Contains(path, "{") {
			parametricPaths = append(parametricPaths, path)
		}
	}

	t.Logf("Found %d paths with parameters:", len(parametricPaths))
	for _, path := range parametricPaths {
		t.Logf("  - %s", path)
	}

	// Ensure no Echo-style paths are registered
	for path := range registeredPaths {
		assert.NotContains(t, path, ":",
			"Path '%s' contains Echo-style :param syntax in OpenAPI spec", path)
	}

	t.Logf("All path parameters properly converted from :param to {param}")
}
