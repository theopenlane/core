//go:build test

package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/taskrules"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/pkg/gala"
)

const taskRuleOrganizationReadyObjectType = "organization.ready"

func TestTaskRuleListenersRealMutations(t *testing.T) {
	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	setup, err := graphapi.SetupListenerRuntime(suite.galaRuntime, hooks.TaskRuleListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	onboarding, err := suite.client.db.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Task Rule Listener Co",
	}).Save(allowCtx)
	assert.NilError(t, err)

	orgID := onboarding.OrganizationID
	orgCtx := privacy.DecisionContext(setContext(auth.NewTestContextWithOrgID(user.ID, orgID), suite.client.db), privacy.Allow)

	taskCount := func(t *testing.T) int {
		t.Helper()

		count, err := suite.client.db.Task.Query().Where(task.OwnerIDEQ(orgID)).Count(orgCtx)
		assert.NilError(t, err)

		return count
	}

	notificationCount := func(t *testing.T) int {
		t.Helper()

		count, err := suite.client.db.Notification.Query().
			Where(
				notification.OwnerIDEQ(orgID),
				notification.ObjectTypeEQ(taskRuleOrganizationReadyObjectType),
			).
			Count(orgCtx)
		assert.NilError(t, err)

		return count
	}

	t.Run("create mutations fire schema task rules", func(t *testing.T) {
		waitForCondition(t, func() bool {
			tasks, err := suite.client.db.Task.Query().Where(task.OwnerIDEQ(orgID)).All(orgCtx)
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

		exists, err := suite.client.db.Task.Query().
			Where(task.IdempotencyKeyEQ(fmt.Sprintf("entityops:organization:%s-%s", orgID, taskrules.RuleSecureOrganization))).
			Exist(orgCtx)
		assert.NilError(t, err)
		assert.Check(t, exists)
	})

	t.Run("organization ready notification emitted", func(t *testing.T) {
		waitForCondition(t, func() bool { return notificationCount(t) == 1 }, "organization ready notification should exist once suggested tasks land")
	})

	t.Run("redelivery does not duplicate tasks", func(t *testing.T) {
		waitForGala(t, setup.Runtime)

		before := taskCount(t)

		entityops.EmitMutation(auth.NewTestContextWithOrgID(user.ID, orgID), []*gala.Gala{setup.Runtime}, entityops.MutationPayload{
			MutationType: generated.TypeOrganization,
			Operation:    entityops.OpCreate,
			EntityID:     orgID,
		})

		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(before, taskCount(t)))
		assert.Check(t, is.Equal(1, notificationCount(t)))
	})
}
