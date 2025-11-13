package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mcuadros/go-defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
)

// TestOpenAPISpecValidation creates a full server with all routes registered,
// fetches the OpenAPI spec from /api-docs, and validates it for compliance
func TestOpenAPISpecValidation(t *testing.T) {
	// Create base configuration with defaults
	cfg := config.Config{}
	defaults.SetDefaults(&cfg)
	cfg.Server.Listen = "localhost:0"
	cfg.Server.MetricsPort = ":0"
	cfg.Server.CSRFProtection.Enabled = false // Disable CSRF for test simplicity
	cfg.ObjectStorage.Enabled = false
	cfg.DB.RunMigrations = false // Skip migrations for test
	cfg.DB.Debug = false

	so := &serveropts.ServerOptions{
		Config: serverconfig.Config{Settings: cfg},
	}

	// Create the server with full configuration
	srv, err := server.NewServer(so.Config)
	require.NoError(t, err, "failed to create server")

	// Apply middleware to echo instance
	for _, m := range so.Config.Handler.AdditionalMiddleware {
		if m != nil {
			srv.Router.Echo.Use(m)
		}
	}

	// CRITICAL: Set the handler on the router before registering routes
	// This is normally done in StartEchoServer, but we need it earlier for testing
	srv.Router.Handler = &so.Config.Handler

	// Register ALL routes - this is the key part that tests our actual route definitions
	err = route.RegisterRoutes(srv.Router)
	require.NoError(t, err, "failed to register routes")

	// Get the OpenAPI specification directly from the router (same data as /api-docs endpoint)
	// Note: We skip the HTTP request to avoid middleware dependencies in tests
	spec := srv.Router.OAS
	require.NotNil(t, spec, "OpenAPI spec should be generated")

	// Convert to JSON bytes for openapi3 loader (same as what /api-docs would return)
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

	// Create router and register routes to test actual conversion
	router, err := server.NewRouter(server.LogConfig{})
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
		"/.well-known/acme-challenge/{path}",
		// Note: /files/{name} is only registered when LocalFilePath is set
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
