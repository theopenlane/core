package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryMappableDomainByID(t *testing.T) {
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name         string
		expectedName string
		queryID      string
		client       *testclient.TestClient
		ctx          context.Context
		errorMsg     string
	}{
		{
			name:         "happy path",
			expectedName: mappableDomain.Name,
			queryID:      mappableDomain.ID,
			client:       suite.client.api,
			ctx:          testUser1.UserCtx,
		},
		{
			name:     "done",
			queryID:  "iddne",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetMappableDomainByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.MappableDomain.ID))
			assert.Check(t, is.Equal(tc.expectedName, resp.MappableDomain.Name))
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestQueryMappableDomains(t *testing.T) {
	mappableDomain1 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	mappableDomain2 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	bologneName := "bologne.io"

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		where           *testclient.MappableDomainWhereInput
	}{
		{
			name:            "return all",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 2,
		},
		{
			name:   "query by name",
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
			where: &testclient.MappableDomainWhereInput{
				Name: &mappableDomain1.Name,
			},
			expectedResults: 1,
		},
		{
			name:   "query by name, does not exist",
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
			where: &testclient.MappableDomainWhereInput{
				Name: &bologneName,
			},
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetMappableDomains(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.MappableDomains.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(int64(tc.expectedResults), resp.MappableDomains.TotalCount))
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain1.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain2.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateMappableDomain(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateMappableDomainInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: testclient.CreateMappableDomainInput{
				Name:   "trust.theopenlane.io",
				ZoneID: "trust-zone-id",
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "invalid domain",
			request: testclient.CreateMappableDomainInput{
				Name:   "!not-a-domain",
				ZoneID: "trust-zone-id",
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "not system admin, unauthorized",
			request: testclient.CreateMappableDomainInput{
				Name:   "trust.theopenlane.io",
				ZoneID: "trust-zone-id",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
	}
	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateMappableDomain(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Name, resp.CreateMappableDomain.MappableDomain.Name))

			(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: resp.CreateMappableDomain.MappableDomain.ID}).MustDelete(tc.ctx, t)
		})
	}
}

func TestMutationCreateBulkMappableDomain(t *testing.T) {
	testCases := []struct {
		name        string
		requests    []*testclient.CreateMappableDomainInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
		numExpected int
	}{
		{
			name: "happy path - multiple domains",
			requests: []*testclient.CreateMappableDomainInput{
				{
					Name:   "bulk1.theopenlane.io",
					ZoneID: "bulk1-zone-id",
				},
				{
					Name:   "bulk2.theopenlane.io",
					ZoneID: "bulk2-zone-id",
				},
				{
					Name:   "bulk3.theopenlane.io",
					ZoneID: "bulk3-zone-id",
				},
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			numExpected: 3,
		},
		{
			name: "happy path - single domain",
			requests: []*testclient.CreateMappableDomainInput{
				{
					Name:   "singlebulk.theopenlane.io",
					ZoneID: "singlebulk-zone-id",
				},
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			numExpected: 1,
		},
		{
			name: "invalid domain in batch",
			requests: []*testclient.CreateMappableDomainInput{
				{
					Name:   "valid.theopenlane.io",
					ZoneID: "singlebulk-zone-id",
				},
				{
					Name:   "!invalid-domain",
					ZoneID: "singlebulk-zone-id",
				},
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: "invalid or unparsable field: url",
		},
		{
			name: "not system admin, unauthorized",
			requests: []*testclient.CreateMappableDomainInput{
				{
					Name:   "unauthorized.theopenlane.io",
					ZoneID: "singlebulk-zone-id",
				},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:        "empty input",
			requests:    []*testclient.CreateMappableDomainInput{},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: "input is required",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateBulkMappableDomain(tc.ctx, tc.requests)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.CreateBulkMappableDomain.MappableDomains, tc.numExpected))

			// Verify each domain was created correctly
			for i, request := range tc.requests {
				assert.Check(t, is.Equal(request.Name, resp.CreateBulkMappableDomain.MappableDomains[i].Name))
			}

			// Clean up created domains
			for _, domain := range resp.CreateBulkMappableDomain.MappableDomains {
				(&Cleanup[*generated.MappableDomainDeleteOne]{
					client: suite.client.db.MappableDomain,
					ID:     domain.ID,
				}).MustDelete(tc.ctx, t)
			}
		})
	}
}

func TestUpdateMappableDomain(t *testing.T) {
	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
		input    testclient.UpdateMappableDomainInput
	}{
		{
			name:    "happy path",
			queryID: mappableDomain.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			input: testclient.UpdateMappableDomainInput{
				Tags: []string{"hello"},
			},
		},
		{
			name:     "does not exist",
			queryID:  "iddne",
			client:   suite.client.api,
			ctx:      systemAdminUser.UserCtx,
			errorMsg: notFoundErrorMsg,
			input: testclient.UpdateMappableDomainInput{
				Tags: []string{"hello"},
			},
		},
		{
			name:     "not allowed",
			queryID:  mappableDomain.ID,
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
			input: testclient.UpdateMappableDomainInput{
				Tags: []string{"hello"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateMappableDomain(tc.ctx, tc.queryID, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.DeepEqual(tc.input.Tags, resp.UpdateMappableDomain.MappableDomain.Tags))
		})
	}

	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestGetAllMappableDomains(t *testing.T) {
	// Create test mappable domains with different users
	mappableDomain1 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	mappableDomain2 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	mappableDomain3 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		expectedErr     string
	}{
		{
			name:            "happy path - system admin can see all domains",
			client:          suite.client.api,
			ctx:             systemAdminUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "regular user",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllMappableDomains(tc.ctx)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.MappableDomains.Edges != nil)

			// Verify the number of results
			assert.Check(t, is.Len(resp.MappableDomains.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(tc.expectedResults, int(resp.MappableDomains.TotalCount)))

			// Verify pagination info
			assert.Check(t, resp.MappableDomains.PageInfo.StartCursor != nil)

			// If we have results, verify the structure of the first result
			if tc.expectedResults > 0 {
				firstNode := resp.MappableDomains.Edges[0].Node
				assert.Check(t, len(firstNode.ID) != 0)
				assert.Check(t, len(firstNode.Name) != 0)
				assert.Check(t, firstNode.CreatedAt != nil)
			}
		})
	}

	// Clean up created domains
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, IDs: []string{mappableDomain1.ID, mappableDomain2.ID, mappableDomain3.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}
