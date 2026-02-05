package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type normalizedConfig struct {
	Name  TrimmedString   `mapstructure:"name"`
	Label LowerString     `mapstructure:"label"`
	Code  UpperString     `mapstructure:"code"`
	Tags  []TrimmedString `mapstructure:"tags"`
	Modes []LowerString   `mapstructure:"modes"`
	Flags []UpperString   `mapstructure:"flags"`
}

func TestDecodeConfigNormalizedStrings(t *testing.T) {
	config := map[string]any{
		"name":  "  Alice ",
		"label": "  FooBar ",
		"code":  "  ab-12 ",
		"tags":  []any{" one ", "two", "", "  "},
		"modes": "  ON, Off ",
		"flags": []string{" aa ", "BB "},
	}

	var decoded normalizedConfig
	require.NoError(t, DecodeConfig(config, &decoded))

	require.Equal(t, TrimmedString("Alice"), decoded.Name)
	require.Equal(t, LowerString("foobar"), decoded.Label)
	require.Equal(t, UpperString("AB-12"), decoded.Code)
	require.Equal(t, []TrimmedString{"one", "two"}, decoded.Tags)
	require.Equal(t, []LowerString{"on", "off"}, decoded.Modes)
	require.Equal(t, []UpperString{"AA", "BB"}, decoded.Flags)
}
