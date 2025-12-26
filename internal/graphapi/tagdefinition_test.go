package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/common/enums"
)

func TestQueryTagDefinition(t *testing.T) {

	// create an tagDef to be queried using testUser1
	tagDef := (&TagDefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	systemTagDef := (&TagDefinitionBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	// add test cases for querying the TagDefinition
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: tagDef.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path",
			queryID: systemTagDef.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: tagDef.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: tagDef.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "tag definition not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "tag definition not found, using not authorized user",
			queryID:  tagDef.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTagDefinitionByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.TagDefinition.ID))

			if tc.queryID == systemTagDef.ID {
				assert.Check(t, resp.TagDefinition.Name == systemTagDef.Name)
				assert.Check(t, *resp.TagDefinition.Color == systemTagDef.Color)
				assert.Check(t, *resp.TagDefinition.SystemOwned == true)
			} else {
				assert.Check(t, resp.TagDefinition.Name == tagDef.Name)
				assert.Check(t, *resp.TagDefinition.Color == tagDef.Color)
				assert.Check(t, *resp.TagDefinition.SystemOwned == false)
			}
		})
	}

	(&Cleanup[*generated.TagDefinitionDeleteOne]{client: suite.client.db.TagDefinition, ID: tagDef.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TagDefinitionDeleteOne]{client: suite.client.db.TagDefinition, ID: systemTagDef.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateTagDefinition(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateTagDefinitionInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateTagDefinitionInput{
				Name: "mitb",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateTagDefinitionInput{
				Name:        "mitb",
				Aliases:     []string{"matt-is-the-best"},
				Description: lo.ToPtr("Use to mark objects as the premium level of quality"),
				Color:       lo.ToPtr("#08800a"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateTagDefinitionInput{
				Name: "blue-sky",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateTagDefinitionInput{
				Name: "sunshine",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "view only user not allowed",
			request: testclient.CreateTagDefinitionInput{
				Name: "sames",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},

		{
			name:        "missing required field",
			request:     testclient.CreateTagDefinitionInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid color format",
			request: testclient.CreateTagDefinitionInput{
				Name:  "invalid-color-tag",
				Color: lo.ToPtr("invalid"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "field is not a valid hex color code",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTagDefinition(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateTagDefinition.TagDefinition.Name))

			// check optional fields
			if tc.request.Aliases != nil {
				assert.Check(t, is.DeepEqual(tc.request.Aliases, resp.CreateTagDefinition.TagDefinition.Aliases))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateTagDefinition.TagDefinition.Description))
			}

			if tc.request.Color != nil {
				assert.Check(t, is.Equal(*tc.request.Color, *resp.CreateTagDefinition.TagDefinition.Color))
			} else {
				// ensure a default color was set
				assert.Assert(t, resp.CreateTagDefinition.TagDefinition.Color != nil)
			}

			// cleanup each TagDefinition created
			(&Cleanup[*generated.TagDefinitionDeleteOne]{client: suite.client.db.TagDefinition, ID: resp.CreateTagDefinition.TagDefinition.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationCreateTagDefinitionWithAliasLookup(t *testing.T) {
	baseTagResp, err := suite.client.api.CreateTagDefinition(testUser1.UserCtx, testclient.CreateTagDefinitionInput{
		Name:        "red",
		Aliases:     []string{"maroon", "brick", "crimson"},
		Description: lo.ToPtr("Red color tag with aliases"),
		Color:       lo.ToPtr("#ff0000"),
	})
	assert.NilError(t, err)
	assert.Assert(t, baseTagResp != nil)
	baseTagID := baseTagResp.CreateTagDefinition.TagDefinition.ID

	testCases := []struct {
		name                 string
		request              testclient.CreateTagDefinitionInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedName         string
		expectedID           string
		shouldReturnOriginal bool
	}{
		{
			name: "create with alias name returns existing tag - maroon",
			request: testclient.CreateTagDefinitionInput{
				Name: "maroon",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
		{
			name: "create with alias name returns existing tag - brick",
			request: testclient.CreateTagDefinitionInput{
				Name: "brick",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
		{
			name: "create with alias name returns existing tag - crimson",
			request: testclient.CreateTagDefinitionInput{
				Name: "crimson",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
		{
			name: "create with alias name case insensitive - MAROON",
			request: testclient.CreateTagDefinitionInput{
				Name: "MAROON",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
		{
			name: "create with alias name case insensitive - BrIcK",
			request: testclient.CreateTagDefinitionInput{
				Name: "BrIcK",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
		{
			name: "create with actual name returns existing tag",
			request: testclient.CreateTagDefinitionInput{
				Name: "red",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
		{
			name: "create with non-alias name creates new tag",
			request: testclient.CreateTagDefinitionInput{
				Name: "blue",
			},
			client:               suite.client.api,
			ctx:                  testUser1.UserCtx,
			expectedName:         "blue",
			shouldReturnOriginal: false,
		},
		{
			name: "create with alias using PAT returns existing tag",
			request: testclient.CreateTagDefinitionInput{
				Name: "brick",
			},
			client:               suite.client.apiWithPAT,
			ctx:                  context.Background(),
			expectedName:         "red",
			expectedID:           baseTagID,
			shouldReturnOriginal: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTagDefinition(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// verify that the stored tag has the expected name
			assert.Check(t, is.Equal(tc.expectedName, resp.CreateTagDefinition.TagDefinition.Name))

			if tc.shouldReturnOriginal {
				assert.Check(t, is.Equal(tc.expectedID, resp.CreateTagDefinition.TagDefinition.ID))
				assert.Check(t, is.Equal("red", resp.CreateTagDefinition.TagDefinition.Name))
				assert.Check(t, is.DeepEqual([]string{"maroon", "brick", "crimson"}, resp.CreateTagDefinition.TagDefinition.Aliases))
				assert.Check(t, is.Equal("#ff0000", *resp.CreateTagDefinition.TagDefinition.Color))
			} else {
				// new tag ( new id )
				assert.Check(t, resp.CreateTagDefinition.TagDefinition.ID != baseTagID)

				(&Cleanup[*generated.TagDefinitionDeleteOne]{client: suite.client.db.TagDefinition, ID: resp.CreateTagDefinition.TagDefinition.ID}).MustDelete(testUser1.UserCtx, t)
			}
		})
	}

	(&Cleanup[*generated.TagDefinitionDeleteOne]{client: suite.client.db.TagDefinition, ID: baseTagID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTagDefinition(t *testing.T) {
	tagDefinition := (&TagDefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	systemTagDefinition := (&TagDefinitionBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateTagDefinitionInput
		reqID       string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateTagDefinitionInput{
				Color: lo.ToPtr("#abcef0"),
			},
			reqID:  tagDefinition.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "not allowed to update system tag definition",
			request: testclient.UpdateTagDefinitionInput{
				Color: lo.ToPtr("#abcef0"),
			},
			reqID:       systemTagDefinition.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateTagDefinitionInput{
				Aliases:     []string{"something-else"},
				Color:       lo.ToPtr("#abcef1"),
				Description: lo.ToPtr("tag for something cool, yo"),
			},
			reqID:  tagDefinition.ID,
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed by view only user",
			request: testclient.UpdateTagDefinitionInput{
				Color:         lo.ToPtr("#accef0"),
				AppendAliases: []string{"new-alias"},
			},
			reqID:       tagDefinition.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, clear aliases",
			request: testclient.UpdateTagDefinitionInput{
				ClearAliases: lo.ToPtr(true),
			},
			reqID:  tagDefinition.ID,
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateTagDefinitionInput{
				Color: lo.ToPtr("#ddce19"),
			},
			reqID:       tagDefinition.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTagDefinition(tc.ctx, tc.reqID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check updated fields
			if tc.request.Aliases != nil {
				assert.Check(t, is.DeepEqual(tc.request.Aliases, resp.UpdateTagDefinition.TagDefinition.Aliases))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateTagDefinition.TagDefinition.Description))
			}

			if tc.request.Color != nil {
				assert.Check(t, is.Equal(*tc.request.Color, *resp.UpdateTagDefinition.TagDefinition.Color))
			}

			if tc.request.AppendAliases != nil {
				for _, alias := range tc.request.AppendAliases {
					assert.Check(t, lo.Contains(resp.UpdateTagDefinition.TagDefinition.Aliases, alias))
				}
			}

			if tc.request.ClearAliases != nil && *tc.request.ClearAliases {
				assert.Check(t, len(resp.UpdateTagDefinition.TagDefinition.Aliases) == 0)
			}
		})
	}

	(&Cleanup[*generated.TagDefinitionDeleteOne]{client: suite.client.db.TagDefinition, ID: tagDefinition.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTagDefinition(t *testing.T) {
	// create TagDefinitions to be deleted
	tagDefinition1 := (&TagDefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	tagDefinition2 := (&TagDefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	tagDefinition3 := (&TagDefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	tagDefinition4 := (&TagDefinitionBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  tagDefinition1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "view only user cannot delete, not authorized",
			idToDelete:  tagDefinition1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete tagDefinition1",
			idToDelete: tagDefinition1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  tagDefinition1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete tagDefinition2",
			idToDelete: tagDefinition2.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: tagDefinition3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: tagDefinition4.ID,
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
			resp, err := tc.client.DeleteTagDefinition(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTagDefinition.DeletedID))
		})
	}
}

func TestMutationDeleteTagDefinitionInUse(t *testing.T) {
	// create a tag definition
	tagDef := (&TagDefinitionBuilder{
		client: suite.client,
		Name:   "test-tag",
	}).MustNew(testUser1.UserCtx, t)

	// create a workflow definition that uses the tag definition
	ctx := setContext(testUser1.UserCtx, suite.client.db)
	workflowResp, err := suite.client.db.WorkflowDefinition.Create().
		SetName("Test Workflow").
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("control").
		AddTagDefinitionIDs(tagDef.ID).
		Save(ctx)
	assert.NilError(t, err)

	workflowID := workflowResp.ID

	t.Run("delete tag definition in use by workflow definition", func(t *testing.T) {
		_, err := suite.client.api.DeleteTagDefinition(testUser1.UserCtx, tagDef.ID)
		assert.ErrorContains(t, err, "tag definition is in use")
	})

	// clean up the workflow definition using the tag
	(&Cleanup[*generated.WorkflowDefinitionDeleteOne]{client: suite.client.db.WorkflowDefinition, ID: workflowID}).MustDelete(testUser1.UserCtx, t)

	t.Run("tag definition deletion works if no workflow definition using it", func(t *testing.T) {
		resp, err := suite.client.api.DeleteTagDefinition(testUser1.UserCtx, tagDef.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(tagDef.ID, resp.DeleteTagDefinition.DeletedID))
	})
}
