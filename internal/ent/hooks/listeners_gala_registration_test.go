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

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaEntitlementListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 3)

	require.True(t, runtime.InterestedIn(gala.TopicName(entgen.TypeOrganization), ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(gala.TopicName(entgen.TypeOrganization), eventqueue.SoftDeleteOne))
	require.False(t, runtime.InterestedIn(gala.TopicName(entgen.TypeOrganization), ent.OpUpdate.String()))

	require.True(t, runtime.InterestedIn(gala.TopicName(entgen.TypeOrganizationSetting), ent.OpUpdate.String()))
	require.False(t, runtime.InterestedIn(gala.TopicName(entgen.TypeOrganizationSetting), ent.OpDelete.String()))
}

func TestRegisterGalaOrganizationAvatarListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaOrganizationAvatarListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeOrganization)
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
}

func TestRegisterGalaTaskRuleListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaTaskRuleListeners(runtime)
	require.NoError(t, err)
	require.NotEmpty(t, ids)

	for _, schemaType := range []string{entgen.TypeOnboarding, entgen.TypeOrganization, entgen.TypeNotification} {
		topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, schemaType)
		require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()), "expected %s to subscribe to create", schemaType)
		require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()), "expected %s not to subscribe to update", schemaType)
	}
}

func TestRegisterGalaTrustCenterCacheListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaTrustCenterCacheListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 10)

	require.True(t, runtime.InterestedIn(gala.TopicName(entgen.TypeTrustCenterDoc), ent.OpUpdate.String()))
	require.True(t, runtime.InterestedIn(gala.TopicName(entgen.TypeTrustCenterFAQ), ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(gala.TopicName(entgen.TypeTrustCenter), eventqueue.SoftDeleteOne))
}

func TestRegisterGalaTrustCenterWatermarkListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaTrustCenterWatermarkListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeTrustCenterDoc)
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestRegisterGalaWorkflowMutationListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaWorkflowMutationListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, len(enums.WorkflowObjectTypes)+1)

	for _, schemaType := range enums.WorkflowObjectTypes {
		topic := eventqueue.MutationTopicName(eventqueue.MutationConcernWorkflow, schemaType)
		require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
		require.True(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
		require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
	}

	assignmentTopic := eventqueue.MutationTopicName(eventqueue.MutationConcernWorkflow, entgen.TypeWorkflowAssignment)
	require.True(t, runtime.InterestedIn(assignmentTopic, ent.OpUpdate.String()))
	require.False(t, runtime.InterestedIn(assignmentTopic, ent.OpCreate.String()))
}

func TestRegisterGalaWorkflowListenersRegistersCommandTopics(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaWorkflowListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, len(enums.WorkflowObjectTypes)+6)

	require.True(t, runtime.InterestedIn(gala.TopicWorkflowTriggered, ""))
	require.True(t, runtime.InterestedIn(gala.TopicWorkflowActionStarted, ""))
	require.True(t, runtime.InterestedIn(gala.TopicWorkflowActionCompleted, ""))
	require.True(t, runtime.InterestedIn(gala.TopicWorkflowAssignmentCompleted, ""))
	require.True(t, runtime.InterestedIn(gala.TopicWorkflowInstanceCompleted, ""))
	require.False(t, runtime.InterestedIn(gala.TopicName("workflows.command.triggered"), ""))
}

func TestRegisterGalaIntegrationCleanupListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaIntegrationCleanupListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeIntegration)
	require.True(t, runtime.InterestedIn(topic, ent.OpDeleteOne.String()))
	require.True(t, runtime.InterestedIn(topic, eventqueue.SoftDeleteOne))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
}

func TestRegisterGalaVendorScoringListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaVendorScoringListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	require.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeVendorScoringConfig), ent.OpUpdate.String()))
	require.True(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeVendorScoringConfig), ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeVendorScoringConfig), ent.OpCreate.String()))
}

func TestRegisterGalaIdentityResolutionListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaIdentityResolutionListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeDirectoryAccount)
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestRegisterGalaDocumentAssociationListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaDocumentAssociationListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 3)

	for _, schemaType := range []string{entgen.TypeActionPlan, entgen.TypeInternalPolicy, entgen.TypeProcedure} {
		topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, schemaType)
		require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
		require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
	}
}

func TestRegisterGalaCampaignRecurringListeners(t *testing.T) {
	t.Parallel()

	registry := gala.NewRegistry()

	ids, err := RegisterGalaCampaignRecurringListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeCampaign)
	require.True(t, registry.InterestedIn(topic, ent.OpUpdate.String()))
	require.True(t, registry.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, registry.InterestedIn(topic, ent.OpCreate.String()))
	require.False(t, registry.InterestedIn(topic, ent.OpDelete.String()))
}

func TestRegisterGalaQuestionnaireTransformListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := RegisterGalaQuestionnaireTransformListeners(runtime)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, entgen.TypeAssessmentResponse)
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
}
