//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
)

func TestOrganizationCleanupListenerCascadeWithIntegrations(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	orgID := org.OrganizationID
	ownerCtx := org.UserCtx
	allowCtx := privacy.DecisionContext(th.SetInternalContext(ownerCtx, suite.Client.DB), privacy.Allow)

	waitForEvents()

	// the suite mocks no stripe subscription cancel call, so keep the entitlements_deleted
	// listener on its skip path to avoid parking a retrying job on the shared runtime
	assert.NilError(t, suite.Client.DB.Organization.UpdateOneID(orgID).ClearStripeCustomerID().Exec(allowCtx))

	task1 := (&th.TaskBuilder{Client: suite.Client}).MustNew(ownerCtx, t)
	contact1 := (&th.ContactBuilder{Client: suite.Client}).MustNew(ownerCtx, t)

	installation, fragment := seedHarnessLoop(t, allowCtx)
	assert.Equal(t, orgID, installation.OwnerID)

	workflowEngine, workflowRuntime := acquireWorkflowRuntime(t)

	approvalParams, err := json.Marshal(workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: org.ID}},
		},
		Label: "Cleanup Approval",
	})
	assert.NilError(t, err)

	workflowDef := createWorkflowDefinition(allowCtx, t, orgID, "Control", enums.WorkflowKindApproval, models.WorkflowDefinitionDocument{
		Triggers:   []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"status"}}},
		Conditions: []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    "cleanup_approval",
			Params: approvalParams,
		}},
	})

	control, err := suite.Client.DB.Control.Create().
		SetRefCode("CTL-" + ulids.New().String()).
		SetTitle("Cleanup Control").
		SetStatus(enums.ControlStatusNotImplemented).
		SetOwnerID(orgID).
		Save(allowCtx)
	assert.NilError(t, err)

	instance, err := workflowEngine.TriggerWorkflow(allowCtx, workflowDef, &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}, engine.TriggerInput{EventType: "UPDATE", ChangedFields: []string{"status"}})
	assert.NilError(t, err)

	waitForGala(t, workflowRuntime)

	assignments, err := graphapi.WaitForAssignments(allowCtx, suite.Client.DB, instance.ID, 1)
	assert.NilError(t, err)

	resp, err := suite.Client.API.DeleteOrganization(ownerCtx, orgID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(orgID, resp.DeleteOrganization.DeletedID))

	waitForEvents()

	purgedCtx := entx.SkipSoftDelete(allowCtx)

	waitForCondition(t, func() bool {
		exists, err := suite.Client.DB.Organization.Query().Where(organization.ID(orgID)).Exist(purgedCtx)
		return err == nil && !exists
	}, "organization row should be hard deleted by the cascade")

	assert.Equal(t, 0, activeReconcileJobs(t, fragment))

	taskExists, err := suite.Client.DB.Task.Query().Where(task.ID(task1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !taskExists)

	contactExists, err := suite.Client.DB.Contact.Query().Where(contact.ID(contact1.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !contactExists)

	groupExists, err := suite.Client.DB.Group.Query().Where(group.ID(org.GroupID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !groupExists)

	installationExists, err := suite.Client.DB.Integration.Query().Where(integration.ID(installation.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !installationExists)

	instanceExists, err := suite.Client.DB.WorkflowInstance.Query().Where(workflowinstance.ID(instance.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !instanceExists)

	assignmentExists, err := suite.Client.DB.WorkflowAssignment.Query().Where(workflowassignment.ID(assignments[0].ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !assignmentExists)

	targetExists, err := suite.Client.DB.WorkflowAssignmentTarget.Query().Where(workflowassignmenttarget.WorkflowAssignmentID(assignments[0].ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !targetExists)

	objectRefExists, err := suite.Client.DB.WorkflowObjectRef.Query().Where(workflowobjectref.WorkflowInstanceID(instance.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !objectRefExists)

	eventExists, err := suite.Client.DB.WorkflowEvent.Query().Where(workflowevent.WorkflowInstanceID(instance.ID)).Exist(purgedCtx)
	assert.NilError(t, err)
	assert.Check(t, !eventExists)
}
