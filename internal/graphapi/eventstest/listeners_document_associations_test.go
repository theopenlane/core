//go:build test

package eventstest_test

import (
	"context"
	"errors"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

const (
	listenerPollTimeout  = 10 * time.Second
	listenerPollInterval = 25 * time.Millisecond
)

var errListenerPollTimeout = errors.New("listener poll timed out")

func listenerPoll[T any](query func() (T, error), condition func(T) bool) (T, error) {
	deadline := time.Now().Add(listenerPollTimeout)

	var (
		value T
		err   error
	)

	for time.Now().Before(deadline) {
		value, err = query()
		if err == nil && condition(value) {
			return value, nil
		}

		time.Sleep(listenerPollInterval)
	}

	if err != nil {
		return value, err
	}

	return value, errListenerPollTimeout
}

func TestDocumentAssociationListeners(t *testing.T) {
	docUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(docUser.UserCtx, suite.Client.DB)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.DocumentAssociationListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	control := (&th.ControlBuilder{Client: suite.Client, RefCode: "CC-2"}).MustNew(docUser.UserCtx, t)
	parentControl := (&th.ControlBuilder{Client: suite.Client}).MustNew(docUser.UserCtx, t)
	subcontrol := (&th.SubcontrolBuilder{Client: suite.Client, Name: "AC-1.1", ControlID: parentControl.ID}).MustNew(docUser.UserCtx, t)

	t.Run("policy create links referenced controls and subcontrols without bumping revision", func(t *testing.T) {
		details := "This policy is governed by CC-2\nand implements AC-1.1 for access reviews"

		resp, err := suite.Client.API.CreateInternalPolicy(docUser.UserCtx, testclient.CreateInternalPolicyInput{
			Name:    "doc assoc policy",
			Details: &details,
		})
		assert.NilError(t, err)

		policyID := resp.CreateInternalPolicy.InternalPolicy.ID

		controlIDs, err := listenerPoll(func() ([]string, error) {
			policy, err := suite.Client.DB.InternalPolicy.Get(ctx, policyID)
			if err != nil {
				return nil, err
			}

			return suite.Client.DB.InternalPolicy.QueryControls(policy).IDs(ctx)
		}, func(ids []string) bool {
			return len(ids) > 0
		})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]string{control.ID}, controlIDs))

		policy, err := suite.Client.DB.InternalPolicy.Get(ctx, policyID)
		assert.NilError(t, err)

		subcontrolIDs, err := suite.Client.DB.InternalPolicy.QuerySubcontrols(policy).IDs(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]string{subcontrol.ID}, subcontrolIDs))

		// re-asserting the loaded revision matches main's SetRevision(doc.Revision): the
		// revision hook treats a same-value set as not-manually-set and still bumps patch
		bumped, err := models.BumpPatch(lo.FromPtr(resp.CreateInternalPolicy.InternalPolicy.Revision))
		assert.NilError(t, err)
		assert.Check(t, is.Equal(bumped, policy.Revision))
	})

	t.Run("procedure create links referenced controls", func(t *testing.T) {
		details := "Follow the steps required by CC-2 during onboarding"

		resp, err := suite.Client.API.CreateProcedure(docUser.UserCtx, testclient.CreateProcedureInput{
			Name:    "doc assoc procedure",
			Details: &details,
		})
		assert.NilError(t, err)

		procedureID := resp.CreateProcedure.Procedure.ID

		controlIDs, err := listenerPoll(func() ([]string, error) {
			procedure, err := suite.Client.DB.Procedure.Get(ctx, procedureID)
			if err != nil {
				return nil, err
			}

			return suite.Client.DB.Procedure.QueryControls(procedure).IDs(ctx)
		}, func(ids []string) bool {
			return len(ids) > 0
		})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]string{control.ID}, controlIDs))
	})

	t.Run("action plan create links referenced controls", func(t *testing.T) {
		details := "Remediation tracked against CC-2"

		resp, err := suite.Client.API.CreateActionPlan(docUser.UserCtx, testclient.CreateActionPlanInput{
			Name:    "doc assoc action plan",
			Title:   "doc assoc action plan",
			Details: &details,
		})
		assert.NilError(t, err)

		actionPlanID := resp.CreateActionPlan.ActionPlan.ID

		controlIDs, err := listenerPoll(func() ([]string, error) {
			actionPlan, err := suite.Client.DB.ActionPlan.Get(ctx, actionPlanID)
			if err != nil {
				return nil, err
			}

			return suite.Client.DB.ActionPlan.QueryControls(actionPlan).IDs(ctx)
		}, func(ids []string) bool {
			return len(ids) > 0
		})
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual([]string{control.ID}, controlIDs))
	})

	t.Run("no control references means no association write", func(t *testing.T) {
		details := "General guidance without any framework references"

		resp, err := suite.Client.API.CreateInternalPolicy(docUser.UserCtx, testclient.CreateInternalPolicyInput{
			Name:    "doc assoc no match policy",
			Details: &details,
		})
		assert.NilError(t, err)

		policyID := resp.CreateInternalPolicy.InternalPolicy.ID

		th.WaitForGala(t, setup.Runtime)

		policy, err := suite.Client.DB.InternalPolicy.Get(ctx, policyID)
		assert.NilError(t, err)

		controlIDs, err := suite.Client.DB.InternalPolicy.QueryControls(policy).IDs(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(controlIDs, 0))

		assert.Check(t, is.Equal(lo.FromPtr(resp.CreateInternalPolicy.InternalPolicy.Revision), policy.Revision))
	})

	th.CleanupOrganizationDataWithContext(docUser.UserCtx, t)
}
