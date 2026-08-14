package serveropts

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
)

// TestJobTopicRenames verifies the merged map carries the bare-schema mutation renames and
// the legacy reconcile entry, and never maps a topic to itself
func TestJobTopicRenames(t *testing.T) {
	t.Parallel()

	renames := jobTopicRenames()

	require.Equal(t, gala.TopicName("mutation.Task"), renames[gala.TopicName("Task")])
	require.Equal(t, entityops.MutationTopicName(entityops.MutationConcernDirect, "Task"), renames[gala.TopicName("Task")])
	require.Equal(t, operations.ReconcileTopic.Name, renames[gala.TopicName("integration.ReconcileEnvelope")])

	for legacyTopic, designated := range renames {
		require.NotEqual(t, legacyTopic, designated)
	}
}
