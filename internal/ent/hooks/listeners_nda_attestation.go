package hooks

import (
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated/documentdata"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// NDAAttestationListeners attests signed trust center NDAs after document data creation
func NDAAttestationListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaDocumentData,
			Operations: []string{entityops.OpCreate},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapBypassOrgFilter)
			},
			Handle: handleNDAAttestationCreated,
		},
	}
}

// handleNDAAttestationCreated attests a signed trust center NDA, records the signed file on
// the NDA request, and emails the signer a copy
func handleNDAAttestationCreated(inv entityops.Invocation, payload entityops.MutationPayload) error {
	templateID, _ := payload.StringValue(documentdata.FieldTemplateID)
	if templateID == "" {
		logx.FromContext(inv.Context).Error().Msg("nda attestation listener: no template")
		return nil
	}

	docTemplate, ok, err := entityops.LoadEntity(inv.Context, templateID, inv.Client.Template.Get)
	if err != nil || !ok {
		return err
	}

	if docTemplate.Kind != enums.TemplateKindTrustCenterNda {
		return nil
	}

	if inv.Caller.SubjectEmail == "" {
		logx.FromContext(inv.Context).Error().Msg("nda attestation listener: caller not available in restored context")

		return nil
	}

	docData, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.DocumentData.Get)
	if err != nil || !ok {
		return err
	}

	var ndaMetadata signedNDADocumentData
	if err := jsonx.RoundTrip(docData.Data, &ndaMetadata); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("nda attestation listener: failed to unmarshal nda metadata from document data")

		return nil
	}

	tcID := ndaMetadata.TrustCenterID
	if tcID == "" {
		logx.FromContext(inv.Context).Error().Msg("nda attestation listener: trust center id not found in document data")

		return nil
	}

	logCtx := logx.WithFields(inv.Context, map[string]any{"email": inv.Caller.SubjectEmail, "trust_center_id": tcID})

	result, err := attestNDADocument(inv.Context, inv.Client, docData, templateID, tcID)
	if err != nil {
		logx.FromContext(logCtx).Error().Err(err).Msg("nda attestation listener: failed to attest NDA document")

		return err
	}

	allowCtx := privacy.DecisionContext(inv.Context, privacy.Allow)

	requestID, err := inv.Client.TrustCenterNDARequest.Query().Where(
		trustcenterndarequest.EmailEqualFold(inv.Caller.SubjectEmail),
		trustcenterndarequest.TrustCenterID(tcID),
		trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusSigned),
	).FirstID(allowCtx)
	if err != nil {
		logx.FromContext(logCtx).Error().Err(err).Msg("nda attestation listener: failed to resolve nda request id for email")

		return err
	}

	if err := inv.Client.TrustCenterNDARequest.UpdateOneID(requestID).
		SetFileID(result.TemplateFileID).
		Exec(allowCtx); err != nil {
		logx.FromContext(logCtx).Error().Err(err).Msg("nda attestation listener: failed to set file ID on nda request")

		return err
	}

	if err := sendSystemEmail(inv.Context, inv.Client, emaildef.TCNDASignedOp.Name(), emaildef.TrustCenterNDASignedEmail{
		RecipientInfo:      emaildef.RecipientInfo{Email: inv.Caller.SubjectEmail},
		OrgName:            result.OrgName,
		RequestID:          requestID,
		TrustCenterID:      tcID,
		AttachmentFilename: "signed_nda_file.pdf",
		AttachmentData:     result.AttestedPDF,
	}); err != nil {
		logx.FromContext(logCtx).Error().Err(err).Msg("nda attestation listener: failed to send NDA signed email")

		return err
	}

	return nil
}
