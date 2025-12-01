package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryCustomDomainByID(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name           string
		expectedDomain string
		queryID        string
		client         *testclient.TestClient
		ctx            context.Context
		errorMsg       string
	}{
		{
			name:           "happy path",
			expectedDomain: customDomain.CnameRecord,
			queryID:        customDomain.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name:           "happy path, view only user",
			expectedDomain: customDomain.CnameRecord,
			queryID:        customDomain.ID,
			client:         suite.client.api,
			ctx:            viewOnlyUser.UserCtx,
		},
		{
			name:     "domain not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:           "not authorized to query org",
			expectedDomain: customDomain.CnameRecord,
			queryID:        customDomain.ID,
			client:         suite.client.api,
			ctx:            testUser2.UserCtx,
			errorMsg:       notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetCustomDomainByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.CustomDomain.ID))
			assert.Check(t, is.Equal(tc.expectedDomain, resp.CustomDomain.CnameRecord))
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryCustomDomains(t *testing.T) {
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	mappableDomain2 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	customDomain1 := (&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain2.ID}).MustNew(testUser1.UserCtx, t)
	customDomain2 := (&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain.ID}).MustNew(testUser1.UserCtx, t)
	customDomain3 := (&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain.ID}).MustNew(testUser2.UserCtx, t)

	nonExistentDomain := "nonexistent.example.com"

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.CustomDomainWhereInput
	}{
		{
			name:            "return all",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "return all, ro user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:   "query by domain",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.CustomDomainWhereInput{
				CnameRecord: &customDomain1.CnameRecord,
			},
			expectedResults: 1,
		},
		{
			name:   "query by domain, not found",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.CustomDomainWhereInput{
				CnameRecord: &nonExistentDomain,
			},
			expectedResults: 0,
		},
		{
			name:   "query by mappable domain",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.CustomDomainWhereInput{
				MappableDomainID: &customDomain1.MappableDomainID,
			},
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetCustomDomains(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResults, resp.CustomDomains.TotalCount))

			for _, domain := range resp.CustomDomains.Edges {
				assert.Check(t, is.Equal(*domain.Node.OwnerID, testUser1.OrganizationID))
			}
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, IDs: []string{mappableDomain.ID, mappableDomain2.ID}}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, IDs: []string{customDomain1.ID, customDomain2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain3.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateCustomDomain(t *testing.T) {
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateCustomDomainInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: testclient.CreateCustomDomainInput{
				CnameRecord:      "test.example.com",
				MappableDomainID: mappableDomain.ID,
				OwnerID:          lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, adminUser",
			request: testclient.CreateCustomDomainInput{
				CnameRecord:      "test.example.com",
				MappableDomainID: mappableDomain.ID,
				OwnerID:          lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "not authorized",
			request: testclient.CreateCustomDomainInput{
				CnameRecord:      "test.example.com",
				MappableDomainID: mappableDomain.ID,
				OwnerID:          lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "invalid domain",
			request: testclient.CreateCustomDomainInput{
				CnameRecord:      "!invalid-domain",
				MappableDomainID: mappableDomain.ID,
				OwnerID:          lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "missing mappable domain",
			request: testclient.CreateCustomDomainInput{
				CnameRecord: "test2.example.com",
				OwnerID:     lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "mappable_domain_id",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateCustomDomain(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.CnameRecord, resp.CreateCustomDomain.CustomDomain.CnameRecord))
			assert.Check(t, is.Equal(tc.request.MappableDomainID, resp.CreateCustomDomain.CustomDomain.MappableDomainID))

			// Clean up
			(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: resp.CreateCustomDomain.CustomDomain.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteCustomDomain(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	customDomain2 := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	customDomain3 := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	anotherUser := suite.userBuilder(context.Background(), t)
	custDomainForTrustCenter := (&CustomDomainBuilder{client: suite.client}).MustNew(anotherUser.UserCtx, t)
	trustCenter := (&TrustCenterBuilder{client: suite.client, CustomDomainID: custDomainForTrustCenter.ID}).MustNew(anotherUser.UserCtx, t)
	nonExistentID := "non-existent-id"

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "delete domain, system admin user",
			id:     customDomain.ID,
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name:   "delete domain, owner user",
			id:     customDomain3.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "delete domain, owner user with trust center",
			id:     custDomainForTrustCenter.ID,
			client: suite.client.api,
			ctx:    anotherUser.UserCtx,
		},
		{
			name:        "unauthorized",
			id:          customDomain2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "domain not found",
			id:          nonExistentID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			// if trust center domain was delete, verify trust center has a custom domain first
			if tc.id == custDomainForTrustCenter.ID {
				tcResp, err := tc.client.GetTrustCenterByID(tc.ctx, trustCenter.ID)
				assert.NilError(t, err)
				assert.Assert(t, tcResp != nil)
				assert.Assert(t, tcResp.TrustCenter.CustomDomainID != nil)
				assert.Check(t, is.Equal(*tcResp.TrustCenter.CustomDomainID, custDomainForTrustCenter.ID))
			}

			resp, err := tc.client.DeleteCustomDomain(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.id, resp.DeleteCustomDomain.DeletedID))

			// Verify the domain is deleted
			_, err = tc.client.GetCustomDomainByID(tc.ctx, tc.id)
			assert.ErrorContains(t, err, notFoundErrorMsg)

			// if trust center domain was delete, verify trust center no longer has custom domain
			if tc.id == custDomainForTrustCenter.ID {
				tcResp, err := tc.client.GetTrustCenterByID(tc.ctx, trustCenter.ID)
				assert.NilError(t, err)
				assert.Assert(t, tcResp != nil)
				assert.Check(t, tcResp.TrustCenter.CustomDomainID == nil)
			}
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, IDs: []string{customDomain.MappableDomainID, customDomain2.MappableDomainID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain2.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestUpdateCustomDomain(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	dnsVerification := (&DNSVerificationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		errorMsg    string
		updateInput testclient.UpdateCustomDomainInput
	}{
		{
			name:    "happy path",
			queryID: customDomain.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			updateInput: testclient.UpdateCustomDomainInput{
				Tags: []string{"hello"},
			},
		},
		{
			name:    "update dns verification id",
			queryID: customDomain.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			updateInput: testclient.UpdateCustomDomainInput{
				DNSVerificationID: &dnsVerification.ID,
			},
		},
		{
			name:    "clear dns verification",
			queryID: customDomain.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			updateInput: testclient.UpdateCustomDomainInput{
				ClearDNSVerification: lo.ToPtr(true),
			},
		},
		{
			name:    "not allowed",
			queryID: customDomain.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: testclient.UpdateCustomDomainInput{
				Tags: []string{"hello"},
			},
			errorMsg: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateCustomDomain(tc.ctx, tc.queryID, tc.updateInput)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}
	(&Cleanup[*generated.DNSVerificationDeleteOne]{client: suite.client.db.DNSVerification, ID: dnsVerification.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateBulkCustomDomain(t *testing.T) {
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		requests    []*testclient.CreateCustomDomainInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
		numExpected int
	}{
		{
			name: "happy path - multiple domains",
			requests: []*testclient.CreateCustomDomainInput{
				{
					CnameRecord:      "bulk1.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
				{
					CnameRecord:      "bulk2.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
				{
					CnameRecord:      "bulk3.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			numExpected: 3,
		},
		{
			name: "happy path - single domain",
			requests: []*testclient.CreateCustomDomainInput{
				{
					CnameRecord:      "singlebulk.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			numExpected: 1,
		},
		{
			name: "happy path - admin user",
			requests: []*testclient.CreateCustomDomainInput{
				{
					CnameRecord:      "adminbulk.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			numExpected: 1,
		},
		{
			name: "invalid domain in batch",
			requests: []*testclient.CreateCustomDomainInput{
				{
					CnameRecord:      "valid.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
				{
					CnameRecord:      "!invalid-domain",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "not authorized",
			requests: []*testclient.CreateCustomDomainInput{
				{
					CnameRecord:      "unauthorized.example.com",
					MappableDomainID: mappableDomain.ID,
					OwnerID:          lo.ToPtr(testUser1.OrganizationID),
				},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "empty input",
			requests:    []*testclient.CreateCustomDomainInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "input is required",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateBulkCustomDomain(tc.ctx, tc.requests)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.CreateBulkCustomDomain.CustomDomains, tc.numExpected))

			// Verify each domain was created correctly
			for i, request := range tc.requests {
				assert.Check(t, is.Equal(request.CnameRecord, resp.CreateBulkCustomDomain.CustomDomains[i].CnameRecord))
				assert.Check(t, is.Equal(request.MappableDomainID, resp.CreateBulkCustomDomain.CustomDomains[i].MappableDomainID))
			}

			// Clean up created domains
			for _, domain := range resp.CreateBulkCustomDomain.CustomDomains {
				(&Cleanup[*generated.CustomDomainDeleteOne]{
					client: suite.client.db.CustomDomain,
					ID:     domain.ID,
				}).MustDelete(tc.ctx, t)
			}
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestGetAllCustomDomains(t *testing.T) {
	// Create test mappable domain
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	deletectx := setContext(systemAdminUser.UserCtx, suite.client.db)
	d, err := suite.client.db.CustomDomain.Query().All(deletectx)
	assert.Assert(t, err == nil)

	for _, cd := range d {
		suite.client.db.CustomDomain.DeleteOneID(cd.ID).ExecX(deletectx)
	}

	// Create test custom domains with different users
	customDomain1 := (&CustomDomainBuilder{
		client:           suite.client,
		MappableDomainID: mappableDomain.ID,
	}).MustNew(testUser1.UserCtx, t)

	customDomain2 := (&CustomDomainBuilder{
		client:           suite.client,
		MappableDomainID: mappableDomain.ID,
	}).MustNew(testUser1.UserCtx, t)

	customDomain3 := (&CustomDomainBuilder{
		client:           suite.client,
		MappableDomainID: mappableDomain.ID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their domains",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2, // Should see only domains owned by testUser1
		},
		{
			name:            "happy path - admin user sees all domains",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 2, // Should see all owned by testUser
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2, // Should see only domains from their organization
		},
		{
			name:            "happy path - different user sees only their domains",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1, // Should see only domains owned by testUser2
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllCustomDomains(tc.ctx)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.CustomDomains.Edges != nil)

			// Verify the number of results
			assert.Check(t, is.Len(resp.CustomDomains.Edges, int(tc.expectedResults)))
			assert.Check(t, is.Equal(tc.expectedResults, resp.CustomDomains.TotalCount))

			// Verify pagination info
			assert.Check(t, resp.CustomDomains.PageInfo.StartCursor != nil)

			// If we have results, verify the structure of the first result
			if tc.expectedResults > 0 {
				firstNode := resp.CustomDomains.Edges[0].Node
				assert.Check(t, len(firstNode.ID) != 0)
				assert.Check(t, len(firstNode.CnameRecord) != 0)
				assert.Check(t, len(firstNode.MappableDomainID) != 0)
				assert.Check(t, firstNode.OwnerID != nil)
				assert.Check(t, firstNode.CreatedAt != nil)
			}

			// Verify that users only see domains from their organization
			if tc.ctx == testUser1.UserCtx || tc.ctx == viewOnlyUser.UserCtx {
				for _, edge := range resp.CustomDomains.Edges {
					assert.Check(t, is.Equal(testUser1.OrganizationID, *edge.Node.OwnerID))
				}
			} else if tc.ctx == testUser2.UserCtx {
				for _, edge := range resp.CustomDomains.Edges {
					assert.Check(t, is.Equal(testUser2.OrganizationID, *edge.Node.OwnerID))
				}
			}
		})
	}

	// Clean up created domains
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, IDs: []string{customDomain1.ID, customDomain2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain3.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationDeleteCustomDomainWithTrustCenter(t *testing.T) {
	// This test validates the fix for the bug where deleting a custom domain
	// was causing trust center FGA tuples to be deleted, making the trust center inaccessible.
	// The bug occurred because the DeleteTuplesFirstKey context marker was being propagated
	// to the trust center update operation when clearing the custom domain reference.

	// Create a custom domain
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create a trust center with the custom domain
	trustCenter := (&TrustCenterBuilder{client: suite.client, CustomDomainID: customDomain.ID}).MustNew(testUser1.UserCtx, t)

	// Verify the trust center has the expected FGA tuples before deletion
	// Check for wildcard user can_view tuple
	userWildcardCheck, err := suite.client.fga.CheckAccess(testUser1.UserCtx, fgax.AccessCheck{
		SubjectID:   "*",
		SubjectType: "user",
		Relation:    "can_view",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
	})
	assert.NilError(t, err)
	assert.Check(t, userWildcardCheck, "trust center should have user:* can_view tuple before custom domain deletion")

	// Check for wildcard service can_view tuple
	serviceWildcardCheck, err := suite.client.fga.CheckAccess(testUser1.UserCtx, fgax.AccessCheck{
		SubjectID:   "*",
		SubjectType: "service",
		Relation:    "can_view",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
	})
	assert.NilError(t, err)
	assert.Check(t, serviceWildcardCheck, "trust center should have service:* can_view tuple before custom domain deletion")

	// Check for system tuple
	systemCheck, err := suite.client.fga.CheckAccess(testUser1.UserCtx, fgax.AccessCheck{
		SubjectID:   "openlane_core",
		SubjectType: "system",
		Relation:    "system",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
	})
	assert.NilError(t, err)
	assert.Check(t, systemCheck, "trust center should have system:openlane_core system tuple before custom domain deletion")

	// Delete the custom domain
	resp, err := suite.client.api.DeleteCustomDomain(testUser1.UserCtx, customDomain.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Check(t, is.Equal(customDomain.ID, resp.DeleteCustomDomain.DeletedID))

	// Verify the custom domain is deleted
	_, err = suite.client.api.GetCustomDomainByID(testUser1.UserCtx, customDomain.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	// Verify the trust center still exists and is accessible
	tcResp, err := suite.client.api.GetTrustCenterByID(testUser1.UserCtx, trustCenter.ID)
	assert.NilError(t, err)
	assert.Assert(t, tcResp != nil)
	assert.Check(t, is.Equal(trustCenter.ID, tcResp.TrustCenter.ID))

	// Verify the trust center's custom domain reference has been cleared
	assert.Check(t, tcResp.TrustCenter.CustomDomainID == nil || *tcResp.TrustCenter.CustomDomainID == "", "trust center custom domain reference should be cleared")

	// Verify the trust center's FGA tuples are still present after custom domain deletion
	// Check for wildcard user can_view tuple
	userWildcardCheckAfter, err := suite.client.fga.CheckAccess(testUser1.UserCtx, fgax.AccessCheck{
		SubjectID:   "*",
		SubjectType: "user",
		Relation:    "can_view",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
	})
	assert.NilError(t, err)
	assert.Check(t, userWildcardCheckAfter, "trust center should still have user:* can_view tuple after custom domain deletion")

	// Check for wildcard service can_view tuple
	serviceWildcardCheckAfter, err := suite.client.fga.CheckAccess(testUser1.UserCtx, fgax.AccessCheck{
		SubjectID:   "*",
		SubjectType: "service",
		Relation:    "can_view",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
	})
	assert.NilError(t, err)
	assert.Check(t, serviceWildcardCheckAfter, "trust center should still have service:* can_view tuple after custom domain deletion")

	// Check for system tuple
	systemCheckAfter, err := suite.client.fga.CheckAccess(testUser1.UserCtx, fgax.AccessCheck{
		SubjectID:   "openlane_core",
		SubjectType: "system",
		Relation:    "system",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
	})
	assert.NilError(t, err)
	assert.Check(t, systemCheckAfter, "trust center should still have system:openlane_core system tuple after custom domain deletion")

	// Cleanup
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(testUser1.UserCtx, t)
}
