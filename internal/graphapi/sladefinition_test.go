package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQuerySLADefinition(t *testing.T) {
	sla := (&th.SLADefinitionBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: sla.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: sla.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: sla.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "not found, using not authorized user",
			queryID:  sla.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetSLADefinitionByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.SLADefinition.ID))
			assert.Check(t, resp.SLADefinition.SLADays > 0)
		})
	}

	(&th.Cleanup[*generated.SLADefinitionDeleteOne]{Client: suite.Client.DB.SLADefinition, ID: sla.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQuerySLADefinitions(t *testing.T) {
	sla1 := (&th.SLADefinitionBuilder{Client: suite.Client, SLADays: 7, SecurityLevel: enums.SecurityLevelNone}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 5,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 5,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 5,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 5,
		},
		{
			name:            "another user, no results from this org",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 4,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllSLADefinitions(tc.ctx, nil, nil, nil, nil, []*testclient.SLADefinitionOrder{})
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.SLADefinitions.Edges, tc.expectedResults))
		})
	}

	(&th.Cleanup[*generated.SLADefinitionDeleteOne]{Client: suite.Client.DB.SLADefinition, ID: sla1.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateSLADefinition(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateSLADefinitionInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateSLADefinitionInput{
				SLADays: 30,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateSLADefinitionInput{
				SLADays: 14,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateSLADefinitionInput{
				SLADays: 60,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateSLADefinition(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.CreateSLADefinition.SLADefinition.ID != "")
			assert.Check(t, is.Equal(tc.request.SLADays, resp.CreateSLADefinition.SLADefinition.SLADays))

			(&th.Cleanup[*generated.SLADefinitionDeleteOne]{Client: suite.Client.DB.SLADefinition, ID: resp.CreateSLADefinition.SLADefinition.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateSLADefinition(t *testing.T) {
	sla := (&th.SLADefinitionBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateSLADefinitionInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field by admin user",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(14)),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name: "happy path, update using pat",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(7)),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, update using api token",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(9)),
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions as view only user",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(60)),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(60)),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateSLADefinition(tc.ctx, sla.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.SLADays != nil {
				assert.Check(t, is.Equal(*tc.request.SLADays, resp.UpdateSLADefinition.SLADefinition.SLADays))
			}
		})
	}

	(&th.Cleanup[*generated.SLADefinitionDeleteOne]{Client: suite.Client.DB.SLADefinition, ID: sla.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteSLADefinition(t *testing.T) {
	sla1 := (&th.SLADefinitionBuilder{Client: suite.Client, SecurityLevel: enums.SecurityLevelLow}).MustNew(th.SharedTestUser1.UserCtx, t)
	sla2 := (&th.SLADefinitionBuilder{Client: suite.Client, SecurityLevel: enums.SecurityLevelMedium}).MustNew(th.SharedTestUser1.UserCtx, t)
	sla3 := (&th.SLADefinitionBuilder{Client: suite.Client, SecurityLevel: enums.SecurityLevelHigh}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  sla1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  sla1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: sla1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedAdminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  sla1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: sla2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: sla3.ID,
			client:     suite.Client.APIWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteSLADefinition(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteSLADefinition.DeletedID))
		})
	}
}
