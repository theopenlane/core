package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryCustomDomainByID(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name           string
		expectedDomain string
		queryID        string
		client         *openlaneclient.OpenlaneClient
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
				assert.Check(t, is.Nil(resp))

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
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int64
		where           *openlaneclient.CustomDomainWhereInput
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
			where: &openlaneclient.CustomDomainWhereInput{
				CnameRecord: &customDomain1.CnameRecord,
			},
			expectedResults: 1,
		},
		{
			name:   "query by domain, not found",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.CustomDomainWhereInput{
				CnameRecord: &nonExistentDomain,
			},
			expectedResults: 0,
		},
		{
			name:   "query by mappable domain",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.CustomDomainWhereInput{
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
		request     openlaneclient.CreateCustomDomainInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: openlaneclient.CreateCustomDomainInput{
				CnameRecord:      "test.example.com",
				MappableDomainID: mappableDomain.ID,
				OwnerID:          lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, adminUser",
			request: openlaneclient.CreateCustomDomainInput{
				CnameRecord:      "test.example.com",
				MappableDomainID: mappableDomain.ID,
				OwnerID:          lo.ToPtr(testUser1.OrganizationID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "not authorized",
			request: openlaneclient.CreateCustomDomainInput{
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
			request: openlaneclient.CreateCustomDomainInput{
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
			request: openlaneclient.CreateCustomDomainInput{
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
				assert.Check(t, is.Nil(resp))

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

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(t.Context(), t)
}

func TestMutationDeleteCustomDomain(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	customDomain2 := (&CustomDomainBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	customDomain3 := (&CustomDomainBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t)
	nonExistentID := "non-existent-id"

	testCases := []struct {
		name        string
		id          string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "delete domain",
			id:     customDomain.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "delete domain, admin user",
			id:     customDomain2.ID,
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:        "unauthorized",
			id:          customDomain3.ID,
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
			resp, err := tc.client.DeleteCustomDomain(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.id, resp.DeleteCustomDomain.DeletedID))

			// Verify the domain is deleted
			_, err = tc.client.GetCustomDomainByID(tc.ctx, tc.id)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, IDs: []string{customDomain.MappableDomainID, customDomain2.MappableDomainID, customDomain3.MappableDomainID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain3.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestUpdateCustomDomain(t *testing.T) {
	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		errorMsg    string
		updateInput openlaneclient.UpdateCustomDomainInput
	}{
		{
			name:    "happy path",
			queryID: customDomain.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: openlaneclient.UpdateCustomDomainInput{
				Tags: []string{"hello"},
			},
		},
		{
			name:    "not allowed",
			queryID: customDomain.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
			updateInput: openlaneclient.UpdateCustomDomainInput{
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
				assert.Check(t, is.Nil(resp))

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: customDomain.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateBulkCustomDomain(t *testing.T) {
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		requests    []*openlaneclient.CreateCustomDomainInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
		numExpected int
	}{
		{
			name: "happy path - multiple domains",
			requests: []*openlaneclient.CreateCustomDomainInput{
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
			requests: []*openlaneclient.CreateCustomDomainInput{
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
			requests: []*openlaneclient.CreateCustomDomainInput{
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
			requests: []*openlaneclient.CreateCustomDomainInput{
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
			requests: []*openlaneclient.CreateCustomDomainInput{
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
			requests:    []*openlaneclient.CreateCustomDomainInput{},
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
				assert.Check(t, is.Nil(resp))

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

	// Create test custom domains with different users
	customDomain1 := (&CustomDomainBuilder{
		client:           suite.client,
		MappableDomainID: mappableDomain.ID,
		OwnerID:          testUser1.OrganizationID,
	}).MustNew(testUser1.UserCtx, t)

	customDomain2 := (&CustomDomainBuilder{
		client:           suite.client,
		MappableDomainID: mappableDomain.ID,
		OwnerID:          testUser1.OrganizationID,
	}).MustNew(testUser1.UserCtx, t)

	customDomain3 := (&CustomDomainBuilder{
		client:           suite.client,
		MappableDomainID: mappableDomain.ID,
		OwnerID:          testUser2.OrganizationID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
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
				assert.Check(t, is.Nil(resp))
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
