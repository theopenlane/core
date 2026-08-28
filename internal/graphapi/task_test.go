package graphapi_test

import (
	"context"
	"slices"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryTask(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedOrgOwner(t)
	testUser := localTestOrg.Owner
	patClient := localTestOrg.PatClient

	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext("abc123", testUser.OrganizationID)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: task.ID,
			client:  suite.Client.API,
			ctx:     testUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: task.ID,
			client:  patClient,
			ctx:     context.Background(),
		},
		{
			name:     th.NotFoundErrorMsg,
			queryID:  "notfound",
			client:   suite.Client.API,
			ctx:      testUser.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  task.ID,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTaskByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Task.ID))
			assert.Check(t, len(resp.Task.Title) != 0)
			assert.Check(t, resp.Task.Details != nil)
			assert.Check(t, len(resp.Task.Status) != 0)
		})
	}

	// cleanup
	th.CleanupOrganizationDataWithContext(testUser.UserCtx, t)
}

func TestQueryTasks(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedFreshOrgUsers(t)
	testUser := localTestOrg.Owner
	adminPatClient := localTestOrg.AdminPatClient
	apiClient := localTestOrg.AdminAPIClient
	viewUser := localTestOrg.Member
	adminUser := localTestOrg.Admin
	superAdmin := localTestOrg.SuperAdmin

	anotherUser := suite.UserBuilder(context.Background(), t)

	// create a bunch to test the pagination with different users
	// works with overfetching
	numTasks := 10
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&th.TaskBuilder{Client: suite.Client, Due: gofakeit.Date()}).MustNew(testUser.UserCtx, t)
		t2 := (&th.TaskBuilder{Client: suite.Client, Due: gofakeit.Date()}).MustNew(viewUser.UserCtx, t)
		t3 := (&th.TaskBuilder{Client: suite.Client, Due: gofakeit.Date()}).MustNew(adminUser.UserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID, t3.ID)

		t4 := (&th.TaskBuilder{Client: suite.Client, Due: gofakeit.Date()}).MustNew(anotherUser.UserCtx, t)
		org2TaskIDs = append(org2TaskIDs, t4.ID)
	}

	userCtxPersonalOrg := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// add a task for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	taskPersonal := (&th.TaskBuilder{Client: suite.Client, AssigneeID: testUser.ID}).MustNew(userCtxPersonalOrg, t)

	risk := (&th.RiskBuilder{Client: suite.Client}).MustNew(adminUser.UserCtx, t)
	taskWithRisk := (&th.TaskBuilder{Client: suite.Client, RiskID: risk.ID}).MustNew(testUser.UserCtx, t)

	org1TaskIDs = append(org1TaskIDs, taskWithRisk.ID)

	var (
		startCursorDue     *string
		startCursorCreated *string
	)

	first := 10
	testCases := []struct {
		name            string
		orderBy         []*testclient.TaskOrder
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		setCursor       bool
		useCursor       bool
		totalCount      int64
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			totalCount:      31,
		},
		{
			name:            "happy path, super admin",
			client:          suite.Client.API,
			ctx:             superAdmin.UserCtx,
			expectedResults: first,
			totalCount:      31,
		},
		{
			name:            "happy path, api client",
			client:          apiClient,
			ctx:             context.Background(),
			expectedResults: first,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date, page 1",
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: 1,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date, page 1",
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: 1,
			totalCount:      31,
		},
		{
			name:            "happy path, view only user",
			client:          suite.Client.API,
			ctx:             viewUser.UserCtx,
			expectedResults: first,
			totalCount:      10,
		},
		{
			name:            "happy path, admin user",
			client:          suite.Client.API,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			totalCount:      11,
		},
		{
			name:            "happy path, using admin user pat pat, should only have access to same as admin user",
			client:          adminPatClient,
			ctx:             context.Background(),
			expectedResults: first,
			totalCount:      11,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.Client.API,
			ctx:             anotherUser.UserCtx,
			expectedResults: first,
			totalCount:      10,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			firstInput := int64(first)

			var after *string

			if tc.useCursor {
				switch tc.orderBy[0].Field {
				case testclient.TaskOrderFieldDue:
					after = startCursorDue
				case testclient.TaskOrderFieldCreatedAt:
					after = startCursorCreated
				}
			}

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, tc.orderBy, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Tasks.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(tc.totalCount, resp.Tasks.TotalCount))

			if tc.setCursor {
				// set the start cursor for the next test case
				assert.Assert(t, resp.Tasks.PageInfo.HasNextPage)
				assert.Assert(t, resp.Tasks.PageInfo.EndCursor != nil)

				switch tc.orderBy[0].Field {
				case testclient.TaskOrderFieldDue:
					startCursorDue = resp.Tasks.PageInfo.EndCursor
				case testclient.TaskOrderFieldCreatedAt:
					startCursorCreated = resp.Tasks.PageInfo.EndCursor
				}
			} else if tc.useCursor {
				// if we are using the cursor, but not setting it, we should not have a next page
				assert.Check(t, !(resp.Tasks.PageInfo.HasNextPage))

				// it should still have an end cursor
				assert.Check(t, resp.Tasks.PageInfo.EndCursor != nil)
			}
		})
	}

	// cleanup
	// internal context because personal orgs do not have access to tasks and the creation earlier
	// with th.TaskBuilder used the bypass too. SO use the system admin to remove
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: taskPersonal.ID}).
		MustDelete(th.SharedSystemAdminUser.UserCtx, t)

	th.CleanupOrganizationDataWithContext(testUser.UserCtx, t)
	th.CleanupOrganizationDataWithContext(anotherUser.UserCtx, t)

}

func getFutureDate() time.Time {
	return gofakeit.DateRange(time.Now(), time.Now().AddDate(1, 0, 0)).Truncate(time.Second)
}

func TestQueryTasksPaginationDueDate(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedFreshOrgUsers(t)
	localTestOrg2 := suite.SeedOrgOwner(t)

	// create a bunch to test the pagination with different users
	// to ensure we are paginating correctly when viewing as org admin
	numTasks := 95
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(localTestOrg.Owner.UserCtx, t)
		t2 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(localTestOrg.Member.UserCtx, t)
		t3 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(localTestOrg.Admin.UserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID, t3.ID)

		t4 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(localTestOrg2.Owner.UserCtx, t)
		org2TaskIDs = append(org2TaskIDs, t4.ID)
	}

	var startCursorDue *string

	first := 10
	testCases := []struct {
		name            string
		orderBy         []*testclient.TaskOrder
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		setCursor       bool
		useCursor       bool
		totalCount      int64
	}{
		{
			name:            "happy path, with order by due date, page 1",
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 5",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 6",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 7",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 8",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 9",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 10",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.Client.API,
			ctx:             localTestOrg.Admin.UserCtx,
			expectedResults: 5,
			totalCount:      95,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			firstInput := int64(first)

			var after *string

			if tc.useCursor {
				after = startCursorDue
			}

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, tc.orderBy, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Tasks.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(tc.totalCount, resp.Tasks.TotalCount))

			if tc.setCursor {
				// set the start cursor for the next test case
				assert.Assert(t, resp.Tasks.PageInfo.HasNextPage)
				assert.Assert(t, resp.Tasks.PageInfo.EndCursor != nil)

				startCursorDue = resp.Tasks.PageInfo.EndCursor
			} else if tc.useCursor {
				// if we are using the cursor, but not setting it, we should not have a next page
				assert.Check(t, !(resp.Tasks.PageInfo.HasNextPage))

				// it should still have an end cursor
				assert.Check(t, resp.Tasks.PageInfo.EndCursor != nil)
			}

			// ensure the tasks are sorted correctly
			for i, edge := range resp.Tasks.Edges {
				if i == 0 {
					continue // skip the first one, we don't have a previous one to compare
				}

				prevEdge := resp.Tasks.Edges[i-1]
				assert.Check(t, prevEdge.Node.Due != nil)
				assert.Check(t, edge.Node.Due != nil)
				currentDue := time.Time(*edge.Node.Due)
				previousDue := time.Time(*prevEdge.Node.Due)

				assert.Check(t, currentDue.After(previousDue) || currentDue.Equal(previousDue), "current due date (%s) should be after previous due date (%s)", currentDue, previousDue)

			}
		})
	}

	// cleanup
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(localTestOrg2.Owner.UserCtx, t)
}

func TestQueryTasksPaginationByCreatedDate(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.SeedFreshOrgUsers(t)
	localTestOrg2 := suite.SeedOrgOwner(t)

	testUser := localTestOrg.Owner
	viewOnlyUserCtx := localTestOrg.Member.UserCtx

	numTasks := 93
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(testUser.UserCtx, t)
		t2 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(viewOnlyUserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID)

		t4 := (&th.TaskBuilder{Client: suite.Client, Due: getFutureDate()}).MustNew(localTestOrg2.Owner.UserCtx, t)
		org2TaskIDs = append(org2TaskIDs, t4.ID)
	}

	var startCursorDue *string

	first := 10
	testCases := []struct {
		name            string
		orderBy         []*testclient.TaskOrder
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
		setCursor       bool
		useCursor       bool
		totalCount      int64
	}{
		{
			name:            "happy path, with order by created date, page 1",
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 5",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 6",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 7",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 8",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 9",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 10",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.Client.API,
			ctx:             viewOnlyUserCtx,
			expectedResults: 3,
			totalCount:      93,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			firstInput := int64(first)

			var after *string

			if tc.useCursor {
				after = startCursorDue
			}

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, tc.orderBy, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Tasks.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(tc.totalCount, resp.Tasks.TotalCount))

			if tc.setCursor {
				// set the start cursor for the next test case
				assert.Assert(t, resp.Tasks.PageInfo.HasNextPage)
				assert.Assert(t, resp.Tasks.PageInfo.EndCursor != nil)

				startCursorDue = resp.Tasks.PageInfo.EndCursor
			} else if tc.useCursor {
				// if we are using the cursor, but not setting it, we should not have a next page
				assert.Check(t, !(resp.Tasks.PageInfo.HasNextPage))

				// it should still have an end cursor
				assert.Check(t, resp.Tasks.PageInfo.EndCursor != nil)
			}

			// ensure the tasks are sorted correctly
			for i, edge := range resp.Tasks.Edges {
				if i == 0 {
					continue // skip the first one, we don't have a previous one to compare
				}

				prevEdge := resp.Tasks.Edges[i-1]
				assert.Check(t, prevEdge.Node.CreatedAt != nil)
				assert.Check(t, edge.Node.CreatedAt != nil)
				currentCreatedAt := time.Time(*edge.Node.CreatedAt)
				previousCreatedAt := time.Time(*prevEdge.Node.CreatedAt)

				assert.Check(t, currentCreatedAt.Before(previousCreatedAt) || currentCreatedAt.Equal(previousCreatedAt), "current created at (%s) should be before previous created at (%s)", currentCreatedAt, previousCreatedAt)

			}
		})
	}

	// cleanup
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
	th.CleanupOrganizationDataWithContext(localTestOrg2.Owner.UserCtx, t)
}

func TestMutationCreateTask(t *testing.T) {
	localTestOrg := suite.SeedFreshMinimalOrgUsers(t, true)

	testUser := localTestOrg.Owner

	userCtx := localTestOrg.Member.UserCtx
	adminCtx := localTestOrg.Admin.UserCtx

	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)
	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)

	systemOwnedControl := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	systemOwnedSubcontrol := (&th.SubcontrolBuilder{Client: suite.Client, ControlID: systemOwnedControl.ID}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTaskInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateTaskInput{
				Title: "test-task",
			},
			client: suite.Client.API,
			ctx:    testUser.UserCtx,
		},
		{
			name: "happy path, minimal input by member user with edges",
			request: testclient.CreateTaskInput{
				Title:             "test-task",
				Details:           lo.ToPtr("test details of the task"),
				Status:            &enums.TaskStatusInProgress,
				Due:               lo.ToPtr(models.DateTime(time.Now().Add(time.Hour * 24))),
				ControlIDs:        []string{control.ID},
				InternalPolicyIDs: []string{internalPolicy.ID},
				AssigneeID:        &localTestOrg.Member.ID, // assign the task to self
			},
			client: suite.Client.API,
			ctx:    userCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				Details:    lo.ToPtr("test details of the task"),
				Status:     &enums.TaskStatusInProgress,
				Due:        lo.ToPtr(models.DateTime(time.Now().Add(time.Hour * 24))),
				AssigneeID: &localTestOrg.Member.ID, // assign the task to another user
			},
			client: suite.Client.API,
			ctx:    testUser.UserCtx,
		},
		{
			name: "create with assignee not in org should not allowed",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				AssigneeID: &th.SharedTestUser2.ID,
			},
			client:      suite.Client.API,
			ctx:         testUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateTaskInput{
				Title:   "test-task",
				OwnerID: &testUser.OrganizationID,
			},
			client: localTestOrg.AdminPatClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				AssigneeID: &localTestOrg.Member.ID, // assign the task to another user

			},
			client: localTestOrg.APIClient,
			ctx:    context.Background(),
		},
		{
			name: "missing title, but display name provided",
			request: testclient.CreateTaskInput{
				Details: lo.ToPtr("making a list, checking it twice"),
			},
			client:      suite.Client.API,
			ctx:         testUser.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "not allowed to associated system owned control",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				ControlIDs: []string{systemOwnedControl.ID},
			},
			client:      suite.Client.API,
			ctx:         testUser.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "not allowed to associate system owned subcontrol",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				ControlIDs: []string{systemOwnedSubcontrol.ID},
			},
			client:      suite.Client.API,
			ctx:         testUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not allowed to associate system owned subcontrol with api token either",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				ControlIDs: []string{systemOwnedSubcontrol.ID},
			},
			client:      localTestOrg.APIClient,
			ctx:         context.Background(),
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTask(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Title, resp.CreateTask.Task.Title))

			assert.Check(t, len(resp.CreateTask.Task.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateTask.Task.DisplayID, "TSK-"))

			assert.Check(t, resp.CreateTask.Task.OwnerID != nil)

			if tc.request.Details == nil {
				assert.Check(t, is.Equal(*resp.CreateTask.Task.Details, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.CreateTask.Task.Details))
			}

			if tc.request.Status == nil {
				assert.Check(t, is.Equal(enums.TaskStatusOpen, resp.CreateTask.Task.Status))
			} else {
				assert.Check(t, is.Equal(*tc.request.Status, resp.CreateTask.Task.Status))
			}

			if tc.request.Details == nil {
				assert.Check(t, is.Equal(*resp.CreateTask.Task.Details, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.CreateTask.Task.Details))
			}

			if tc.request.Due == nil {
				assert.Check(t, resp.CreateTask.Task.Due == nil)
			} else {
				assert.Assert(t, resp.CreateTask.Task.Due != nil)
				diff := time.Time(*resp.CreateTask.Task.Due).Sub(time.Time(*tc.request.Due))
				assert.Check(t, diff >= -10*time.Second && diff <= 10*time.Second, "time difference is not within 10 seconds")
			}

			// when using an API token, the assigner is not set
			if tc.client == localTestOrg.APIClient {
				assert.Check(t, is.Nil(resp.CreateTask.Task.Assigner))
			} else {
				// otherwise it defaults to the authorized user
				assert.Check(t, resp.CreateTask.Task.Assigner != nil)
				switch tc.ctx {
				case testUser.UserCtx:
					assert.Check(t, is.Equal(testUser.ID, resp.CreateTask.Task.Assigner.ID))
				case userCtx:
					assert.Check(t, is.Equal(localTestOrg.Member.ID, resp.CreateTask.Task.Assigner.ID))
				}
			}

			if tc.request.AssigneeID == nil {
				assert.Check(t, is.Nil(resp.CreateTask.Task.Assignee))
			} else {
				assert.Assert(t, resp.CreateTask.Task.Assignee != nil)
				assert.Check(t, is.Equal(*tc.request.AssigneeID, resp.CreateTask.Task.Assignee.ID))

				// make sure the assignee can see the task
				taskResp, err := suite.Client.API.GetTaskByID(userCtx, resp.CreateTask.Task.ID)
				assert.NilError(t, err)
				assert.Check(t, taskResp != nil)

				// make sure the another org member cannot see the task if not linked to objects they can see
				if tc.request.ControlIDs == nil {
					taskResp, err = suite.Client.API.GetTaskByID(adminCtx, resp.CreateTask.Task.ID)
				}
			}

			// cleanup
			(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: resp.CreateTask.Task.ID}).MustDelete(testUser.UserCtx, t)
		})
	}

	// cleanup
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
	// cleanup system owned controls
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, ID: systemOwnedSubcontrol.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: systemOwnedControl.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)

}

func TestMutationUpdateTask(t *testing.T) {
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	group := (&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedAdminUser.ID, Role: enums.RoleAdmin.String()}).MustNew(th.SharedAdminUser.UserCtx, t)

	pngFile := th.UploadFile(t, th.LogoFilePath)
	pdfFile := th.UploadFile(t, th.PdfFilePath)

	taskCommentID := ""

	assignee := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &assignee, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	// add parents to ensure permissions are inherited
	risk := (&th.RiskBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)
	taskRisk := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// make sure the user cannot can see the task before they are the assigner
	_, err := suite.Client.API.GetTaskByID(th.SharedViewOnlyUser2.UserCtx, task.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	// make sure the user cannot can see the task before they are the assignee
	_, err = suite.Client.API.GetTaskByID(assignee.UserCtx, task.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	// make sure the user cannot see the task before the risk is added
	_, err = suite.Client.API.GetTaskByID(th.SharedAdminUser.UserCtx, taskRisk.ID)
	assert.ErrorContains(t, err, th.NotFoundErrorMsg)

	// NOTE: the tests and checks are ordered due to dependencies between updates
	// if you update cases, they will most likely need to be added to the end of the list
	testCases := []struct {
		name                 string
		taskID               string
		request              *testclient.UpdateTaskInput
		updateCommentRequest *testclient.UpdateNoteInput
		files                []*graphql.Upload
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
	}{
		{
			name:   "happy path, update details",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				Details:    lo.ToPtr(("making a list, checking it twice")),
				AssigneeID: &th.SharedAdminUser.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "happy path, add comment",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "matt is the best",
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "happy path, update comment with files",
			taskID: task.ID,
			updateCommentRequest: &testclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is better"),
			},
			files: []*graphql.Upload{
				pngFile,
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "happy path, update comment with file using PAT",
			taskID: task.ID,
			updateCommentRequest: &testclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is still better"),
			},
			files: []*graphql.Upload{
				pdfFile,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "happy path, delete comment",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				DeleteComment: &taskCommentID,
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "happy path, add risk",
			taskID: taskRisk.ID,
			request: &testclient.UpdateTaskInput{
				AddRiskIDs: []string{risk.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:    "update category using pat of owner",
			taskID:  task.ID,
			request: &testclient.UpdateTaskInput{},
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:   "update assignee to user not in org should now allowed",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(th.SharedTestUser2.ID),
			},
			client:      suite.Client.API,
			ctx:         th.SharedAdminUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:   "update assignee to view only user",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(assignee.ID),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "update assignee to same user, should not error",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(assignee.ID),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "update status and details",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				Status:  &enums.TaskStatusInProgress,
				Details: lo.ToPtr("do all the things for the thing"),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "add to group",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AddGroupIDs: []string{group.GroupID},
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "update assigner to another org member, no longer see it because no parent linked",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssignerID: lo.ToPtr(th.SharedViewOnlyUser2.ID),
			},
			client: suite.Client.API,
			ctx:    th.SharedAdminUser.UserCtx,
		},
		{
			name:   "clear assignee",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				ClearAssignee: lo.ToPtr(true),
			},
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser2.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			var (
				err         error
				resp        *testclient.UpdateTask
				commentResp *testclient.UpdateTaskComment
			)

			if tc.request != nil {
				resp, err = tc.client.UpdateTask(tc.ctx, tc.taskID, *tc.request)
			} else if tc.updateCommentRequest != nil {
				if len(tc.files) > 0 {
					th.ExpectUploadNillable(t, suite.Client.MockProvider, tc.files)
				}

				commentResp, err = suite.Client.API.UpdateTaskComment(th.SharedTestUser1.UserCtx, taskCommentID, *tc.updateCommentRequest, tc.files)
			}

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)

			if tc.request != nil {
				assert.Assert(t, resp != nil)

				if tc.request.Details != nil {
					assert.Check(t, is.Equal(*tc.request.Details, *resp.UpdateTask.Task.Details))
				}

				if tc.request.Status != nil {
					assert.Check(t, is.Equal(*tc.request.Status, resp.UpdateTask.Task.Status))
				}

				if tc.request.Details != nil {
					assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateTask.Task.Details))
				}

				if tc.request.ClearAssignee != nil {
					assert.Check(t, is.Nil(resp.UpdateTask.Task.Assignee))

					// the previous assignee should no longer be able to see the task
					_, err = suite.Client.API.GetTaskByID(assignee.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.ErrorContains(err, th.NotFoundErrorMsg))
				}

				if tc.request.ClearAssigner != nil {
					assert.Check(t, is.Nil(resp.UpdateTask.Task.Assignee))

					// the previous assigner should no longer be able to see the task
					_, err := suite.Client.API.GetTaskByID(th.SharedViewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.ErrorContains(err, th.NotFoundErrorMsg))
				}

				if tc.request.AddRiskIDs != nil {
					taskResp, err := suite.Client.API.GetTaskByID(th.SharedAdminUser.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.Nil(err))
					assert.Check(t, is.Equal(taskResp.Task.ID, tc.taskID))
				}

				if tc.request.AssignerID != nil {
					assert.Check(t, resp.UpdateTask.Task.Assigner != nil)
					assert.Check(t, is.Equal(*tc.request.AssignerID, resp.UpdateTask.Task.Assigner.ID))

					// make sure the assigner can see the task
					taskResp, err := suite.Client.API.GetTaskByID(th.SharedViewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, err)
					assert.Check(t, taskResp != nil)
				}

				if tc.request.AddComment != nil {
					assert.Check(t, len(resp.UpdateTask.Task.Comments.Edges) != 0)
					assert.Check(t, is.Equal(tc.request.AddComment.Text, resp.UpdateTask.Task.Comments.Edges[0].Node.Text))

					// there should only be one comment
					assert.Assert(t, is.Len(resp.UpdateTask.Task.Comments.Edges, 1))
					taskCommentID = resp.UpdateTask.Task.Comments.Edges[0].Node.ID

					// user shouldn't be able to see the comment
					checkResp, err := suite.Client.API.GetNoteByID(assignee.UserCtx, taskCommentID)
					assert.Check(t, is.ErrorContains(err, th.NotFoundErrorMsg))

					// user should be able to see the comment since they created the task
					checkResp, err = suite.Client.API.GetNoteByID(th.SharedAdminUser.UserCtx, taskCommentID)
					assert.Check(t, err)

					// org owner should be able to see the comment
					checkResp, err = suite.Client.API.GetNoteByID(th.SharedTestUser1.UserCtx, taskCommentID)
					assert.Check(t, err)
					assert.Check(t, checkResp != nil)
				} else if tc.request.DeleteComment != nil {
					// should not have any comments
					assert.Check(t, is.Len(resp.UpdateTask.Task.Comments.Edges, 0))
				}
			} else if tc.updateCommentRequest != nil {
				assert.Assert(t, commentResp != nil)

				// should only have the original comment
				assert.Assert(t, is.Len(commentResp.UpdateTaskComment.Task.Comments.Edges, 1))
				assert.Check(t, is.Equal(*tc.updateCommentRequest.Text, commentResp.UpdateTaskComment.Task.Comments.Edges[0].Node.Text))
			}
		})
	}

	// cleanup
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: task.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: group.GroupID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteTask(t *testing.T) {
	t.Parallel()

	testUser := suite.SeedOrgOwner(t)
	task1 := (&th.TaskBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)
	task2 := (&th.TaskBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete task",
			idToDelete:  task1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete task",
			idToDelete: task1.ID,
			client:     suite.Client.API,
			ctx:        testUser.Owner.UserCtx,
		},
		{
			name:        "task already deleted, not found",
			idToDelete:  task1.ID,
			client:      suite.Client.API,
			ctx:         testUser.Owner.UserCtx,
			expectedErr: "task not found",
		},
		{
			name:       "happy path, delete task using personal access token",
			idToDelete: task2.ID,
			client:     testUser.PatClient,
			ctx:        context.Background(),
		},
		{
			name:        "unknown task, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         testUser.Owner.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTask(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTask.DeletedID))
		})
	}

	th.CleanupOrganizationDataWithContext(testUser.Owner.UserCtx, t)
}

func TestMutationUpdateBulkTask(t *testing.T) {
	testUser := suite.SeedOrgOwner(t)

	task1 := (&th.TaskBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)
	task2 := (&th.TaskBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)
	task3 := (&th.TaskBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)

	taskAnotherUser := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	om := (&th.OrgMemberBuilder{Client: suite.Client}).MustNew(testUser.Owner.UserCtx, t)

	testCases := []struct {
		name                 string
		ids                  []string
		input                testclient.UpdateTaskInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedUpdatedCount int
	}{
		{
			name: "happy path, clear tags on multiple tasks",
			ids:  []string{task1.ID, task2.ID},
			input: testclient.UpdateTaskInput{
				ClearTags: lo.ToPtr(true),
				Details:   lo.ToPtr("Cleared all tags"),
			},
			client:               suite.Client.API,
			ctx:                  testUser.Owner.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateTaskInput{Title: lo.ToPtr("test")},
			client:      suite.Client.API,
			ctx:         testUser.Owner.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some tasks not authorized",
			ids:  []string{task1.ID, taskAnotherUser.ID}, // second task should fail authorization
			input: testclient.UpdateTaskInput{
				Title: lo.ToPtr("Updated by authorized user"),
			},
			client:               suite.Client.API,
			ctx:                  testUser.Owner.UserCtx,
			expectedUpdatedCount: 1, // only task1 should be updated
		},
		{
			name: "update not allowed, no permissions to tasks",
			ids:  []string{task1.ID},
			input: testclient.UpdateTaskInput{
				Title: lo.ToPtr("Should not update"),
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser2.UserCtx,
			expectedUpdatedCount: 0, // should not find any tasks to update
		},
		{
			name: "update status on multiple tasks",
			ids:  []string{task1.ID, task2.ID, task3.ID},
			input: testclient.UpdateTaskInput{
				Status: &enums.TaskStatusInProgress,
			},
			client:               suite.Client.API,
			ctx:                  testUser.Owner.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "assign tasks to org member",
			ids:  []string{task1.ID, task2.ID},
			input: testclient.UpdateTaskInput{
				AssigneeID: &om.UserID,
			},
			client:               suite.Client.API,
			ctx:                  testUser.Owner.UserCtx,
			expectedUpdatedCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateBulkTask(tc.ctx, tc.ids, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.UpdateBulkTask.Tasks, tc.expectedUpdatedCount))
			assert.Check(t, is.Len(resp.UpdateBulkTask.UpdatedIDs, tc.expectedUpdatedCount))

			// verify all returned tasks have the expected values
			for _, task := range resp.UpdateBulkTask.Tasks {
				if tc.input.Title != nil {
					assert.Check(t, is.Equal(*tc.input.Title, task.Title))
				}

				if tc.input.Details != nil {
					assert.Check(t, is.Equal(*tc.input.Details, *task.Details))
				}

				if tc.input.Status != nil {
					assert.Check(t, task.GetStatus() != nil)
					assert.Check(t, is.Equal(*tc.input.Status, *task.GetStatus()))
				}

				if tc.input.AssigneeID != nil {
					assert.Check(t, task.Assignee != nil)
					assert.Check(t, is.Equal(*tc.input.AssigneeID, task.Assignee.ID))
				}

				if tc.input.Due != nil {
					assert.Check(t, task.Due != nil)
				}

				if tc.input.AppendTags != nil {
					for _, tag := range tc.input.AppendTags {
						assert.Check(t, slices.Contains(task.Tags, tag))
					}

					tagDefs, err := tc.client.GetTagDefinitions(tc.ctx, nil, nil, &testclient.TagDefinitionWhereInput{
						NameIn: tc.input.AppendTags,
					})

					assert.NilError(t, err)
					assert.Check(t, is.Len(tagDefs.TagDefinitions.Edges, len(tc.input.AppendTags)))
				}

				if tc.input.ClearTags != nil && *tc.input.ClearTags {
					assert.Check(t, is.Len(task.Tags, 0))
				}

				// ensure the org owner has access to the task that was updated
				res, err := suite.Client.API.GetTaskByID(testUser.Owner.UserCtx, task.ID)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(task.ID, res.Task.ID))
			}

			// verify that the returned IDs match the ones that were actually updated
			for _, updatedID := range resp.UpdateBulkTask.UpdatedIDs {
				found := false
				for _, expectedID := range tc.ids {
					if expectedID == updatedID {
						found = true
						break
					}
				}
				assert.Check(t, found, "Updated ID %s should be in the original request", updatedID)
			}
		})
	}

	// th.Cleanup created tasks
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: taskAnotherUser.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	th.CleanupOrganizationDataWithContext(testUser.Owner.UserCtx, t)
}
