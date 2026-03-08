package operations

import (
	"encoding/json"
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
				Config:         json.RawMessage(`{"a":1}`),
				AllowOverrides: []string{" b ", "b"},
			},
		},
	}

	template, ok := OperationTemplateFor(&cfg, "op")
	if !ok {
		t.Fatalf("expected template to be found")
	}

	var configMap map[string]any
	if err := json.Unmarshal(template.Config, &configMap); err != nil {
		t.Fatalf("expected config to unmarshal: %v", err)
	}
	if configMap["a"] != float64(1) {
		t.Fatalf("expected config value")
	}

	if _, ok := template.AllowedOverrides["b"]; !ok {
		t.Fatalf("expected override to be allowed")
	}
	if _, ok := template.AllowedOverrides[" b "]; !ok {
		t.Fatalf("expected strict override key to be preserved")
	}
}

func TestApplyOperationTemplate(t *testing.T) {
	template := OperationTemplate{
		Config:           json.RawMessage(`{"a":1}`),
		AllowedOverrides: map[string]struct{}{"b": {}},
	}

	out, err := ApplyOperationTemplate(template, json.RawMessage(`{"b":2}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var outMap map[string]any
	if err := json.Unmarshal(out, &outMap); err != nil {
		t.Fatalf("expected output to unmarshal: %v", err)
	}
	if outMap["a"] != float64(1) || outMap["b"] != float64(2) {
		t.Fatalf("expected merged config, got %v", outMap)
	}

	if _, err := ApplyOperationTemplate(template, json.RawMessage(`{"c":3}`)); err != ErrOperationTemplateOverrideNotAllowed {
		t.Fatalf("expected ErrOperationTemplateOverrideNotAllowed, got %v", err)
	}

	template.AllowedOverrides = nil
	if _, err := ApplyOperationTemplate(template, json.RawMessage(`{"b":2}`)); err != ErrOperationTemplateOverridesNotAllowed {
		t.Fatalf("expected ErrOperationTemplateOverridesNotAllowed, got %v", err)
	}
}

func TestResolveOperationConfig(t *testing.T) {
	cfg := openapi.IntegrationConfig{
		OperationTemplates: map[string]openapi.IntegrationOperationTemplate{
			"op": {
				Config:         json.RawMessage(`{"a":1}`),
				AllowOverrides: []string{"b"},
			},
		},
	}

	out, err := ResolveOperationConfig(&cfg, "op", json.RawMessage(`{"b":2}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var outMap map[string]any
	if err := json.Unmarshal(out, &outMap); err != nil {
		t.Fatalf("expected output to unmarshal: %v", err)
	}
	if outMap["a"] != float64(1) || outMap["b"] != float64(2) {
		t.Fatalf("expected merged config, got %v", outMap)
	}

	out, err = ResolveOperationConfig(&cfg, "missing", json.RawMessage(`{"x":1}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := json.Unmarshal(out, &outMap); err != nil {
		t.Fatalf("expected output to unmarshal: %v", err)
	}
	if outMap["x"] != float64(1) {
		t.Fatalf("expected overrides to pass through, got %v", outMap)
	}
}
