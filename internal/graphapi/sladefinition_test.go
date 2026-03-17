package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQuerySLADefinition(t *testing.T) {
	sla := (&SLADefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: sla.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: sla.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, using not authorized user",
			queryID:  sla.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
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

	(&Cleanup[*generated.SLADefinitionDeleteOne]{client: suite.client.db.SLADefinition, ID: sla.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQuerySLADefinitions(t *testing.T) {
	sla1 := (&SLADefinitionBuilder{client: suite.client, SLADays: 7, SecurityLevel: enums.SecurityLevelNone}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 5,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 5,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 5,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 5,
		},
		{
			name:            "another user, no results from this org",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
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

	(&Cleanup[*generated.SLADefinitionDeleteOne]{client: suite.client.db.SLADefinition, ID: sla1.ID}).MustDelete(testUser1.UserCtx, t)
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateSLADefinitionInput{
				SLADays: 14,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateSLADefinitionInput{
				SLADays: 7,
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateSLADefinitionInput{
				SLADays: 60,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
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

			(&Cleanup[*generated.SLADefinitionDeleteOne]{client: suite.client.db.SLADefinition, ID: resp.CreateSLADefinition.SLADefinition.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateSLADefinition(t *testing.T) {
	sla := (&SLADefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update using pat",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(7)),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions as view only user",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(60)),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateSLADefinitionInput{
				SLADays: lo.ToPtr(int64(60)),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
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

	(&Cleanup[*generated.SLADefinitionDeleteOne]{client: suite.client.db.SLADefinition, ID: sla.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteSLADefinition(t *testing.T) {
	sla1 := (&SLADefinitionBuilder{client: suite.client, SecurityLevel: enums.SecurityLevelLow}).MustNew(testUser1.UserCtx, t)
	sla2 := (&SLADefinitionBuilder{client: suite.client, SecurityLevel: enums.SecurityLevelMedium}).MustNew(testUser1.UserCtx, t)
	sla3 := (&SLADefinitionBuilder{client: suite.client, SecurityLevel: enums.SecurityLevelHigh}).MustNew(testUser1.UserCtx, t)

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
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  sla1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: sla1.ID,
			client:     suite.client.api,
			ctx:        adminUser.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  sla1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: sla2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: sla3.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
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
