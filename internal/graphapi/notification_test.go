package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateNotification(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateNotificationInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, create notification as system admin",
			request: testclient.CreateNotificationInput{
				NotificationType: enums.NotificationTypeOrganization,
				ObjectType:       "program",
				Title:            "Test Notification",
				Body:             "This is a test notification body",
				OwnerID:          &testUser1.OrganizationID,
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "not authorized, create notification as member",
			request: testclient.CreateNotificationInput{
				NotificationType: enums.NotificationTypeOrganization,
				ObjectType:       "program",
				Title:            "Test Notification",
				Body:             "This is a test notification body",
				OwnerID:          &viewOnlyUser.OrganizationID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateNotification(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CreateNotification.Notification.Title != "")
			assert.Check(t, is.Equal(tc.request.Title, resp.CreateNotification.Notification.Title))
			assert.Check(t, is.Equal(tc.request.Body, resp.CreateNotification.Notification.Body))
			assert.Check(t, is.Equal(tc.request.ObjectType, resp.CreateNotification.Notification.ObjectType))

			(&Cleanup[*generated.NotificationDeleteOne]{client: suite.client.db.Notification, ID: resp.CreateNotification.Notification.ID}).MustDelete(adminUser.UserCtx, t)
		})
	}
}
