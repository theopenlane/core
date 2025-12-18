package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/entx/history"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestAuditLogList(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	orgUser := suite.userBuilder(context.Background(), t)

	reqCtx := orgUser.UserCtx

	otherUser := suite.userBuilder(context.Background(), t)
	otherUserCtx := otherUser.UserCtx

	// create a program to test that it is returned in the audit logs
	program1 := (&ProgramBuilder{client: suite.client, WithProcedure: true, WithPolicy: true}).MustNew(reqCtx, t)

	// Update the program to ensure that the update is logged
	_, err := suite.client.api.UpdateProgram(reqCtx, program1.ID,
		testclient.UpdateProgramInput{
			Name:        lo.ToPtr("Updated Program"),
			Description: lo.ToPtr("Updated description"),
		})
	assert.NilError(t, err)

	numControls := 22
	for range numControls {
		(&ControlBuilder{client: suite.client}).MustNew(reqCtx, t)
	}

	afterCursor := ""
	updateOp := history.OpTypeUpdate.String()
	createOp := history.OpTypeInsert.String()

	testCases := []struct {
		name          string
		queryID       string
		first         *int64
		last          *int64
		after         *string
		before        *string
		where         *testclient.AuditLogWhereInput
		order         *testclient.AuditLogOrder
		client        *testclient.TestClient
		ctx           context.Context
		expectedCount int
		totalCount    int
		hasNext       bool
		hasPrevious   bool
		checkForID    string
		setAfter      bool
		errorMsg      string
	}{
		{
			name: "happy path, table with no history",
			where: &testclient.AuditLogWhereInput{
				Table: "APIToken",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 0, // No history for APIToken
		},
		{
			name: "happy path, user",
			where: &testclient.AuditLogWhereInput{
				Table: "User",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 2, // user creation should be logged and update after the user is created
			checkForID:    orgUser.ID,
		},
		{
			name: "happy path, user setting",
			where: &testclient.AuditLogWhereInput{
				Table: "UserSetting",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 2, // user setting creation should be logged and update after the user setting is updated
		},
		{
			name: "happy path, program",
			where: &testclient.AuditLogWhereInput{
				Table: "Program",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 2, //program creation, update should be logged
			checkForID:    program1.ID,
		},
		{
			name: "happy path, program update",
			where: &testclient.AuditLogWhereInput{
				Table:     "Program",
				Operation: &updateOp,
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 1,
			checkForID:    program1.ID,
		},
		{
			name: "happy path, program create",
			where: &testclient.AuditLogWhereInput{
				Table:     "Program",
				Operation: &createOp,
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 1,
			checkForID:    program1.ID,
		},
		{
			name: "happy path, procedure",
			where: &testclient.AuditLogWhereInput{
				Table: "Procedure",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 1, // procedure creation should be logged with program creation
		},
		{
			name: "happy path, policy",
			where: &testclient.AuditLogWhereInput{
				Table: "InternalPolicy",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 1, // policy creation should be logged with program creation
		},
		{
			name: "happy path, control",
			where: &testclient.AuditLogWhereInput{
				Table: "Control",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 10,          // default pagination in tests is 10
			totalCount:    numControls, // we created 22 tasks
			hasNext:       true,        // there are more tasks than the default pagination
			hasPrevious:   false,
			setAfter:      true,
		},
		{
			name: "happy path, control page 2",
			where: &testclient.AuditLogWhereInput{
				Table: "Control",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 10,          // default pagination in tests is 10
			totalCount:    numControls, // we created 22 tasks
			hasNext:       true,        // there are more tasks than the default pagination
			hasPrevious:   true,
			after:         &afterCursor,
			setAfter:      true,
		},
		{
			name: "happy path, control page 3",
			where: &testclient.AuditLogWhereInput{
				Table: "Control",
			},
			client:        suite.client.api,
			ctx:           reqCtx,
			expectedCount: 2,
			totalCount:    numControls, // we created 22 tasks
			hasNext:       false,       // page 3 is the last page
			hasPrevious:   true,
			after:         &afterCursor,
		},
		{
			name:   "missing table parameter",
			where:  &testclient.AuditLogWhereInput{},
			client: suite.client.api,
			ctx:    reqCtx,
		},
		{
			name:   "nil where input",
			where:  nil,
			client: suite.client.api,
			ctx:    reqCtx,
		},
		{
			name: "happy path, another user shouldn't see these procedure audit logs",
			where: &testclient.AuditLogWhereInput{
				Table: "Procedure",
			},
			client:        suite.client.api,
			ctx:           otherUserCtx,
			expectedCount: 0, // procedure creation should be logged with program creation
		},
	}

	for _, tc := range testCases {
		t.Run("Audit Logs "+tc.name, func(t *testing.T) {
			resp, err := tc.client.AuditLogs(tc.ctx, tc.first, tc.last, tc.after, tc.before, tc.where, tc.order)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.AuditLogs.Edges, tc.expectedCount), "expected %d audit logs, got %d", tc.expectedCount, len(resp.AuditLogs.Edges))

			// if the expected count is less than the default pagination,
			// then we should have the total count equal to the expected count
			if tc.expectedCount < 10 && tc.totalCount == 0 {
				tc.totalCount = tc.expectedCount
			}

			// check the pagination info
			assert.Check(t, is.Equal(resp.AuditLogs.TotalCount, int64(tc.totalCount)))
			assert.Check(t, is.Equal(resp.AuditLogs.PageInfo.HasNextPage, tc.hasNext))
			assert.Check(t, is.Equal(resp.AuditLogs.PageInfo.HasPreviousPage, tc.hasPrevious))

			if tc.checkForID != "" {
				// check that the audit logs contain the expected IDs
				for _, node := range resp.AuditLogs.Edges {
					assert.Check(t, is.Equal(node.Node.ID, tc.checkForID))
				}
			}

			if tc.setAfter {
				// set the after cursor for the next page
				afterCursor = *resp.AuditLogs.PageInfo.EndCursor
			}
		})
	}
}
