package slack

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/internal/integrations/types"
)

// TestSlackMessageConfigSchema verifies the Slack message schema validates expected payloads.
func TestSlackMessageConfigSchema(t *testing.T) {
	schema := slackMessageConfigSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"channel": "C12345",
			"text": "hello",
			"thread_ts": "1712345678.000100",
			"unfurl_links": false,
			"unfurl_media": false,
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "hello"
					}
				}
			],
			"attachments": [
				{
					"text": "attachment"
				}
			]
		}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{"channel":true}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
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
