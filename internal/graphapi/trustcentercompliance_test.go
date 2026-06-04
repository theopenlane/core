package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/ulids"
)

func TestMutationCreateTrustCenterCompliance(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	// Create test data - standards and trust centers
	standard1 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	standard2 := (&StandardBuilder{client: suite.client}).MustNew(sharedSystemAdminUser.UserCtx, t)
	publicStandard := (&StandardBuilder{client: suite.client, IsPublic: true}).MustNew(sharedSystemAdminUser.UserCtx, t)

	trustCenter1 := tcOrg.trustCenter
	trustCenter2 := tcOrg2.trustCenter

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterComplianceInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input with standard and trust center determined by org",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID: standard1.ID,
			},
			client: suite.client.api,
			ctx:    tcOrg.owner.UserCtx,
		},
		{
			name: "happy path, with trust center and tags",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"compliance", "test"},
			},
			client: suite.client.api,
			ctx:    tcOrg.superAdmin.UserCtx,
		},
		{
			name: "happy path, using public standard",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    publicStandard.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"public", "compliance"},
			},
			client: suite.client.api,
			ctx:    tcOrg.admin.UserCtx,
		},
		{
			name: "happy path, using personal access token",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"pat", "test"},
			},
			client: tcOrg.adminPatClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"api", "token"},
			},
			client: tcOrg.adminApiClient,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, different org standard",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard2.ID,
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "user not authorized, different org trust center",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter2.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: testclient.CreateTrustCenterComplianceInput{
				Tags:          []string{"missing", "standard"},
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid standard id",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    "invalid-id",
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid trust center id",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: lo.ToPtr("invalid-id"),
			},
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenterCompliance(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the created trust center compliance
			assert.Check(t, resp.CreateTrustCenterCompliance.TrustCenterCompliance.ID != "")

			expectedTags := []string{}
			if tc.request.Tags != nil {
				expectedTags = tc.request.Tags
			}
			assert.Check(t, is.DeepEqual(expectedTags, resp.CreateTrustCenterCompliance.TrustCenterCompliance.Tags))

			// Verify standard relationship exists
			assert.Check(t, resp.CreateTrustCenterCompliance.TrustCenterCompliance.Standard.Name != "")

			// cleanup the created trust center compliance
			ctx := tc.ctx
			if tc.client != suite.client.api {
				ctx = tcOrg.owner.UserCtx
			}

			(&Cleanup[*generated.TrustCenterComplianceDeleteOne]{client: suite.client.db.TrustCenterCompliance, ID: resp.CreateTrustCenterCompliance.TrustCenterCompliance.ID}).MustDelete(ctx, t)
		})
	}

	// Cleanup test data
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestQueryTrustCenterCompliance(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAllUserTypes())

	// Create test data
	standard := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	trustCenter := tcOrg.trustCenter

	compliance := (&TrustCenterComplianceBuilder{
		client:        suite.client,
		StandardID:    standard.ID,
		TrustCenterID: trustCenter.ID,
		Tags:          []string{"test", "query"},
	}).MustNew(tcOrg.owner.UserCtx, t)

	users2 := suite.seedFreshOrgUsers(t)

	// Create compliance for different org
	standardOther := (&StandardBuilder{client: suite.client}).MustNew(users2.owner.UserCtx, t)
	complianceOther := (&TrustCenterComplianceBuilder{
		client:     suite.client,
		StandardID: standardOther.ID,
	}).MustNew(users2.owner.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: compliance.ID,
			client:  suite.client.api,
			ctx:     tcOrg.admin.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: compliance.ID,
			client:  suite.client.api,
			ctx:     tcOrg.member.UserCtx,
		},
		{
			name:    "happy path, anonymous user",
			queryID: compliance.ID,
			client:  suite.client.api,
			ctx:     createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.owner.OrganizationID),
		},
		{
			name:    "happy path using personal access token",
			queryID: compliance.ID,
			client:  tcOrg.adminPatClient,
			ctx:     context.Background(),
		},
		{
			name:    "happy path using api token",
			queryID: compliance.ID,
			client:  tcOrg.adminApiClient,
			ctx:     context.Background(),
		},
		{
			name:     "trust center compliance not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      tcOrg.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "trust center compliance not found, using not authorized user",
			queryID:  compliance.ID,
			client:   suite.client.api,
			ctx:      users2.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "trust center compliance not found, different org",
			queryID:  complianceOther.ID,
			client:   suite.client.api,
			ctx:      tcOrg.owner.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterComplianceByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterCompliance.ID))
			assert.Check(t, resp.TrustCenterCompliance.Standard.Name != "")
			assert.Check(t, is.DeepEqual([]string{"test", "query"}, resp.TrustCenterCompliance.Tags))
		})
	}

	// Cleanup
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestUpdateTrustCenterComplianceUpdatesFgaTuples(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	standard1 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	standard2 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)

	resp, err := suite.client.api.CreateTrustCenterCompliance(tcOrg.owner.UserCtx, testclient.CreateTrustCenterComplianceInput{
		TrustCenterID: &trustCenter.ID,
		StandardID:    standard1.ID,
	})
	assert.NilError(t, err)
	complianceID := resp.CreateTrustCenterCompliance.TrustCenterCompliance.ID

	checkTuple := func(standardID string, shouldExist bool) {
		ac := fgax.AccessCheck{
			SubjectID:   trustCenter.ID,
			SubjectType: "trust_center",
			ObjectID:    standardID,
			ObjectType:  "standard",
			Relation:    "associated_with",
		}
		exists, err := suite.client.db.Authz.CheckAccess(tcOrg.owner.UserCtx, ac)
		assert.NilError(t, err)
		if shouldExist {
			assert.Assert(t, exists)
		} else {
			assert.Assert(t, !exists)
		}
	}

	checkTuple(standard1.ID, true)

	_, err = suite.client.api.UpdateTrustCenterCompliance(tcOrg.owner.UserCtx, complianceID, testclient.UpdateTrustCenterComplianceInput{
		StandardID: &standard2.ID,
	})
	assert.NilError(t, err)

	checkTuple(standard1.ID, false)
	checkTuple(standard2.ID, true)

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestQueryTrustCenterCompliances(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAPIClients())
	tcOrg2 := createFreshOrgWithTrustCenter(t)

	// Create test data
	standard1 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	standard2 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	trustCenter := tcOrg.trustCenter

	// Create multiple compliances for tcOrg.owner
	countOrgOwned := 2
	orgOwnedComplianceIDs := []string{}
	for i := range countOrgOwned {
		standardID := standard1.ID
		if i == 1 {
			standardID = standard2.ID
		}
		compliance := (&TrustCenterComplianceBuilder{
			client:        suite.client,
			StandardID:    standardID,
			TrustCenterID: trustCenter.ID,
			Tags:          []string{"org", "test"},
		}).MustNew(tcOrg.owner.UserCtx, t)
		orgOwnedComplianceIDs = append(orgOwnedComplianceIDs, compliance.ID)
	}

	// Create compliance for different org
	standardOther := (&StandardBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)
	trustCenterOther := tcOrg2.trustCenter
	(&TrustCenterComplianceBuilder{
		client:        suite.client,
		StandardID:    standardOther.ID,
		TrustCenterID: trustCenterOther.ID,
		Tags:          []string{"other", "org"},
	}).MustNew(tcOrg2.owner.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path, org user should get all org owned compliances",
			client:          suite.client.api,
			ctx:             tcOrg.owner.UserCtx,
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             tcOrg.member.UserCtx,
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, anonymous user",
			client:          suite.client.api,
			ctx:             createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.owner.OrganizationID),
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, using api token",
			client:          tcOrg.adminApiClient,
			ctx:             context.Background(),
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, using pat",
			client:          tcOrg.adminPatClient,
			ctx:             context.Background(),
			expectedResults: countOrgOwned,
		},
		{
			name:            "another user, should see their own compliance",
			client:          suite.client.api,
			ctx:             tcOrg2.owner.UserCtx,
			expectedResults: 1, // only their own compliance
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllTrustCenterCompliances(tc.ctx)
			assert.NilError(t, err)

			assert.Check(t, is.Len(resp.TrustCenterCompliances.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(int64(tc.expectedResults), resp.TrustCenterCompliances.TotalCount))

			// under the max results in tests (10), has next should be false
			assert.Check(t, !resp.TrustCenterCompliances.PageInfo.HasNextPage)

			// Verify each compliance has required fields
			for _, edge := range resp.TrustCenterCompliances.Edges {
				assert.Check(t, edge.Node.ID != "")
				assert.Check(t, edge.Node.Standard.Name != "")
				assert.Check(t, edge.Node.Tags != nil)
			}
		})
	}

	// Cleanup
	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterCompliance(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t, withAPIClients())

	// Create test data for deletion
	standard1 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	standard2 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	standard3 := (&StandardBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
	trustCenter1 := tcOrg.trustCenter

	// Create compliance objects to delete
	compliance1 := (&TrustCenterComplianceBuilder{
		client:        suite.client,
		StandardID:    standard1.ID,
		TrustCenterID: trustCenter1.ID,
		Tags:          []string{"delete", "test1"},
	}).MustNew(tcOrg.owner.UserCtx, t)

	compliance2 := (&TrustCenterComplianceBuilder{
		client:        suite.client,
		StandardID:    standard2.ID,
		TrustCenterID: trustCenter1.ID,
		Tags:          []string{"delete", "test2"},
	}).MustNew(tcOrg.owner.UserCtx, t)

	compliance3 := (&TrustCenterComplianceBuilder{
		client:        suite.client,
		StandardID:    standard3.ID,
		TrustCenterID: trustCenter1.ID,
		Tags:          []string{"delete", "test3"},
	}).MustNew(tcOrg.owner.UserCtx, t)

	tcOrg2 := createFreshOrgWithTrustCenter(t, withAPIClients())

	// Create compliance for different org
	standardOther := (&StandardBuilder{client: suite.client}).MustNew(tcOrg2.owner.UserCtx, t)
	trustCenterOther := tcOrg2.trustCenter
	complianceOther := (&TrustCenterComplianceBuilder{
		client:        suite.client,
		StandardID:    standardOther.ID,
		TrustCenterID: trustCenterOther.ID,
		Tags:          []string{"other", "org"},
	}).MustNew(tcOrg2.owner.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path, delete trust center compliance",
			idToDelete: compliance1.ID,
			client:     suite.client.api,
			ctx:        tcOrg.owner.UserCtx,
		},
		{
			name:        "not authorized, different org compliance api token",
			idToDelete:  compliance2.ID,
			client:      tcOrg2.adminApiClient,
			ctx:         context.Background(),
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: compliance2.ID,
			client:     tcOrg.adminPatClient,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: compliance3.ID,
			client:     tcOrg.adminApiClient,
			ctx:        context.Background(),
		},
		{
			name:        "not authorized, different org compliance via jwt",
			idToDelete:  complianceOther.ID,
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, view only user",
			idToDelete:  complianceOther.ID, // use different org compliance to test permissions
			client:      suite.client.api,
			ctx:         tcOrg.member.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center compliance not found, invalid ID",
			idToDelete:  "invalid-id",
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center compliance not found, non-existent ID",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         tcOrg.owner.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterCompliance(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustCenterCompliance.DeletedID))

			// Verify the trust center compliance is actually deleted
			_, err = tc.client.GetTrustCenterComplianceByID(tc.ctx, tc.idToDelete)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}
