package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
)

func TestQueryTrustCenterNDARequest(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate(), th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	ndaRequest, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
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
			client:  suite.Client.API,
			ctx:     tcOrg.Admin.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Member.UserCtx,
		},
		{
			name:     "not found, different org",
			queryID:  ndaRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not found, invalid id",
			queryID:  ulids.New().String(),
			client:   suite.Client.API,
			ctx:      tcOrg.SuperAdmin.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenterNDARequests(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter

	_, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
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
			client:      suite.Client.API,
			ctx:         tcOrg.Admin.UserCtx,
			expectCount: 2,
		},
		{
			name:        "happy path, view only user",
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectCount: 2,
		},
		{
			name:        "different org, no results",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationCreateTrustCenterNDARequestDuplicateEmail(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter

	email := gofakeit.Email()

	originalRequest, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         email,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	t.Run("duplicate email returns existing request with REQUESTED status", func(t *testing.T) {
		resp, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("duplicate email returns existing request with NEEDS_APPROVAL status", func(t *testing.T) {
		_, err := suite.Client.API.UpdateTrustCenterNDARequest(tcOrg.Owner.UserCtx, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
			Status: lo.ToPtr(enums.TrustCenterNDARequestStatusNeedsApproval),
		})
		assert.NilError(t, err)

		resp, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("duplicate email returns existing request with APPROVED status", func(t *testing.T) {
		_, err := suite.Client.API.UpdateTrustCenterNDARequest(tcOrg.Owner.UserCtx, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, testclient.UpdateTrustCenterNDARequestInput{
			Status: lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
		})
		assert.NilError(t, err)

		resp, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Equal(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	t.Run("different email creates new request", func(t *testing.T) {
		resp, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         gofakeit.Email(),
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)
		assert.Assert(t, originalRequest.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID != resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	})

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterNDARequest(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate(), th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	ndaRequest1, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaRequest2, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
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
			name:        "member cannot delete",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "different org cannot delete",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "admin can delete",
			idToDelete: ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:     suite.Client.API,
			ctx:        tcOrg.Admin.UserCtx,
		},
		{
			name:       "super admin can delete",
			idToDelete: ndaRequest2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:     suite.Client.API,
			ctx:        tcOrg.SuperAdmin.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  ndaRequest1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "invalid id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationRevokeNDARequestsRemovesDocAccess(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter

	protectedDoc := (&th.TrustCenterDocBuilder{
		Client:        suite.Client,
		TrustCenterID: trustCenter.ID,
		Visibility:    enums.TrustCenterDocumentVisibilityProtected,
	}).MustNew(tcOrg.Owner.UserCtx, t)

	// create two NDA requests
	ndaReq1, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaReq2, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.Email(),
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	ndaReqIDs := []string{
		ndaReq1.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
		ndaReq2.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID,
	}

	// simulate signed NDA access by creating anonymous contexts with subject IDs
	// matching the trust center JWT subject format: AnonTrustCenterJWTPrefix + ndaRequestID
	anonCtxs := make([]context.Context, 0, len(ndaReqIDs))

	for _, id := range ndaReqIDs {
		subjectID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, id)

		anonCaller := auth.NewTrustCenterCaller(trustCenter.OwnerID, subjectID, "Anonymous User", "")
		anonCtx := th.NewAnonTrustCenterCtxFromCaller(anonCaller, trustCenter.ID)

		anonCtxs = append(anonCtxs, anonCtx)

		tuple := fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   subjectID,
			SubjectType: "user",
			ObjectID:    trustCenter.ID,
			ObjectType:  "trust_center",
			Relation:    "nda_signed",
		})

		_, err := suite.Client.DB.Authz.WriteTupleKeys(tcOrg.Owner.UserCtx, []fgax.TupleKey{tuple}, nil)
		assert.NilError(t, err)
	}

	// verify both anon users CAN see protected doc file details
	for _, ctx := range anonCtxs {
		resp, err := suite.Client.API.GetTrustCenterDocByID(ctx, protectedDoc.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp.TrustCenterDoc.OriginalFile != nil, "anon user should see file details before revocation")
	}

	// now we are revoking the NDA requests
	revokeResp, err := suite.Client.API.DeleteBulkTrustCenterNDARequest(tcOrg.Owner.UserCtx, ndaReqIDs)
	assert.NilError(t, err)
	assert.Equal(t, len(ndaReqIDs), len(revokeResp.DeleteBulkTrustCenterNDARequest.DeletedIDs))

	// anon users should not be able to see the file after we revoked their access
	for _, anonCtx := range anonCtxs {
		resp, err := suite.Client.API.GetTrustCenterDocByID(anonCtx, protectedDoc.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp.TrustCenterDoc.OriginalFile == nil, "anon user should NOT see file details after revocation")
	}

	// verify the NDA requests are actually deleted
	for _, id := range ndaReqIDs {
		_, err = suite.Client.API.GetTrustCenterNDARequestByID(tcOrg.Owner.UserCtx, id)
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestMutationBulkDeleteTrustCenterNDARequest(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	trustCenter := tcOrg.TrustCenter

	count := 5
	// members cannot bulk delete anymore
	expectedDeletedItems := 0

	ids := make([]string, 0, count)
	for range count {
		resp, err := suite.Client.API.CreateTrustCenterNDARequest(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         gofakeit.Email(),
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)

		ids = append(ids, resp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID)
	}

	resp, err := suite.Client.API.DeleteBulkTrustCenterNDARequest(tcOrg.Member.UserCtx, ids)
	assert.NilError(t, err)
	assert.Equal(t, expectedDeletedItems, len(resp.DeleteBulkTrustCenterNDARequest.DeletedIDs))

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
