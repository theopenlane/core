package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func TestMarkNotificationsAsRead(t *testing.T) {
	// Setup test data (initializes testUser1, testUser2, etc.)
	ctx := context.Background()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	suite.setupTestData(ctx, t)

	// Test case 1: Successfully mark multiple notifications as read
	t.Run("MarkMultipleNotificationsAsRead", func(t *testing.T) {
		// Create test notifications directly using ent client
		notif1, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("test.event").
			SetTitle("Test Notification 1").
			SetBody("Test body 1").
			Save(ctx)
		require.NoError(t, err)
		require.Nil(t, notif1.ReadAt, "Notification 1 should be unread initially")

		notif2, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("test.event").
			SetTitle("Test Notification 2").
			SetBody("Test body 2").
			Save(ctx)
		require.NoError(t, err)
		require.Nil(t, notif2.ReadAt, "Notification 2 should be unread initially")

		notif3, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("test.event").
			SetTitle("Test Notification 3").
			SetBody("Test body 3").
			Save(ctx)
		require.NoError(t, err)
		require.Nil(t, notif3.ReadAt, "Notification 3 should be unread initially")

		// Test the bulk mark as read mutation with user context
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)
		resp, err := suite.client.api.MarkNotificationsAsRead(userCtx,
			[]string{notif1.ID, notif2.ID, notif3.ID})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify response contains all the IDs
		assert.Len(t, resp.MarkNotificationsAsRead.ReadIDs, 3)

		// Verify all IDs are in the response
		readIDsMap := make(map[string]bool)
		for _, id := range resp.MarkNotificationsAsRead.ReadIDs {
			require.NotNil(t, id)
			readIDsMap[*id] = true
		}
		assert.True(t, readIDsMap[notif1.ID], "Response should contain notif1 ID")
		assert.True(t, readIDsMap[notif2.ID], "Response should contain notif2 ID")
		assert.True(t, readIDsMap[notif3.ID], "Response should contain notif3 ID")

		// Verify notifications were actually updated in the database
		updated1, err := suite.client.db.Notification.Get(ctx, notif1.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated1.ReadAt, "Notification 1 should be marked as read")

		updated2, err := suite.client.db.Notification.Get(ctx, notif2.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated2.ReadAt, "Notification 2 should be marked as read")

		updated3, err := suite.client.db.Notification.Get(ctx, notif3.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated3.ReadAt, "Notification 3 should be marked as read")

		// Cleanup
		_, err = suite.client.db.Notification.Delete().
			Where().
			Exec(ctx)
		require.NoError(t, err)
	})

	// Test case 2: Mark single notification as read
	t.Run("MarkSingleNotificationAsRead", func(t *testing.T) {
		// Create a single test notification
		notif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("single.test").
			SetTitle("Single Test Notification").
			SetBody("Single test body").
			Save(ctx)
		require.NoError(t, err)
		require.Nil(t, notif.ReadAt)

		// Mark as read
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)
		resp, err := suite.client.api.MarkNotificationsAsRead(userCtx, []string{notif.ID})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify response
		assert.Len(t, resp.MarkNotificationsAsRead.ReadIDs, 1)
		require.NotNil(t, resp.MarkNotificationsAsRead.ReadIDs[0])
		assert.Equal(t, notif.ID, *resp.MarkNotificationsAsRead.ReadIDs[0])

		// Verify database update
		updated, err := suite.client.db.Notification.Get(ctx, notif.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.ReadAt)

		// Cleanup
		err = suite.client.db.Notification.DeleteOne(notif).Exec(ctx)
		require.NoError(t, err)
	})

	// Test case 3: Mark already-read notifications again (idempotent)
	t.Run("MarkAlreadyReadNotifications", func(t *testing.T) {
		// Create notification
		notif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("already.read").
			SetTitle("Already Read Notification").
			SetBody("Already read body").
			Save(ctx)
		require.NoError(t, err)

		// Mark as read first time
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)
		resp1, err := suite.client.api.MarkNotificationsAsRead(userCtx, []string{notif.ID})
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// Get the first read timestamp
		updated1, err := suite.client.db.Notification.Get(ctx, notif.ID)
		require.NoError(t, err)
		require.NotNil(t, updated1.ReadAt)
		firstReadTime := updated1.ReadAt

		// Mark as read again (should be idempotent)
		resp2, err := suite.client.api.MarkNotificationsAsRead(userCtx, []string{notif.ID})
		require.NoError(t, err)
		require.NotNil(t, resp2)

		// Verify the operation succeeded
		assert.Len(t, resp2.MarkNotificationsAsRead.ReadIDs, 1)

		// Get the second read timestamp
		updated2, err := suite.client.db.Notification.Get(ctx, notif.ID)
		require.NoError(t, err)
		require.NotNil(t, updated2.ReadAt)

		// The read_at timestamp should have been updated (verify it's not nil)
		// Both timestamps should be set after marking as read
		assert.NotNil(t, firstReadTime)
		assert.NotNil(t, updated2.ReadAt)

		// Cleanup
		err = suite.client.db.Notification.DeleteOne(notif).Exec(ctx)
		require.NoError(t, err)
	})

	// Test case 4: Empty array
	t.Run("MarkEmptyArray", func(t *testing.T) {
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)
		resp, err := suite.client.api.MarkNotificationsAsRead(userCtx, []string{})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return empty array
		assert.Len(t, resp.MarkNotificationsAsRead.ReadIDs, 0)
	})

	// Test case 5: Non-existent notification IDs
	t.Run("MarkNonExistentNotifications", func(t *testing.T) {
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)

		// Use non-existent IDs
		resp, err := suite.client.api.MarkNotificationsAsRead(userCtx,
			[]string{"01NONEXISTENT1111111111", "01NONEXISTENT2222222222"})

		// Should succeed but not update anything (0 rows affected)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should still return the IDs that were requested
		assert.Len(t, resp.MarkNotificationsAsRead.ReadIDs, 2)
	})

	// Test case 6: Mix of organization and user notifications
	t.Run("MarkMixedNotificationTypes", func(t *testing.T) {
		// Create user notification
		userNotif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("user.event").
			SetTitle("User Notification").
			SetBody("User notification body").
			Save(ctx)
		require.NoError(t, err)

		// Create organization notification
		orgNotif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetNotificationType(enums.NotificationTypeOrganization).
			SetObjectType("org.event").
			SetTitle("Org Notification").
			SetBody("Org notification body").
			Save(ctx)
		require.NoError(t, err)

		// Mark both as read
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)
		resp, err := suite.client.api.MarkNotificationsAsRead(userCtx,
			[]string{userNotif.ID, orgNotif.ID})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify both were updated
		assert.Len(t, resp.MarkNotificationsAsRead.ReadIDs, 2)

		updatedUser, err := suite.client.db.Notification.Get(ctx, userNotif.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedUser.ReadAt)

		updatedOrg, err := suite.client.db.Notification.Get(ctx, orgNotif.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedOrg.ReadAt)

		// Cleanup
		err = suite.client.db.Notification.DeleteOne(userNotif).Exec(ctx)
		require.NoError(t, err)
		err = suite.client.db.Notification.DeleteOne(orgNotif).Exec(ctx)
		require.NoError(t, err)
	})

	// Test case 7: Different notification topics
	t.Run("MarkDifferentTopics", func(t *testing.T) {
		// Create notifications with different topics
		taskNotif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("task.created").
			SetTitle("Task Notification").
			SetBody("Task created").
			SetTopic(enums.NotificationTopicTaskAssignment).
			Save(ctx)
		require.NoError(t, err)

		approvalNotif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("approval.required").
			SetTitle("Approval Notification").
			SetBody("Approval required").
			SetTopic(enums.NotificationTopicApproval).
			Save(ctx)
		require.NoError(t, err)

		// Mark both as read
		userCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID)
		resp, err := suite.client.api.MarkNotificationsAsRead(userCtx,
			[]string{taskNotif.ID, approvalNotif.ID})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify both were updated
		assert.Len(t, resp.MarkNotificationsAsRead.ReadIDs, 2)

		updatedTask, err := suite.client.db.Notification.Get(ctx, taskNotif.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedTask.ReadAt)
		assert.Equal(t, enums.NotificationTopicTaskAssignment, updatedTask.Topic)

		updatedApproval, err := suite.client.db.Notification.Get(ctx, approvalNotif.ID)
		require.NoError(t, err)
		assert.NotNil(t, updatedApproval.ReadAt)
		assert.Equal(t, enums.NotificationTopicApproval, updatedApproval.Topic)

		// Cleanup
		err = suite.client.db.Notification.DeleteOne(taskNotif).Exec(ctx)
		require.NoError(t, err)
		err = suite.client.db.Notification.DeleteOne(approvalNotif).Exec(ctx)
		require.NoError(t, err)
	})
}

func TestUpdateNotification(t *testing.T) {
	// Setup test data (initializes testUser1, testUser2, etc.)
	ctx := context.Background()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)
	suite.setupTestData(ctx, t)

	t.Run("UpdateSingleNotification", func(t *testing.T) {
		// Create a test notification
		notif, err := suite.client.db.Notification.Create().
			SetOwnerID(testUser1.OrganizationID).
			SetUserID(testUser1.ID).
			SetNotificationType(enums.NotificationTypeUser).
			SetObjectType("update.test").
			SetTitle("Update Test Notification").
			SetBody("Update test body").
			Save(ctx)
		require.NoError(t, err)
		require.Nil(t, notif.ReadAt)

		// Update the notification directly via ent client
		updatedNotif, err := suite.client.db.Notification.UpdateOne(notif).
			SetTags([]string{"updated", "test"}).
			Save(ctx)
		require.NoError(t, err)
		require.NotNil(t, updatedNotif)

		// Verify the notification was updated
		assert.Equal(t, notif.ID, updatedNotif.ID)
		assert.Contains(t, updatedNotif.Tags, "updated")
		assert.Contains(t, updatedNotif.Tags, "test")

		// Cleanup
		err = suite.client.db.Notification.DeleteOne(notif).Exec(ctx)
		require.NoError(t, err)
	})
}
