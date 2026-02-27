package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryIdentityHolder(t *testing.T) {
	ih := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: ih.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path, with api token",
			queryID: ih.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path, with pat",
			queryID: ih.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found, no access",
			queryID:  ih.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, invalid id",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
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

	(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: ih.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryIdentityHolders(t *testing.T) {
	ih1 := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ih2 := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name   string
		client *testclient.TestClient
		ctx    context.Context
	}{
		{
			name:   "happy path",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "happy path, view only user",
			client: suite.client.api,
			ctx:    viewOnlyUser.UserCtx,
		},
		{
			name:   "happy path, using api token",
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name:   "happy path, using pat",
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "another user, no identity holders should be returned",
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllIdentityHolders(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}

	(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: ih1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: ih2.ID}).MustDelete(testUser1.UserCtx, t)
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
				OwnerID:  &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "not authorized, view only user",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field, no email",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
			},
			expectedErr: "value is less than the required length",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name: "missing required field, no name",
			request: testclient.CreateIdentityHolderInput{
				Email: gofakeit.Email(),
			},
			expectedErr: "value is less than the required length",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name: "invalid email",
			request: testclient.CreateIdentityHolderInput{
				FullName: gofakeit.Name(),
				Email:    "not-an-email",
			},
			expectedErr: "mail: missing '@' or angle-addr",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
		},
		{
			name: "invalid phone number",
			request: testclient.CreateIdentityHolderInput{
				FullName:    gofakeit.Name(),
				Email:       gofakeit.Email(),
				PhoneNumber: lo.ToPtr("not a phone number"),
			},
			expectedErr: rout.InvalidField("phone_number").Error(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
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
				assert.Check(t, is.Equal(enums.IdentityHolderTypeEmployee, ih.IdentityHolderType))
			} else {
				assert.Check(t, is.Equal(*tc.request.IdentityHolderType, ih.IdentityHolderType))
			}

			if tc.request.Status == nil {
				assert.Check(t, is.Equal(enums.UserStatusActive, ih.Status))
			} else {
				assert.Check(t, is.Equal(*tc.request.Status, ih.Status))
			}

			assert.Check(t, ih.IsActive)

			(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: ih.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateIdentityHolder(t *testing.T) {
	ih := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update email",
			request: testclient.UpdateIdentityHolderInput{
				Email: lo.ToPtr(gofakeit.Email()),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update department and team",
			request: testclient.UpdateIdentityHolderInput{
				Department: lo.ToPtr(gofakeit.JobDescriptor()),
				Team:       lo.ToPtr(gofakeit.AppName()),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update status",
			request: testclient.UpdateIdentityHolderInput{
				Status: &enums.UserStatusInactive,
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update identity holder type",
			request: testclient.UpdateIdentityHolderInput{
				IdentityHolderType: &enums.IdentityHolderTypeContractor,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update title and location",
			request: testclient.UpdateIdentityHolderInput{
				Title:    lo.ToPtr(gofakeit.JobTitle()),
				Location: lo.ToPtr(gofakeit.City()),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "not authorized, view only user",
			request: testclient.UpdateIdentityHolderInput{
				FullName: lo.ToPtr(gofakeit.Name()),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not found, no access",
			request: testclient.UpdateIdentityHolderInput{
				FullName: lo.ToPtr(gofakeit.Name()),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "invalid phone number",
			request: testclient.UpdateIdentityHolderInput{
				PhoneNumber: lo.ToPtr("not a phone number"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: rout.InvalidField("phone_number").Error(),
		},
		{
			name: "invalid email",
			request: testclient.UpdateIdentityHolderInput{
				Email: lo.ToPtr("not-an-email"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
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

	(&Cleanup[*generated.IdentityHolderDeleteOne]{client: suite.client.db.IdentityHolder, ID: ih.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteIdentityHolder(t *testing.T) {
	ih1 := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ih2 := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ih3 := (&IdentityHolderBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete, no access",
			idToDelete:  ih1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: ih1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  ih1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: ih2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using pat",
			idToDelete: ih3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
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
