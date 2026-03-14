package workflows

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
)

func TestResolveSeedSchemaType(t *testing.T) {
	doc := models.WorkflowDefinitionDocument{
		SchemaType: "Control",
		Triggers: []models.WorkflowTrigger{
			{ObjectType: enums.WorkflowObjectTypeEvidence},
		},
	}

	require.Equal(t, "CampaignTarget", resolveSeedSchemaType("CampaignTarget", doc))
	require.Equal(t, "Control", resolveSeedSchemaType("", doc))

	doc.SchemaType = ""
	require.Equal(t, enums.WorkflowObjectTypeEvidence.String(), resolveSeedSchemaType("", doc))
}

func TestResolveSeedWorkflowKind(t *testing.T) {
	require.Equal(t, enums.WorkflowKindApproval, resolveSeedWorkflowKind(enums.WorkflowKindApproval, models.WorkflowDefinitionDocument{}))

	doc := models.WorkflowDefinitionDocument{WorkflowKind: enums.WorkflowKindLifecycle}
	require.Equal(t, enums.WorkflowKindLifecycle, resolveSeedWorkflowKind("", doc))

	doc.WorkflowKind = ""
	doc.Actions = []models.WorkflowAction{{Type: enums.WorkflowActionTypeApproval.String()}}
	require.Equal(t, enums.WorkflowKindApproval, resolveSeedWorkflowKind("", doc))

	doc.Actions = []models.WorkflowAction{{Type: enums.WorkflowActionTypeReview.String()}}
	require.Equal(t, enums.WorkflowKindLifecycle, resolveSeedWorkflowKind("", doc))

	doc.Actions = []models.WorkflowAction{{Type: enums.WorkflowActionTypeSendEmail.String()}}
	require.Equal(t, enums.WorkflowKindNotification, resolveSeedWorkflowKind("", doc))
}

func TestBoolOrDefault(t *testing.T) {
	require.True(t, boolOrDefault(nil, true))
	require.False(t, boolOrDefault(nil, false))

	value := true
	require.True(t, boolOrDefault(&value, false))
}
