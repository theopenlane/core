package operations

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/types"
)

type normalizedConfig struct {
	// Name holds the trimmed name value
	Name types.TrimmedString `json:"name"`
	// Label holds the lower-cased label value
	Label types.LowerString `json:"label"`
	// Code holds the upper-cased code value
	Code types.UpperString `json:"code"`
	// Tags lists trimmed tag values
	Tags []types.TrimmedString `json:"tags"`
	// Modes lists lower-cased mode values
	Modes []types.LowerString `json:"modes"`
	// Flags lists upper-cased flag values
	Flags []types.UpperString `json:"flags"`
}

// TestDecodeConfigNormalizedStrings verifies normalized strings are decoded correctly
func TestDecodeConfigNormalizedStrings(t *testing.T) {
	config := map[string]any{
		"name":  "  Alice ",
		"label": "  FooBar ",
		"code":  "  ab-12 ",
		"tags":  []string{" one ", "two"},
		"modes": []string{"  ON ", " Off "},
		"flags": []string{" aa ", "BB "},
	}

	var decoded normalizedConfig
	require.NoError(t, DecodeConfig(config, &decoded))

	require.Equal(t, types.TrimmedString("Alice"), decoded.Name)
	require.Equal(t, types.LowerString("foobar"), decoded.Label)
	require.Equal(t, types.UpperString("AB-12"), decoded.Code)
	require.Equal(t, []types.TrimmedString{"one", "two"}, decoded.Tags)
	require.Equal(t, []types.LowerString{"on", "off"}, decoded.Modes)
	require.Equal(t, []types.UpperString{"AA", "BB"}, decoded.Flags)
}

func TestDecodeConfigUnknownField(t *testing.T) {
	type sample struct {
		Name string `json:"name"`
	}

	var decoded sample
	err := DecodeConfig(map[string]any{
		"name":  "ok",
		"extra": "nope",
	}, &decoded)
	if err == nil {
		t.Fatalf("expected error for unknown field")
	}
}
