package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryCustomDomainByID() {
	t := suite.T()

	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t, nil)

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.CustomDomain)

			assert.Equal(t, tc.queryID, resp.CustomDomain.ID)
			assert.Equal(t, tc.expectedDomain, resp.CustomDomain.CnameRecord)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(t.Context(), suite)
}

func (suite *GraphTestSuite) TestQueryCustomDomains() {
	t := suite.T()

	mappableDomain := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	mappableDomain2 := (&MappableDomainBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	customDomain1 := (&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain2.ID}).MustNew(testUser1.UserCtx, t, nil)
	(&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain.ID}).MustNew(testUser1.UserCtx, t, lo.ToPtr(enums.CustomDomainStatusVerified))
	(&CustomDomainBuilder{client: suite.client, MappableDomainID: mappableDomain.ID}).MustNew(testUser2.UserCtx, t, lo.ToPtr(enums.CustomDomainStatusVerified))
	nonExistentDomain := "nonexistent.example.com"

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
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
		{
			name:   "query by status",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &openlaneclient.CustomDomainWhereInput{
				Status: lo.ToPtr(enums.CustomDomainStatusPending),
			},
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetCustomDomains(tc.ctx, nil, nil, tc.where)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.CustomDomains)
			assert.Equal(t, int64(tc.expectedResults), resp.CustomDomains.TotalCount)
			for _, domain := range resp.CustomDomains.Edges {
				assert.Equal(t, *domain.Node.OwnerID, testUser1.OrganizationID)
			}
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(t.Context(), suite)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain2.ID}).MustDelete(t.Context(), suite)
}

func (suite *GraphTestSuite) TestMutationCreateCustomDomain() {
	t := suite.T()

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.CnameRecord, resp.CreateCustomDomain.CustomDomain.CnameRecord)
			assert.Equal(t, tc.request.MappableDomainID, resp.CreateCustomDomain.CustomDomain.MappableDomainID)
			assert.Equal(t, enums.CustomDomainStatusPending, resp.CreateCustomDomain.CustomDomain.Status)
			assert.Equal(t, resp.CreateCustomDomain.CustomDomain.TxtRecordSubdomain, "_olverify")
			assert.NotEmpty(t, resp.CreateCustomDomain.CustomDomain.TxtRecordSubdomain)
			assert.NotEmpty(t, resp.CreateCustomDomain.CustomDomain.TxtRecordValue)

			// Clean up
			(&Cleanup[*generated.CustomDomainDeleteOne]{client: suite.client.db.CustomDomain, ID: resp.CreateCustomDomain.CustomDomain.ID}).MustDelete(tc.ctx, suite)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: mappableDomain.ID}).MustDelete(t.Context(), suite)
}

func (suite *GraphTestSuite) TestMutationDeleteCustomDomain() {
	t := suite.T()

	customDomain := (&CustomDomainBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t, nil)
	customDomain2 := (&CustomDomainBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t, nil)
	customDomain3 := (&CustomDomainBuilder{client: suite.client, OwnerID: testUser1.OrganizationID}).MustNew(testUser1.UserCtx, t, nil)
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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.id, resp.DeleteCustomDomain.DeletedID)

			// Verify the domain is deleted
			_, err = tc.client.GetCustomDomainByID(tc.ctx, tc.id)
			require.Error(t, err)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(t.Context(), suite)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain2.MappableDomainID}).MustDelete(t.Context(), suite)
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain3.MappableDomainID}).MustDelete(t.Context(), suite)
}

func (suite *GraphTestSuite) TestUpdateCustomDomain() {
	t := suite.T()

	customDomain := (&CustomDomainBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t, nil)

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
	(&Cleanup[*generated.MappableDomainDeleteOne]{client: suite.client.db.MappableDomain, ID: customDomain.MappableDomainID}).MustDelete(t.Context(), suite)
}
