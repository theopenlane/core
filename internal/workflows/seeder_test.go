package workflows

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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

	assert.Assert(t, is.Equal("CampaignTarget", resolveSeedSchemaType("CampaignTarget", doc)))
	assert.Assert(t, is.Equal("Control", resolveSeedSchemaType("", doc)))

	doc.SchemaType = ""
	assert.Assert(t, is.Equal(enums.WorkflowObjectTypeEvidence.String(), resolveSeedSchemaType("", doc)))
}

func TestResolveSeedWorkflowKind(t *testing.T) {
	assert.Assert(t, is.Equal(enums.WorkflowKindApproval, resolveSeedWorkflowKind(enums.WorkflowKindApproval, models.WorkflowDefinitionDocument{})))

	doc := models.WorkflowDefinitionDocument{WorkflowKind: enums.WorkflowKindLifecycle}
	assert.Assert(t, is.Equal(enums.WorkflowKindLifecycle, resolveSeedWorkflowKind("", doc)))

	doc.WorkflowKind = ""
	doc.Actions = []models.WorkflowAction{{Type: enums.WorkflowActionTypeApproval.String()}}
	assert.Assert(t, is.Equal(enums.WorkflowKindApproval, resolveSeedWorkflowKind("", doc)))

	doc.Actions = []models.WorkflowAction{{Type: enums.WorkflowActionTypeReview.String()}}
	assert.Assert(t, is.Equal(enums.WorkflowKindLifecycle, resolveSeedWorkflowKind("", doc)))

	doc.Actions = []models.WorkflowAction{{Type: enums.WorkflowActionTypeSendEmail.String()}}
	assert.Assert(t, is.Equal(enums.WorkflowKindNotification, resolveSeedWorkflowKind("", doc)))
}

func TestBoolOrDefault(t *testing.T) {
	assert.Assert(t, boolOrDefault(nil, true))
	assert.Assert(t, !boolOrDefault(nil, false))

	value := true
	assert.Assert(t, boolOrDefault(&value, false))
}
