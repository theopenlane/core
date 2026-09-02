package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationTransferOrganizationOwnership(t *testing.T) {
	// Create an existing member user to transfer ownership to
	existingMember := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	memberRole := enums.RoleMember.String()
	membershipID := (&th.OrgMemberBuilder{
		Client: suite.Client,
		UserID: existingMember.ID,
		Role:   memberRole,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	// Create a non-member user (exists but not in the org)
	nonMemberUser := (&th.UserBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// Create a different organization owner for negative test cases
	otherOwner := suite.UserBuilder(context.Background(), t)

	testCases := []struct {
		name           string
		newOwnerEmail  string
		client         *testclient.TestClient
		ctx            context.Context
		expectedInvite bool
		expectedErr    string
		checkTransfer  bool
	}{
		{
			name:           "happy path, transfer to existing member",
			newOwnerEmail:  existingMember.Email,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedInvite: false,
			checkTransfer:  true,
		},
		{
			name:           "happy path, transfer to non-member (sends invitation)",
			newOwnerEmail:  nonMemberUser.Email,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedInvite: true,
		},
		{
			name:           "happy path, transfer to new user (sends invitation)",
			newOwnerEmail:  "new-owner@theopenlane.io",
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedInvite: true,
		},
		{
			name:          "not owner, permission denied",
			newOwnerEmail: "someone@theopenlane.io",
			client:        suite.Client.API,
			ctx:           th.SharedViewOnlyUser.UserCtx,
			expectedErr:   th.NotAuthorizedErrorMsg,
		},
		{
			name:          "different org owner, no access",
			newOwnerEmail: "someone@theopenlane.io",
			client:        suite.Client.API,
			ctx:           auth.NewTestContextWithOrgID(otherOwner.ID, th.SharedTestUser1.OrganizationID),
			expectedErr:   th.NotFoundErrorMsg,
		},
		{
			name:          "invalid email",
			newOwnerEmail: "invalid-email",
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			expectedErr:   th.InvalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Transfer "+tc.name, func(t *testing.T) {
			resp, err := tc.client.TransferOrganizationOwnership(tc.ctx, tc.newOwnerEmail)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, resp.TransferOrganizationOwnership.Organization.ID))
			assert.Check(t, is.Equal(tc.expectedInvite, resp.TransferOrganizationOwnership.InvitationSent))

			// If checkTransfer is true, verify the ownership was actually transferred
			if tc.checkTransfer {
				allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

				// Verify new owner has OWNER role
				newOwnerMembership, err := suite.Client.DB.OrgMembership.Query().
					Where(
						orgmembership.OrganizationID(th.SharedTestUser1.OrganizationID),
						orgmembership.UserID(existingMember.ID),
					).
					Only(allowCtx)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(enums.RoleOwner, newOwnerMembership.Role))
				// verify new owner is sso exempt
				assert.Check(t, newOwnerMembership.SSOExempt)

				// Verify old owner has SUPER_ADMIN role
				oldOwnerMembership, err := suite.Client.DB.OrgMembership.Query().
					Where(
						orgmembership.OrganizationID(th.SharedTestUser1.OrganizationID),
						orgmembership.UserID(th.SharedTestUser1.ID),
					).
					Only(allowCtx)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(enums.RoleSuperAdmin, oldOwnerMembership.Role))
				// verify old owner is no longer sso exempt
				assert.Check(t, !oldOwnerMembership.SSOExempt)

				// Transfer back to original owner for other tests
				// Use auth context with proper org ID
				transferBackCtx := auth.NewTestContextWithOrgID(existingMember.ID, th.SharedTestUser1.OrganizationID)
				_, err = suite.Client.API.TransferOrganizationOwnership(transferBackCtx, th.SharedTestUser1.UserInfo.Email)
				assert.NilError(t, err)
			}
		})
	}

	// th.Cleanup
	(&th.Cleanup[*generated.OrgMembershipDeleteOne]{Client: suite.Client.DB.OrgMembership, ID: membershipID.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
