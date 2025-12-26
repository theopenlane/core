package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGlobalSearch(t *testing.T) {
	// create a new user for this test
	testSearchUser := suite.userBuilder(context.Background(), t)

	testViewOnlyUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testSearchUser.UserCtx, t, &testViewOnlyUser, enums.RoleMember, testSearchUser.OrganizationID)

	testAnotherUser := suite.userBuilder(context.Background(), t)

	// create multiple objects to be searched by testSearchUser
	// dont use objects that might be created by the system, that could be returned such as system owned controls, policies, etc
	// also don't use objects that are created automatically, like groups
	numContacts := 10
	contactIDs := []string{}
	for i := range numContacts {
		contact := (&ContactBuilder{client: suite.client, Name: fmt.Sprintf("Test A1CD2D Contact %d", i)}).MustNew(testSearchUser.UserCtx, t)
		contactIDs = append(contactIDs, contact.ID)
	}

	numPrograms := 3
	programIDs := []string{}
	for i := range numPrograms {
		program := (&ProgramBuilder{client: suite.client, Name: fmt.Sprintf("Test A1CD2D Program %d", i)}).MustNew(testSearchUser.UserCtx, t)
		programIDs = append(programIDs, program.ID)
	}

	testCases := []struct {
		name             string
		client           *testclient.TestClient
		ctx              context.Context
		query            string
		expectResults    bool
		expectedContacts int
		expectedPrograms int
		errExpected      string
	}{
		{
			name:             "happy path",
			client:           suite.client.api,
			ctx:              testSearchUser.UserCtx,
			query:            "A1CD2D",
			expectResults:    true,
			expectedContacts: 10,
			expectedPrograms: 3,
		},
		{
			name:             "happy path, case insensitive with both",
			client:           suite.client.api,
			ctx:              testSearchUser.UserCtx,
			query:            "a1cd2d",
			expectResults:    true,
			expectedContacts: 10,
			expectedPrograms: 3,
		},
		{
			name:             "happy path, case insensitive just contacts",
			client:           suite.client.api,
			ctx:              testSearchUser.UserCtx,
			query:            "a1cd2d con",
			expectResults:    true,
			expectedContacts: 10,
			expectedPrograms: 0,
		},
		{
			name:             "happy path, view only user",
			client:           suite.client.api,
			ctx:              testViewOnlyUser.UserCtx,
			query:            "A1CD2D",
			expectResults:    true,
			expectedContacts: 10,
			expectedPrograms: 0, // no access to the programs by the view only user
		},
		{
			name:          "no results",
			client:        suite.client.api,
			ctx:           testSearchUser.UserCtx,
			query:         "NonExistent RAnD0M Str!ng F0r Sanity",
			expectResults: false,
		},
		{
			name:          "no results, another user",
			client:        suite.client.api,
			ctx:           testAnotherUser.UserCtx,
			query:         "A1CD2D",
			expectResults: false,
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

			if tc.expectResults {
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

			} else {
				assert.Check(t, is.Nil(resp.Search.Contacts))
				assert.Check(t, is.Nil(resp.Search.Programs))
			}
		})
	}

	// clean up the created objects
	(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, IDs: contactIDs}).MustDelete(testSearchUser.UserCtx, t)
	(&Cleanup[*generated.ProgramDeleteOne]{client: suite.client.db.Program, IDs: programIDs}).MustDelete(testSearchUser.UserCtx, t)
}
