package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQuerySubprocessorByID(t *testing.T) {
	subprocessor := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create a file to be used as logo and ensure file is returned properly
	logoFile := th.UploadFile(t, th.LogoFilePath)
	input := testclient.CreateSubprocessorInput{
		Name: "test subprocessor" + ulids.New().String(),
	}

	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*logoFile})

	systemOwnedSubprocessor, err := suite.Client.API.CreateSubprocessor(th.SharedSystemAdminUser.UserCtx, input, logoFile, nil)
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
			client:       suite.Client.API,
			ctx:          th.SharedTestUser1.UserCtx,
		},
		{
			name:         "happy path, view only user",
			expectedName: subprocessor.Name,
			queryID:      subprocessor.ID,
			client:       suite.Client.API,
			ctx:          th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:     "subprocessor not found",
			queryID:  "non-existent-id",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:         "not authorized to query org",
			expectedName: subprocessor.Name,
			queryID:      subprocessor.ID,
			client:       suite.Client.API,
			ctx:          th.SharedTestUser2.UserCtx,
			errorMsg:     th.NotFoundErrorMsg,
		},
		{
			name:         "happy path, system owned",
			queryID:      systemSubprocessr.ID,
			client:       suite.Client.API,
			ctx:          th.SharedSystemAdminUser.UserCtx,
			expectedName: systemSubprocessr.Name,
		},
		{
			name:         "happy path, system owned, regular only user",
			queryID:      systemSubprocessr.ID,
			client:       suite.Client.API,
			ctx:          th.SharedTestUser1.UserCtx,
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

	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: systemSubprocessr.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestQuerySubprocessors(t *testing.T) {
	localTestOrg := suite.SeedFreshMinimalOrgUsers(t, false)
	testUser := localTestOrg.Owner
	viewUser := localTestOrg.Member

	subprocessor1 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)
	(&th.SubprocessorBuilder{Client: suite.Client}).MustNew(testUser.UserCtx, t)
	subprocessor3 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)
	subprocessor4 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

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
			client:          suite.Client.API,
			ctx:             testUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "return all, ro user",
			client:          suite.Client.API,
			ctx:             viewUser.UserCtx,
			expectedResults: 3,
		},
		{
			name:   "query by name",
			client: suite.Client.API,
			ctx:    testUser.UserCtx,
			where: &testclient.SubprocessorWhereInput{
				Name: &subprocessor1.Name,
			},
			expectedResults: 1,
		},
		{
			name:   "query by name, not found",
			client: suite.Client.API,
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

	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor3.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor4.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)
}

func TestMutationCreateSubprocessor(t *testing.T) {
	createImageUpload := th.LogoFileFunc(t)

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
			expectedOwnerID: lo.ToPtr(th.SharedTestUser1.OrganizationID),
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path with description and logo upload",
			request: testclient.CreateSubprocessorInput{
				Name:        "Test Subprocessor with Description",
				Description: lo.ToPtr("This is a test subprocessor"),
			},
			upload:          createImageUpload(),
			expectedOwnerID: lo.ToPtr(th.SharedTestUser1.OrganizationID),
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, adminUser",
			request: testclient.CreateSubprocessorInput{
				Name: "Admin Test Subprocessor",
			},
			expectedOwnerID: lo.ToPtr(th.SharedTestUser1.OrganizationID),
			client:          suite.Client.API,
			ctx:             th.SharedAdminUser.UserCtx,
		},
		{
			name: "not authorized",
			request: testclient.CreateSubprocessorInput{
				Name:    "Unauthorized Subprocessor",
				OwnerID: lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing name",
			request: testclient.CreateSubprocessorInput{
				OwnerID: lo.ToPtr(th.SharedTestUser1.OrganizationID),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "name",
		},
		{
			name: "happy path, sysadmin",
			request: testclient.CreateSubprocessorInput{
				Name: "Sys Admin Test Subprocessor",
			},
			client:          suite.Client.API,
			ctx:             th.SharedSystemAdminUser.UserCtx,
			expectedOwnerID: nil,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.upload != nil {
				th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*tc.upload})
			}

			resp, err := tc.client.CreateSubprocessor(tc.ctx, tc.request, tc.upload, nil)
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
			(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: resp.CreateSubprocessor.Subprocessor.ID}).MustDelete(tc.ctx, t)
		})
	}
}

func TestMutationDeleteSubprocessor(t *testing.T) {
	subprocessor := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subprocessor2 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subprocessor3 := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:        "unauthorized",
			id:          subprocessor3.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:        "subprocessor not found",
			id:          nonExistentID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
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
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)
		})
	}

	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, IDs: []string{subprocessor2.ID, subprocessor3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestUpdateSubprocessor(t *testing.T) {
	subprocessor := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	systemOwnedSubprocessor := (&th.SubprocessorBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

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
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Tags: []string{"updated", "test"},
			},
		},
		{
			name:    "update name and description",
			queryID: subprocessor.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Name:        lo.ToPtr("Updated Subprocessor Name"),
				Description: lo.ToPtr("Updated description"),
			},
		},
		{
			name:    "update name and description, system owned",
			queryID: systemOwnedSubprocessor.ID,
			client:  suite.Client.API,
			ctx:     th.SharedSystemAdminUser.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Name:        lo.ToPtr("Updated System Owned Subprocessor Name"),
				Description: lo.ToPtr("Updated system owned description"),
			},
		},
		{
			name:    "update logo remote URL",
			queryID: subprocessor.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				LogoRemoteURL: lo.ToPtr("https://example.com/new-logo.png"),
			},
		},
		{
			name:    "not allowed",
			queryID: subprocessor.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Tags: []string{"unauthorized"},
			},
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name:    "not allowed to update system owned",
			queryID: systemOwnedSubprocessor.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			updateInput: testclient.UpdateSubprocessorInput{
				Tags: []string{"unauthorized"},
			},
			errorMsg: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateSubprocessor(tc.ctx, tc.queryID, tc.updateInput, nil, nil)

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

	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: systemOwnedSubprocessor.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestGetAllSubprocessors(t *testing.T) {
	// Clean up any existing subprocessors to ensure clean test state
	deletectx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)
	existingSubprocessors, err := suite.Client.DB.Subprocessor.Query().All(deletectx)
	assert.NilError(t, err)
	for _, sp := range existingSubprocessors {
		suite.Client.DB.Subprocessor.DeleteOneID(sp.ID).ExecX(deletectx)
	}

	// Create test subprocessors with different users
	subprocessor1 := (&th.SubprocessorBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	subprocessor2 := (&th.SubprocessorBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	subprocessor3 := (&th.SubprocessorBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedTestUser2.UserCtx, t)

	subprocessor4 := (&th.SubprocessorBuilder{
		Client: suite.Client,
	}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their subprocessors",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 3, // Should see only subprocessors owned by th.SharedTestUser1
		},
		{
			name:            "happy path - admin user sees all subprocessors",
			client:          suite.Client.API,
			ctx:             th.SharedAdminUser.UserCtx,
			expectedResults: 3, // Should see all owned by th.SharedTestUser1
		},
		{
			name:            "happy path - view only user",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 3, // Should see only subprocessors from their organization
		},
		{
			name:            "happy path - different user sees only their subprocessors",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 2, // Should see only subprocessors owned by testUser2
		},
		{
			name:   "happy path - sysadmin",
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
			// a system admin holds CapBypassOrgFilter, which the subprocessor interceptor honors, so it
			// sees every organization's subprocessors (all four created above), not just system-owned ones
			expectedResults: 4,
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
			if tc.ctx == th.SharedTestUser1.UserCtx || tc.ctx == th.SharedViewOnlyUser.UserCtx {
				for _, edge := range resp.Subprocessors.Edges {
					if edge.Node.Owner == nil {
						continue
					}
					assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, edge.Node.Owner.ID))
				}
			} else if tc.ctx == th.SharedTestUser2.UserCtx {
				for _, edge := range resp.Subprocessors.Edges {
					if edge.Node.Owner == nil {
						continue
					}
					assert.Check(t, is.Equal(th.SharedTestUser2.OrganizationID, edge.Node.Owner.ID))
				}
			}
		})
	}

	// Clean up created subprocessors
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, IDs: []string{subprocessor1.ID, subprocessor2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor3.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor4.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}
