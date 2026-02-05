package slack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlackMessageConfigSchema(t *testing.T) {
	schema := slackMessageConfigSchema
	require.NotNil(t, schema)
	require.Equal(t, "object", schema["type"])

	props := schemaProperties(t, schema)
	for _, key := range []string{
		"channel",
		"text",
		"blocks",
		"attachments",
		"thread_ts",
		"unfurl_links",
		"unfurl_media",
	} {
		require.Contains(t, props, key)
	}

	required := schemaRequired(t, schema)
	require.Contains(t, required, "channel")
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "expected properties to be a map")
	return props
}

func schemaRequired(t *testing.T, schema map[string]any) []string {
	t.Helper()

	raw, ok := schema["required"]
	if !ok {
		return nil
	}

	values, ok := raw.([]any)
	require.True(t, ok, "expected required to be a slice")

	out := make([]string, 0, len(values))
	for _, value := range values {
		if name, ok := value.(string); ok {
			out = append(out, name)
		}
	}
	return out
}
