package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
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
				OwnerID:          &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
		{
			name: "not authorized, create notification as member",
			request: testclient.CreateNotificationInput{
				NotificationType: enums.NotificationTypeOrganization,
				ObjectType:       "program",
				Title:            "Test Notification",
				Body:             "This is a test notification body",
				OwnerID:          &th.SharedViewOnlyUser.OrganizationID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "not authorized, create notification as org owner/admin",
			request: testclient.CreateNotificationInput{
				NotificationType: enums.NotificationTypeOrganization,
				ObjectType:       "program",
				Title:            "Test Notification",
				Body:             "This is a test notification body",
				OwnerID:          &th.SharedTestUser1.OrganizationID,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

			(&th.Cleanup[*generated.NotificationDeleteOne]{Client: suite.Client.DB.Notification, ID: resp.CreateNotification.Notification.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
		})
	}
}
