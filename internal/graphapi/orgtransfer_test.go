package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationTransferOrganizationOwnership(t *testing.T) {
	// Create an existing member user to transfer ownership to
	existingMember := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	memberRole := enums.RoleMember.String()
	membershipID := (&OrgMemberBuilder{
		client: suite.client,
		UserID: existingMember.ID,
		Role:   memberRole,
	}).MustNew(testUser1.UserCtx, t)

	// Create a non-member user (exists but not in the org)
	nonMemberUser := (&UserBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create a different organization owner for negative test cases
	otherOwner := suite.userBuilder(context.Background(), t)

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
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedInvite: false,
			checkTransfer:  true,
		},
		{
			name:           "happy path, transfer to non-member (sends invitation)",
			newOwnerEmail:  nonMemberUser.Email,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedInvite: true,
		},
		{
			name:           "happy path, transfer to new user (sends invitation)",
			newOwnerEmail:  "new-owner@theopenlane.io",
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedInvite: true,
		},
		{
			name:          "not owner, permission denied",
			newOwnerEmail: "someone@theopenlane.io",
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
			expectedErr:   notAuthorizedErrorMsg,
		},
		{
			name:          "different org owner, no access",
			newOwnerEmail: "someone@theopenlane.io",
			client:        suite.client.api,
			ctx:           auth.NewTestContextWithOrgID(otherOwner.ID, testUser1.OrganizationID),
			expectedErr:   notFoundErrorMsg,
		},
		{
			name:          "invalid email",
			newOwnerEmail: "invalid-email",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedErr:   "email domain not allowed in organization",
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
			assert.Check(t, is.Equal(testUser1.OrganizationID, resp.TransferOrganizationOwnership.Organization.ID))
			assert.Check(t, is.Equal(tc.expectedInvite, resp.TransferOrganizationOwnership.InvitationSent))

			// If checkTransfer is true, verify the ownership was actually transferred
			if tc.checkTransfer {
				allowCtx := privacy.DecisionContext(context.Background(), privacy.Allow)

				// Verify new owner has OWNER role
				newOwnerMembership, err := suite.client.db.OrgMembership.Query().
					Where(
						orgmembership.OrganizationID(testUser1.OrganizationID),
						orgmembership.UserID(existingMember.ID),
					).
					Only(allowCtx)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(enums.RoleOwner, newOwnerMembership.Role))

				// Verify old owner has ADMIN role
				oldOwnerMembership, err := suite.client.db.OrgMembership.Query().
					Where(
						orgmembership.OrganizationID(testUser1.OrganizationID),
						orgmembership.UserID(testUser1.ID),
					).
					Only(allowCtx)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(enums.RoleAdmin, oldOwnerMembership.Role))

				// Transfer back to original owner for other tests
				// Use auth context with proper org ID
				transferBackCtx := auth.NewTestContextWithOrgID(existingMember.ID, testUser1.OrganizationID)
				_, err = suite.client.api.TransferOrganizationOwnership(transferBackCtx, testUser1.UserInfo.Email)
				assert.NilError(t, err)
			}
		})
	}

	// Cleanup
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: membershipID.ID}).MustDelete(testUser1.UserCtx, t)
}
