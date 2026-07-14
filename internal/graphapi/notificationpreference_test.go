package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateNotificationPreference(t *testing.T) {
	orgMember := (&OrgMemberBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	orgMemberCtx := auth.NewTestContextWithOrgID(orgMember.UserID, sharedTestUser1.OrganizationID)

	testCases := []struct {
		name        string
		request     testclient.CreateNotificationPreferenceInput
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, create preference for self",
			request: testclient.CreateNotificationPreferenceInput{
				Channel: enums.ChannelEmail,
				OwnerID: &sharedTestUser1.OrganizationID,
				Priority: &enums.PriorityMedium,
				UserID:  orgMember.UserID,
			},
			ctx: orgMemberCtx,
		},
		{
			name: "not authorized, user not in org",
			request: testclient.CreateNotificationPreferenceInput{
				Channel: enums.ChannelSlack,
				OwnerID: &sharedTestUser1.OrganizationID,
				Priority: &enums.PriorityMedium,
				UserID:  sharedTestUser2.ID,
			},
			ctx:         sharedTestUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.CreateNotificationPreference(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			preference := resp.GetCreateNotificationPreference().GetNotificationPreference()
			assert.Assert(t, preference != nil)
			assert.Check(t, is.Equal(tc.request.UserID, preference.UserID))
			assert.Check(t, is.Equal(tc.request.Channel, preference.Channel))

			(&Cleanup[*generated.NotificationPreferenceDeleteOne]{client: suite.client.db.NotificationPreference, ID: preference.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: orgMember.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}
