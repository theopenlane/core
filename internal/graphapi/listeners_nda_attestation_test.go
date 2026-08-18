//go:build test

package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/newman"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

const signedNDAAttachmentName = "signed_nda_file.pdf"

func TestNDAAttestationListener(t *testing.T) {
	t.Run("signed nda stamps file on the signed request and emails the signer", func(t *testing.T) {
		tcOrg := createFreshOrgWithTrustCenter(t, withNDATemplate())
		trustCenter := tcOrg.trustCenter
		allowCtx := setContext(tcOrg.owner.UserCtx, suite.client.db)

		signerEmail := "nda-signer@listenerpin.io"
		bystanderEmail := "nda-bystander@listenerpin.io"

		signerCtx, signerCaller := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, signerEmail)
		bystanderCtx, _ := createAnonymousTrustCenterContextWithEmail(trustCenter.ID, trustCenter.OwnerID, bystanderEmail)

		_, err := suite.client.api.CreateTrustCenterNDARequest(signerCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     "Signer",
			LastName:      "User",
			CompanyName:   lo.ToPtr("Signer Co"),
			Email:         signerEmail,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)

		_, err = suite.client.api.CreateTrustCenterNDARequest(bystanderCtx, testclient.CreateTrustCenterNDARequestInput{
			FirstName:     "Bystander",
			LastName:      "User",
			CompanyName:   lo.ToPtr("Bystander Co"),
			Email:         bystanderEmail,
			TrustCenterID: &trustCenter.ID,
		})
		assert.NilError(t, err)

		suite.WaitForEvents()
		suite.mockEmailSender().Reset()
		expectAttestedUpload(t, suite.client.mockProvider)

		resp, err := suite.client.api.SubmitTrustCenterNDAResponse(signerCtx, testclient.SubmitTrustCenterNDAResponseInput{
			TemplateID: *tcOrg.ndaTemplateID,
			Response: map[string]any{
				"signatory_info": map[string]any{
					"email": signerEmail,
				},
				"acknowledgment": true,
				"signature_metadata": map[string]any{
					"ip_address": "192.168.1.100",
					"timestamp":  "2025-09-22T19:37:59.988Z",
					"pdf_hash":   getMD5Hash(t, pdfFilePath),
					"user_id":    signerCaller.SubjectID,
				},
				"pdf_file_id":     *tcOrg.ndaFileID,
				"trust_center_id": trustCenter.ID,
			},
		})
		assert.NilError(t, err)

		docDataID := resp.SubmitTrustCenterNDAResponse.DocumentData.ID

		suite.WaitForEvents()

		signed, err := suite.client.db.TrustCenterNDARequest.Query().Where(
			trustcenterndarequest.EmailEqualFold(signerEmail),
			trustcenterndarequest.TrustCenterID(trustCenter.ID),
		).Only(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.TrustCenterNDARequestStatusSigned, signed.Status))
		assert.Assert(t, signed.FileID != nil)
		assert.Check(t, is.Equal(*tcOrg.ndaFileID, *signed.FileID))

		bystander, err := suite.client.db.TrustCenterNDARequest.Query().Where(
			trustcenterndarequest.EmailEqualFold(bystanderEmail),
			trustcenterndarequest.TrustCenterID(trustCenter.ID),
		).Only(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.TrustCenterNDARequestStatusRequested, bystander.Status))
		assert.Check(t, bystander.FileID == nil)

		docData, err := suite.client.db.DocumentData.Get(allowCtx, docDataID)
		assert.NilError(t, err)

		attestedHash, _ := docData.Data["attested_pdf_hash"].(string)
		assert.Check(t, attestedHash != "")

		fileCount, err := suite.client.db.DocumentData.QueryFiles(docData).Count(allowCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, fileCount))

		signedEmails := lo.Filter(suite.mockEmailSender().Messages(), func(msg *newman.EmailMessage, _ int) bool {
			return lo.Contains(msg.To, signerEmail) &&
				len(msg.Attachments) == 1 &&
				msg.Attachments[0].Filename == signedNDAAttachmentName
		})
		assert.Assert(t, is.Len(signedEmails, 1))
		assert.Check(t, len(signedEmails[0].Attachments[0].Content) > 0)

		cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
	})

	t.Run("document data for a non-nda template is skipped", func(t *testing.T) {
		docUser := suite.userBuilder(context.Background(), t)
		ctx := setContext(docUser.UserCtx, suite.client.db)

		tmpl := (&TemplateBuilder{client: suite.client}).MustNew(docUser.UserCtx, t)

		docData, err := suite.client.db.DocumentData.Create().
			SetTemplateID(tmpl.ID).
			SetOwnerID(docUser.OrganizationID).
			SetData(map[string]any{"question": "answer"}).
			Save(ctx)
		assert.NilError(t, err)

		suite.WaitForEvents()

		reloaded, err := suite.client.db.DocumentData.Get(ctx, docData.ID)
		assert.NilError(t, err)

		_, hasAttestedHash := reloaded.Data["attested_pdf_hash"]
		assert.Check(t, !hasAttestedHash)

		fileCount, err := suite.client.db.DocumentData.QueryFiles(reloaded).Count(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, fileCount))

		cleanupOrganizationDataWithContext(docUser.UserCtx, t)
	})
}
