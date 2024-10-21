package graphapi_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/openlaneclient"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
	"github.com/theopenlane/utils/ulids"
)

func (suite *GraphTestSuite) TestQueryEntitlement() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	entitlement := (&EntitlementBuilder{client: suite.client, OrganizationID: testOrgID}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *openlaneclient.OpenlaneClient
		listObjects bool
		ctx         context.Context
		errorMsg    string
	}{
		{
			name:        "happy path",
			queryID:     entitlement.ID,
			client:      suite.client.api,
			listObjects: true,
			ctx:         reqCtx,
		},
		{
			name:        "happy path, using api token",
			queryID:     entitlement.ID,
			client:      suite.client.apiWithToken,
			listObjects: false, // api token already restricts to a single org
			ctx:         context.Background(),
		},
		{
			name:        "happy path, using personal access token",
			queryID:     entitlement.ID,
			client:      suite.client.apiWithPAT,
			listObjects: true,
			ctx:         context.Background(),
		},
		{
			name:        "not found",
			queryID:     "notfound",
			client:      suite.client.api,
			listObjects: false,
			ctx:         reqCtx,
			errorMsg:    "not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.errorMsg == "" {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			if tc.listObjects {
				mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID})
			}

			resp, err := tc.client.GetEntitlementByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.queryID, resp.Entitlement.ID)

			require.NotEmpty(t, resp.Entitlement.GetPlan())
			assert.NotEmpty(t, resp.Entitlement.Plan.ID)

			require.NotEmpty(t, resp.Entitlement.GetOrganization())
			assert.NotEmpty(t, resp.Entitlement.Organization.ID)
		})
	}
}

func (suite *GraphTestSuite) TestQueryEntitlements() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	_ = (&EntitlementBuilder{client: suite.client, OrganizationID: testOrgID}).MustNew(reqCtx, t)

	otherUser := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	otherCtx, err := userContextWithID(otherUser.ID)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		listObjects     bool
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			listObjects:     true,
			ctx:             reqCtx,
			expectedResults: 1,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			listObjects:     false, // api token already restricts to a single org
			ctx:             context.Background(),
			expectedResults: 1,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			listObjects:     true,
			ctx:             context.Background(),
			expectedResults: 1,
		},
		{
			name:            "another user, no entitlements should be returned",
			client:          suite.client.api,
			listObjects:     false,
			ctx:             otherCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.listObjects {
				mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID})
			}

			resp, err := tc.client.GetAllEntitlements(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Entitlements.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateEntitlement() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	// setup for entitlement creation
	org1 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	org2 := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	plan := (&EntitlementPlanBuilder{client: suite.client}).MustNew(reqCtx, t)

	expiresAt := time.Now().Add(time.Hour * 24 * 365) // 1 year from now

	testCases := []struct {
		name        string
		request     openlaneclient.CreateEntitlementInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateEntitlementInput{
				PlanID:         plan.ID,
				OrganizationID: org1.ID,
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateEntitlementInput{
				PlanID:                 plan.ID,
				OrganizationID:         org2.ID,
				ExternalCustomerID:     lo.ToPtr("customer-123"),
				ExternalSubscriptionID: lo.ToPtr("sub-123"),
				Cancelled:              lo.ToPtr(false),
				ExpiresAt:              &expiresAt,
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "organization already has active entitlement",
			request: openlaneclient.CreateEntitlementInput{
				PlanID:         plan.ID,
				OrganizationID: org1.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "entitlement already exists",
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateEntitlementInput{
				PlanID:         plan.ID,
				OrganizationID: org1.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: create on entitlement",
		},
		{
			name: "missing required field, organization",
			request: openlaneclient.CreateEntitlementInput{
				PlanID: plan.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "value is less than the required length",
		},
		{
			name: "missing required field, plan",
			request: openlaneclient.CreateEntitlementInput{
				OrganizationID: org1.ID,
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			if tc.expectedErr == "" {
				mock_fga.ListAny(t, suite.client.fga, []string{fmt.Sprintf("organization:%s", testOrgID),
					fmt.Sprintf("organization:%s", org1.ID),
					fmt.Sprintf("organization:%s", org2.ID)})
			}

			resp, err := tc.client.CreateEntitlement(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.CreateEntitlement.Entitlement.GetPlan())
			assert.Equal(t, tc.request.PlanID, resp.CreateEntitlement.Entitlement.Plan.ID)

			if tc.request.ExpiresAt != nil {
				assert.WithinDuration(t, *tc.request.ExpiresAt, *resp.CreateEntitlement.Entitlement.ExpiresAt, time.Second)
				assert.True(t, resp.CreateEntitlement.Entitlement.Expires)
			} else {
				assert.False(t, resp.CreateEntitlement.Entitlement.Expires)
			}

			if tc.request.ExternalCustomerID != nil {
				assert.Equal(t, *tc.request.ExternalCustomerID, *resp.CreateEntitlement.Entitlement.ExternalCustomerID)
			}

			if tc.request.ExternalSubscriptionID != nil {
				assert.Equal(t, *tc.request.ExternalSubscriptionID, *resp.CreateEntitlement.Entitlement.ExternalSubscriptionID)
			}

			if tc.request.Cancelled != nil {
				assert.Equal(t, *tc.request.Cancelled, resp.CreateEntitlement.Entitlement.Cancelled)
			} else {
				assert.False(t, resp.CreateEntitlement.Entitlement.Cancelled)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateEntitlement() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	entitlement := (&EntitlementBuilder{client: suite.client}).MustNew(reqCtx, t)

	expiresAt := time.Now().Add(time.Hour * 24 * 30) // 30 days from now

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateEntitlementInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, update external customer id",
			request: openlaneclient.UpdateEntitlementInput{
				ExternalCustomerID: lo.ToPtr("customer-123"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, update external customer id using api token",
			request: openlaneclient.UpdateEntitlementInput{
				ExternalSubscriptionID: lo.ToPtr("sub-123"),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, expire entitlement using pat",
			request: openlaneclient.UpdateEntitlementInput{
				ExpiresAt: &expiresAt,
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "cancel entitlement",
			request: openlaneclient.UpdateEntitlementInput{
				Cancelled: lo.ToPtr(true),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateEntitlementInput{
				Cancelled: lo.ToPtr(false),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: update on entitlement",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.UpdateEntitlement(tc.ctx, entitlement.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.request.Cancelled != nil {
				assert.Equal(t, *tc.request.Cancelled, resp.UpdateEntitlement.Entitlement.GetCancelled())
			}

			if tc.request.ExternalCustomerID != nil {
				assert.Equal(t, *tc.request.ExternalCustomerID, *resp.UpdateEntitlement.Entitlement.GetExternalCustomerID())
			}

			if tc.request.ExternalSubscriptionID != nil {
				assert.Equal(t, *tc.request.ExternalSubscriptionID, *resp.UpdateEntitlement.Entitlement.GetExternalSubscriptionID())
			}

			if tc.request.ExpiresAt != nil {
				assert.WithinDuration(t, *tc.request.ExpiresAt, *resp.UpdateEntitlement.Entitlement.GetExpiresAt(), time.Second)
				assert.True(t, resp.UpdateEntitlement.Entitlement.GetExpires())
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteEntitlement() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	entitlement1 := (&EntitlementBuilder{client: suite.client}).MustNew(reqCtx, t)
	entitlement2 := (&EntitlementBuilder{client: suite.client}).MustNew(reqCtx, t)
	entitlement3 := (&EntitlementBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		checkAccess bool
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  entitlement1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: delete on entitlement",
		},
		{
			name:        "happy path, delete entitlement",
			idToDelete:  entitlement1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "entitlement already deleted, not found",
			idToDelete:  entitlement1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "entitlement not found",
		},
		{
			name:        "happy path, delete entitlement using api token",
			idToDelete:  entitlement2.ID,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "happy path, delete entitlement using pat",
			idToDelete:  entitlement3.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "unknown entitlement, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "entitlement not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization if entitlement exists
			if tc.checkAccess {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			resp, err := tc.client.DeleteEntitlement(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteEntitlement.DeletedID)
		})
	}
}
