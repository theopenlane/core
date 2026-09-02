package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationCreateTrustCenterSubprocessor(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	// Create subprocessors for testing
	subprocessor1 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	subprocessor2 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg2.Owner.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterSubprocessorInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - org owner can create trust center subprocessor",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor1.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US", "CA"},
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "not authorized - view only user cannot create trust center subprocessor",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor1.ID,
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not authorized - different org user cannot create trust center subprocessor",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor2.ID,
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.Client.API,
			ctx:         tcOrg2.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "trust center not found",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor1.ID,
				TrustCenterID:                   lo.ToPtr("non-existent-trust-center-id"),
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "subprocessor not found",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  "non-existent-subprocessor-id",
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.InvalidInputErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenterSubprocessor(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID != "")
			if tc.request.TrustCenterID != nil {
				assert.Check(t, is.Equal(*tc.request.TrustCenterID, *resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.TrustCenterID))
			}
			if tc.request.TrustCenterSubprocessorKindName != nil {
				assert.Check(t, is.Equal(*tc.request.TrustCenterSubprocessorKindName, *resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.TrustCenterSubprocessorKindName))
			}
			assert.Check(t, is.DeepEqual(tc.request.Countries, resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.Countries))

			// Verify subprocessor details are included
			assert.Check(t, resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.Subprocessor.Name != "")
		})
	}

	// Clean up test data
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationCreateTrustCenterSubprocessorAsAnonymousUser(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	// Create a subprocessor for testing
	subprocessor := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Admin.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	testCases := []struct {
		name           string
		request        testclient.CreateTrustCenterSubprocessorInput
		trustCenterID  string
		organizationID string
		client         *testclient.TestClient
		expectedErr    string
	}{
		{
			name: "anonymous user cannot create trust center subprocessor",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor.ID,
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			trustCenterID:  trustCenter.ID,
			organizationID: tcOrg.Owner.OrganizationID,
			client:         suite.Client.API,
			expectedErr:    "could not identify authenticated user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context should fail
			anonCtx := th.CreateAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

			resp, err := tc.client.CreateTrustCenterSubprocessor(anonCtx, tc.request)

			assert.ErrorContains(t, err, tc.expectedErr)
			assert.Check(t, resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID == "")
		})
	}

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenterSubprocessorByID(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	subprocessor := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	// Create a trust center subprocessor using GraphQL mutation
	createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Admin.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"US", "CA"},
	})
	assert.NilError(t, err)
	tcSubprocessor := createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	// Create another trust center subprocessor for different org
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.TrustCenter

	subprocessor2 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg2.Owner.UserCtx, t)

	kind2 := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.Owner.UserCtx, t)

	_, err = suite.Client.API.CreateTrustCenterSubprocessor(tcOrg2.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor2.ID,
		TrustCenterID:                   &trustCenter2.ID,
		TrustCenterSubprocessorKindName: &kind2.Name,
		Countries:                       []string{"EU"},
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
			name:    "happy path - get trust center subprocessor",
			queryID: tcSubprocessor.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Owner.UserCtx,
		},
		{
			name:    "happy path - view only user can get trust center subprocessor",
			queryID: tcSubprocessor.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Member.UserCtx,
		},
		{
			name:    "happy path - anon user",
			queryID: tcSubprocessor.ID,
			client:  suite.Client.API,
			ctx:     th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.OrganizationID),
		},
		{
			name:     "not found - different org user cannot access trust center subprocessor",
			queryID:  tcSubprocessor.ID,
			client:   suite.Client.API,
			ctx:      tcOrg2.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not found - different anonymous user cannot access trust center subprocessor",
			queryID:  tcSubprocessor.ID,
			client:   suite.Client.API,
			ctx:      th.CreateAnonymousTrustCenterContext(trustCenter2.ID, tcOrg2.OrganizationID),
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not found - non-existent ID",
			queryID:  "non-existent-id",
			client:   suite.Client.API,
			ctx:      tcOrg.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterSubprocessorByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterSubprocessor.ID))
			assert.Check(t, resp.TrustCenterSubprocessor.TrustCenterSubprocessorKindName != nil && *resp.TrustCenterSubprocessor.TrustCenterSubprocessorKindName != "")
			assert.Check(t, resp.TrustCenterSubprocessor.Subprocessor.Name != "")
		})
	}

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterSubprocessor(t *testing.T) {
	t.Parallel()
	// Create test data
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	// Create test data
	subprocessor1 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	(&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	(&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	subprocessor4 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Admin.UserCtx, t)
	subprocessor5 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	subprocessor6 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.SuperAdmin.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	// Create another trust center subprocessor for different org
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter2 := tcOrg2.TrustCenter

	subprocessorOtherOrg := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg2.Admin.UserCtx, t)

	kind2 := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.SuperAdmin.UserCtx, t)

	_, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg2.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessorOtherOrg.ID,
		TrustCenterID:                   &trustCenter2.ID,
		TrustCenterSubprocessorKindName: &kind2.Name,
		Countries:                       []string{"EU"},
	})
	assert.NilError(t, err)

	newKind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Admin.UserCtx, t)
	newCountries := []string{"US", "CA", "EU"}

	testCases := []struct {
		name        string
		setupFunc   func() string // Function to create and return the ID of the trust center subprocessor
		request     testclient.UpdateTrustCenterSubprocessorInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - update kind and countries",
			setupFunc: func() string {
				createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
					SubprocessorID:                  subprocessor1.ID,
					TrustCenterSubprocessorKindName: &kind.Name,
					Countries:                       []string{"US"},
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID
			},
			request: testclient.UpdateTrustCenterSubprocessorInput{
				TrustCenterSubprocessorKindName: &newKind.Name,
				Countries:                       newCountries,
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "happy path - append countries",
			setupFunc: func() string {
				createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
					SubprocessorID:                  subprocessor4.ID,
					TrustCenterID:                   &trustCenter.ID,
					TrustCenterSubprocessorKindName: &kind.Name,
					Countries:                       []string{"US"},
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID
			},
			request: testclient.UpdateTrustCenterSubprocessorInput{
				AppendCountries: []string{"MX"},
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name: "happy path - clear countries",
			setupFunc: func() string {
				createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
					SubprocessorID:                  subprocessor5.ID,
					TrustCenterID:                   &trustCenter.ID,
					TrustCenterSubprocessorKindName: &kind.Name,
					Countries:                       []string{"US", "CA"},
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID
			},
			request: testclient.UpdateTrustCenterSubprocessorInput{
				ClearCountries: lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name: "not authorized - view only user cannot update",
			setupFunc: func() string {
				createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
					SubprocessorID:                  subprocessor6.ID,
					TrustCenterID:                   &trustCenter.ID,
					TrustCenterSubprocessorKindName: &kind.Name,
					Countries:                       []string{"US"},
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID
			},
			request: testclient.UpdateTrustCenterSubprocessorInput{
				TrustCenterSubprocessorKindName: &newKind.Name,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not authorized - anon user cannot update",
			setupFunc: func() string {
				subprocessoranon := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
				createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
					SubprocessorID:                  subprocessoranon.ID,
					TrustCenterID:                   &trustCenter.ID,
					TrustCenterSubprocessorKindName: &kind.Name,
					Countries:                       []string{"US"},
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID
			},
			request: testclient.UpdateTrustCenterSubprocessorInput{
				TrustCenterSubprocessorKindName: &newKind.Name,
			},
			client:      suite.Client.API,
			ctx:         th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.OrganizationID),
			expectedErr: "could not identify authenticated user",
		},
		{
			name: "not authorized - different org user cannot update",
			setupFunc: func() string {
				// Create a separate subprocessor for this test to avoid conflicts
				subprocessor7 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
				createResp, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
					SubprocessorID:                  subprocessor7.ID,
					TrustCenterID:                   &trustCenter.ID,
					TrustCenterSubprocessorKindName: &kind.Name,
					Countries:                       []string{"US"},
				})
				assert.NilError(t, err)
				return createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID
			},
			request: testclient.UpdateTrustCenterSubprocessorInput{
				TrustCenterSubprocessorKindName: &newKind.Name,
			},
			client:      suite.Client.API,
			ctx:         tcOrg2.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:      "not found - non-existent ID",
			setupFunc: func() string { return "non-existent-id" },
			request: testclient.UpdateTrustCenterSubprocessorInput{
				TrustCenterSubprocessorKindName: &newKind.Name,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	var createdIDs []string

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			id := tc.setupFunc()
			if tc.expectedErr == "" {
				createdIDs = append(createdIDs, id)
			}

			resp, err := tc.client.UpdateTrustCenterSubprocessor(tc.ctx, id, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(id, resp.UpdateTrustCenterSubprocessor.TrustCenterSubprocessor.ID))

			// Verify specific updates
			if tc.request.TrustCenterSubprocessorKindName != nil {
				assert.Check(t, is.Equal(*tc.request.TrustCenterSubprocessorKindName, *resp.UpdateTrustCenterSubprocessor.TrustCenterSubprocessor.TrustCenterSubprocessorKindName))
			}
			if tc.request.Countries != nil {
				assert.Check(t, is.DeepEqual(tc.request.Countries, resp.UpdateTrustCenterSubprocessor.TrustCenterSubprocessor.Countries))
			}
			if tc.request.ClearCountries != nil && *tc.request.ClearCountries {
				assert.Check(t, is.Len(resp.UpdateTrustCenterSubprocessor.TrustCenterSubprocessor.Countries, 0))
			}
		})
	}

	// Clean up created org
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterSubprocessor(t *testing.T) {
	t.Parallel()
	// Create test data
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	subprocessor1 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	subprocessor2 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	// Create trust center subprocessors to delete
	createResp1, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor1.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"US"},
	})
	assert.NilError(t, err)
	tcSubprocessor1 := createResp1.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	createResp2, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor2.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"CA"},
	})
	assert.NilError(t, err)
	tcSubprocessor2 := createResp2.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	// Create another trust center subprocessor for different org
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.TrustCenter
	subprocessor3 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg2.Owner.UserCtx, t)

	kindAnother := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.Owner.UserCtx, t)

	createResp3, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg2.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor3.ID,
		TrustCenterID:                   &trustCenter2.ID,
		TrustCenterSubprocessorKindName: &kindAnother.Name,
		Countries:                       []string{"EU"},
	})
	assert.NilError(t, err)
	tcSubprocessor3 := createResp3.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "happy path - delete trust center subprocessor",
			id:     tcSubprocessor1.ID,
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name:        "not authorized - view only user cannot delete",
			id:          tcSubprocessor2.ID,
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different org user cannot delete",
			id:          tcSubprocessor3.ID,
			client:      suite.Client.API,
			ctx:         tcOrg.Admin.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not found - non-existent ID",
			id:          "non-existent-id",
			client:      suite.Client.API,
			ctx:         tcOrg.Admin.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterSubprocessor(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.id, resp.DeleteTrustCenterSubprocessor.DeletedID))

			// Verify the trust center subprocessor is deleted
			_, err = tc.client.GetTrustCenterSubprocessorByID(tc.ctx, tc.id)
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)
		})
	}

	// Clean up remaining data
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestQueryTrustCenterSubprocessors(t *testing.T) {
	t.Parallel()
	// Create test data
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	trustCenter := tcOrg.TrustCenter

	subprocessor1 := (&th.SubprocessorBuilder{Client: suite.Client, Description: gofakeit.Sentence()}).MustNew(tcOrg.Owner.UserCtx, t)

	kind := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.Owner.UserCtx, t)

	createLogoUpload := th.LogoFileFunc(t)
	logoFile := createLogoUpload()

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*logoFile})

	subprocessorWithFile, err := suite.Client.API.CreateSubprocessor(tcOrg.Owner.UserCtx, testclient.CreateSubprocessorInput{
		Name:        "Subprocessor With File",
		Description: lo.ToPtr("A subprocessor with a logo file"),
	}, logoFile, nil)
	assert.NilError(t, err)
	assert.Assert(t, subprocessorWithFile != nil)
	assert.Assert(t, subprocessorWithFile.CreateSubprocessor.Subprocessor.ID != "")
	assert.Assert(t, subprocessorWithFile.CreateSubprocessor.Subprocessor.LogoFile.ID != "")

	// Create trust center subprocessors
	_, err = suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.Admin.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor1.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"US"},
	})
	assert.NilError(t, err)

	createResp2, err := suite.Client.API.CreateTrustCenterSubprocessor(tcOrg.SuperAdmin.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessorWithFile.CreateSubprocessor.Subprocessor.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"CA"},
	})
	assert.NilError(t, err)
	tcSubprocessor2 := createResp2.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	// Create another trust center subprocessor for different org
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.TrustCenter

	subprocessor3 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(tcOrg2.Owner.UserCtx, t)

	kindAnother := (&th.CustomTypeEnumBuilder{
		Client:     suite.Client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.Owner.UserCtx, t)

	_, err = suite.Client.API.CreateTrustCenterSubprocessor(tcOrg2.Owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor3.ID,
		TrustCenterID:                   &trustCenter2.ID,
		TrustCenterSubprocessorKindName: &kindAnother.Name,
		Countries:                       []string{"EU"},
	})
	assert.NilError(t, err)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.TrustCenterSubprocessorWhereInput
	}{
		{
			name:            "get all trust center subprocessors for user1",
			client:          suite.Client.API,
			ctx:             tcOrg.Admin.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "get all trust center subprocessors for user2",
			client:          suite.Client.API,
			ctx:             tcOrg2.Admin.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "view only user can see trust center subprocessors",
			client:          suite.Client.API,
			ctx:             tcOrg.Member.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "anonymous user can see trust center subprocessors",
			client:          suite.Client.API,
			ctx:             th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.OrganizationID),
			expectedResults: 2,
		},
		{
			name:   "filter by kind name",
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
			where: &testclient.TrustCenterSubprocessorWhereInput{
				TrustCenterSubprocessorKindName: &kind.Name,
			},
			expectedResults: 2,
		},
		{
			name:   "filter by trust center ID",
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
			where: &testclient.TrustCenterSubprocessorWhereInput{
				TrustCenterID: &trustCenter.ID,
			},
			expectedResults: 2,
		},
		{
			name:   "filter by non-existent kind name",
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
			where: &testclient.TrustCenterSubprocessorWhereInput{
				TrustCenterSubprocessorKindName: lo.ToPtr("Non-existent"),
			},
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Query "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterSubprocessors(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenterSubprocessors.TotalCount))

			if tc.expectedResults > 0 {
				assert.Check(t, is.Len(resp.TrustCenterSubprocessors.Edges, int(tc.expectedResults)))
				// Verify that each result has the expected fields
				for _, edge := range resp.TrustCenterSubprocessors.Edges {
					assert.Check(t, edge.Node.ID != "")
					assert.Check(t, edge.Node.TrustCenterSubprocessorKindName != nil && *edge.Node.TrustCenterSubprocessorKindName != "")
					assert.Check(t, edge.Node.Subprocessor.Name != "")
					assert.Assert(t, edge.Node.Subprocessor.Description != nil)

					if edge.Node.ID == tcSubprocessor2.ID {
						assert.Check(t, *edge.Node.Subprocessor.Description == *tcSubprocessor2.Subprocessor.Description)
						// Verify that the subprocessor with file has logo file details
						assert.Assert(t, edge.Node.Subprocessor.LogoFile != nil)
						assert.Check(t, edge.Node.Subprocessor.LogoFile.PresignedURL != nil)
					}

				}
			}
		})
	}

	// Clean up
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}
