package server

import "testing"

func TestNewOpenAPISpec(t *testing.T) {
	spec, err := NewOpenAPISpec()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.OpenAPI == "" || spec.Info == nil {
		t.Fatalf("spec not initialized")
	}
}
