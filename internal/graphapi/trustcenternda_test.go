package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestMutationSubmitTrustCenterNDADocAccess(t *testing.T) {
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	trustCenterDocProtected := (&TrustCenterDocBuilder{client: suite.client, TrustCenterID: trustCenter.ID, Visibility: enums.TrustCenterDocumentVisibilityProtected}).MustNew(tcOrg.owner.UserCtx, t)

	up := uploadFile(t, pdfFilePath)
	pdfHash := getMD5Hash(t, pdfFilePath)
	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*up})

	trustCenterNDA, err := suite.client.api.CreateTrustCenterNda(tcOrg.owner.UserCtx, testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}, []*graphql.Upload{up})

	assert.NilError(t, err)
	assert.Assert(t, trustCenterNDA != nil)

	// Create anonymous trust center context helper
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())

	email := "test@example.com"

	anonUser := auth.NewTrustCenterCaller(trustCenter.OwnerID, anonUserID, "Anonymous User", email)

	anonCtxForRequest := newAnonTrustCenterCtxFromCaller(anonUser, trustCenter.ID)
	ndaCreateResp, err := suite.client.api.CreateTrustCenterNDARequest(anonCtxForRequest, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     "Test",
		LastName:      "User",
		CompanyName:   lo.ToPtr("Test Company"),
		Email:         email,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	assert.Assert(t, ndaCreateResp != nil)
	// make sure the nda request is in requested status, the approval is off by default
	assert.Check(t, *ndaCreateResp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status == enums.TrustCenterNDARequestStatusRequested)

	input := testclient.SubmitTrustCenterNDAResponseInput{
		TemplateID: trustCenterNDA.CreateTrustCenterNda.Template.ID,
		Response: map[string]any{
			"signatory_info": map[string]any{
				"email": email,
			},
			"acknowledgment": true,
			"signature_metadata": map[string]any{
				"ip_address": "192.168.1.100",
				"timestamp":  "2025-09-22T19:37:59.988Z",
				"pdf_hash":   pdfHash,
				"user_id":    anonUserID,
			},
			"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
			"trust_center_id": trustCenter.ID,
		},
	}

	anonCtx := newAnonTrustCenterCtxFromCaller(anonUser, trustCenter.ID)

	// check that the anonymous user can't query the protected doc's files
	getTrustCenterDocResp, err := suite.client.api.GetTrustCenterDocByID(anonCtx, trustCenterDocProtected.ID)
	assert.NilError(t, err)
	assert.Assert(t, getTrustCenterDocResp.TrustCenterDoc.OriginalFile == nil)

	// Clear any existing jobs and emails before submitting
	err = suite.client.db.Job.TruncateRiverTables(tcOrg.owner.UserCtx)
	assert.NilError(t, err)

	suite.mockEmailSender().Reset()
	expectAttestedUpload(t, suite.client.mockProvider)

	resp, err := suite.client.api.SubmitTrustCenterNDAResponse(anonCtx, input)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// make sure the nda request is marked as signed
	ndaRequest, err := suite.client.api.GetTrustCenterNDARequests(tcOrg.owner.UserCtx, nil, nil, nil, nil, []*testclient.TrustCenterNDARequestOrder{}, &testclient.TrustCenterNDARequestWhereInput{
		Email: &email,
	})
	assert.NilError(t, err)
	assert.Assert(t, len(ndaRequest.TrustCenterNdaRequests.Edges) == 1)
	assert.Equal(t, ndaRequest.TrustCenterNdaRequests.Edges[0].Node.Status.String(), enums.TrustCenterNDARequestStatusSigned.String())
	assert.Check(t, ndaRequest.TrustCenterNdaRequests.Edges[0].Node.SignedAt != nil)

	// wait for the NDA attestation listener to process the document data creation
	suite.WaitForEvents()

	// verify the signed NDA email was sent with the attested PDF attached
	msgs := suite.mockEmailSender().Messages()
	assert.Assert(t, len(msgs) == 1, "expected 1 email after NDA signing, got %d", len(msgs))
	assert.Assert(t, len(msgs[0].Attachments) == 1, "expected signed PDF attachment")
	assert.Equal(t, "signed_nda_file.pdf", msgs[0].Attachments[0].Filename)
	assert.Assert(t, len(msgs[0].Attachments[0].Content) > 0, "expected non-empty PDF content in attachment")

	// now, check that the anonymous user can query the protected doc's files
	getTrustCenterDocResp, err = suite.client.api.GetTrustCenterDocByID(anonCtx, trustCenterDocProtected.ID)
	assert.NilError(t, err)
	assert.Assert(t, getTrustCenterDocResp.TrustCenterDoc.OriginalFile != nil)

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestCreateTrustCenterNDA(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	testCases := []struct {
		name     string
		ctx      context.Context
		input    testclient.CreateTrustCenterNDAInput
		errorMsg string
		uploads  []string
	}{
		{
			name: "happy path",
			ctx:  tcOrg.admin.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: trustCenter.ID,
			},
			uploads: []string{pdfFilePath},
		},
		{
			name: "missing upload",
			ctx:  tcOrg.owner.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: trustCenter.ID,
			},
			errorMsg: "one NDA file is required",
		},
		{
			name: "Other user cannot create NDA",
			ctx:  sharedTestUser2.UserCtx,
			input: testclient.CreateTrustCenterNDAInput{
				TrustCenterID: trustCenter.ID,
			},
			uploads:  []string{pdfFilePath},
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uploads := []*graphql.Upload{}
			expectUploads := []graphql.Upload{}
			for _, file := range tc.uploads {
				uploadFile, err := storage.NewUploadFile(file)
				assert.NilError(t, err)
				up := graphql.Upload{
					File:        uploadFile.RawFile,
					Filename:    uploadFile.OriginalName,
					Size:        uploadFile.Size,
					ContentType: uploadFile.ContentType,
				}

				expectUploads = append(expectUploads, up)
				uploads = append(uploads, &up)
			}
			if len(uploads) > 0 {
				expectUpload(t, suite.client.mockProvider, expectUploads)
			}

			if tc.errorMsg != "" {
				expectDelete(t, suite.client.mockProvider, expectUploads)
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

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}

func TestAnonymousUserCanQueryTrustCenterNDA(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	input := testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}

	uploadFiles := []string{pdfFilePath}
	uploads := []*graphql.Upload{}
	expectUploads := []graphql.Upload{}

	for _, file := range uploadFiles {
		up := uploadFile(t, file)
		expectUploads = append(expectUploads, *up)
		uploads = append(uploads, up)
	}

	if len(uploads) > 0 {
		expectUpload(t, suite.client.mockProvider, expectUploads)
	}

	resp, err := suite.client.api.CreateTrustCenterNda(tcOrg.owner.UserCtx, input, uploads)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// check we can't create a second NDA
	// expect an upload and a delete since the upload will be rolled back on error
	expectUpload(t, suite.client.mockProvider, expectUploads)
	expectDelete(t, suite.client.mockProvider, expectUploads)
	_, err = suite.client.api.CreateTrustCenterNda(tcOrg.owner.UserCtx, input, uploads)
	assert.ErrorContains(t, err, "template already exists")

	// check an anonymous user for this trust center can query the NDA
	queryResp, err := suite.client.api.GetAllTemplates(createAnonymousTrustCenterContext(trustCenter.ID, tcOrg.organizationID))
	assert.NilError(t, err)
	assert.Assert(t, queryResp != nil)
	assert.Check(t, len(queryResp.Templates.Edges) == 1)
	assert.Check(t, queryResp.Templates.Edges[0].Node.ID == resp.CreateTrustCenterNda.Template.ID)

	// ... and that an anonymous user for a different trust center cannot query the NDA
	queryResp2, err := suite.client.api.GetAllTemplates(createAnonymousTrustCenterContext(trustCenter2.ID, tcOrg2.organizationID))

	assert.NilError(t, err)
	assert.Assert(t, queryResp2 != nil)
	assert.Check(t, len(queryResp2.Templates.Edges) == 0)

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestSubmitTrustCenterNDAResponse(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	tcOrg2 := createFreshOrgWithTrustCenter(t)
	trustCenter2 := tcOrg2.trustCenter

	up := uploadFile(t, pdfFilePath)
	pdfHash := getMD5Hash(t, pdfFilePath)

	// the happy path triggers attestNDADocument which uploads the attested PDF
	expectAttestedUpload(t, suite.client.mockProvider)

	trustCenterNDA, err := suite.client.api.CreateTrustCenterNda(tcOrg.owner.UserCtx, testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}, []*graphql.Upload{up})

	assert.NilError(t, err)
	assert.Assert(t, trustCenterNDA != nil)

	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	anonUserID2 := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())

	anonUser := auth.NewTrustCenterCaller(trustCenter.OwnerID, anonUserID, "Anonymous User", "test@example.com")
	anonUser2 := auth.NewTrustCenterCaller(trustCenter2.OwnerID, anonUserID2, "Anonymous User", "testother@example.com")

	anonCtx := newAnonTrustCenterCtxFromCaller(anonUser, trustCenter.ID)
	anonCtx2 := newAnonTrustCenterCtxFromCaller(anonUser2, trustCenter2.ID)

	_, err = suite.client.api.CreateTrustCenterNDARequest(anonCtx, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     "Test",
		LastName:      "User",
		CompanyName:   lo.ToPtr("Test Company"),
		Email:         "test@example.com",
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	testCases := []struct {
		name     string
		ctx      context.Context
		input    testclient.SubmitTrustCenterNDAResponseInput
		errorMsg string
	}{
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
						"pdf_hash":   pdfHash,
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: notFoundErrorMsg,
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
						"pdf_hash":   pdfHash,
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": "test123",
				},
			},
			errorMsg: "NDA submission does not match authenticated user",
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
						"pdf_hash":   pdfHash,
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
						"pdf_hash":   pdfHash,
						"user_id":    "abc123",
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: "NDA submission does not match authenticated user",
		},
		{
			name: "invalid pdf file ID",
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
						"timestamp":  "2026-06-29T19:37:59.988Z",
						"pdf_hash":   pdfHash,
						"user_id":    anonUserID,
					},
					"pdf_file_id":     "non-existent-id",
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: "NDA PDF file does not match the template",
		},
		{
			name: "invalid pdf hash",
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
						"timestamp":  "2026-06-29T20:37:59.988Z",
						"pdf_hash":   "invalid hash",
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
			errorMsg: "NDA PDF hash does not match template",
		},
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
						"pdf_hash":   pdfHash,
						"user_id":    anonUserID,
					},
					"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
					"trust_center_id": trustCenter.ID,
				},
			},
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
			(&Cleanup[*generated.DocumentDataDeleteOne]{client: suite.client.db.DocumentData, ID: resp.SubmitTrustCenterNDAResponse.DocumentData.ID}).MustDelete(tcOrg.owner.UserCtx, t)
		})
	}

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(tcOrg2.owner.UserCtx, t)
}

func TestUpdateTrustCenterNDA(t *testing.T) {
	t.Parallel()
	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter

	up1 := uploadFile(t, pdfFilePath)
	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*up1})

	createResp, err := suite.client.api.CreateTrustCenterNda(tcOrg.owner.UserCtx, testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}, []*graphql.Upload{up1})

	assert.NilError(t, err)
	assert.Assert(t, createResp != nil)
	assert.Check(t, len(createResp.CreateTrustCenterNda.Template.Files.Edges) == 1)

	fileID := createResp.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID

	secondUpload := uploadFile(t, logoFilePath)
	expectUpload(t, suite.client.mockProvider, []graphql.Upload{*secondUpload})

	updateResp, err := suite.client.api.UpdateTrustCenterNda(tcOrg.owner.UserCtx, trustCenter.ID, []*graphql.Upload{secondUpload})

	assert.NilError(t, err)
	assert.Assert(t, updateResp != nil)

	assert.Check(t, len(updateResp.UpdateTrustCenterNda.Template.Files.Edges) == 1)

	newFileID := updateResp.UpdateTrustCenterNda.Template.Files.Edges[0].Node.ID
	assert.Check(t, newFileID != fileID)

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}
