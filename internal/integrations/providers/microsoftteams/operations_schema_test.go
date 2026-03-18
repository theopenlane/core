package microsoftteams

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providers/schematest"
)

// TestTeamsMessageConfigSchema verifies the Teams message schema has expected fields
func TestTeamsMessageConfigSchema(t *testing.T) {
	schema := teamsMessageConfigSchema
	require.NotNil(t, schema)

	props := schematest.Properties(t, schema)
	for _, key := range []string{
		"team_id",
		"channel_id",
		"body",
		"body_format",
		"subject",
	} {
		require.Contains(t, props, key)
	}

	required := schematest.Required(t, schema)
	require.Contains(t, required, "team_id")
	require.Contains(t, required, "channel_id")
	require.Contains(t, required, "body")
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
