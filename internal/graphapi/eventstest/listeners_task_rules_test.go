//go:build test

package eventstest_test

import (
	"context"
	"fmt"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const taskRuleOrganizationReadyObjectType = "organization.ready"

func TestTaskRuleListenersRealMutations(t *testing.T) {
	user := suite.UserBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(th.SetContext(user.UserCtx, suite.Client.DB), privacy.Allow)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.TaskRuleListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	onboarding, err := suite.Client.DB.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Task Rule Listener Co",
	}).Save(allowCtx)
	assert.NilError(t, err)

	orgID := onboarding.OrganizationID
	orgCtx := privacy.DecisionContext(th.SetContext(auth.NewTestContextWithOrgID(user.ID, orgID), suite.Client.DB), privacy.Allow)

	taskCount := func(t *testing.T) int {
		t.Helper()

		count, err := suite.Client.DB.Task.Query().Where(task.OwnerIDEQ(orgID)).Count(orgCtx)
		assert.NilError(t, err)

		return count
	}

	notificationCount := func(t *testing.T) int {
		t.Helper()

		count, err := suite.Client.DB.Notification.Query().
			Where(
				notification.OwnerIDEQ(orgID),
				notification.ObjectTypeEQ(taskRuleOrganizationReadyObjectType),
			).
			Count(orgCtx)
		assert.NilError(t, err)

		return count
	}

	t.Run("create mutations fire schema task rules", func(t *testing.T) {
		th.WaitForCondition(t, func() bool {
			tasks, err := suite.Client.DB.Task.Query().Where(task.OwnerIDEQ(orgID)).All(orgCtx)
			if err != nil {
				return false
			}

			keys := make(map[string]struct{}, len(tasks))
			for _, tk := range tasks {
				keys[tk.SourceKey] = struct{}{}
			}

			_, orgRule := keys["organization-"+taskrules.RuleSecureOrganization]
			_, onboardingRule := keys["onboarding-"+taskrules.RuleImportTemplateControls]

			return orgRule && onboardingRule
		}, "organization and onboarding task rules should create suggested tasks")

		exists, err := suite.Client.DB.Task.Query().
			Where(task.IdempotencyKeyEQ(fmt.Sprintf("entityops:organization:%s-%s", orgID, taskrules.RuleSecureOrganization))).
			Exist(orgCtx)
		assert.NilError(t, err)
		assert.Check(t, exists)
	})

	t.Run("organization ready notification emitted", func(t *testing.T) {
		th.WaitForCondition(t, func() bool { return notificationCount(t) == 1 }, "organization ready notification should exist once suggested tasks land")
	})

	t.Run("redelivery does not duplicate tasks", func(t *testing.T) {
		th.WaitForGala(t, setup.Runtime)

		before := taskCount(t)

		entityops.EmitMutation(auth.NewTestContextWithOrgID(user.ID, orgID), []*gala.Gala{setup.Runtime}, entityops.MutationPayload{
			MutationType: generated.TypeOrganization,
			Operation:    entityops.OpCreate,
			EntityID:     orgID,
		})

		th.WaitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(before, taskCount(t)))
		assert.Check(t, is.Equal(1, notificationCount(t)))
	})
}
