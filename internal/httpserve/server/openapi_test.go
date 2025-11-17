package server

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/theopenlane/core/internal/httpserve/specs"
)

func TestNewOpenAPISpec(t *testing.T) {
	t.Parallel()

	spec, err := NewOpenAPISpec()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.OpenAPI == "" || spec.Info == nil {
		t.Fatalf("spec not initialized")
	}
}

func TestGetBasePathFromServers(t *testing.T) {
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(specs.SCIMSpec)
	if err != nil {
		t.Fatalf("failed to load SCIM spec: %v", err)
	}

	basePath := getBasePathFromServers(spec.Servers)
	if basePath != "/scim/v2" {
		t.Fatalf("expected base path /scim/v2, got %q", basePath)
	}
}

func TestApplyBasePath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		path     string
		basePath string
		expected string
	}{
		{
			name:     "with leading slash",
			path:     "/Users",
			basePath: "/scim/v2",
			expected: "/scim/v2/Users",
		},
		{
			name:     "without leading slash",
			path:     "Users",
			basePath: "/scim/v2",
			expected: "/scim/v2/Users",
		},
		{
			name:     "base path root",
			path:     "/Users",
			basePath: "/scim",
			expected: "/scim/Users",
		},
		{
			name:     "empty base path",
			path:     "/Users",
			basePath: "",
			expected: "/Users",
		},
		{
			name:     "slash termination",
			path:     "/",
			basePath: "/scim/v2",
			expected: "/scim/v2",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := applyBasePath(tc.path, tc.basePath)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestMergeSCIMSpecPrefixesPaths(t *testing.T) {
	t.Parallel()

	spec, err := NewOpenAPISpec()
	if err != nil {
		t.Fatalf("failed to build spec: %v", err)
	}

	if _, ok := spec.Paths.Map()["/scim/v2/Users"]; !ok {
		t.Fatalf("expected /scim/v2/Users path to exist")
	}

	if _, ok := spec.Paths.Map()["/Users"]; ok {
		t.Fatalf("unexpected root-level /Users path found")
	}
}
