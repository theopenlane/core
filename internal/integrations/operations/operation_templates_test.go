package operations

import (
	"testing"

	openapi "github.com/theopenlane/core/common/openapi"
)

func TestOperationTemplateFor(t *testing.T) {
	if _, ok := OperationTemplateFor(nil, "op"); ok {
		t.Fatalf("expected false for nil config")
	}

	cfg := openapi.IntegrationConfig{
		OperationTemplates: map[string]openapi.IntegrationOperationTemplate{
			"op": {
				Config:         map[string]any{"a": 1},
				AllowOverrides: []string{"b"},
			},
		},
	}

	template, ok := OperationTemplateFor(&cfg, "op")
	if !ok {
		t.Fatalf("expected template to be found")
	}
	if template.Config["a"] != 1 {
		t.Fatalf("expected config value")
	}
	if _, ok := template.AllowedOverrides["b"]; !ok {
		t.Fatalf("expected override to be allowed")
	}
}

func TestApplyOperationTemplate(t *testing.T) {
	template := OperationTemplate{
		Config:           map[string]any{"a": 1},
		AllowedOverrides: map[string]struct{}{"b": {}},
	}

	out, err := ApplyOperationTemplate(template, map[string]any{"b": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["a"] != 1 || out["b"] != 2 {
		t.Fatalf("expected merged config")
	}

	if _, err := ApplyOperationTemplate(template, map[string]any{"c": 3}); err != ErrOperationTemplateOverrideNotAllowed {
		t.Fatalf("expected ErrOperationTemplateOverrideNotAllowed, got %v", err)
	}

	template.AllowedOverrides = nil
	if _, err := ApplyOperationTemplate(template, map[string]any{"b": 2}); err != ErrOperationTemplateOverridesNotAllowed {
		t.Fatalf("expected ErrOperationTemplateOverridesNotAllowed, got %v", err)
	}
}

func TestResolveOperationConfig(t *testing.T) {
	cfg := openapi.IntegrationConfig{
		OperationTemplates: map[string]openapi.IntegrationOperationTemplate{
			"op": {
				Config:         map[string]any{"a": 1},
				AllowOverrides: []string{"b"},
			},
		},
	}

	out, err := ResolveOperationConfig(&cfg, "op", map[string]any{"b": 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["a"] != 1 || out["b"] != 2 {
		t.Fatalf("expected merged config")
	}

	out, err = ResolveOperationConfig(&cfg, "missing", map[string]any{"x": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["x"] != 1 {
		t.Fatalf("expected overrides to pass through")
	}
}
