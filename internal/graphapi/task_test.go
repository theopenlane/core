package graphapi_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestQueryTask(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(testUser, t)

	task := (&TaskBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	anonymousContext := createAnonymousTrustCenterContext("abc123", testUser.OrganizationID)

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
			client:  suite.client.api,
			ctx:     testUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: task.ID,
			client:  patClient,
			ctx:     context.Background(),
		},
		{
			name:     notFoundErrorMsg,
			queryID:  "notfound",
			client:   suite.client.api,
			ctx:      testUser.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.client.api,
			ctx:      anonymousContext,
			queryID:  task.ID,
			errorMsg: notFoundErrorMsg,
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
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser.UserCtx, t)
}

func TestQueryTasks(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(testUser, t)

	viewUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewUser, enums.RoleMember, testUser.OrganizationID)

	adminUser1 := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &adminUser1, enums.RoleAdmin, testUser.OrganizationID)

	anotherUser := suite.userBuilder(context.Background(), t)

	// create a bunch to test the pagination with different users
	// works with overfetching
	numTasks := 10
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(testUser.UserCtx, t)
		t2 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(viewUser.UserCtx, t)
		t3 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(adminUser1.UserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID, t3.ID)

		t4 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(anotherUser.UserCtx, t)
		org2TaskIDs = append(org2TaskIDs, t4.ID)
	}

	userCtxPersonalOrg := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// add a task for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	taskPersonal := (&TaskBuilder{client: suite.client, AssigneeID: testUser.ID}).MustNew(userCtxPersonalOrg, t)

	risk := (&RiskBuilder{client: suite.client}).MustNew(adminUser1.UserCtx, t)
	taskWithRisk := (&TaskBuilder{client: suite.client, RiskID: risk.ID}).MustNew(testUser.UserCtx, t)

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
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date, page 1",
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: 1,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date, page 1",
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: 1,
			totalCount:      31,
		},
		{
			name:            "happy path, view only user",
			client:          suite.client.api,
			ctx:             viewUser.UserCtx,
			expectedResults: first,
			totalCount:      10,
		},
		{
			name:            "happy path, admin user",
			client:          suite.client.api,
			ctx:             adminUser1.UserCtx,
			expectedResults: first,
			totalCount:      11,
		},
		{
			name:            "happy path, using pat - which should have access to all tasks because its authorized to the personal org",
			client:          patClient,
			ctx:             context.Background(),
			expectedResults: first,
			totalCount:      32,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.client.api,
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

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, nil, tc.orderBy)
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
	// with TaskBuilder used the bypass too. SO use the system admin to remove
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskPersonal.ID}).
		MustDelete(systemAdminUser.UserCtx, t)

	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org1TaskIDs}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org2TaskIDs}).MustDelete(anotherUser.UserCtx, t)
}

func getFutureDate() time.Time {
	return gofakeit.DateRange(time.Now(), time.Now().AddDate(1, 0, 0)).Truncate(time.Second)
}

func TestQueryTasksPaginationDueDate(t *testing.T) {
	// create a bunch to test the pagination with different users
	// to ensure we are paginating correctly when viewing as org admin
	numTasks := 95
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(testUser1.UserCtx, t)
		t2 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(viewOnlyUser2.UserCtx, t)
		t3 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(adminUser.UserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID, t3.ID)

		t4 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(testUser2.UserCtx, t)
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
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 5",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 6",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 7",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 8",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 9",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      95,
		},
		{
			name:            "happy path, with order by due date and cursor, page 10",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldDue, Direction: testclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
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

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, nil, tc.orderBy)
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
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org1TaskIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org2TaskIDs}).MustDelete(testUser2.UserCtx, t)
}

func TestQueryTasksPaginationByCreatedDate(t *testing.T) {
	// create a bunch to test the pagination with different users
	// to ensure we are paginating correctly when viewing as org admin
	testUser := suite.userBuilder(context.Background(), t)
	om := (&OrgMemberBuilder{client: suite.client, Role: enums.RoleMember.String()}).MustNew(testUser.UserCtx, t)

	viewOnlyUserCtx := auth.NewTestContextWithOrgID(om.UserID, testUser.OrganizationID)

	numTasks := 93
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(testUser.UserCtx, t)
		t2 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(viewOnlyUserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID)

		t4 := (&TaskBuilder{client: suite.client, Due: getFutureDate()}).MustNew(testUser2.UserCtx, t)
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
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 5",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 6",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 7",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 8",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 9",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             viewOnlyUserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      93,
		},
		{
			name:            "happy path, with order by created date and cursor, page 10",
			useCursor:       true,
			orderBy:         []*testclient.TaskOrder{{Field: testclient.TaskOrderFieldCreatedAt, Direction: testclient.OrderDirectionDesc}},
			client:          suite.client.api,
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

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, nil, tc.orderBy)
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
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org1TaskIDs}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org2TaskIDs}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTask(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(testUser, t)
	apiClient := suite.setupAPITokenClient(testUser.UserCtx, t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	om2 := (&OrgMemberBuilder{client: suite.client, Role: enums.RoleAdmin.String()}).MustNew(testUser.UserCtx, t)

	userCtx := auth.NewTestContextWithOrgID(om.UserID, om.OrganizationID)
	adminCtx := auth.NewTestContextWithOrgID(om2.UserID, om2.OrganizationID)

	control := (&ControlBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	internalPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	systemOwnedControl := (&ControlBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	systemOwnedSubcontrol := (&SubcontrolBuilder{client: suite.client, ControlID: systemOwnedControl.ID}).MustNew(systemAdminUser.UserCtx, t)

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
			client: suite.client.api,
			ctx:    testUser.UserCtx,
		},
		{
			name: "happy path, minimal input by member user with edges",
			request: testclient.CreateTaskInput{
				Title:             "test-task",
				Details:           lo.ToPtr("test details of the task"),
				Status:            &enums.TaskStatusInProgress,
				Category:          lo.ToPtr("evidence upload"),
				Due:               lo.ToPtr(models.DateTime(time.Now().Add(time.Hour * 24))),
				ControlIDs:        []string{control.ID},
				InternalPolicyIDs: []string{internalPolicy.ID},
				AssigneeID:        &om.UserID, // assign the task to self
			},
			client: suite.client.api,
			ctx:    userCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				Details:    lo.ToPtr("test details of the task"),
				Status:     &enums.TaskStatusInProgress,
				Category:   lo.ToPtr("evidence upload"),
				Due:        lo.ToPtr(models.DateTime(time.Now().Add(time.Hour * 24))),
				AssigneeID: &om.UserID, // assign the task to another user
			},
			client: suite.client.api,
			ctx:    testUser.UserCtx,
		},
		{
			name: "create with assignee not in org should fail",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				AssigneeID: &testUser2.ID,
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: "user not in organization",
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateTaskInput{
				Title:   "test-task",
				OwnerID: &testUser.OrganizationID,
			},
			client: patClient,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTaskInput{
				Title: "test-task",
			},
			client: apiClient,
			ctx:    context.Background(),
		},
		{
			name: "missing title, but display name provided",
			request: testclient.CreateTaskInput{
				Details: lo.ToPtr("making a list, checking it twice"),
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "not allowed to associated system owned control",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				ControlIDs: []string{systemOwnedControl.ID},
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not allowed to associated system owned subcontrol",
			request: testclient.CreateTaskInput{
				Title:      "test-task",
				ControlIDs: []string{systemOwnedSubcontrol.ID},
			},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
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

			if tc.request.Category == nil {
				assert.Check(t, is.Equal(*resp.CreateTask.Task.Category, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Category, *resp.CreateTask.Task.Category))
			}

			if tc.request.Due == nil {
				assert.Check(t, resp.CreateTask.Task.Due == nil)
			} else {
				assert.Assert(t, resp.CreateTask.Task.Due != nil)
				diff := time.Time(*resp.CreateTask.Task.Due).Sub(time.Time(*tc.request.Due))
				assert.Check(t, diff >= -10*time.Second && diff <= 10*time.Second, "time difference is not within 10 seconds")
			}

			// when using an API token, the assigner is not set
			if tc.client == apiClient {
				assert.Check(t, is.Nil(resp.CreateTask.Task.Assigner))
			} else {
				// otherwise it defaults to the authorized user
				assert.Check(t, resp.CreateTask.Task.Assigner != nil)
				switch tc.ctx {
				case testUser.UserCtx:
					assert.Check(t, is.Equal(testUser.ID, resp.CreateTask.Task.Assigner.ID))
				case userCtx:
					assert.Check(t, is.Equal(om.UserID, resp.CreateTask.Task.Assigner.ID))
				}
			}

			if tc.request.AssigneeID == nil {
				assert.Check(t, is.Nil(resp.CreateTask.Task.Assignee))
			} else {
				assert.Assert(t, resp.CreateTask.Task.Assignee != nil)
				assert.Check(t, is.Equal(*tc.request.AssigneeID, resp.CreateTask.Task.Assignee.ID))

				// make sure the assignee can see the task
				taskResp, err := suite.client.api.GetTaskByID(userCtx, resp.CreateTask.Task.ID)
				assert.NilError(t, err)
				assert.Check(t, taskResp != nil)

				// make sure the another org member cannot see the task if not linked to objects they can see
				if tc.request.ControlIDs == nil {
					taskResp, err = suite.client.api.GetTaskByID(adminCtx, resp.CreateTask.Task.ID)
				}
			}

			// cleanup
			(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: resp.CreateTask.Task.ID}).MustDelete(testUser.UserCtx, t)
		})
	}

	// cleanup
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: om.ID}).MustDelete(testUser.UserCtx, t)
	// cleanup controls and policies
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: internalPolicy.ID}).MustDelete(testUser.UserCtx, t)
	// cleanup system owned controls
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: systemOwnedSubcontrol.ID}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: systemOwnedControl.ID}).MustDelete(systemAdminUser.UserCtx, t)

}

func TestMutationUpdateTask(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	group := (&GroupMemberBuilder{client: suite.client, UserID: adminUser.ID, Role: enums.RoleAdmin.String()}).MustNew(adminUser.UserCtx, t)

	pngFile, err := storage.NewUploadFile("testdata/uploads/logo.png")
	assert.NilError(t, err)

	pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
	assert.NilError(t, err)

	taskCommentID := ""

	assignee := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &assignee, enums.RoleMember, testUser1.OrganizationID)

	// add parents to ensure permissions are inherited
	risk := (&RiskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	taskRisk := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// make sure the user cannot can see the task before they are the assigner
	_, err = suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, task.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	// make sure the user cannot can see the task before they are the assignee
	_, err = suite.client.api.GetTaskByID(assignee.UserCtx, task.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

	// make sure the user cannot see the task before the risk is added
	_, err = suite.client.api.GetTaskByID(adminUser.UserCtx, taskRisk.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)

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
				AssigneeID: &adminUser.ID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, add comment",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "matt is the best",
				},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, update comment with files",
			taskID: task.ID,
			updateCommentRequest: &testclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is better"),
			},
			files: []*graphql.Upload{
				{
					File:        pngFile.RawFile,
					Filename:    pngFile.OriginalName,
					Size:        pngFile.Size,
					ContentType: pngFile.ContentType,
				},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, update comment with file using PAT",
			taskID: task.ID,
			updateCommentRequest: &testclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is still better"),
			},
			files: []*graphql.Upload{
				{
					File:        pdfFile.RawFile,
					Filename:    pdfFile.OriginalName,
					Size:        pdfFile.Size,
					ContentType: pdfFile.ContentType,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "happy path, delete comment",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				DeleteComment: &taskCommentID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, add risk",
			taskID: taskRisk.ID,
			request: &testclient.UpdateTaskInput{
				AddRiskIDs: []string{risk.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "update category using pat of owner",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				Category: lo.ToPtr("risk review"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "update assignee to user not in org should fail",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(testUser2.ID),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: "user not in organization",
		},
		{
			name:   "update assignee to view only user",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(assignee.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "update assignee to same user, should not error",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(assignee.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "update status and details",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				Status:  &enums.TaskStatusInProgress,
				Details: lo.ToPtr("do all the things for the thing"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "add to group",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AddGroupIDs: []string{group.GroupID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "update assigner to another org member, this user should still be able to see it because they originally created it",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				AssignerID: lo.ToPtr(viewOnlyUser2.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "clear assignee",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				ClearAssignee: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "clear assigner",
			taskID: task.ID,
			request: &testclient.UpdateTaskInput{
				ClearAssigner: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
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
					expectUploadNillable(t, suite.client.mockProvider, tc.files)
				}

				commentResp, err = suite.client.api.UpdateTaskComment(testUser1.UserCtx, taskCommentID, *tc.updateCommentRequest, tc.files)
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

				if tc.request.Category != nil {
					assert.Check(t, is.Equal(*tc.request.Category, *resp.UpdateTask.Task.Category))
				}

				if tc.request.ClearAssignee != nil {
					assert.Check(t, is.Nil(resp.UpdateTask.Task.Assignee))

					// the previous assignee should no longer be able to see the task
					_, err = suite.client.api.GetTaskByID(assignee.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))
				}

				if tc.request.ClearAssigner != nil {
					assert.Check(t, is.Nil(resp.UpdateTask.Task.Assignee))

					// the previous assigner should no longer be able to see the task
					_, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))
				}

				if tc.request.AddRiskIDs != nil {
					taskResp, err := suite.client.api.GetTaskByID(adminUser.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.Nil(err))
					assert.Check(t, is.Equal(taskResp.Task.ID, tc.taskID))
				}

				if tc.request.AssignerID != nil {
					assert.Check(t, resp.UpdateTask.Task.Assigner != nil)
					assert.Check(t, is.Equal(*tc.request.AssignerID, resp.UpdateTask.Task.Assigner.ID))

					// make sure the assigner can see the task
					taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, err)
					assert.Check(t, taskResp != nil)

					// make sure the original creator can still see the task
					taskResp, err = suite.client.api.GetTaskByID(adminUser.UserCtx, resp.UpdateTask.Task.ID)
					assert.NilError(t, err)
					assert.Check(t, taskResp != nil)
				}

				if tc.request.AddComment != nil {
					assert.Check(t, len(resp.UpdateTask.Task.Comments.Edges) != 0)
					assert.Check(t, is.Equal(tc.request.AddComment.Text, resp.UpdateTask.Task.Comments.Edges[0].Node.Text))

					// there should only be one comment
					assert.Assert(t, is.Len(resp.UpdateTask.Task.Comments.Edges, 1))
					taskCommentID = resp.UpdateTask.Task.Comments.Edges[0].Node.ID

					// user shouldn't be able to see the comment
					checkResp, err := suite.client.api.GetNoteByID(assignee.UserCtx, taskCommentID)
					assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))

					// user should be able to see the comment since they created the task
					checkResp, err = suite.client.api.GetNoteByID(adminUser.UserCtx, taskCommentID)
					assert.Check(t, err)

					// org owner should be able to see the comment
					checkResp, err = suite.client.api.GetNoteByID(testUser1.UserCtx, taskCommentID)
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
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group.GroupID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTask(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	patClient := suite.setupPatClient(testUser, t)
	task1 := (&TaskBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	task2 := (&TaskBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

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
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete task",
			idToDelete: task1.ID,
			client:     suite.client.api,
			ctx:        testUser.UserCtx,
		},
		{
			name:        "task already deleted, not found",
			idToDelete:  task1.ID,
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: "task not found",
		},
		{
			name:       "happy path, delete task using personal access token",
			idToDelete: task2.ID,
			client:     patClient,
			ctx:        context.Background(),
		},
		{
			name:        "unknown task, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
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
}

func TestMutationUpdateBulkTask(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)

	task1 := (&TaskBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	task2 := (&TaskBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	task3 := (&TaskBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

	taskAnotherUser := (&TaskBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)

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
			client:               suite.client.api,
			ctx:                  testUser.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateTaskInput{Title: lo.ToPtr("test")},
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some tasks not authorized",
			ids:  []string{task1.ID, taskAnotherUser.ID}, // second task should fail authorization
			input: testclient.UpdateTaskInput{
				Title: lo.ToPtr("Updated by authorized user"),
			},
			client:               suite.client.api,
			ctx:                  testUser.UserCtx,
			expectedUpdatedCount: 1, // only task1 should be updated
		},
		{
			name: "update not allowed, no permissions to tasks",
			ids:  []string{task1.ID},
			input: testclient.UpdateTaskInput{
				Title: lo.ToPtr("Should not update"),
			},
			client:               suite.client.api,
			ctx:                  testUser2.UserCtx,
			expectedUpdatedCount: 0, // should not find any tasks to update
		},
		{
			name: "update status on multiple tasks",
			ids:  []string{task1.ID, task2.ID, task3.ID},
			input: testclient.UpdateTaskInput{
				Status: &enums.TaskStatusInProgress,
			},
			client:               suite.client.api,
			ctx:                  testUser.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "assign tasks to org member",
			ids:  []string{task1.ID, task2.ID},
			input: testclient.UpdateTaskInput{
				AssigneeID: &om.UserID,
			},
			client:               suite.client.api,
			ctx:                  testUser.UserCtx,
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

				if tc.input.Category != nil {
					assert.Check(t, is.Equal(*tc.input.Category, *task.Category))
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
				res, err := suite.client.api.GetTaskByID(testUser.UserCtx, task.ID)
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

	// Cleanup created tasks
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: []string{task1.ID, task2.ID, task3.ID}}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskAnotherUser.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: om.ID}).MustDelete(testUser.UserCtx, t)
}
