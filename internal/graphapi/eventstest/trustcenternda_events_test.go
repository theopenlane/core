package eventstest_test

import (
	"fmt"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
)

func TestMutationSubmitTrustCenterNDADocAccess(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.NDAAutoApprovalListeners())
	assert.NilError(t, err)

	t.Cleanup(setup.Teardown)

	tcOrg := th.CreateFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.TrustCenter

	_, err = suite.Client.API.UpdateTrustCenter(tcOrg.Owner.UserCtx, trustCenter.ID, testclient.UpdateTrustCenterInput{
		UpdateTrustCenterSetting: &testclient.UpdateTrustCenterSettingInput{
			EnableAutoApproval: lo.ToPtr(true),
		},
	})
	assert.NilError(t, err)

	trustCenterDocProtected := (&th.TrustCenterDocBuilder{Client: suite.Client, TrustCenterID: trustCenter.ID, Visibility: enums.TrustCenterDocumentVisibilityProtected}).MustNew(tcOrg.Owner.UserCtx, t)

	up := th.UploadFile(t, th.PdfFilePath)
	pdfHash := th.GetMD5Hash(t, th.PdfFilePath)
	th.ExpectUpload(t, suite.Client.MockProvider, []graphql.Upload{*up})

	trustCenterNDA, err := suite.Client.API.CreateTrustCenterNda(tcOrg.Owner.UserCtx, testclient.CreateTrustCenterNDAInput{
		TrustCenterID: trustCenter.ID,
	}, []*graphql.Upload{up})

	assert.NilError(t, err)
	assert.Assert(t, trustCenterNDA != nil)

	// Create anonymous trust center context helper
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())

	email := "test@example.com"

	anonUser := auth.NewTrustCenterCaller(trustCenter.OwnerID, anonUserID, "Anonymous User", email)

	anonCtxForRequest := th.NewAnonTrustCenterCtxFromCaller(anonUser, trustCenter.ID)
	ndaCreateResp, err := suite.Client.API.CreateTrustCenterNDARequest(anonCtxForRequest, testclient.CreateTrustCenterNDARequestInput{
		FirstName:     "Test",
		LastName:      "User",
		CompanyName:   lo.ToPtr("Test Company"),
		Email:         email,
		TrustCenterID: &trustCenter.ID,
	})
	assert.NilError(t, err)

	assert.Assert(t, ndaCreateResp != nil)
	assert.Check(t, *ndaCreateResp.CreateTrustCenterNDARequest.TrustCenterNDARequest.Status == enums.TrustCenterNDARequestStatusRequested)

	// the access email mints the signing identity from the request id, not the browsing session
	ndaRequestID := ndaCreateResp.CreateTrustCenterNDARequest.TrustCenterNDARequest.ID

	waitForEvents()

	approvedRequest, err := suite.Client.API.GetTrustCenterNDARequestByID(tcOrg.Owner.UserCtx, ndaRequestID)
	assert.NilError(t, err)
	assert.Equal(t, enums.TrustCenterNDARequestStatusApproved, *approvedRequest.TrustCenterNDARequest.Status)
	anonCtx, signer := th.CreateAnonymousTrustCenterContextForSubject(trustCenter.ID, trustCenter.OwnerID, ndaRequestID, email)

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
				"user_id":    signer.SubjectID,
			},
			"pdf_file_id":     trustCenterNDA.CreateTrustCenterNda.Template.Files.Edges[0].Node.ID,
			"trust_center_id": trustCenter.ID,
		},
	}

	// check that the anonymous user can't query the protected doc's files
	getTrustCenterDocResp, err := suite.Client.API.GetTrustCenterDocByID(anonCtx, trustCenterDocProtected.ID)
	assert.NilError(t, err)
	assert.Assert(t, getTrustCenterDocResp.TrustCenterDoc.OriginalFile == nil)

	// Clear any existing jobs and emails before submitting
	err = suite.Client.DB.Job.TruncateRiverTables(tcOrg.Owner.UserCtx)
	assert.NilError(t, err)

	mockEmailSender().Reset()
	th.ExpectAttestedUpload(t, suite.Client.MockProvider)

	resp, err := suite.Client.API.SubmitTrustCenterNDAResponse(anonCtx, input)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	// make sure the nda request is marked as signed
	ndaRequest, err := suite.Client.API.GetTrustCenterNDARequests(tcOrg.Owner.UserCtx, nil, nil, nil, nil, []*testclient.TrustCenterNDARequestOrder{}, &testclient.TrustCenterNDARequestWhereInput{
		Email: &email,
	})
	assert.NilError(t, err)
	assert.Assert(t, len(ndaRequest.TrustCenterNdaRequests.Edges) == 1)
	assert.Equal(t, ndaRequest.TrustCenterNdaRequests.Edges[0].Node.Status.String(), enums.TrustCenterNDARequestStatusSigned.String())
	assert.Check(t, ndaRequest.TrustCenterNdaRequests.Edges[0].Node.SignedAt != nil)

	// wait for the NDA attestation listener to process the document data creation
	waitForEvents()

	// verify the signed NDA email was sent with the attested PDF attached
	msgs := mockEmailSender().Messages()
	assert.Assert(t, len(msgs) == 2, "expected access and signed NDA emails, got %d", len(msgs))
	accessEmails := lo.Filter(msgs, func(msg *newman.EmailMessage, _ int) bool {
		return lo.Contains(msg.To, email) && strings.HasPrefix(msg.Subject, "Access ")
	})
	assert.Assert(t, len(accessEmails) == 1, "expected one access email")
	signedEmails := lo.Filter(msgs, func(msg *newman.EmailMessage, _ int) bool {
		return lo.Contains(msg.To, email) &&
			len(msg.Attachments) == 1 &&
			msg.Attachments[0].Filename == signedNDAAttachmentName
	})
	assert.Assert(t, len(signedEmails) == 1, "expected one email with the signed PDF attachment")
	assert.Assert(t, len(signedEmails[0].Attachments[0].Content) > 0, "expected non-empty PDF content in attachment")

	// now, check that the anonymous user can query the protected doc's files
	getTrustCenterDocResp, err = suite.Client.API.GetTrustCenterDocByID(anonCtx, trustCenterDocProtected.ID)
	assert.NilError(t, err)
	assert.Assert(t, getTrustCenterDocResp.TrustCenterDoc.OriginalFile != nil)

	th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
}
