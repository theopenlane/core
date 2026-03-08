package microsoftteams

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// TestTeamsMessageConfigSchema verifies the Teams message schema validates expected payloads.
func TestTeamsMessageConfigSchema(t *testing.T) {
	schema := teamsMessageConfigSchema
	require.NotNil(t, schema)

	validResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"team_id": "team-1",
			"channel_id": "channel-1",
			"body": "hello",
			"body_format": "text",
			"subject": "test message"
		}`)),
	)
	require.NoError(t, err)
	require.True(t, validResult.Valid(), "expected valid config, got errors: %v", validResult.Errors())

	invalidResult, err := gojsonschema.Validate(
		gojsonschema.NewBytesLoader(schema),
		gojsonschema.NewBytesLoader([]byte(`{
			"team_id": 1,
			"channel_id": "channel-1",
			"body": "hello"
		}`)),
	)
	require.NoError(t, err)
	require.False(t, invalidResult.Valid(), "expected invalid config to fail schema validation")
}

// TestTeamsOperationsIncludeMessageSend verifies message.send operation registration
func TestTeamsOperationsIncludeMessageSend(t *testing.T) {
	ops := teamsOperations()
	seen := map[string]struct{}{}
	for _, op := range ops {
		seen[string(op.Name)] = struct{}{}
	}

	_, ok := seen[string(teamsMessageSendOp)]
	require.True(t, ok, "expected message.send operation")
}
