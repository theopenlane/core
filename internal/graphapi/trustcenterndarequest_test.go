package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateTrustCenterNDARequest(t *testing.T) {
	trustCenterNoApproval := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	trustCenterWithApproval := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	_, err := suite.client.api.UpdateTrustCenter(testUser2.UserCtx, trustCenterWithApproval.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	testCases := []struct {
		name           string
		input          testclient.CreateTrustCenterNDARequestInput
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
		expectedStatus enums.TrustCenterNDARequestStatus
	}{
		{
			name: "happy path - no approval required, status should be REQUESTED",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedStatus: enums.TrustCenterNDARequestStatusRequested,
		},
		{
			name: "happy path - approval required, status should be NEEDS_APPROVAL",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterWithApproval.ID,
			},
			client:         suite.client.api,
			ctx:            testUser2.UserCtx,
			expectedStatus: enums.TrustCenterNDARequestStatusNeedsApproval,
		},
		{
			name: "happy path - with company name and reason",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				CompanyName:   lo.ToPtr(gofakeit.Company()),
				Reason:        lo.ToPtr("Need access to security documentation"),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedStatus: enums.TrustCenterNDARequestStatusRequested,
		},
		{
			name: "view only user cannot create",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user cannot create in another org's trust center",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterWithApproval.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid email",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         "invalid-email",
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "validator failed",
		},
		{
			name: "missing first name",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     "",
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenterNoApproval.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "first_name",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenterNDARequest(tc.ctx, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Equal(t, tc.input.FirstName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.FirstName)
			assert.Equal(t, tc.input.LastName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.LastName)
			assert.Equal(t, tc.input.Email, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Email)
			assert.Equal(t, tc.expectedStatus, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

			if tc.input.CompanyName != nil {
				assert.Equal(t, *tc.input.CompanyName, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.CompanyName)
			}

			if tc.input.Reason != nil {
				assert.Equal(t, *tc.input.Reason, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Reason)
			}

			(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenterNoApproval.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenterWithApproval.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestQueryTrustCenterNDARequest(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaRequest, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:     "not found, different org",
			queryID:  ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, invalid id",
			queryID:  ulids.New().String(),
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterNDARequestByID(tc.ctx, tc.queryID)
			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.queryID, resp.TrustCenterNDARequest.ID)
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenterNDARequests(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaRequest1, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaRequest2, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name        string
		client      *testclient.TestClient
		ctx         context.Context
		expectCount int
	}{
		{
			name:        "happy path",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectCount: 2,
		},
		{
			name:        "happy path, view only user",
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectCount: 2,
		},
		{
			name:        "different org, no results",
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllTrustCenterNDARequests(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.expectCount, len(resp.TrustCenterNdaRequests.Edges))
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTrustCenterNDARequest(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	_, err := suite.client.api.UpdateTrustCenter(testUser1.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			NdaApprovalRequired: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	ndaRequest, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Equal(t, enums.TrustCenterNDARequestStatusNeedsApproval, *ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

	testCases := []struct {
		name        string
		input       testclient.UpdateTrustCenterNDARequestInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - update first name",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("UpdatedFirstName"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path - update status to approved",
			input: testclient.UpdateTrustCenterNDARequestInput{
				Status: lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "view only user cannot update",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("ShouldNotUpdate"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "different org cannot update",
			input: testclient.UpdateTrustCenterNDARequestInput{
				FirstName: lo.ToPtr("ShouldNotUpdate"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTrustCenterNDARequest(tc.ctx, ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.input.FirstName != nil {
				assert.Equal(t, *tc.input.FirstName, resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.FirstName)
			}

			if tc.input.Status != nil {
				assert.Equal(t, *tc.input.Status, *resp.UpdateTrustCenterNDARequest.TrustCenterNDARequest.Status)
			}
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateTrustCenterNDARequestAsAnonymousUser(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	otherTrustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	anonCtx := createAnonymousTrustCenterContext(trustCenter.ID, trustCenter.OwnerID)
	wrongTrustCenterAnonCtx := createAnonymousTrustCenterContext(otherTrustCenter.ID, otherTrustCenter.OwnerID)

	testCases := []struct {
		name           string
		input          testclient.CreateTrustCenterNDARequestInput
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
		expectedStatus enums.TrustCenterNDARequestStatus
	}{
		{
			name: "happy path - anonymous user can create NDA request",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenter.ID,
			},
			client:         suite.client.api,
			ctx:            anonCtx,
			expectedStatus: enums.TrustCenterNDARequestStatusRequested,
		},
		{
			name: "anonymous user cannot create NDA request for different trust center",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &otherTrustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         anonCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "anonymous user with wrong trust center context cannot create",
			input: testclient.CreateTrustCenterNDARequestInput{
				FirstName:     gofakeit.FirstName(),
				LastName:      gofakeit.LastName(),
				Email:         gofakeit.Email(),
				TrustCenterID: &trustCenter.ID,
			},
			client:      suite.client.api,
			ctx:         wrongTrustCenterAnonCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenterNDARequest(tc.ctx, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Equal(t, tc.input.FirstName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.FirstName)
			assert.Equal(t, tc.input.LastName, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.LastName)
			assert.Equal(t, tc.input.Email, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Email)
			assert.Equal(t, tc.expectedStatus, *resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status)

			(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: otherTrustCenter.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTrustCenterNDARequestDuplicateEmail(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	email := gofakeit.Email()

	originalRequest, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         email,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	t.Run("duplicate email returns existing request with REQUESTED status", func(t *testing.T) {
		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("duplicate email returns existing request with NEEDS_APPROVAL status", func(t *testing.T) {
		_, err := suite.client.api.UpdateTrustCenterNDARequest(testUser1.UserCtx, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
			Status: lo.ToPtr(enums.TrustCenterNDARequestStatusNeedsApproval),
		})
		assert.NilError(t, err)

		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("duplicate email returns existing request with APPROVED status", func(t *testing.T) {
		_, err := suite.client.api.UpdateTrustCenterNDARequest(testUser1.UserCtx, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
			Status: lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
		})
		assert.NilError(t, err)

		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("different email creates new request", func(t *testing.T) {
		resp, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         gofakeit.Email(),
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Assert(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID != resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)

		(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	})

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTrustCenterNDARequest(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	ndaRequest1, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaRequest2, err := suite.client.api.CreateTrustCenterNDARequest(testUser1.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "view only user cannot delete",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "different org cannot delete",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path - delete nda request",
			idToDelete: ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "invalid id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterNDARequest(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.idToDelete, resp.DeleteTrustCenterNDARequest.DeletedID)
		})
	}

	(&Cleanup[*generated.TrustCenterNDARequestDeleteOne]{client: suite.client.db.TrustCenterNDARequest, ID: ndaRequest2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
