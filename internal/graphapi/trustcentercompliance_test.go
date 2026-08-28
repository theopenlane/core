package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/ulids"
)

func TestMutationCreateTrustCenterCompliance(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	// Create test data - standards and trust centers
	standard1 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	standard2 := (&th.StandardBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	publicStandard := (&th.StandardBuilder{Client: suite.Client, IsPublic: true}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	trustCenter1 := tcOrg.TrustCenter
	trustCenter2 := tcOrg2.TrustCenter

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
			client: suite.Client.API,
			ctx:    tcOrg.Owner.UserCtx,
		},
		{
			name: "happy path, with trust center and tags",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"compliance", "test"},
			},
			client: suite.Client.API,
			ctx:    tcOrg.SuperAdmin.UserCtx,
		},
		{
			name: "happy path, using public standard",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    publicStandard.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"public", "compliance"},
			},
			client: suite.Client.API,
			ctx:    tcOrg.Admin.UserCtx,
		},
		{
			name: "happy path, using personal access token",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"pat", "test"},
			},
			client: tcOrg.AdminPatClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
				Tags:          []string{"api", "token"},
			},
			client: tcOrg.AdminAPIClient,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, different org standard",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard2.ID,
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "user not authorized, different org trust center",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter2.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing required field",
			request: testclient.CreateTrustCenterComplianceInput{
				Tags:          []string{"missing", "standard"},
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid standard id",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    "invalid-id",
				TrustCenterID: &trustCenter1.ID,
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "invalid trust center id",
			request: testclient.CreateTrustCenterComplianceInput{
				StandardID:    standard1.ID,
				TrustCenterID: lo.ToPtr("invalid-id"),
			},
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
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
			if tc.client != suite.Client.API {
				ctx = tcOrg.Owner.UserCtx
			}

			(&th.Cleanup[*generated.TrustCenterComplianceDeleteOne]{Client: suite.Client.DB.TrustCenterCompliance, ID: resp.CreateTrustCenterCompliance.TrustCenterCompliance.ID}).MustDelete(ctx, t)
		})
	}

	// th.Cleanup test data
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenterCompliance(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAllUserTypes())

	// Create test data
	standard := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	trustCenter := tcOrg.TrustCenter

	compliance := (&th.TrustCenterComplianceBuilder{
		Client:        suite.Client,
		StandardID:    standard.ID,
		TrustCenterID: trustCenter.ID,
		Tags:          []string{"test", "query"},
	}).MustNew(tcOrg.Owner.UserCtx, t)

	users2 := suite.SeedFreshOrgUsers(t)

	// Create compliance for different org
	standardOther := (&th.StandardBuilder{Client: suite.Client}).MustNew(users2.Owner.UserCtx, t)
	complianceOther := (&th.TrustCenterComplianceBuilder{
		Client:     suite.Client,
		StandardID: standardOther.ID,
	}).MustNew(users2.Owner.UserCtx, t)

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
			client:  suite.Client.API,
			ctx:     tcOrg.Admin.UserCtx,
		},
		{
			name:    "happy path, view only user",
			queryID: compliance.ID,
			client:  suite.Client.API,
			ctx:     tcOrg.Member.UserCtx,
		},
		{
			name:    "happy path, anonymous user",
			queryID: compliance.ID,
			client:  suite.Client.API,
			ctx:     th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.Owner.OrganizationID),
		},
		{
			name:    "happy path using personal access token",
			queryID: compliance.ID,
			client:  tcOrg.AdminPatClient,
			ctx:     context.Background(),
		},
		{
			name:    "happy path using api token",
			queryID: compliance.ID,
			client:  tcOrg.AdminAPIClient,
			ctx:     context.Background(),
		},
		{
			name:     "trust center compliance not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      tcOrg.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "trust center compliance not found, using not authorized user",
			queryID:  compliance.ID,
			client:   suite.Client.API,
			ctx:      users2.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "trust center compliance not found, different org",
			queryID:  complianceOther.ID,
			client:   suite.Client.API,
			ctx:      tcOrg.Owner.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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

	// th.Cleanup
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestUpdateTrustCenterComplianceUpdatesFgaTuples(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	standard1 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	standard2 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)

	resp, err := suite.Client.API.CreateTrustCenterCompliance(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterComplianceInput{
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
		exists, err := suite.Client.DB.Authz.CheckAccess(tcOrg.Owner.UserCtx, ac)
		assert.NilError(t, err)
		if shouldExist {
			assert.Assert(t, exists)
		} else {
			assert.Assert(t, !exists)
		}
	}

	checkTuple(standard1.ID, true)

	_, err = suite.Client.API.UpdateTrustCenterCompliance(tcOrg.Owner.UserCtx, complianceID, testclient.UpdateTrustCenterComplianceInput{
		StandardID: &standard2.ID,
	})
	assert.NilError(t, err)

	checkTuple(standard1.ID, false)
	checkTuple(standard2.ID, true)

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}

func TestQueryTrustCenterCompliances(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())
	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t)

	// Create test data
	standard1 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	standard2 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	trustCenter := tcOrg.TrustCenter

	// Create multiple compliances for tcOrg.Owner
	countOrgOwned := 2
	orgOwnedComplianceIDs := []string{}
	for i := range countOrgOwned {
		standardID := standard1.ID
		if i == 1 {
			standardID = standard2.ID
		}
		compliance := (&th.TrustCenterComplianceBuilder{
			Client:        suite.Client,
			StandardID:    standardID,
			TrustCenterID: trustCenter.ID,
			Tags:          []string{"org", "test"},
		}).MustNew(tcOrg.Owner.UserCtx, t)
		orgOwnedComplianceIDs = append(orgOwnedComplianceIDs, compliance.ID)
	}

	// Create compliance for different org
	standardOther := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg2.Owner.UserCtx, t)
	trustCenterOther := tcOrg2.TrustCenter
	(&th.TrustCenterComplianceBuilder{
		Client:        suite.Client,
		StandardID:    standardOther.ID,
		TrustCenterID: trustCenterOther.ID,
		Tags:          []string{"other", "org"},
	}).MustNew(tcOrg2.Owner.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path, org user should get all org owned compliances",
			client:          suite.Client.API,
			ctx:             tcOrg.Owner.UserCtx,
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             tcOrg.Member.UserCtx,
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, anonymous user",
			client:          suite.Client.API,
			ctx:             th.CreateAnonymousTrustCenterContext(trustCenter.ID, tcOrg.Owner.OrganizationID),
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, using api token",
			client:          tcOrg.AdminAPIClient,
			ctx:             context.Background(),
			expectedResults: countOrgOwned,
		},
		{
			name:            "happy path, using pat",
			client:          tcOrg.AdminPatClient,
			ctx:             context.Background(),
			expectedResults: countOrgOwned,
		},
		{
			name:            "another user, should see their own compliance",
			client:          suite.Client.API,
			ctx:             tcOrg2.Owner.UserCtx,
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

	// th.Cleanup
	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(tcOrg2.Owner.UserCtx, t)
}

func TestMutationDeleteTrustCenterCompliance(t *testing.T) {
	t.Parallel()
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())

	// Create test data for deletion
	standard1 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	standard2 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	standard3 := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg.Owner.UserCtx, t)
	trustCenter1 := tcOrg.TrustCenter

	// Create compliance objects to delete
	compliance1 := (&th.TrustCenterComplianceBuilder{
		Client:        suite.Client,
		StandardID:    standard1.ID,
		TrustCenterID: trustCenter1.ID,
		Tags:          []string{"delete", "test1"},
	}).MustNew(tcOrg.Owner.UserCtx, t)

	compliance2 := (&th.TrustCenterComplianceBuilder{
		Client:        suite.Client,
		StandardID:    standard2.ID,
		TrustCenterID: trustCenter1.ID,
		Tags:          []string{"delete", "test2"},
	}).MustNew(tcOrg.Owner.UserCtx, t)

	compliance3 := (&th.TrustCenterComplianceBuilder{
		Client:        suite.Client,
		StandardID:    standard3.ID,
		TrustCenterID: trustCenter1.ID,
		Tags:          []string{"delete", "test3"},
	}).MustNew(tcOrg.Owner.UserCtx, t)

	tcOrg2 := th.CreateFreshOrgWithTrustCenter(t, th.WithAPIClients())

	// Create compliance for different org
	standardOther := (&th.StandardBuilder{Client: suite.Client}).MustNew(tcOrg2.Owner.UserCtx, t)
	trustCenterOther := tcOrg2.TrustCenter
	complianceOther := (&th.TrustCenterComplianceBuilder{
		Client:        suite.Client,
		StandardID:    standardOther.ID,
		TrustCenterID: trustCenterOther.ID,
		Tags:          []string{"other", "org"},
	}).MustNew(tcOrg2.Owner.UserCtx, t)

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
			client:     suite.Client.API,
			ctx:        tcOrg.Owner.UserCtx,
		},
		{
			name:        "not authorized, different org compliance api token",
			idToDelete:  compliance2.ID,
			client:      tcOrg2.AdminAPIClient,
			ctx:         context.Background(),
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: compliance2.ID,
			client:     tcOrg.AdminPatClient,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: compliance3.ID,
			client:     tcOrg.AdminAPIClient,
			ctx:        context.Background(),
		},
		{
			name:        "not authorized, different org compliance via jwt",
			idToDelete:  complianceOther.ID,
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, view only user",
			idToDelete:  complianceOther.ID, // use different org compliance to test permissions
			client:      suite.Client.API,
			ctx:         tcOrg.Member.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "trust center compliance not found, invalid ID",
			idToDelete:  "invalid-id",
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "trust center compliance not found, non-existent ID",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         tcOrg.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)
		})
	}

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
