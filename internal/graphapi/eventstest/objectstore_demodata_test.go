//go:build test

package eventstest_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/v2/internal/ent/generated/procedure"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

const (
	demoTestDataset   = "minimal"
	demoTestSourceKey = "demo/minimal"
)

type demoRowCounts struct {
	Policies   int
	Procedures int
	Tasks      int
}

func deleteTestDemoData(t *testing.T, ctx context.Context, keep []string) {
	t.Helper()

	_, err := suite.Client.API.DeleteDemoData(ctx, testclient.DeleteDemoDataInput{Dataset: demoTestDataset, Keep: keep})
	th.RequireNoError(t, err)

	waitForEvents()
}

func countDemoRows(t *testing.T, ctx context.Context) demoRowCounts {
	t.Helper()

	policies, err := suite.Client.DB.InternalPolicy.Query().Where(internalpolicy.SourceKey(demoTestSourceKey)).Count(ctx)
	th.RequireNoError(t, err)

	procedures, err := suite.Client.DB.Procedure.Query().Where(procedure.SourceKey(demoTestSourceKey)).Count(ctx)
	th.RequireNoError(t, err)

	tasks, err := suite.Client.DB.Task.Query().Where(task.SourceKey(demoTestSourceKey)).Count(ctx)
	th.RequireNoError(t, err)

	return demoRowCounts{Policies: policies, Procedures: procedures, Tasks: tasks}
}

func TestDemoDataSeedAndDelete(t *testing.T) {
	testCases := []struct {
		name          string
		ownControlRef string
		keep          []string
		afterDelete   demoRowCounts
	}{
		{
			name:        "delete removes every seeded record",
			afterDelete: demoRowCounts{},
		},
		{
			name:        "delete keeps the listed schemas",
			keep:        []string{"internal_policy"},
			afterDelete: demoRowCounts{Policies: 1},
		},
		{
			name:          "seed and delete leave the organization's own rows alone",
			ownControlRef: "MIN-01",
			afterDelete:   demoRowCounts{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			org := suite.UserBuilder(context.Background(), t)

			var own *ent.Control
			if tc.ownControlRef != "" {
				own = (&th.ControlBuilder{Client: suite.Client, RefCode: tc.ownControlRef, Title: "Our access review"}).MustNew(org.UserCtx, t)
			}

			_, err := suite.Client.API.SeedDemoData(org.UserCtx, testclient.SeedDemoDataInput{Dataset: demoTestDataset})
			th.RequireNoError(t, err)

			waitForEvents()

			t.Cleanup(func() {
				deleteTestDemoData(t, org.UserCtx, nil)
			})

			assert.Check(t, is.DeepEqual(demoRowCounts{Policies: 1, Procedures: 1, Tasks: 2}, countDemoRows(t, org.UserCtx)))

			deleteTestDemoData(t, org.UserCtx, tc.keep)

			assert.Check(t, is.DeepEqual(tc.afterDelete, countDemoRows(t, org.UserCtx)))

			if own == nil {
				return
			}

			remaining, err := suite.Client.DB.Control.Get(org.UserCtx, own.ID)
			th.RequireNoError(t, err)
			assert.Check(t, is.Equal("Our access review", remaining.Title))
			assert.Check(t, is.Equal("", remaining.SourceKey))
		})
	}
}

func TestDemoDataRejected(t *testing.T) {
	owner := suite.UserBuilder(context.Background(), t)
	member := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(owner.UserCtx, t, &member, enums.RoleMember, owner.OrganizationID)

	testCases := []struct {
		name        string
		ctx         context.Context
		dataset     string
		seed        bool
		expectedErr string
	}{
		{
			name:        "member cannot seed",
			ctx:         member.UserCtx,
			dataset:     demoTestDataset,
			seed:        true,
			expectedErr: ent.ErrPermissionDenied.Error(),
		},
		{
			name:        "member cannot delete",
			ctx:         member.UserCtx,
			dataset:     demoTestDataset,
			expectedErr: ent.ErrPermissionDenied.Error(),
		},
		{
			name:        "seed of an invalid dataset name",
			ctx:         owner.UserCtx,
			dataset:     "Not A Dataset",
			seed:        true,
			expectedErr: graphapi.ErrDemoDatasetNotFound.Error(),
		},
		{
			name:        "delete of an invalid dataset name",
			ctx:         owner.UserCtx,
			dataset:     "../vendors",
			expectedErr: graphapi.ErrDemoDatasetNotFound.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			if tc.seed {
				_, err = suite.Client.API.SeedDemoData(tc.ctx, testclient.SeedDemoDataInput{Dataset: tc.dataset})
			} else {
				_, err = suite.Client.API.DeleteDemoData(tc.ctx, testclient.DeleteDemoDataInput{Dataset: tc.dataset})
			}

			assert.Check(t, is.ErrorContains(err, tc.expectedErr))
		})
	}
}
