//go:build test

package hooks_test

import (
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func (suite *HookTestSuite) TestHookWorkflowProposalInvalidateAssignments() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].ID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)

	control := suite.client.Control.Create().
		SetRefCode("CTL-TEST").
		SaveX(userCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetControlID(control.ID).
		SaveX(userCtx)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("text,description").
		SetState(enums.WorkflowProposalStateSubmitted).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID("def-123").
		SetState(enums.WorkflowInstanceStateRunning).
		SetWorkflowProposalID(proposal.ID).
		SetControlID(control.ID).
		SaveX(userCtx)

	assignment1 := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SaveX(userCtx)

	assignment2 := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-2").
		SetStatus(enums.WorkflowAssignmentStatusPending).
		SaveX(userCtx)

	assignment3 := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-3").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SaveX(userCtx)

	_, err := suite.client.WorkflowProposal.UpdateOne(proposal).
		SetChanges(map[string]any{"text": "modified"}).
		Save(userCtx)
	suite.NoError(err)

	reloaded1, err := suite.client.WorkflowAssignment.Get(userCtx, assignment1.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusPending, reloaded1.Status)
	suite.NotEmpty(reloaded1.InvalidationMetadata.Reason)
	suite.Equal("proposal changes edited after approval", reloaded1.InvalidationMetadata.Reason)

	reloaded2, err := suite.client.WorkflowAssignment.Get(userCtx, assignment2.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusPending, reloaded2.Status)

	reloaded3, err := suite.client.WorkflowAssignment.Get(userCtx, assignment3.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusPending, reloaded3.Status)
	suite.NotEmpty(reloaded3.InvalidationMetadata.Reason)
	suite.Equal("proposal changes edited after approval", reloaded3.InvalidationMetadata.Reason)
}

func (suite *HookTestSuite) TestHookWorkflowProposalInvalidateAssignments_DraftState() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].ID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)

	control := suite.client.Control.Create().
		SetRefCode("CTL-TEST-2").
		SaveX(userCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetControlID(control.ID).
		SaveX(userCtx)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("text,description").
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID("def-123").
		SetState(enums.WorkflowInstanceStateRunning).
		SetWorkflowProposalID(proposal.ID).
		SetControlID(control.ID).
		SaveX(userCtx)

	assignment := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SaveX(userCtx)

	_, err := suite.client.WorkflowProposal.UpdateOne(proposal).
		SetChanges(map[string]any{"text": "modified"}).
		Save(userCtx)
	suite.NoError(err)

	reloaded, err := suite.client.WorkflowAssignment.Get(userCtx, assignment.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status)
}

func (suite *HookTestSuite) TestHookWorkflowProposalInvalidateAssignments_NonChangesField() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].ID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)

	control := suite.client.Control.Create().
		SetRefCode("CTL-TEST-3").
		SaveX(userCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetControlID(control.ID).
		SaveX(userCtx)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("text,description").
		SetState(enums.WorkflowProposalStateSubmitted).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID("def-123").
		SetState(enums.WorkflowInstanceStateRunning).
		SetWorkflowProposalID(proposal.ID).
		SetControlID(control.ID).
		SaveX(userCtx)

	assignment := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SaveX(userCtx)

	_, err := suite.client.WorkflowProposal.UpdateOne(proposal).
		SetRevision(2).
		Save(userCtx)
	suite.NoError(err)

	reloaded, err := suite.client.WorkflowAssignment.Get(userCtx, assignment.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status)
}

func (suite *HookTestSuite) TestHookWorkflowProposalTriggerOnSubmitResumesInstance() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].ID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)

	wfEngine, err := engine.NewWorkflowEngine(suite.client, nil)
	suite.NoError(err)
	suite.client.WorkflowEngine = wfEngine

	params := struct {
		Targets []workflows.TargetConfig `json:"targets"`
		Fields  []string                 `json:"fields"`
	}{
		Targets: []workflows.TargetConfig{
			{Type: enums.WorkflowTargetTypeUser, ID: user.ID},
		},
		Fields: []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	suite.NoError(err)

	def := suite.client.WorkflowDefinition.Create().
		SetName("Submit Resume " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"status"}},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "status_approval",
					Params: paramsBytes,
				},
			},
		}).
		SaveX(userCtx)

	control := suite.client.Control.Create().
		SetRefCode("CTL-RESUME-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStatePaused).
		SetControlID(control.ID).
		SaveX(userCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SaveX(userCtx)

	proposedHash, err := workflows.ComputeProposalHash(map[string]any{"status": enums.ControlStatusApproved})
	suite.NoError(err)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("status").
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(map[string]any{"status": enums.ControlStatusApproved}).
		SetProposedHash(proposedHash).
		SetSubmittedByUserID(user.ID).
		SaveX(userCtx)

	instance = suite.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(userCtx)

	_, err = suite.client.WorkflowProposal.UpdateOne(proposal).
		SetState(enums.WorkflowProposalStateSubmitted).
		SetSubmittedByUserID(user.ID).
		Save(userCtx)
	suite.NoError(err)

	updated, err := suite.client.WorkflowInstance.Get(userCtx, instance.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowInstanceStateRunning, updated.State)
	suite.Equal(0, updated.CurrentActionIndex)

	count, err := suite.client.WorkflowInstance.Query().
		Where(workflowinstance.WorkflowProposalID(proposal.ID)).
		Count(userCtx)
	suite.NoError(err)
	suite.Equal(1, count)
}
