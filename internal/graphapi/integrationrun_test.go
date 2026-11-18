//go:build test

package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func TestQueryAllIntegrationRuns(t *testing.T) {
	integration := (&th.IntegrationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	run1 := (&th.IntegrationRunBuilder{Client: suite.Client, IntegrationID: integration.ID}).MustNew(th.SharedTestUser1.UserCtx, t)
	run2 := (&th.IntegrationRunBuilder{Client: suite.Client, IntegrationID: integration.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	systemCtx := generated.NewContext(auth.NewTestContextWithOrgID(th.SharedTestUser1.ID, "", auth.WithCapabilities(auth.CapSystemAdmin|auth.CapBypassOrgFilter|auth.CapInternalOperation|auth.CapSystemSweep)), suite.Client.DB)

	_, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(integration.ID).
		SetOperationName("DirectorySync").
		Save(systemCtx)
	assert.Assert(t, err != nil)

	t.Cleanup(func() {
		(&th.Cleanup[*generated.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{run1.ID, run2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	where := &testclient.IntegrationRunWhereInput{IntegrationIDIn: []string{integration.ID}}

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
			expectedResults: 2,
		},
		{
			name:            "read only member of the same org cannot see runs",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 0,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no integration runs should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllIntegrationRuns(tc.ctx, where)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.IntegrationRuns.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(int64(tc.expectedResults), resp.IntegrationRuns.TotalCount))

			for _, edge := range resp.IntegrationRuns.Edges {
				assert.Check(t, is.Equal(integration.OwnerID, *edge.Node.OwnerID))
			}
		})
	}
}
