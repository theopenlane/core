package hooks

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterGalaEntitlementListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, EntitlementListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 3)

	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganization), ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganization), ""))
	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganization), entityops.OpSoftDelete))
	require.False(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganization), ent.OpUpdate.String()))

	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganizationSetting), ent.OpUpdate.String()))
	require.False(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganizationSetting), ent.OpDelete.String()))
}

func TestRegisterGalaOrganizationCleanupListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, OrganizationCleanupListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganization)
	require.True(t, runtime.InterestedIn(topic, ""))
	require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpDeleteOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
	require.True(t, runtime.InterestedIn(topic, entityops.OpSoftDelete))
}

func TestRegisterGalaOrganizationAvatarListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, OrganizationAvatarListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeOrganization)
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
}

func TestRegisterGalaTaskRuleListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, TaskRuleListeners()...)
	require.NoError(t, err)
	require.NotEmpty(t, ids)

	for _, schemaType := range []string{entgen.TypeOnboarding, entgen.TypeOrganization, entgen.TypeNotification} {
		topic := entityops.MutationTopicName(entityops.MutationConcernDirect, schemaType)
		require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()), "expected %s to subscribe to create", schemaType)
		require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()), "expected %s not to subscribe to update", schemaType)
	}
}

func TestRegisterGalaTrustCenterCacheListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, TrustCenterCacheListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 10)

	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeTrustCenterDoc), ent.OpUpdate.String()))
	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeTrustCenterFAQ), ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeTrustCenter), ""))
}

func TestRegisterGalaTrustCenterWatermarkListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, TrustCenterWatermarkListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeTrustCenterDoc)
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestRegisterGalaWorkflowMutationListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, WorkflowMutationListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, len(enums.WorkflowObjectTypes)+1)

	for _, schemaType := range enums.WorkflowObjectTypes {
		topic := entityops.MutationTopicName(entityops.MutationConcernWorkflow, schemaType)
		require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
		require.True(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
		require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
	}

	assignmentTopic := entityops.MutationTopicName(entityops.MutationConcernWorkflow, entgen.TypeWorkflowAssignment)
	require.True(t, runtime.InterestedIn(assignmentTopic, ent.OpUpdate.String()))
	require.False(t, runtime.InterestedIn(assignmentTopic, ent.OpCreate.String()))
}

func TestRegisterGalaWorkflowListenersRegistersCommandTopics(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, WorkflowListeners()...)
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

	ids, err := gala.Register(runtime, IntegrationCleanupListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 2)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeIntegration)
	require.True(t, runtime.InterestedIn(topic, ""))
	require.False(t, runtime.InterestedIn(topic, ent.OpDeleteOne.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
}

func TestRegisterGalaVendorScoringListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, VendorScoringListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeVendorScoringConfig), ent.OpUpdate.String()))
	require.True(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeVendorScoringConfig), ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeVendorScoringConfig), ent.OpCreate.String()))
}

func TestRegisterGalaIdentityResolutionListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, IdentityResolutionListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeDirectoryAccount)
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestRegisterGalaDocumentAssociationListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, DocumentAssociationListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 3)

	for _, schemaType := range []string{entgen.TypeActionPlan, entgen.TypeInternalPolicy, entgen.TypeProcedure} {
		topic := entityops.MutationTopicName(entityops.MutationConcernDirect, schemaType)
		require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
		require.False(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
	}
}

func TestRegisterGalaCampaignRecurringListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, CampaignRecurringListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeCampaign)
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	require.False(t, runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestRegisterGalaQuestionnaireTransformListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	require.NoError(t, err)

	ids, err := gala.Register(runtime, QuestionnaireTransformListeners()...)
	require.NoError(t, err)
	require.Len(t, ids, 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, entgen.TypeAssessmentResponse)
	require.True(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	require.True(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
}
