package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationCreateTrustCenterSubprocessor(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)

	// Create subprocessors for testing
	subprocessor1 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	subprocessor2 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	kind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)

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
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name: "not authorized - view only user cannot create trust center subprocessor",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor1.ID,
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - different org user cannot create trust center subprocessor",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor2.ID,
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.client.api,
			ctx:         tcOrg2.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "trust center not found",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  subprocessor1.ID,
				TrustCenterID:                   lo.ToPtr("non-existent-trust-center-id"),
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "subprocessor not found",
			request: testclient.CreateTrustCenterSubprocessorInput{
				SubprocessorID:                  "non-existent-subprocessor-id",
				TrustCenterID:                   &trustCenter.ID,
				TrustCenterSubprocessorKindName: &kind.Name,
				Countries:                       []string{"US"},
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: invalidInputErrorMsg,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationCreateTrustCenterSubprocessorAsAnonymousUser(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	// Create a subprocessor for testing
	subprocessor := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.admin.UserCtx, t)

	kind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)

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
			organizationID: tcOrg.owner.OrganizationID,
			client:         suite.client.api,
			expectedErr:    "could not identify authenticated user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create anonymous trust center context should fail
			anonCtx := createAnonymousTrustCenterContext(tc.trustCenterID, tc.organizationID)

			resp, err := tc.client.CreateTrustCenterSubprocessor(anonCtx, tc.request)

			assert.ErrorContains(t, err, tc.expectedErr)
			assert.Check(t, resp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor.ID == "")
		})
	}

	// Clean up
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestQueryTrustCenterSubprocessorByID(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	subprocessor := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	kind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)

	// Create a trust center subprocessor using GraphQL mutation
	createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.admin.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"US", "CA"},
	})
	assert.NilError(t, err)
	tcSubprocessor := createResp.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	// Create another trust center subprocessor for different org
	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	subprocessor2 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	kind2 := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.owner.UserCtx, t)

	_, err = suite.client.api.CreateTrustCenterSubprocessor(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client:  suite.client.api,
			ctx:     tcOrg.owner.UserCtx,
		},
		{
			name:    "happy path - view only user can get trust center subprocessor",
			queryID: tcSubprocessor.ID,
			client:  suite.client.api,
			ctx:     tcOrg.member.UserCtx,
		},
		{
			name:    "happy path - anon user",
			queryID: tcSubprocessor.ID,
			client:  suite.client.api,
			ctx:     createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID),
		},
		{
			name:     "not found - different org user cannot access trust center subprocessor",
			queryID:  tcSubprocessor.ID,
			client:   suite.client.api,
			ctx:      tcOrg2.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found - different anonymous user cannot access trust center subprocessor",
			queryID:  tcSubprocessor.ID,
			client:   suite.client.api,
			ctx:      createAnonymousTrustCenterContext(trustCenter2.ID, tcOrg2.organizationID),
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found - non-existent ID",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      tcOrg.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationUpdateTrustCenterSubprocessor(t *testing.T) {
	t.Parallel()
	// Create test data
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	trustCenter := tcOrg.trustCenter

	// Create test data
	subprocessor1 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	(&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	(&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	subprocessor4 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.admin.UserCtx, t)
	subprocessor5 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	subprocessor6 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.superAdmin.UserCtx, t)

	kind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)

	// Create another trust center subprocessor for different org
	tcOrg2 := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	trustCenter2 := tcOrg2.trustCenter

	subprocessorOtherOrg := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg2.admin.UserCtx, t)

	kind2 := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.superAdmin.UserCtx, t)

	_, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessorOtherOrg.ID,
		TrustCenterID:                   &trustCenter2.ID,
		TrustCenterSubprocessorKindName: &kind2.Name,
		Countries:                       []string{"EU"},
	})
	assert.NilError(t, err)

	newKind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.admin.UserCtx, t)
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
				createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name: "happy path - append countries",
			setupFunc: func() string {
				createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name: "happy path - clear countries",
			setupFunc: func() string {
				createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name: "not authorized - view only user cannot update",
			setupFunc: func() string {
				createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - anon user cannot update",
			setupFunc: func() string {
				subprocessoranon := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client:      suite.client.api,
			ctx:         createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID),
			expectedErr: "could not identify authenticated user",
		},
		{
			name: "not authorized - different org user cannot update",
			setupFunc: func() string {
				// Create a separate subprocessor for this test to avoid conflicts
				subprocessor7 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
				createResp, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client:      suite.client.api,
			ctx:         tcOrg2.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:      "not found - non-existent ID",
			setupFunc: func() string { return "non-existent-id" },
			request: testclient.UpdateTrustCenterSubprocessorInput{
				TrustCenterSubprocessorKindName: &newKind.Name,
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterSubprocessor(t *testing.T) {
	t.Parallel()
	// Create test data
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	subprocessor1 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	subprocessor2 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	kind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)

	// Create trust center subprocessors to delete
	createResp1, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor1.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"US"},
	})
	assert.NilError(t, err)
	tcSubprocessor1 := createResp1.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	createResp2, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor2.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"CA"},
	})
	assert.NilError(t, err)
	tcSubprocessor2 := createResp2.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	// Create another trust center subprocessor for different org
	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter
	subprocessor3 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	kindAnother := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.owner.UserCtx, t)

	createResp3, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name:        "not authorized - view only user cannot delete",
			id:          tcSubprocessor2.ID,
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different org user cannot delete",
			id:          tcSubprocessor3.ID,
			client:      suite.client.api,
			ctx:         tcOrg.admin.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not found - non-existent ID",
			id:          "non-existent-id",
			client:      suite.client.api,
			ctx:         tcOrg.admin.UserCtx,
			expectedErr: notFoundErrorMsg,
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
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}

	// Clean up remaining data
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestQueryTrustCenterSubprocessors(t *testing.T) {
	t.Parallel()
	// Create test data
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	trustCenter := tcOrg.trustCenter

	subprocessor1 := (&SubprocessorBuilder{client: suite.client, Description: gofakeit.Sentence()}).MustNew(tcOrg.owner.UserCtx, t)

	kind := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg.owner.UserCtx, t)

	createLogoUpload := logoFileFunc(t)
	logoFile := createLogoUpload()

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*logoFile})

	subprocessorWithFile, err := suite.client.api.CreateSubprocessor(tcOrg.owner.UserCtx, testclient.CreateSubprocessorInput{
		Name:        "Subprocessor With File",
		Description: lo.ToPtr("A subprocessor with a logo file"),
	}, logoFile, nil)
	assert.NilError(t, err)
	assert.Assert(t, subprocessorWithFile != nil)
	assert.Assert(t, subprocessorWithFile.CreateSubprocessor.Subprocessor.ID != "")
	assert.Assert(t, subprocessorWithFile.CreateSubprocessor.Subprocessor.LogoFile.ID != "")

	// Create trust center subprocessors
	_, err = suite.client.api.CreateTrustCenterSubprocessor(tcOrg.admin.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessor1.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"US"},
	})
	assert.NilError(t, err)

	createResp2, err := suite.client.api.CreateTrustCenterSubprocessor(tcOrg.superAdmin.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
		SubprocessorID:                  subprocessorWithFile.CreateSubprocessor.Subprocessor.ID,
		TrustCenterID:                   &trustCenter.ID,
		TrustCenterSubprocessorKindName: &kind.Name,
		Countries:                       []string{"CA"},
	})
	assert.NilError(t, err)
	tcSubprocessor2 := createResp2.CreateTrustCenterSubprocessor.TrustCenterSubprocessor

	// Create another trust center subprocessor for different org
	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	subprocessor3 := (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)

	kindAnother := (&CustomTypeEnumBuilder{
		client:     suite.client,
		ObjectType: "trust_center_subprocessor",
	}).MustNew(tcOrg2.owner.UserCtx, t)

	_, err = suite.client.api.CreateTrustCenterSubprocessor(tcOrg2.owner.UserCtx, testclient.CreateTrustCenterSubprocessorInput{
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
			client:          suite.client.api,
			ctx:             tcOrg.admin.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "get all trust center subprocessors for user2",
			client:          suite.client.api,
			ctx:             tcOrg2.admin.UserCtx,
			expectedResults: 1,
		},
		{
			name:            "view only user can see trust center subprocessors",
			client:          suite.client.api,
			ctx:             tcOrg.member.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "anonymous user can see trust center subprocessors",
			client:          suite.client.api,
			ctx:             createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID),
			expectedResults: 2,
		},
		{
			name:   "filter by kind name",
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
			where: &testclient.TrustCenterSubprocessorWhereInput{
				TrustCenterSubprocessorKindName: &kind.Name,
			},
			expectedResults: 2,
		},
		{
			name:   "filter by trust center ID",
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
			where: &testclient.TrustCenterSubprocessorWhereInput{
				TrustCenterID: &trustCenter.ID,
			},
			expectedResults: 2,
		},
		{
			name:   "filter by non-existent kind name",
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
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
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}
