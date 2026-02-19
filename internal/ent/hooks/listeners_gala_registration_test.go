package hooks

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterGalaEntitlementListeners(t *testing.T) {
	t.Parallel()

	registry := gala.NewRegistry()

	ids, err := RegisterGalaEntitlementListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeOrganization), ent.OpCreate.String()))
	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeOrganization), SoftDeleteOne))
	require.False(t, registry.InterestedIn(gala.TopicName(entgen.TypeOrganization), ent.OpUpdate.String()))

	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeOrganizationSetting), ent.OpUpdate.String()))
	require.False(t, registry.InterestedIn(gala.TopicName(entgen.TypeOrganizationSetting), ent.OpDelete.String()))
}

func TestRegisterGalaTrustCenterCacheListeners(t *testing.T) {
	t.Parallel()

	registry := gala.NewRegistry()

	ids, err := RegisterGalaTrustCenterCacheListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, 9)

	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeTrustCenterDoc), ent.OpUpdate.String()))
	require.True(t, registry.InterestedIn(gala.TopicName(entgen.TypeTrustCenter), SoftDeleteOne))
}

func TestRegisterGalaWorkflowMutationListeners(t *testing.T) {
	t.Parallel()

	registry := gala.NewRegistry()

	ids, err := RegisterGalaWorkflowMutationListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, len(enums.WorkflowObjectTypes)+1)

	for _, schemaType := range enums.WorkflowObjectTypes {
		topic := eventqueue.WorkflowMutationTopicName(schemaType)
		require.True(t, registry.InterestedIn(topic, ent.OpCreate.String()))
		require.True(t, registry.InterestedIn(topic, ent.OpUpdate.String()))
		require.False(t, registry.InterestedIn(topic, ent.OpDelete.String()))
	}

	assignmentTopic := eventqueue.WorkflowMutationTopicName(entgen.TypeWorkflowAssignment)
	require.True(t, registry.InterestedIn(assignmentTopic, ent.OpUpdate.String()))
	require.False(t, registry.InterestedIn(assignmentTopic, ent.OpCreate.String()))
}

func TestRegisterGalaNotificationListeners(t *testing.T) {
	t.Parallel()

	registry := gala.NewRegistry()

	ids, err := RegisterGalaNotificationListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, 6)

	require.True(t, registry.InterestedIn(eventqueue.NotificationMutationTopicName(entgen.TypeTask), ent.OpCreate.String()))
	require.True(t, registry.InterestedIn(eventqueue.NotificationMutationTopicName(entgen.TypeInternalPolicy), ent.OpUpdate.String()))
	require.True(t, registry.InterestedIn(eventqueue.NotificationMutationTopicName(entgen.TypeRisk), ent.OpDelete.String()))
	require.True(t, registry.InterestedIn(eventqueue.NotificationMutationTopicName(entgen.TypeProcedure), ent.OpUpdateOne.String()))
	require.True(t, registry.InterestedIn(eventqueue.NotificationMutationTopicName(entgen.TypeNote), ent.OpCreate.String()))
	require.True(t, registry.InterestedIn(eventqueue.NotificationMutationTopicName(entgen.TypeExport), ent.OpUpdate.String()))
}
