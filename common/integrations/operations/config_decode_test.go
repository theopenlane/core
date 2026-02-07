package operations

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/types"
)

type normalizedConfig struct {
	Name  types.TrimmedString   `json:"name"`
	Label types.LowerString     `json:"label"`
	Code  types.UpperString     `json:"code"`
	Tags  []types.TrimmedString `json:"tags"`
	Modes []types.LowerString   `json:"modes"`
	Flags []types.UpperString   `json:"flags"`
}

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
