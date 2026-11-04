//go:build test

package hooks_test

import (
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestHookWorkflowProposalInvalidateAssignments(t *testing.T) {
	s := suite.setupTest(t)
	user := s.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, s.client)
	internalCtx := auth.WithInternalCrossOrgContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(s.client, nil)
	assert.Check(t, err)
	engine.SetDefault(wfEngine)
	t.Cleanup(func() { engine.SetDefault(nil) })

	def := s.client.WorkflowDefinition.Create().
		SetName("Invalidate Test " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SaveX(internalCtx)

	control := s.client.Control.Create().
		SetRefCode("CTL-TEST").
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := s.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := s.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposal := s.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey(workflows.DeriveDomainKey(enums.WorkflowObjectTypeControl, []string{"text", "description"})).
		SetState(enums.WorkflowProposalStateSubmitted).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = s.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	assignment1 := s.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	assignment2 := s.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-2").
		SetStatus(enums.WorkflowAssignmentStatusPending).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	assignment3 := s.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-3").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	_, err = s.client.WorkflowProposal.UpdateOne(proposal).
		SetChanges(map[string]any{"text": "modified"}).
		Save(internalCtx)
	assert.Check(t, err)

	reloaded1, err := s.client.WorkflowAssignment.Get(internalCtx, assignment1.ID)
	assert.Check(t, err)
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, reloaded1.Status))
	assert.Check(t, reloaded1.InvalidationMetadata.Reason != "")
	assert.Check(t, is.Equal("proposal changes edited after approval", reloaded1.InvalidationMetadata.Reason))

	reloaded2, err := s.client.WorkflowAssignment.Get(internalCtx, assignment2.ID)
	assert.Check(t, err)
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, reloaded2.Status))

	reloaded3, err := s.client.WorkflowAssignment.Get(internalCtx, assignment3.ID)
	assert.Check(t, err)
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, reloaded3.Status))
	assert.Check(t, reloaded3.InvalidationMetadata.Reason != "")
	assert.Check(t, is.Equal("proposal changes edited after approval", reloaded3.InvalidationMetadata.Reason))
}

func TestHookWorkflowProposalInvalidateAssignments_DraftState(t *testing.T) {
	s := suite.setupTest(t)
	user := s.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, s.client)
	internalCtx := auth.WithInternalCrossOrgContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(s.client, nil)
	assert.Check(t, err)
	engine.SetDefault(wfEngine)
	t.Cleanup(func() { engine.SetDefault(nil) })

	def := s.client.WorkflowDefinition.Create().
		SetName("Draft Test " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SaveX(internalCtx)

	control := s.client.Control.Create().
		SetRefCode("CTL-TEST-2").
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := s.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := s.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposal := s.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey(workflows.DeriveDomainKey(enums.WorkflowObjectTypeControl, []string{"text", "description"})).
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = s.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	assignment := s.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	_, err = s.client.WorkflowProposal.UpdateOne(proposal).
		SetChanges(map[string]any{"text": "modified"}).
		Save(internalCtx)
	assert.Check(t, err)

	reloaded, err := s.client.WorkflowAssignment.Get(internalCtx, assignment.ID)
	assert.Check(t, err)
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status))
}

func TestHookWorkflowProposalInvalidateAssignments_NonChangesField(t *testing.T) {
	s := suite.setupTest(t)
	user := s.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, s.client)
	internalCtx := auth.WithInternalCrossOrgContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(s.client, nil)
	assert.Check(t, err)
	engine.SetDefault(wfEngine)
	t.Cleanup(func() { engine.SetDefault(nil) })

	def := s.client.WorkflowDefinition.Create().
		SetName("NonChanges Test " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SaveX(internalCtx)

	control := s.client.Control.Create().
		SetRefCode("CTL-TEST-3").
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := s.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStateRunning).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := s.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposal := s.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey(workflows.DeriveDomainKey(enums.WorkflowObjectTypeControl, []string{"text", "description"})).
		SetState(enums.WorkflowProposalStateSubmitted).
		SetChanges(map[string]any{"text": "original"}).
		SetProposedHash("hash1").
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = s.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	assignment := s.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instance.ID).
		SetAssignmentKey("approver-1").
		SetStatus(enums.WorkflowAssignmentStatusApproved).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	_, err = s.client.WorkflowProposal.UpdateOne(proposal).
		SetRevision(2).
		Save(internalCtx)
	assert.Check(t, err)

	reloaded, err := s.client.WorkflowAssignment.Get(internalCtx, assignment.ID)
	assert.Check(t, err)
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status))
}

func TestHookWorkflowProposalTriggerOnSubmitResumesInstance(t *testing.T) {
	s := suite.setupTest(t)
	user := s.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	userCtx = generated.NewContext(userCtx, s.client)
	internalCtx := auth.WithInternalCrossOrgContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(s.client, nil)
	assert.Check(t, err)
	engine.SetDefault(wfEngine)
	t.Cleanup(func() { engine.SetDefault(nil) })

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
	assert.Check(t, err)

	def := s.client.WorkflowDefinition.Create().
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

	control := s.client.Control.Create().
		SetRefCode("CTL-RESUME-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(userCtx)

	instance := s.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(def.ID).
		SetState(enums.WorkflowInstanceStatePaused).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	objRef := s.client.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	proposedHash, err := workflows.ComputeProposalHash(map[string]any{"status": enums.ControlStatusApproved})
	assert.Check(t, err)

	proposal := s.client.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey(workflows.DeriveDomainKey(enums.WorkflowObjectTypeControl, []string{"status"})).
		SetState(enums.WorkflowProposalStateDraft).
		SetChanges(map[string]any{"status": enums.ControlStatusApproved}).
		SetProposedHash(proposedHash).
		SetSubmittedByUserID(user.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	instance = s.client.WorkflowInstance.UpdateOne(instance).
		SetWorkflowProposalID(proposal.ID).
		SaveX(internalCtx)

	_, err = s.client.WorkflowProposal.UpdateOne(proposal).
		SetState(enums.WorkflowProposalStateSubmitted).
		SetSubmittedByUserID(user.ID).
		Save(internalCtx)
	assert.Check(t, err)

	updated, err := s.client.WorkflowInstance.Get(internalCtx, instance.ID)
	assert.Check(t, err)
	assert.Check(t, is.Equal(enums.WorkflowInstanceStateRunning, updated.State))
	assert.Check(t, is.Equal(0, updated.CurrentActionIndex))

	count, err := s.client.WorkflowInstance.Query().
		Where(workflowinstance.WorkflowProposalID(proposal.ID)).
		Count(internalCtx)
	assert.Check(t, err)
	assert.Check(t, is.Equal(1, count))
}
