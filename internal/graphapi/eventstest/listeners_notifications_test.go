//go:build test

package eventstest_test

import (
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/notifications"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestNotificationListenerTaskAssignment(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	assigneeCtx := th.SetContext(org.Member.UserCtx, suite.Client.DB)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, notifications.Listeners())
	assert.NilError(t, err)
	t.Cleanup(setup.Teardown)

	_, err = suite.Client.API.CreateTask(org.Owner.UserCtx, testclient.CreateTaskInput{
		Title:      "notification listener task",
		AssigneeID: &org.Member.ID,
	})
	assert.NilError(t, err)

	waitForGala(t, setup.Runtime)

	waitForCondition(t, func() bool {
		exists, err := suite.Client.DB.Notification.Query().
			Where(
				notification.UserID(org.Member.ID),
				notification.TopicEQ(enums.NotificationTopicTaskAssignment),
			).
			Exist(assigneeCtx)

		return err == nil && exists
	}, "assignee should receive a task assignment notification")
}

func TestNotificationListenerStandardUpdate(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	allowCtx := th.SetContext(org.Owner.UserCtx, suite.Client.DB)
	adminCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, notifications.Listeners())
	assert.NilError(t, err)
	t.Cleanup(setup.Teardown)

	std := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	assert.Assert(t, std.SystemOwned)

	assert.NilError(t, suite.Client.DB.Standard.UpdateOneID(std.ID).SetRevision("v1.0.0").Exec(adminCtx))

	ctrl := (&th.ControlBuilder{Client: suite.Client, StandardID: std.ID}).MustNew(org.Owner.UserCtx, t)
	assert.NilError(t, suite.Client.DB.Control.UpdateOneID(ctrl.ID).SetReferenceFrameworkRevision("v1.0.0").Exec(allowCtx))

	_, err = suite.Client.API.UpdateStandard(th.SharedSystemAdminUser.UserCtx, std.ID, testclient.UpdateStandardInput{
		Revision: lo.ToPtr("v2.0.0"),
	}, nil, nil)
	assert.NilError(t, err)

	waitForGala(t, setup.Runtime)

	waitForCondition(t, func() bool {
		exists, err := suite.Client.DB.Notification.Query().
			Where(
				notification.UserID(org.Owner.ID),
				notification.TopicEQ(enums.NotificationTopicStandardUpdate),
			).
			Exist(allowCtx)

		return err == nil && exists
	}, "org owner should receive a standard update notification")
}
