package server

import "testing"

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
