package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestAuditLogList(t *testing.T) {
	testCases := []struct {
		name     string
		queryID  string
		first    *int64
		last     *int64
		after    *string
		before   *string
		where    *openlaneclient.AuditLogWhereInput
		order    []*openlaneclient.AuditLogOrder
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name: "happy path, table with no history",
			where: &openlaneclient.AuditLogWhereInput{
				Table: "APIToken",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, user",
			where: &openlaneclient.AuditLogWhereInput{
				Table: "User",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, user setting",
			where: &openlaneclient.AuditLogWhereInput{
				Table: "UserSetting",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, program",
			where: &openlaneclient.AuditLogWhereInput{
				Table: "Program",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		// figure out why the json unmarshal is failing for this one
		// {
		// 	name: "happy path, group",
		// 	where: &openlaneclient.AuditLogWhereInput{
		// 		Table: "Group",
		// 	},
		// 	client: suite.client.api,
		// 	ctx:    testUser1.UserCtx,
		// },
	}

	for _, tc := range testCases {
		t.Run("Audit Logs "+tc.name, func(t *testing.T) {
			resp, err := tc.client.AuditLogs(testUser1.UserCtx, tc.first, tc.last, tc.after, tc.before, tc.where, tc.order)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}
}
