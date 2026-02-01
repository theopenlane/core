package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQuerySubprocessorByID(t *testing.T) {
	subprocessor := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// create a file to be used as logo and ensure file is returned properly
	logoFile := uploadFile(t, logoFilePath)
	input := testclient.CreateSubprocessorInput{
		Name: "test subprocessor" + ulids.New().String(),
	}

	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*logoFile})

	systemOwnedSubprocessor, err := suite.client.api.CreateSubprocessor(systemAdminUser.UserCtx, input, logoFile)
	assert.NilError(t, err)

	systemSubprocessr := systemOwnedSubprocessor.CreateSubprocessor.Subprocessor

	testCases := []struct {
		name         string
		expectedName string
		queryID      string
		client       *testclient.TestClient
		ctx          context.Context
		errorMsg     string
	}{
		{
			name:         "happy path",
			expectedName: subprocessor.Name,
			queryID:      subprocessor.ID,
			client:       suite.client.api,
			ctx:          testUser1.UserCtx,
		},
		{
			name:         "happy path, view only user",
			expectedName: subprocessor.Name,
			queryID:      subprocessor.ID,
			client:       suite.client.api,
			ctx:          viewOnlyUser.UserCtx,
		},
		{
			name:     "subprocessor not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:         "not authorized to query org",
			expectedName: subprocessor.Name,
			queryID:      subprocessor.ID,
			client:       suite.client.api,
			ctx:          testUser2.UserCtx,
			errorMsg:     notFoundErrorMsg,
		},
		{
			name:         "happy path, system owned",
			queryID:      systemSubprocessr.ID,
			client:       suite.client.api,
			ctx:          systemAdminUser.UserCtx,
			expectedName: systemSubprocessr.Name,
		},
		{
			name:         "happy path, system owned, regular only user",
			queryID:      systemSubprocessr.ID,
			client:       suite.client.api,
			ctx:          testUser1.UserCtx,
			expectedName: systemSubprocessr.Name,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetSubprocessorByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Subprocessor.ID))
			assert.Check(t, is.Equal(tc.expectedName, resp.Subprocessor.Name))

			if tc.queryID == systemSubprocessr.ID {
				// make sure the file info is present
				assert.Check(t, resp.Subprocessor.LogoFile != nil)
				assert.Check(t, resp.Subprocessor.LogoFile.ID != "")
				assert.Check(t, *resp.Subprocessor.LogoFileID != "")
			}
		})
	}

	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: subprocessor.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: systemSubprocessr.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestQuerySubprocessors(t *testing.T) {
	testUser := suite.userBuilder(context.Background(), t)
	viewUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser.UserCtx, t, &viewUser, enums.RoleMember, testUser.OrganizationID)

	subprocessor1 := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	subprocessor2 := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	subprocessor3 := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	subprocessor4 := (&SubprocessorBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	nonExistentName := "nonexistent-subprocessor"

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		where           *testclient.SubprocessorWhereInput
	}{
		{
			name:            "return all",
			client:          suite.client.api,
			ctx:             testUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "return all, ro user",
			client:          suite.client.api,
			ctx:             viewUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:   "query by name",
			client: suite.client.api,
			ctx:    testUser.UserCtx,
			where: &testclient.SubprocessorWhereInput{
				Name: &subprocessor1.Name,
			},
			expectedResults: 1,
		},
		{
			name:   "query by name, not found",
			client: suite.client.api,
			ctx:    testUser.UserCtx,
			where: &testclient.SubprocessorWhereInput{
				Name: &nonExistentName,
			},
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetSubprocessors(tc.ctx, nil, nil, tc.where)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResults, resp.Subprocessors.TotalCount))

			for _, subprocessor := range resp.Subprocessors.Edges {
				if subprocessor.Node.Name != subprocessor4.Name {
					assert.Check(t, subprocessor.Node.Owner != nil)
					assert.Check(t, is.Equal(subprocessor.Node.Owner.ID, testUser.OrganizationID))
				} else {
					assert.Check(t, subprocessor.Node.Owner == nil)
				}
			}
		})
	}

	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, IDs: []string{subprocessor1.ID, subprocessor2.ID}}).MustDelete(testUser.UserCtx, t)
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: subprocessor3.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: subprocessor4.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateSubprocessor(t *testing.T) {
	createImageUpload := logoFileFunc(t)

	testCases := []struct {
		name            string
		request         testclient.CreateSubprocessorInput
		upload          *graphql.Upload
		client          *testclient.TestClient
		ctx             context.Context
		expectedErr     string
		expectedOwnerID *string
	}{
		{
			name: "happy path",
			request: testclient.CreateSubprocessorInput{
				Name: "Test Subprocessor",
			},
			expectedOwnerID: lo.ToPtr(testUser1.OrganizationID),
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
		},
		{
			name: "happy path with description and logo upload",
			request: testclient.CreateSubprocessorInput{
				Name:        "Test Subprocessor with Description",
				Description: lo.ToPtr("This is a test subprocessor"),
			},
			upload:          createImageUpload(),
			expectedOwnerID: lo.ToPtr(testUser1.OrganizationID),
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
		},
		{
			name: "happy path, adminUser",
			request: testclient.CreateSubprocessorInput{
				Name: "Admin Test Subprocessor",
			},
			expectedOwnerID: lo.ToPtr(testUser1.OrganizationID),
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
		},
		{
			name: "not authorized",
			request: testclient.CreateSubprocessorInput{
				Name:    "Unauthorized Subprocessor",
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing name",
			request: testclient.CreateSubprocessorInput{
				OwnerID: lo.ToPtr(testUser1.OrganizationID),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "name",
		},
		{
			name: "happy path, sysadmin",
			request: testclient.CreateSubprocessorInput{
				Name: "Sys Admin Test Subprocessor",
			},
			client:          suite.client.api,
			ctx:             systemAdminUser.UserCtx,
			expectedOwnerID: nil,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.upload != nil {
				expectUpload(t, suite.client.mockProvider, []graphql.Upload{*tc.upload})
			}

			resp, err := tc.client.CreateSubprocessor(tc.ctx, tc.request, tc.upload)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Name, resp.CreateSubprocessor.Subprocessor.Name))
			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateSubprocessor.Subprocessor.Description))
			}
			if tc.expectedOwnerID == nil {
				assert.Check(t, resp.CreateSubprocessor.Subprocessor.Owner == nil)

			} else {
				assert.Check(t, resp.CreateSubprocessor.Subprocessor.Owner != nil)
				assert.Check(t, is.Equal(*tc.expectedOwnerID, resp.CreateSubprocessor.Subprocessor.Owner.ID))
			}

			if tc.upload != nil {
				assert.Check(t, resp.CreateSubprocessor.Subprocessor.LogoFileID != nil)
				assert.Assert(t, resp.CreateSubprocessor.Subprocessor.LogoFile != nil)
				assert.Check(t, resp.CreateSubprocessor.Subprocessor.LogoFile.ID != "")
			}

			// Clean up
			(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: resp.CreateSubprocessor.Subprocessor.ID}).MustDelete(tc.ctx, t)
		})
	}
}

func TestMutationDeleteSubprocessor(t *testing.T) {
	subprocessor := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subprocessor2 := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subprocessor3 := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	nonExistentID := "non-existent-id"

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "delete subprocessor",
			id:     subprocessor.ID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:        "unauthorized",
			id:          subprocessor3.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "subprocessor not found",
			id:          nonExistentID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteSubprocessor(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.id, resp.DeleteSubprocessor.DeletedID))

			// Verify the subprocessor is deleted
			_, err = tc.client.GetSubprocessorByID(tc.ctx, tc.id)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}

	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, IDs: []string{subprocessor2.ID, subprocessor3.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestUpdateSubprocessor(t *testing.T) {
	subprocessor := (&SubprocessorBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	systemOwnedSubprocessor := (&SubprocessorBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		errorMsg    string
		updateInput testclient.UpdateSubprocessorInput
	}{
		{
			name:    "happy path",
			queryID: subprocessor.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Tags: []string{"updated", "test"},
			},
		},
		{
			name:    "update name and description",
			queryID: subprocessor.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Name:        lo.ToPtr("Updated Subprocessor Name"),
				Description: lo.ToPtr("Updated description"),
			},
		},
		{
			name:    "update name and description, system owned",
			queryID: systemOwnedSubprocessor.ID,
			client:  suite.client.api,
			ctx:     systemAdminUser.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Name:        lo.ToPtr("Updated System Owned Subprocessor Name"),
				Description: lo.ToPtr("Updated system owned description"),
			},
		},
		{
			name:    "update logo remote URL",
			queryID: subprocessor.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				LogoRemoteURL: lo.ToPtr("https://example.com/new-logo.png"),
			},
		},
		{
			name:    "not allowed",
			queryID: subprocessor.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Tags: []string{"unauthorized"},
			},
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:    "not allowed to update system owned",
			queryID: systemOwnedSubprocessor.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Tags: []string{"unauthorized"},
			},
			errorMsg: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateSubprocessor(tc.ctx, tc.queryID, tc.updateInput, nil)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.updateInput.Name != nil {
				assert.Check(t, is.Equal(*tc.updateInput.Name, resp.UpdateSubprocessor.Subprocessor.Name))
			}
			if tc.updateInput.Description != nil {
				assert.Check(t, is.Equal(*tc.updateInput.Description, *resp.UpdateSubprocessor.Subprocessor.Description))
			}
			if tc.updateInput.LogoRemoteURL != nil {
				assert.Check(t, is.Equal(*tc.updateInput.LogoRemoteURL, *resp.UpdateSubprocessor.Subprocessor.LogoRemoteURL))
			}
		})
	}

	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: subprocessor.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: systemOwnedSubprocessor.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestGetAllSubprocessors(t *testing.T) {
	// Clean up any existing subprocessors to ensure clean test state
	deletectx := setContext(systemAdminUser.UserCtx, suite.client.db)
	existingSubprocessors, err := suite.client.db.Subprocessor.Query().All(deletectx)
	assert.NilError(t, err)
	for _, sp := range existingSubprocessors {
		suite.client.db.Subprocessor.DeleteOneID(sp.ID).ExecX(deletectx)
	}

	// Create test subprocessors with different users
	subprocessor1 := (&SubprocessorBuilder{
		client: suite.client,
	}).MustNew(testUser1.UserCtx, t)

	subprocessor2 := (&SubprocessorBuilder{
		client: suite.client,
	}).MustNew(testUser1.UserCtx, t)

	subprocessor3 := (&SubprocessorBuilder{
		client: suite.client,
	}).MustNew(testUser2.UserCtx, t)

	subprocessor4 := (&SubprocessorBuilder{
		client: suite.client,
	}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their subprocessors",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3, // Should see only subprocessors owned by testUser1
		},
		{
			name:            "happy path - admin user sees all subprocessors",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 3, // Should see all owned by testUser1
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 3, // Should see only subprocessors from their organization
		},
		{
			name:            "happy path - different user sees only their subprocessors",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 2, // Should see only subprocessors owned by testUser2
		},
		{
			name:            "happy path - sysadmin",
			client:          suite.client.api,
			ctx:             systemAdminUser.UserCtx,
			expectedResults: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllSubprocessors(tc.ctx)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.Subprocessors.Edges != nil)

			// Verify the number of results
			assert.Check(t, is.Len(resp.Subprocessors.Edges, int(tc.expectedResults)))
			assert.Check(t, is.Equal(tc.expectedResults, resp.Subprocessors.TotalCount))

			// Verify pagination info
			assert.Check(t, resp.Subprocessors.PageInfo.StartCursor != nil)

			// If we have results, verify the structure of the first result
			if tc.expectedResults > 0 {
				firstNode := resp.Subprocessors.Edges[0].Node
				assert.Check(t, len(firstNode.ID) != 0)
				assert.Check(t, len(firstNode.Name) != 0)
				assert.Check(t, firstNode.CreatedAt != nil)
			}

			// Verify that users only see subprocessors from their organization
			if tc.ctx == testUser1.UserCtx || tc.ctx == viewOnlyUser.UserCtx {
				for _, edge := range resp.Subprocessors.Edges {
					if edge.Node.Owner == nil {
						continue
					}
					assert.Check(t, is.Equal(testUser1.OrganizationID, edge.Node.Owner.ID))
				}
			} else if tc.ctx == testUser2.UserCtx {
				for _, edge := range resp.Subprocessors.Edges {
					if edge.Node.Owner == nil {
						continue
					}
					assert.Check(t, is.Equal(testUser2.OrganizationID, edge.Node.Owner.ID))
				}
			}
		})
	}

	// Clean up created subprocessors
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, IDs: []string{subprocessor1.ID, subprocessor2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: subprocessor3.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.SubprocessorDeleteOne]{client: suite.client.db.Subprocessor, ID: subprocessor4.ID}).MustDelete(systemAdminUser.UserCtx, t)
}
