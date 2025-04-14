package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestGlobalSearch() {
	t := suite.T()

	// create multiple objects to be searched by testUser1
	numGroups := 12
	for i := range numGroups {
		(&GroupBuilder{client: suite.client, Name: fmt.Sprintf("Test Group %d", i)}).MustNew(testUser1.UserCtx, t)
	}

	numContacts := 3
	for i := range numContacts {
		(&ContactBuilder{client: suite.client, Name: fmt.Sprintf("Test Contact %d", i)}).MustNew(testUser1.UserCtx, t)
	}

	numPrograms := 3
	for i := range numPrograms {
		(&ProgramBuilder{client: suite.client, Name: fmt.Sprintf("Test Program %d", i)}).MustNew(testUser1.UserCtx, t)
	}

	numControls := 3
	for i := range numControls {
		(&ControlBuilder{client: suite.client, Name: fmt.Sprintf("Test Control %d", i)}).MustNew(testUser1.UserCtx, t)
	}

	testCases := []struct {
		name             string
		client           *openlaneclient.OpenlaneClient
		ctx              context.Context
		query            string
		expectedResults  int
		expectedGroups   int
		expectedContacts int
		expectedPrograms int
		expectedControls int
		errExpected      string
	}{
		{
			name:             "happy path",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			query:            "Test",
			expectedResults:  21, // this is total count of all objects searched
			expectedGroups:   10, // 12 groups created by max results in tests is 10 and we are testing len of edges
			expectedContacts: 3,
			expectedPrograms: 3,
			expectedControls: 3,
		},
		{
			name:             "happy path, case insensitive",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			query:            "TEST",
			expectedResults:  21, // this is total count of all objects searched
			expectedGroups:   10, // 12 groups created by max results in tests is 10 and we are testing len of edges
			expectedContacts: 3,
			expectedPrograms: 3,
			expectedControls: 3,
		},
		{
			name:             "happy path, case insensitive",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			query:            "con",
			expectedResults:  6, // this is total count of all objects searched
			expectedGroups:   0,
			expectedContacts: 3,
			expectedPrograms: 0,
			expectedControls: 3,
		},
		{
			name:             "happy path, view only user",
			client:           suite.client.api,
			ctx:              viewOnlyUser.UserCtx,
			query:            "Test",
			expectedResults:  15, // this is total count of all objects searched
			expectedGroups:   10, // 12 groups created by max results in tests is 10 and we are testing len of edges
			expectedContacts: 3,
			expectedPrograms: 0, // no access to the programs by the view only user
			expectedControls: 0, // no access to the controls by the view only user
		},
		{
			name:            "no results",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			query:           "NonExistent",
			expectedResults: 0,
		},
		{
			name:            "no results, another user",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			query:           "Test",
			expectedResults: 0,
		},
		{
			name:        "empty query",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			query:       "",
			errExpected: "search query is too short",
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GlobalSearch(tc.ctx, tc.query)
			if tc.errExpected != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errExpected)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.expectedResults > 0 {
				assert.Equal(t, resp.Search.TotalCount, int64(tc.expectedResults))

				if tc.expectedGroups > 0 {
					assert.Len(t, resp.Search.Groups.Edges, tc.expectedGroups)
					assert.Equal(t, resp.Search.Groups.TotalCount, int64(numGroups))
					assert.True(t, resp.Search.Groups.PageInfo.HasNextPage)
					assert.False(t, resp.Search.Groups.PageInfo.HasPreviousPage)
				} else {
					assert.Empty(t, resp.Search.Groups)
				}

				if tc.expectedContacts > 0 {
					assert.Len(t, resp.Search.Contacts.Edges, tc.expectedContacts)
				} else {
					assert.Empty(t, resp.Search.Contacts)
				}

				if tc.expectedPrograms > 0 {
					assert.Len(t, resp.Search.Programs.Edges, tc.expectedPrograms)
				} else {
					assert.Empty(t, resp.Search.Programs)
				}

				if tc.expectedControls > 0 {
					assert.Len(t, resp.Search.Controls.Edges, tc.expectedControls)
				} else {
					assert.Empty(t, resp.Search.Controls)
				}
			} else {
				assert.Empty(t, resp.Search.Groups)
				assert.Empty(t, resp.Search.Contacts)
				assert.Empty(t, resp.Search.Programs)
				assert.Empty(t, resp.Search.Controls)
			}
		})
	}
}
