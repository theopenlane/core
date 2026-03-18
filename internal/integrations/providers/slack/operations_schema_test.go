package slack

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers/schematest"
)

// TestSlackMessageConfigSchema verifies the Slack message schema has expected fields
func TestSlackMessageConfigSchema(t *testing.T) {
	schema := slackMessageConfigSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
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

	required := schematest.Required(t, schema)
	require.Contains(t, required, "channel")
}

// TestSlackOperationsIncludeMessageSend verifies message.send operation registration
func TestSlackOperationsIncludeMessageSend(t *testing.T) {
	ops := slackOperations()
	seen := map[types.OperationName]types.OperationDescriptor{}
	for _, op := range ops {
		seen[op.Name] = op
	}

	desc, ok := seen[slackOperationMessageSend]
	require.True(t, ok, "expected message.send operation")
	require.Equal(t, slackMessageConfigSchema, desc.ConfigSchema)
}
