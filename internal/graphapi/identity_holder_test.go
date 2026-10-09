package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestIdentityHolder_IsOpenlaneUser(t *testing.T) {
	org := suite.SeedOrgOwner(t)

	client := suite.Client.API

	orgID := org.Owner.OrganizationID

	ctx := org.Owner.UserCtx

	t.Cleanup(func() {
		_, err := client.DeleteOrganization(ctx, orgID)
		assert.NilError(t, err)
	})

	users := []th.TestUserDetails{th.SharedTestUser2, th.SharedViewOnlyUser}

	orgMemberIDs := make([]string, 0, len(users))

	// add 2 users to the new org
	for _, user := range users {
		member, err := client.AddUserToOrgWithRole(ctx, testclient.CreateOrgMembershipInput{
			OrganizationID: orgID,
			UserID:         user.ID,
			Role:           &enums.RoleMember,
		})
		assert.NilError(t, err)
		assert.Assert(t, member != nil)

		orgMemberIDs = append(orgMemberIDs, member.CreateOrgMembership.OrgMembership.ID)
	}

	// create identity holders, make 2 use the emails of the previously added org users and a completely new email
	// then verify isOpenlaneUser is set correctly
	idsToCreate := []struct {
		email          string
		isOpenlaneUser bool
	}{
		{
			isOpenlaneUser: true,
			email:          users[0].UserInfo.Email,
		},
		{
			isOpenlaneUser: true,
			email:          users[1].UserInfo.Email,
		},
		{
			isOpenlaneUser: false,
			email:          gofakeit.Email(),
		},
	}

	for _, val := range idsToCreate {
		holder, err := client.CreateIdentityHolder(ctx, testclient.CreateIdentityHolderInput{
			FullName: gofakeit.Name(),
			Email:    val.email,
			OwnerID:  &orgID,
		})
		assert.NilError(t, err)
		assert.Assert(t, holder != nil)

		assert.Check(t, is.Equal(val.isOpenlaneUser, *holder.CreateIdentityHolder.IdentityHolder.IsOpenlaneUser))
	}

	_, err := client.RemoveUserFromOrg(ctx, orgMemberIDs[0])
	assert.NilError(t, err)

	holders, err := client.GetIdentityHolders(ctx, nil, nil, nil, nil, nil, &testclient.IdentityHolderWhereInput{
		IsOpenlaneUser: lo.ToPtr(true),
	})
	assert.NilError(t, err)

	// we created 2 holders with isOpenlaneUser set to true and removed a member from the org
	// so we should be left with just 1
	assert.Check(t, is.Len(holders.IdentityHolders.Edges, 1))
}

func TestIdentityHolder_AutoSetOpenlaneUserFromOrgMembership(t *testing.T) {
	org := suite.SeedOrgOwner(t)

	client := suite.Client.API

	orgID := org.Owner.OrganizationID

	ctx := org.Owner.UserCtx

	t.Cleanup(func() {
		_, err := client.DeleteOrganization(ctx, orgID)
		assert.NilError(t, err)
	})

	_, err := client.AddUserToOrgWithRole(ctx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgID,
		UserID:         th.SharedTestUser2.ID,
		Role:           lo.ToPtr(enums.RoleMember),
	})

	assert.NilError(t, err)

	// create a new identity holder with an email that is not a member of the org yet
	// then add the org member using SharedTestUser1

	holder, err := client.CreateIdentityHolder(ctx, testclient.CreateIdentityHolderInput{
		Email:    th.SharedTestUser1.UserInfo.Email,
		FullName: gofakeit.Name(),
		OwnerID:  &orgID,
	})
	assert.NilError(t, err)

	assert.Check(t, is.Equal(false, *holder.CreateIdentityHolder.IdentityHolder.IsOpenlaneUser))

	// now add the user to the org
	_, err = client.AddUserToOrgWithRole(ctx, testclient.CreateOrgMembershipInput{
		OrganizationID: orgID,
		Role:           lo.ToPtr(enums.RoleMember),
		UserID:         th.SharedTestUser1.ID,
	})
	assert.NilError(t, err)

	// verify the identity holder got toggled
	newHolder, err := client.GetIdentityHolderByID(ctx, holder.CreateIdentityHolder.IdentityHolder.ID)
	assert.NilError(t, err)

	assert.Check(t, is.Equal(true, *newHolder.IdentityHolder.IsOpenlaneUser))
}

func TestQueryIdentityHolder(t *testing.T) {
	ih := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: ih.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: ih.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path, with api token",
			queryID: ih.ID,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, with pat",
			queryID: ih.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found, no access",
			queryID:  ih.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not found, invalid id",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetIdentityHolderByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.queryID, resp.IdentityHolder.ID))
		})
	}

	(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: ih.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryIdentityHolders(t *testing.T) {
	ih1 := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	ih2 := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name   string
		client *testclient.TestClient
		ctx    context.Context
	}{
		{
			name:   "happy path",
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:   "happy path, view only user",
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:   "happy path, using api token",
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name:   "happy path, using pat",
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "another user, no identity holders should be returned",
			client: suite.Client.API,
			ctx:    th.SharedTestUser2.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllIdentityHolders(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}

	(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: ih1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: ih2.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateIdentityHolder(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateIdentityHolderInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateIdentityHolderInput{
				FullName:           gofakeit.Name(),
				Email:              gofakeit.Email(),
				AlternateEmail:     lo.ToPtr(gofakeit.Email()),
				PhoneNumber:        lo.ToPtr(gofakeit.Phone()),
				Title:              lo.ToPtr(gofakeit.JobTitle()),
				Department:         lo.ToPtr(gofakeit.JobDescriptor()),
				Team:               lo.ToPtr(gofakeit.AppName()),
				Location:           lo.ToPtr(gofakeit.City()),
				IdentityHolderType: &enums.IdentityHolderTypeContractor,
				Status:             &enums.UserStatusOnboarding,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
				OwnerID:  &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "not authorized, view only user",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing required field, no email",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
			},
			expectedErr: "value is less than the required length",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
		},
		{
			name: "missing required field, no name",
			request: testclient.CreateIdentityHolderInput{
				Email: gofakeit.Email(),
			},
			expectedErr: "value is less than the required length",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
		},
		{
			name: "invalid email",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    "not-an-email",
			},
			expectedErr: "mail: missing '@' or angle-addr",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
		},
		{
			name: "invalid phone number",
			request: testclient.CreateIdentityHolderInput{
				FullName:    gofakeit.Name(),
				Email:       gofakeit.Email(),
				PhoneNumber: lo.ToPtr("not a phone number"),
			},
			expectedErr: rout.InvalidField("phone_number").Error(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateIdentityHolder(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			ih := resp.CreateIdentityHolder.IdentityHolder

			assert.Check(t, is.Equal(tc.request.FullName, ih.FullName))
			assert.Check(t, is.Equal(tc.request.Email, ih.Email))

			if tc.request.Title != nil {
				assert.Check(t, is.Equal(*tc.request.Title, *ih.Title))
			}

			if tc.request.Department != nil {
				assert.Check(t, is.Equal(*tc.request.Department, *ih.Department))
			}

			if tc.request.Team != nil {
				assert.Check(t, is.Equal(*tc.request.Team, *ih.Team))
			}

			if tc.request.Location != nil {
				assert.Check(t, is.Equal(*tc.request.Location, *ih.Location))
			}

			if tc.request.PhoneNumber != nil {
				assert.Check(t, is.Equal(*tc.request.PhoneNumber, *ih.PhoneNumber))
			}

			if tc.request.AlternateEmail != nil {
				assert.Check(t, is.Equal(*tc.request.AlternateEmail, *ih.AlternateEmail))
			}

			// defaults
			if tc.request.IdentityHolderType == nil {
				assert.Check(t, is.Equal(enums.IdentityHolderTypeUnspecified, ih.IdentityHolderType))
			} else {
				assert.Check(t, is.Equal(*tc.request.IdentityHolderType, ih.IdentityHolderType))
			}

			if tc.request.Status == nil {
				assert.Check(t, is.Equal(enums.UserStatusActive, ih.Status))
			} else {
				assert.Check(t, is.Equal(*tc.request.Status, ih.Status))
			}

			assert.Check(t, ih.IsActive)

			(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: ih.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateIdentityHolder(t *testing.T) {
	ih := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateIdentityHolderInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name",
			request: testclient.UpdateIdentityHolderInput{
				FullName: lo.ToPtr(gofakeit.Name()),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update email",
			request: testclient.UpdateIdentityHolderInput{
				Email: lo.ToPtr(gofakeit.Email()),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update department and team",
			request: testclient.UpdateIdentityHolderInput{
				Department: lo.ToPtr(gofakeit.JobDescriptor()),
				Team:       lo.ToPtr(gofakeit.AppName()),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update status",
			request: testclient.UpdateIdentityHolderInput{
				Status: &enums.UserStatusInactive,
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update identity holder type",
			request: testclient.UpdateIdentityHolderInput{
				IdentityHolderType: &enums.IdentityHolderTypeContractor,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update title and location",
			request: testclient.UpdateIdentityHolderInput{
				Title:    lo.ToPtr(gofakeit.JobTitle()),
				Location: lo.ToPtr(gofakeit.City()),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "not authorized, view only user",
			request: testclient.UpdateIdentityHolderInput{
				FullName: lo.ToPtr(gofakeit.Name()),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not found, no access",
			request: testclient.UpdateIdentityHolderInput{
				FullName: lo.ToPtr(gofakeit.Name()),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "invalid phone number",
			request: testclient.UpdateIdentityHolderInput{
				PhoneNumber: lo.ToPtr("not a phone number"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: rout.InvalidField("phone_number").Error(),
		},
		{
			name: "invalid email",
			request: testclient.UpdateIdentityHolderInput{
				Email: lo.ToPtr("not-an-email"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "validator failed for field",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateIdentityHolder(tc.ctx, ih.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			updated := resp.UpdateIdentityHolder.IdentityHolder

			if tc.request.FullName != nil {
				assert.Check(t, is.Equal(*tc.request.FullName, updated.FullName))
			}

			if tc.request.Email != nil {
				assert.Check(t, is.Equal(*tc.request.Email, updated.Email))
			}

			if tc.request.Department != nil {
				assert.Check(t, is.Equal(*tc.request.Department, *updated.Department))
			}

			if tc.request.Team != nil {
				assert.Check(t, is.Equal(*tc.request.Team, *updated.Team))
			}

			if tc.request.Title != nil {
				assert.Check(t, is.Equal(*tc.request.Title, *updated.Title))
			}

			if tc.request.Location != nil {
				assert.Check(t, is.Equal(*tc.request.Location, *updated.Location))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, updated.Status))
			}

			if tc.request.IdentityHolderType != nil {
				assert.Check(t, is.Equal(*tc.request.IdentityHolderType, updated.IdentityHolderType))
			}

			if tc.request.PhoneNumber != nil {
				assert.Check(t, is.Equal(*tc.request.PhoneNumber, *updated.PhoneNumber))
			}
		})
	}

	(&th.Cleanup[*generated.IdentityHolderDeleteOne]{Client: suite.Client.DB.IdentityHolder, ID: ih.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteIdentityHolder(t *testing.T) {
	ih1 := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	ih2 := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	ih3 := (&th.IdentityHolderBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete, view only user",
			idToDelete:  ih1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete, no access",
			idToDelete:  ih1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: ih1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  ih1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: ih2.ID,
			client:     suite.Client.APIWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using pat",
			idToDelete: ih3.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteIdentityHolder(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteIdentityHolder.DeletedID))
		})
	}
}
