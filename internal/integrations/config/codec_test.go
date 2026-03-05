package config

import (
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"

	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

type decodeHookTarget struct {
	Duration time.Duration                  `json:"duration"`
	Label    integrationtypes.TrimmedString `json:"label"`
}

func TestDefaultMapstructureDecodeHook(t *testing.T) {
	target := decodeHookTarget{}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &target,
		TagName:          "json",
		WeaklyTypedInput: true,
		DecodeHook:       DefaultMapstructureDecodeHook(),
	})
	if err != nil {
		t.Fatalf("unexpected decoder construction error: %v", err)
	}

	if err := decoder.Decode(map[string]any{
		"duration": "2m",
		"label":    "  hello  ",
	}); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if target.Duration != 2*time.Minute {
		t.Fatalf("expected duration decode hook, got %v", target.Duration)
	}
	if target.Label.String() != "hello" {
		t.Fatalf("expected text unmarshal hook, got %q", target.Label.String())
	}
}

type testOverlayConfig struct {
	Name string         `json:"name,omitempty"`
	Meta map[string]any `json:"meta,omitempty"`
	Flag bool           `json:"flag,omitempty"`
}

func TestJSONValueDeepMergeWithPruneZero(t *testing.T) {
	base := map[string]any{
		"name": "base",
		"meta": map[string]any{
			"region": "us-east-1",
			"tier":   "gold",
		},
	}

	out, err := JSONValue(base, testOverlayConfig{
		Meta: map[string]any{
			"tier": "platinum",
		},
	}, MapOptions{
		PruneZero: true,
		DeepMerge: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out["name"] != "base" {
		t.Fatalf("expected base name to remain, got %v", out["name"])
	}

	meta, ok := out["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected merged nested meta map, got %#v", out["meta"])
	}
	if meta["region"] != "us-east-1" || meta["tier"] != "platinum" {
		t.Fatalf("expected nested deep merge, got %#v", meta)
	}
}

func TestJSONValueShallowMerge(t *testing.T) {
	base := map[string]any{
		"meta": map[string]any{
			"region": "us-east-1",
		},
	}

	out, err := JSONValue(base, map[string]any{
		"meta": map[string]any{
			"tier": "gold",
		},
	}, MapOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta, ok := out["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta map, got %#v", out["meta"])
	}
	if _, exists := meta["region"]; exists {
		t.Fatalf("expected shallow merge to replace nested map, got %#v", meta)
	}
	if meta["tier"] != "gold" {
		t.Fatalf("expected shallow merge value, got %#v", meta)
	}
}
