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
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)
	internalCtx := workflows.AllowContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(suite.client, nil)
	suite.NoError(err)
	suite.client.WorkflowEngine = wfEngine

	def := suite.client.WorkflowDefinition.Create().
		SetName("Invalidate Test " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SaveX(internalCtx)

	control := suite.client.Control.Create().
		SetRefCode("CTL-TEST").
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("text,description").
		SetState(enums.WorkflowProposalStateSubmitted).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = suite.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	assignment1 := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	assignment2 := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-2").
		SetStatus(enums.WorkflowAssignmentStatusPending).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	assignment3 := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-3").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	_, err = suite.client.WorkflowProposal.UpdateOne(proposal).
		SetChanges(map[string]any{"text": "modified"}).
		Save(internalCtx)
	suite.NoError(err)

	reloaded1, err := suite.client.WorkflowAssignment.Get(internalCtx, assignment1.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusPending, reloaded1.Status)
	suite.NotEmpty(reloaded1.InvalidationMetadata.Reason)
	suite.Equal("proposal changes edited after approval", reloaded1.InvalidationMetadata.Reason)

	reloaded2, err := suite.client.WorkflowAssignment.Get(internalCtx, assignment2.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusPending, reloaded2.Status)

	reloaded3, err := suite.client.WorkflowAssignment.Get(internalCtx, assignment3.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusPending, reloaded3.Status)
	suite.NotEmpty(reloaded3.InvalidationMetadata.Reason)
	suite.Equal("proposal changes edited after approval", reloaded3.InvalidationMetadata.Reason)
}

func (suite *HookTestSuite) TestHookWorkflowProposalInvalidateAssignments_DraftState() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)
	internalCtx := workflows.AllowContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(suite.client, nil)
	suite.NoError(err)
	suite.client.WorkflowEngine = wfEngine

	def := suite.client.WorkflowDefinition.Create().
		SetName("Draft Test " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SaveX(internalCtx)

	control := suite.client.Control.Create().
		SetRefCode("CTL-TEST-2").
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("text,description").
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = suite.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	assignment := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	_, err = suite.client.WorkflowProposal.UpdateOne(proposal).
		SetChanges(map[string]any{"text": "modified"}).
		Save(internalCtx)
	suite.NoError(err)

	reloaded, err := suite.client.WorkflowAssignment.Get(internalCtx, assignment.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status)
}

func (suite *HookTestSuite) TestHookWorkflowProposalInvalidateAssignments_NonChangesField() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)
	internalCtx := workflows.AllowContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(suite.client, nil)
	suite.NoError(err)
	suite.client.WorkflowEngine = wfEngine

	def := suite.client.WorkflowDefinition.Create().
		SetName("NonChanges Test " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SaveX(internalCtx)

	control := suite.client.Control.Create().
		SetRefCode("CTL-TEST-3").
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("text,description").
		SetState(enums.WorkflowProposalStateSubmitted).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = suite.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	assignment := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	_, err = suite.client.WorkflowProposal.UpdateOne(proposal).
		SetRevision(2).
		Save(internalCtx)
	suite.NoError(err)

	reloaded, err := suite.client.WorkflowAssignment.Get(internalCtx, assignment.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status)
}

func (suite *HookTestSuite) TestHookWorkflowProposalTriggerOnSubmitResumesInstance() {
	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)
	internalCtx := workflows.AllowContext(userCtx)

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
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"status"}).
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
		SaveX(internalCtx)

	control := suite.client.Control.Create().
		SetRefCode("CTL-RESUME-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStatePaused).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := suite.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposedHash, err := workflows.ComputeProposalHash(map[string]any{"status": enums.ControlStatusApproved})
	suite.NoError(err)

	proposal := suite.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("status").
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(map[string]any{"status": enums.ControlStatusApproved}).
		SetProposedHash(proposedHash).
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = suite.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	_, err = suite.client.WorkflowProposal.UpdateOne(proposal).
		SetState(enums.WorkflowProposalStateSubmitted).
		SetSubmittedByUserID(user.ID).
		Save(internalCtx)
	suite.NoError(err)

	updated, err := suite.client.WorkflowInstance.Get(internalCtx, instance.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowInstanceStateRunning, updated.State)
	suite.Equal(0, updated.CurrentActionIndex)

	count, err := suite.client.WorkflowInstance.Query().
		Where(workflowinstance.WorkflowProposalID(proposal.ID)).
		Count(internalCtx)
	suite.NoError(err)
	suite.Equal(1, count)
}
