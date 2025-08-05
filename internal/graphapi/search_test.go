package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGlobalSearch(t *testing.T) {
	t.Parallel()

	// create a new user for this test
	testSearchUser := suite.userBuilder(context.Background(), t)

	testViewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testSearchUser.UserCtx, t, &testViewOnlyUser, enums.RoleMember, testSearchUser.OrganizationID)

	// create multiple objects to be searched by testSearchUser
	numGroups := 12
	groupIDs := []string{}
	for i := range numGroups {
		group := (&GroupBuilder{client: suite.client, Name: fmt.Sprintf("Test Group %d", i)}).MustNew(testSearchUser.UserCtx, t)
		groupIDs = append(groupIDs, group.ID)
	}

	numContacts := 3
	contactIDs := []string{}
	for i := range numContacts {
		contact := (&ContactBuilder{client: suite.client, Name: fmt.Sprintf("Test Contact %d", i)}).MustNew(testSearchUser.UserCtx, t)
		contactIDs = append(contactIDs, contact.ID)
	}

	numPrograms := 3
	programIDs := []string{}
	for i := range numPrograms {
		program := (&ProgramBuilder{client: suite.client, Name: fmt.Sprintf("Test Program %d", i)}).MustNew(testSearchUser.UserCtx, t)
		programIDs = append(programIDs, program.ID)
	}

	numControls := 3
	controlIDs := []string{}
	for i := range numControls {
		control := (&ControlBuilder{client: suite.client, Name: fmt.Sprintf("Test Control %d", i)}).MustNew(testSearchUser.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	testCases := []struct {
		name             string
		client           *testclient.TestClient
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
			ctx:              testSearchUser.UserCtx,
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
			ctx:              testSearchUser.UserCtx,
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
			ctx:              testSearchUser.UserCtx,
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
			ctx:              testViewOnlyUser.UserCtx,
			query:            "Test",
			expectedResults:  18, // this is total count of all objects searched
			expectedGroups:   10, // 12 groups created by max results in tests is 10 and we are testing len of edges
			expectedContacts: 3,
			expectedPrograms: 0, // no access to the programs by the view only user
			expectedControls: 3,
		},
		{
			name:            "no results",
			client:          suite.client.api,
			ctx:             testSearchUser.UserCtx,
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
			ctx:         testSearchUser.UserCtx,
			query:       "",
			errExpected: "search query is too short",
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GlobalSearch(tc.ctx, tc.query)
			if tc.errExpected != "" {

				assert.Assert(t, is.Contains(err.Error(), tc.errExpected))
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.expectedResults > 0 {
				assert.Check(t, is.Equal(resp.Search.TotalCount, int64(tc.expectedResults)))

				if tc.expectedGroups > 0 {
					assert.Assert(t, resp.Search.Groups != nil)
					assert.Check(t, is.Len(resp.Search.Groups.Edges, tc.expectedGroups))
					assert.Check(t, is.Equal(resp.Search.Groups.TotalCount, int64(numGroups)))
					assert.Check(t, resp.Search.Groups.PageInfo.HasNextPage)
					assert.Check(t, !resp.Search.Groups.PageInfo.HasPreviousPage)
				} else {
					assert.Check(t, is.Nil(resp.Search.Groups))
				}

				if tc.expectedContacts > 0 {
					assert.Assert(t, resp.Search.Contacts != nil)
					assert.Check(t, is.Len(resp.Search.Contacts.Edges, tc.expectedContacts))
				} else {
					assert.Check(t, is.Nil(resp.Search.Contacts))
				}

				if tc.expectedPrograms > 0 {
					assert.Assert(t, resp.Search.Programs != nil)
					assert.Check(t, is.Len(resp.Search.Programs.Edges, tc.expectedPrograms))
				} else {
					assert.Check(t, is.Nil(resp.Search.Programs))
				}

				if tc.expectedControls > 0 {
					assert.Assert(t, resp.Search.Controls != nil)
					assert.Check(t, is.Len(resp.Search.Controls.Edges, tc.expectedControls))
				} else {
					assert.Check(t, is.Nil(resp.Search.Controls))
				}
			} else {
				assert.Check(t, is.Nil(resp.Search.Groups))
				assert.Check(t, is.Nil(resp.Search.Contacts))
				assert.Check(t, is.Nil(resp.Search.Programs))
				assert.Check(t, is.Nil(resp.Search.Controls))
			}
		})
	}

	// clean up the created objects
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: groupIDs}).MustDelete(testSearchUser.UserCtx, t)
	(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, IDs: contactIDs}).MustDelete(testSearchUser.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: programIDs}).MustDelete(testSearchUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: controlIDs}).MustDelete(testSearchUser.UserCtx, t)
}
