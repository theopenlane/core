package graphapi_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

func TestQueryTrustCenterDocByID(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterDocProtected := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter.ID, Visibility: enums.TrustCenterDocumentVisibilityProtected}).MustNew(testUser1.UserCtx, t)
	trustCenterDocPublic := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter.ID, Visibility: enums.TrustCenterDocumentVisibilityPubliclyVisible}).MustNew(testUser1.UserCtx, t)
	trustCenterDocNotVisible := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter.ID, Visibility: enums.TrustCenterDocumentVisibilityNotVisible}).MustNew(testUser1.UserCtx, t)

	signedNdaAnonCtx := createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID)
	signedNDAAnonUser, _ := auth.AnonymousTrustCenterUserFromContext(signedNdaAnonCtx)
	req := fgax.TupleRequest{
		SubjectID:   signedNDAAnonUser.SubjectID,
		SubjectType: "user",
		ObjectID:    trustCenter.ID,
		ObjectType:  "trust_center",
		Relation:    "nda_signed",
	}

	tuple := fgax.GetTupleKey(req)
	if _, err := suite.client.db.Authz.WriteTupleKeys(testUser1.UserCtx, []fgax.TupleKey{tuple}, nil); err != nil {
		requireNoError(err)
	}
	testCases := []struct {
		name                  string
		queryID               string
		client                *testclient.TestClient
		ctx                   context.Context
		errorMsg              string
		shouldShowFileDetails bool
	}{
		{
			name:                  "happy path",
			queryID:               trustCenterDocProtected.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			shouldShowFileDetails: true,
		},
		{
			name:                  "happy path, view only user",
			queryID:               trustCenterDocProtected.ID,
			client:                suite.client.api,
			ctx:                   viewOnlyUser.UserCtx,
			shouldShowFileDetails: true,
		},
		{
			name:     "trust center doc not found",
			queryID:  "non-existent-id",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not authorized to query org",
			queryID:  trustCenterDocProtected.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:                  "anonymous user can query trust center doc, but can't see files",
			queryID:               trustCenterDocProtected.ID,
			client:                suite.client.api,
			ctx:                   createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID),
			shouldShowFileDetails: false,
		},
		{
			name:                  "anonymous user can query trust center doc and files with signed nda",
			queryID:               trustCenterDocProtected.ID,
			client:                suite.client.api,
			ctx:                   signedNdaAnonCtx,
			shouldShowFileDetails: true,
		},
		{
			name:                  "anonymous user can query trust center doc, can see public files",
			queryID:               trustCenterDocPublic.ID,
			client:                suite.client.api,
			ctx:                   createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID),
			shouldShowFileDetails: true,
		},
		{
			name:                  "nda anonymous user can query trust center doc, can see public files",
			queryID:               trustCenterDocPublic.ID,
			client:                suite.client.api,
			ctx:                   signedNdaAnonCtx,
			shouldShowFileDetails: true,
		},
		{
			name:     "anonymous user can't see not visible docs",
			queryID:  trustCenterDocNotVisible.ID,
			client:   suite.client.api,
			ctx:      createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID),
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "nda anonymous user can't see not visible docs",
			queryID:  trustCenterDocNotVisible.ID,
			client:   suite.client.api,
			ctx:      signedNdaAnonCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:                  "org user can see not visible docs",
			queryID:               trustCenterDocNotVisible.ID,
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			shouldShowFileDetails: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterDocByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterDoc.ID))
			assert.Check(t, resp.TrustCenterDoc.Title != "")
			assert.Check(t, resp.TrustCenterDoc.Category != "")
			assert.Check(t, resp.TrustCenterDoc.TrustCenterID != nil)
			assert.Check(t, resp.TrustCenterDoc.OriginalFileID != nil)
			if tc.shouldShowFileDetails {
				assert.Check(t, resp.TrustCenterDoc.OriginalFile != nil)
			} else {
				assert.Check(t, resp.TrustCenterDoc.OriginalFile == nil)
			}

		})
	}

	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocProtected.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocPublic.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocNotVisible.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateTrustCenterDoc(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Helper function to create fresh file uploads for each test case
	createPDFUpload := func() *graphql.Upload {
		pdfFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pdfFile.File,
			Filename:    pdfFile.Filename,
			Size:        pdfFile.Size,
			ContentType: pdfFile.ContentType,
		}
	}

	createTXTUpload := func() *graphql.Upload {
		txtFile, err := objects.NewUploadFile("testdata/uploads/hello.txt")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        txtFile.File,
			Filename:    txtFile.Filename,
			Size:        txtFile.Size,
			ContentType: txtFile.ContentType,
		}
	}

	testCases := []struct {
		name        string
		input       testclient.CreateTrustCenterDocInput
		file        *graphql.Upload
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, create trust center doc with PDF file",
			input: testclient.CreateTrustCenterDocInput{
				Title:         "Test Document",
				Category:      "Policy",
				TrustCenterID: &trustCenter.ID,
				Tags:          []string{"test", "document"},
				Visibility:    &enums.TrustCenterDocumentVisibilityPubliclyVisible,
			},
			file:   createPDFUpload(),
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, minimal required fields",
			input: testclient.CreateTrustCenterDocInput{
				Title:         "Minimal Document",
				Category:      "General",
				TrustCenterID: &trustCenter.ID,
			},
			file:   createPDFUpload(),
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using personal access token",
			input: testclient.CreateTrustCenterDocInput{
				Title:         "PAT Document",
				Category:      "Policy",
				TrustCenterID: &trustCenter.ID,
			},
			file:   createPDFUpload(),
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "invalid file type",
			input: testclient.CreateTrustCenterDocInput{
				Title:    "Invalid File Document",
				Category: "Policy",
			},
			file:        createTXTUpload(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "unsupported mime type",
		},
		{
			name: "missing title",
			input: testclient.CreateTrustCenterDocInput{
				Category:      "Policy",
				TrustCenterID: &trustCenter.ID,
			},
			file:        createPDFUpload(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "title",
		},
		{
			name: "missing category",
			input: testclient.CreateTrustCenterDocInput{
				Title:         "Test Document",
				TrustCenterID: &trustCenter.ID,
			},
			file:        createPDFUpload(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "category",
		},
		{
			name: "not authorized, view only user",
			input: testclient.CreateTrustCenterDocInput{
				Title:         "Unauthorized Document",
				Category:      "Policy",
				TrustCenterID: &trustCenter.ID,
			},
			file:        createPDFUpload(),
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized, different user",
			input: testclient.CreateTrustCenterDocInput{
				Title:         "Different User Document",
				Category:      "Policy",
				TrustCenterID: &trustCenter.ID,
			},
			file:        createPDFUpload(),
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.file != nil {
				expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{*tc.file})
			}

			resp, err := tc.client.CreateTrustCenterDoc(tc.ctx, tc.input, *tc.file)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the created trust center doc
			trustCenterDoc := resp.CreateTrustCenterDoc.TrustCenterDoc
			assert.Check(t, trustCenterDoc.ID != "")
			assert.Check(t, is.Equal(tc.input.Title, trustCenterDoc.Title))
			assert.Check(t, is.Equal(tc.input.Category, trustCenterDoc.Category))

			// Check trust center ID
			if tc.input.TrustCenterID != nil {
				assert.Check(t, is.Equal(*tc.input.TrustCenterID, *trustCenterDoc.TrustCenterID))
			}

			// Check visibility
			if tc.input.Visibility != nil {
				assert.Check(t, is.Equal(*tc.input.Visibility, *trustCenterDoc.Visibility))
			} else {
				// Default visibility should be NOT_VISIBLE
				assert.Check(t, is.Equal(enums.TrustCenterDocumentVisibilityNotVisible, *trustCenterDoc.Visibility))
			}

			// Check tags
			if len(tc.input.Tags) > 0 {
				assert.Check(t, is.DeepEqual(tc.input.Tags, trustCenterDoc.Tags))
			}

			// Verify file was uploaded and associated
			assert.Check(t, trustCenterDoc.OriginalFileID != nil)
			assert.Check(t, trustCenterDoc.OriginalFile != nil)
			assert.Check(t, trustCenterDoc.OriginalFile.ID != "")

			// Verify timestamps
			assert.Check(t, trustCenterDoc.CreatedAt != nil)
			assert.Check(t, trustCenterDoc.UpdatedAt != nil)
			assert.Check(t, trustCenterDoc.CreatedBy != nil)

			// Clean up the created trust center doc
			(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDoc.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// Clean up the trust center
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenterDocs(t *testing.T) {
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	trustCenterDoc1 := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter1.ID, Visibility: enums.TrustCenterDocumentVisibilityProtected}).MustNew(testUser1.UserCtx, t)
	trustCenterDoc2 := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter2.ID}).MustNew(testUser2.UserCtx, t)
	signedNdaAnonCtx := createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID)
	signedNDAAnonUser, _ := auth.AnonymousTrustCenterUserFromContext(signedNdaAnonCtx)
	req := fgax.TupleRequest{
		SubjectID:   signedNDAAnonUser.SubjectID,
		SubjectType: "user",
		ObjectID:    trustCenter1.ID,
		ObjectType:  "trust_center",
		Relation:    "nda_signed",
	}

	tuple := fgax.GetTupleKey(req)
	if _, err := suite.client.db.Authz.WriteTupleKeys(testUser1.UserCtx, []fgax.TupleKey{tuple}, nil); err != nil {
		requireNoError(err)
	}

	testCases := []struct {
		name                  string
		client                *testclient.TestClient
		ctx                   context.Context
		expectedResults       int64
		where                 *testclient.TrustCenterDocWhereInput
		errorMsg              string
		shouldShowFileDetails bool
	}{
		{
			name:                  "return all for user1",
			client:                suite.client.api,
			ctx:                   testUser1.UserCtx,
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
		{
			name:                  "return all, ro user",
			client:                suite.client.api,
			ctx:                   viewOnlyUser.UserCtx,
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
		{
			name:   "query by trust center ID",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterDocWhereInput{
				TrustCenterID: &trustCenter1.ID,
			},
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
		{
			name:   "query by title",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterDocWhereInput{
				Title: &trustCenterDoc1.Title,
			},
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
		{
			name:   "query by category",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterDocWhereInput{
				Category: &trustCenterDoc1.Category,
			},
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
		{
			name:   "query by visibility",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterDocWhereInput{
				Visibility: &trustCenterDoc1.Visibility,
			},
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
		{
			name:   "query by non-existent title",
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
			where: &testclient.TrustCenterDocWhereInput{
				Title: lo.ToPtr("non-existent-title"),
			},
			expectedResults: 0,
		},
		{
			name:   "anonymous user can query trust center docs",
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
			where: &testclient.TrustCenterDocWhereInput{
				Title: &trustCenterDoc1.Title,
			},
			expectedResults:       1,
			shouldShowFileDetails: false,
		},
		{
			name:   "anonymous user with signed can query trust center docs and view files",
			client: suite.client.api,
			ctx:    signedNdaAnonCtx,
			where: &testclient.TrustCenterDocWhereInput{
				Title: &trustCenterDoc1.Title,
			},
			expectedResults:       1,
			shouldShowFileDetails: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterDocs(tc.ctx, nil, nil, tc.where)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenterDocs.TotalCount))
			for _, node := range resp.TrustCenterDocs.Edges {
				assert.Check(t, node.Node != nil)
				assert.Check(t, node.Node.Title != "")
				assert.Check(t, node.Node.Category != "")
				assert.Check(t, node.Node.TrustCenterID != nil)
				assert.Check(t, node.Node.OriginalFileID != nil)
				if tc.shouldShowFileDetails {
					assert.Check(t, node.Node.OriginalFile != nil)
				} else {
					assert.Check(t, node.Node.OriginalFile == nil)
				}
			}
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDoc1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDoc2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationUpdateTrustCenterDoc(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterDoc := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name             string
		trustCenterDocID string
		request          testclient.UpdateTrustCenterDocInput
		client           *testclient.TestClient
		ctx              context.Context
		expectedErr      string
	}{
		{
			name:             "happy path, update title",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Title: lo.ToPtr("Updated Document Title"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:             "happy path, update category",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Category: lo.ToPtr("UpdatedCategory"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:             "happy path, update visibility",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Visibility: &enums.TrustCenterDocumentVisibilityProtected,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:             "happy path, update multiple fields",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Title:      lo.ToPtr("Multi-field Update Title"),
				Category:   lo.ToPtr("MultiCategory"),
				Visibility: &enums.TrustCenterDocumentVisibilityPubliclyVisible,
				Tags:       []string{"multi", "update"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:             "happy path, using personal access token",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Title: lo.ToPtr("PAT Updated Title"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:             "not authorized, view only user",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Title: lo.ToPtr("Unauthorized Update"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:             "not authorized, different org user",
			trustCenterDocID: trustCenterDoc.ID,
			request: testclient.UpdateTrustCenterDocInput{
				Title: lo.ToPtr("Unauthorized Update"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:             "trust center doc not found",
			trustCenterDocID: "non-existent-id",
			request: testclient.UpdateTrustCenterDocInput{
				Title: lo.ToPtr("Update Non-existent"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTrustCenterDoc(tc.ctx, tc.trustCenterDocID, tc.request, nil, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.trustCenterDocID, resp.UpdateTrustCenterDoc.TrustCenterDoc.ID))

			// Check updated fields
			if tc.request.Title != nil {
				assert.Check(t, is.Equal(*tc.request.Title, resp.UpdateTrustCenterDoc.TrustCenterDoc.Title))
			}

			if tc.request.Category != nil {
				assert.Check(t, is.Equal(*tc.request.Category, resp.UpdateTrustCenterDoc.TrustCenterDoc.Category))
			}

			if tc.request.Visibility != nil {
				assert.Check(t, is.Equal(*tc.request.Visibility, *resp.UpdateTrustCenterDoc.TrustCenterDoc.Visibility))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateTrustCenterDoc.TrustCenterDoc.Tags))
			}
		})
	}

	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDoc.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTrustCenterDoc(t *testing.T) {
	t.Parallel()
	// Create new test users
	testUser := suite.userBuilder(context.Background(), t)
	testUserOther := suite.userBuilder(testUser.UserCtx, t)

	// create objects to be deleted
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUserOther.UserCtx, t)

	trustCenterDoc1 := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter1.ID}).MustNew(testUser.UserCtx, t)
	trustCenterDoc2 := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter2.ID}).MustNew(testUserOther.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path, delete trust center doc",
			idToDelete: trustCenterDoc1.ID,
			client:     suite.client.api,
			ctx:        testUser.UserCtx,
		},
		{
			name:        "not authorized, different org user",
			idToDelete:  trustCenterDoc2.ID,
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "trust center doc not found",
			idToDelete:  "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterDoc(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustCenterDoc.DeletedID))
		})
	}
}

func TestGetAllTrustCenterDocs(t *testing.T) {
	// Clean up any existing trust center docs
	deletectx := setContext(systemAdminUser.UserCtx, suite.client.db)
	d, err := suite.client.db.TrustCenterDoc.Query().All(deletectx)
	require.Nil(t, err)
	for _, doc := range d {
		suite.client.db.TrustCenterDoc.DeleteOneID(doc.ID).ExecX(deletectx)
	}

	// Create test trust center docs with different users
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	trustCenterDocProtected := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter1.ID, Visibility: enums.TrustCenterDocumentVisibilityProtected}).MustNew(testUser1.UserCtx, t)
	trustCenterDocPublic := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter1.ID, Visibility: enums.TrustCenterDocumentVisibilityPubliclyVisible}).MustNew(testUser1.UserCtx, t)
	trustCenterDocNotVisible := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter1.ID, Visibility: enums.TrustCenterDocumentVisibilityNotVisible}).MustNew(testUser1.UserCtx, t)

	trustCenterDoc2 := (&TrustCenterDocBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter2.ID,
	}).MustNew(testUser2.UserCtx, t)

	signedNdaAnonCtx := createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID)
	signedNDAAnonUser, _ := auth.AnonymousTrustCenterUserFromContext(signedNdaAnonCtx)
	req := fgax.TupleRequest{
		SubjectID:   signedNDAAnonUser.SubjectID,
		SubjectType: "user",
		ObjectID:    trustCenter1.ID,
		ObjectType:  "trust_center",
		Relation:    "nda_signed",
	}

	tuple := fgax.GetTupleKey(req)
	if _, err := suite.client.db.Authz.WriteTupleKeys(testUser1.UserCtx, []fgax.TupleKey{tuple}, nil); err != nil {
		requireNoError(err)
	}

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int64
		expectedErr     string
	}{
		{
			name:            "happy path - regular user sees only their trust center docs",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3, // Should see only trust center docs owned by testUser1
		},
		{
			name:            "happy path - admin user sees all trust center docs",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: 3, // Should see all owned by testUser
		},
		{
			name:            "happy path - view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 3, // Should see only trust center docs from their organization
		},
		{
			name:            "happy path - different user sees only their trust center docs",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 1, // Should see only trust center docs owned by testUser2
		},
		{
			name:            "anonymous user can't see not visible docs",
			client:          suite.client.api,
			ctx:             createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
			expectedResults: 2,
		},
		{
			name:            "nda anonymous user can't see not visible docs",
			client:          suite.client.api,
			ctx:             signedNdaAnonCtx,
			expectedResults: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllTrustCenterDocs(tc.ctx)

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.TrustCenterDocs.Edges != nil)

			// Verify the number of results
			assert.Check(t, is.Len(resp.TrustCenterDocs.Edges, int(tc.expectedResults)))
			assert.Check(t, is.Equal(tc.expectedResults, resp.TrustCenterDocs.TotalCount))

			// Verify pagination info
			assert.Check(t, resp.TrustCenterDocs.PageInfo.StartCursor != nil)

			// If we have results, verify the structure of the first result
			if tc.expectedResults > 0 {
				firstNode := resp.TrustCenterDocs.Edges[0].Node
				assert.Check(t, len(firstNode.ID) != 0)
				assert.Check(t, len(firstNode.Title) != 0)
				assert.Check(t, len(firstNode.Category) != 0)
				assert.Check(t, firstNode.TrustCenterID != nil)
				assert.Check(t, firstNode.CreatedAt != nil)
			}

			// Verify that users only see trust center docs from their organization
			if tc.ctx == testUser1.UserCtx || tc.ctx == viewOnlyUser.UserCtx {
				for _, edge := range resp.TrustCenterDocs.Edges {
					assert.Check(t, is.Equal(trustCenter1.ID, *edge.Node.TrustCenterID))
				}
			} else if tc.ctx == testUser2.UserCtx {
				for _, edge := range resp.TrustCenterDocs.Edges {
					assert.Check(t, is.Equal(trustCenter2.ID, *edge.Node.TrustCenterID))
				}
			}
		})
	}

	// Clean up created trust center docs
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDoc2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocProtected.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocPublic.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocNotVisible.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}
