package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
)

func TestMutationSubmitTrustCenterNDADocAccess(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterDocProtected := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter.ID, Visibility: enums.TrustCenterDocumentVisibilityProtected}).MustNew(testUser1.UserCtx, t)

	uploadFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
	require.Nil(t, err)
	up := graphql.Upload{
		File:        uploadFile.File,
		Filename:    uploadFile.Filename,
		Size:        uploadFile.Size,
		ContentType: uploadFile.ContentType,
	}
	expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{up})

	trustCenterNDA, err := suite.client.api.CreateTrustCenterNda(testUser1.UserCtx, testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}, []*graphql.Upload{&up})

	require.Nil(t, err)
	require.NotNil(t, trustCenterNDA)

	// Create anonymous trust center context helper
	anonUserID := fmt.Sprintf("anon_%s", ulids.New().String())

	anonUser := &auth.AnonymousTrustCenterUser{
		SubjectID:          anonUserID,
		SubjectName:        "Anonymous User",
		OrganizationID:     trustCenter.OwnerID,
		AuthenticationType: auth.JWTAuthentication,
		TrustCenterID:      trustCenter.ID,
		SubjectEmail:       "test@example.com",
	}

	input := testclient.SubmitTrustCenterNDAResponseInput{
		TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
		Response: map[string]any{
			"signatory_info": map[string]any{
				"email": "test@example.com",
			},
			"acknowledgment": true,
			"signature_metadata": map[string]any{
				"ip_address": "192.168.1.100",
				"timestamp":  "2025-09-22T19:37:59.988Z",
				"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
				"user_id":    anonUserID,
			},
			"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
			"trust_center_id": trustCenter.ID,
		},
	}

	ctx := context.Background()
	anonCtx := auth.WithAnonymousTrustCenterUser(ctx, anonUser)

	// check that the anonymous user can't query the protected doc's files
	getTrustCenterDocResp, err := suite.client.api.GetTrustCenterDocByID(anonCtx, trustCenterDocProtected.ID)
	require.Nil(t, err)
	require.Nil(t, getTrustCenterDocResp.TrustCenterDoc.File)

	resp, err := suite.client.api.SubmitTrustCenterNDAResponse(anonCtx, input)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// now, check that the anonymous user can query the protected doc's files
	getTrustCenterDocResp, err = suite.client.api.GetTrustCenterDocByID(anonCtx, trustCenterDocProtected.ID)
	require.Nil(t, err)
	assert.Assert(t, getTrustCenterDocResp.TrustCenterDoc.File != nil)

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: trustCenterNDA.CreateTrustCenterNda.Template.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: trustCenterDocProtected.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationSendTrustCenterNDAEmail(t *testing.T) {
	// Create test trust centers
	trustCenter1 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	// Create anonymous trust center context helper
	createAnonymousTrustCenterContext := func(trustCenterID, organizationID string) context.Context {
		anonUserID := fmt.Sprintf("anon_%s", ulids.New().String())

		anonUser := &auth.AnonymousTrustCenterUser{
			SubjectID:          anonUserID,
			SubjectName:        "Anonymous User",
			OrganizationID:     organizationID,
			AuthenticationType: auth.JWTAuthentication,
			TrustCenterID:      trustCenterID,
		}

		ctx := context.Background()
		return auth.WithAnonymousTrustCenterUser(ctx, anonUser)
	}

	testCases := []struct {
		name        string
		input       testclient.SendTrustCenterNDAInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - anonymous user sends email for their trust center",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
		},
		{
			name: "happy path - system admin can send email for any trust center",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client: suite.client.api,
			ctx:    systemAdminUser.UserCtx,
		},
		{
			name: "not authorized - anonymous user tries to send for different trust center",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter2.ID, // Different trust center
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID), // Anonymous user for trustCenter1
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - regular user cannot send NDA email",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - view only user cannot send NDA email",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "trust center not found",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: "non-existent-id",
				Email:         gofakeit.Email(),
			},
			client:      suite.client.api,
			ctx:         systemAdminUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "invalid email format",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         "invalid-email",
			},
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
		},
		{
			name: "empty email",
			input: testclient.SendTrustCenterNDAInput{
				TrustCenterID: trustCenter1.ID,
				Email:         "",
			},
			client: suite.client.api,
			ctx:    createAnonymousTrustCenterContext(trustCenter1.ID, testUser1.OrganizationID),
			// Note: Empty email validation might be handled at GraphQL schema level
			expectedErr: "email is required",
		},
	}

	for _, tc := range testCases {
		t.Run("Send "+tc.name, func(t *testing.T) {
			resp, err := tc.client.SendTrustCenterNDAEmail(tc.ctx, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify the response indicates success
			assert.Check(t, resp.SendTrustCenterNDAEmail.Success)
		})
	}

	// Clean up trust centers
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestCreateTrustCenterNDA(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	testCases := []struct {
		name     string
		ctx      context.Context
		input    testclient.CreateTrustCenterNDAInput
		errorMsg string
		uploads  []string
	}{
		{
			name: "happy path",
			ctx:  testUser1.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: trustCenter.ID,
			},
			uploads: []string{"testdata/uploads/hello.pdf"},
		},
		{
			name: "missing upload",
			ctx:  testUser1.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: trustCenter.ID,
			},
			errorMsg: "one NDA file is required",
		},
		{
			name: "missing trust center",
			ctx:  testUser1.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: "",
			},
			uploads:  []string{"testdata/uploads/hello.pdf"},
			errorMsg: notFoundErrorMsg,
		},
		{
			name: "Other user cannot create NDA",
			ctx:  testUser2.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: trustCenter.ID,
			},
			uploads:  []string{"testdata/uploads/hello.pdf"},
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uploads := []*graphql.Upload{}
			expectUploads := []graphql.Upload{}
			for _, file := range tc.uploads {
				uploadFile, err := objects.NewUploadFile(file)
				assert.NilError(t, err)
				up := graphql.Upload{
					File:        uploadFile.File,
					Filename:    uploadFile.Filename,
					Size:        uploadFile.Size,
					ContentType: uploadFile.ContentType,
				}

				expectUploads = append(expectUploads, up)
				uploads = append(uploads, &up)
			}
			if len(uploads) > 0 {
				expectUpload(t, suite.client.objectStore.Storage, expectUploads)
			}
			resp, err := suite.client.api.CreateTrustCenterNda(tc.ctx, tc.input, uploads)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: resp.CreateTrustCenterNda.Template.ID}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestAnonymousUserCanQueryTrustCenterNDA(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	input := testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}
	uploadFiles := []string{"testdata/uploads/hello.pdf"}
	uploads := []*graphql.Upload{}
	expectUploads := []graphql.Upload{}
	for _, file := range uploadFiles {
		uploadFile, err := objects.NewUploadFile(file)
		assert.NilError(t, err)
		up := graphql.Upload{
			File:        uploadFile.File,
			Filename:    uploadFile.Filename,
			Size:        uploadFile.Size,
			ContentType: uploadFile.ContentType,
		}

		expectUploads = append(expectUploads, up)
		uploads = append(uploads, &up)
	}
	if len(uploads) > 0 {
		expectUpload(t, suite.client.objectStore.Storage, expectUploads)
	}
	resp, err := suite.client.api.CreateTrustCenterNda(testUser1.UserCtx, input, uploads)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// check we can't create a second NDA
	_, err = suite.client.api.CreateTrustCenterNda(testUser1.UserCtx, input, uploads)
	assert.ErrorContains(t, err, "template already exists")

	// check an anonymous user for this trust center can query the NDA
	queryResp, err := suite.client.api.GetAllTemplates(createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID))
	assert.NilError(t, err)
	assert.Assert(t, queryResp != nil)
	assert.Check(t, len(queryResp.Templates.Edges) == 1)
	assert.Check(t, queryResp.Templates.Edges[0].Node.ID == resp.CreateTrustCenterNda.Template.ID)

	// ... and that an anonymous user for a different trust center cannot query the NDA
	queryResp2, err := suite.client.api.GetAllTemplates(createAnonymousTrustCenterContext(trustCenter2.ID, testUser2.OrganizationID))

	assert.NilError(t, err)
	assert.Assert(t, queryResp2 != nil)
	assert.Check(t, len(queryResp2.Templates.Edges) == 0)

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: resp.CreateTrustCenterNda.Template.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestSubmitTrustCenterNDAResponse(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	uploadFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
	require.Nil(t, err)
	up := graphql.Upload{
		File:        uploadFile.File,
		Filename:    uploadFile.Filename,
		Size:        uploadFile.Size,
		ContentType: uploadFile.ContentType,
	}
	expectUpload(t, suite.client.objectStore.Storage, []graphql.Upload{up})

	trustCenterNDA, err := suite.client.api.CreateTrustCenterNda(testUser1.UserCtx, testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}, []*graphql.Upload{&up})

	require.Nil(t, err)
	require.NotNil(t, trustCenterNDA)

	anonUserID := fmt.Sprintf("anon_%s", ulids.New().String())
	anonUserID2 := fmt.Sprintf("anon_%s", ulids.New().String())

	anonUser := &auth.AnonymousTrustCenterUser{
		SubjectID:          anonUserID,
		SubjectName:        "Anonymous User",
		OrganizationID:     trustCenter.OwnerID,
		AuthenticationType: auth.JWTAuthentication,
		TrustCenterID:      trustCenter.ID,
		SubjectEmail:       "test@example.com",
	}
	anonUser2 := &auth.AnonymousTrustCenterUser{
		SubjectID:          anonUserID2,
		SubjectName:        "Anonymous User",
		OrganizationID:     trustCenter2.OwnerID,
		AuthenticationType: auth.JWTAuthentication,
		TrustCenterID:      trustCenter2.ID,
		SubjectEmail:       "testother@example.com",
	}

	ctx := context.Background()
	anonCtx := auth.WithAnonymousTrustCenterUser(ctx, anonUser)
	anonCtx2 := auth.WithAnonymousTrustCenterUser(ctx, anonUser2)

	testCases := []struct {
		name     string
		ctx      context.Context
		input    testclient.SubmitTrustCenterNDAResponseInput
		errorMsg string
	}{
		{
			name: "happy path",
			ctx:  anonCtx,
			input: testclient.SubmitTrustCenterNDAResponseInput{
				TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
				Response: map[string]any{
					"signatory_info": map[string]any{
						"email": "test@example.com",
					},
					"acknowledgment": true,
					"signature_metadata": map[string]any{
						"ip_address": "192.168.1.100",
						"timestamp":  "2025-09-22T19:37:59.988Z",
						"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
		},
		{
			name: "Does not conform to format",
			ctx:  anonCtx,
			input: testclient.SubmitTrustCenterNDAResponseInput{
				TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
				Response: map[string]any{
					"signatory_info": map[string]any{
						"email": "test@example.com",
					},
				},
			},
			errorMsg: "validation failed:",
		},
		{
			name: "authed to wrong trust center",
			ctx:  anonCtx2,
			input: testclient.SubmitTrustCenterNDAResponseInput{
				TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
				Response: map[string]any{
					"signatory_info": map[string]any{
						"email": "test@example.com",
					},
					"acknowledgment": true,
					"signature_metadata": map[string]any{
						"ip_address": "192.168.1.100",
						"timestamp":  "2025-09-22T19:37:59.988Z",
						"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: "NDA submission does not match authenticated user",
		},
		{
			name: "wrong trust center ID",
			ctx:  anonCtx,
			input: testclient.SubmitTrustCenterNDAResponseInput{
				TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
				Response: map[string]any{
					"signatory_info": map[string]any{
						"email": "test@example.com",
					},
					"acknowledgment": true,
					"signature_metadata": map[string]any{
						"ip_address": "192.168.1.100",
						"timestamp":  "2025-09-22T19:37:59.988Z",
						"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": "test123",
				},
			},
			errorMsg: "validation failed:",
		},
		{
			name: "email mismatch",
			ctx:  anonCtx,
			input: testclient.SubmitTrustCenterNDAResponseInput{
				TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
				Response: map[string]any{
					"signatory_info": map[string]any{
						"email": "wrongemail@yahoo.com",
					},
					"acknowledgment": true,
					"signature_metadata": map[string]any{
						"ip_address": "192.168.1.100",
						"timestamp":  "2025-09-22T19:37:59.988Z",
						"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: "NDA submission does not match authenticated user",
		},
		{
			name: "wrong user ID",
			ctx:  anonCtx,
			input: testclient.SubmitTrustCenterNDAResponseInput{
				TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
				Response: map[string]any{
					"signatory_info": map[string]any{
						"email": "test@example.com",
					},
					"acknowledgment": true,
					"signature_metadata": map[string]any{
						"ip_address": "192.168.1.100",
						"timestamp":  "2025-09-22T19:37:59.988Z",
						"pdf_hash":   "a1b2c3d4e5f6789012345678901234567890abcd",
						"user_id":    "abc123",
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: "NDA submission does not match authenticated user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.client.api.SubmitTrustCenterNDAResponse(tc.ctx, tc.input)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			(&Cleanup[*generated.DocumentDataDeleteOne]{client: suite.client.db.DocumentData, ID: resp.SubmitTrustCenterNDAResponse.DocumentData.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: trustCenterNDA.CreateTrustCenterNda.Template.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
}
