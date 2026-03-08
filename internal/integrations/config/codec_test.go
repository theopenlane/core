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
