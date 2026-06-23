package hooks

import (
	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaNDAAttestationListeners registers listeners that process NDA attestation
// asynchronously after document data creation
func RegisterGalaNDAAttestationListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeDocumentData),
			Name:       "nda.attestation",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleNDAAttestationCreated,
		},
	)
}

func handleNDAAttestationCreated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		logx.FromContext(ctx.Context).Error().Msg("nda attestation listener: no client in context")
		return nil
	}

	docDataID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || docDataID == "" {
		logx.FromContext(ctx.Context).Error().Msg("nda attestation listener: no document data id")
		return nil
	}

	templateID, _ := eventqueue.MutationStringValue(payload, "template_id")
	if templateID == "" {
		logx.FromContext(ctx.Context).Error().Msg("nda attestation listener: no template")
		return nil
	}

	docTemplate, err := client.Template.Query().Where(template.ID(templateID)).Only(ctx.Context)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("nda attestation listener: cannot get template")
		return nil
	}

	if docTemplate.Kind != enums.TemplateKindTrustCenterNda {
		return nil
	}

	caller, hasCaller := auth.CallerFromContext(ctx.Context)
	if !hasCaller || caller == nil || caller.SubjectEmail == "" {
		logx.FromContext(ctx.Context).Error().Msg("nda attestation listener: caller not available in restored context")

		return nil
	}

	// bypass the org context filter when calling from inside the listener
	allowCtx := auth.WithCaller(ctx.Context, caller.WithCapabilities(auth.CapBypassOrgFilter))

	docData, err := client.DocumentData.Get(allowCtx, docDataID)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("document_data_id", docDataID).Msg("nda attestation listener: failed to get document data for nda attestation")

		return nil
	}

	var ndaMetadata signedNDADocumentData
	if err := jsonx.RoundTrip(docData.Data, &ndaMetadata); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("nda attestation listener: failed to unmarshal nda metadata from document data")

		return nil
	}

	tcID := ndaMetadata.TrustCenterID
	if tcID == "" {
		logx.FromContext(ctx.Context).Error().Msg("nda attestation listener: nda attestation listener: trust center id not found in document data")

		return nil
	}

	result, err := attestNDADocument(allowCtx, client, docData, templateID, tcID)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("nda attestation listener: failed to attest NDA document")

		return err
	}

	allowCtx = privacy.DecisionContext(allowCtx, privacy.Allow)
	if err := client.TrustCenterNDARequest.Update().Where(
		trustcenterndarequest.EmailEqualFold(caller.SubjectEmail),
		trustcenterndarequest.TrustCenterID(tcID),
	).SetFileID(result.TemplateFileID).Exec(allowCtx); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("email", caller.SubjectEmail).Str("trust_center_id", tcID).Msg("nda attestation listener: failed to set file ID on nda request")

		return err
	}

	requestID, err := client.TrustCenterNDARequest.Query().Where(
		trustcenterndarequest.EmailEqualFold(caller.SubjectEmail),
		trustcenterndarequest.TrustCenterID(tcID),
		trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusSigned),
	).FirstID(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Str("email", caller.SubjectEmail).Str("trust_center_id", tcID).Msg("nda attestation listener: failed to resolve nda request id for email")

		return err
	}

	if err := sendSystemEmail(ctx.Context, client, emaildef.TCNDASignedOp.Name(), emaildef.TrustCenterNDASignedEmail{
		RecipientInfo:      emaildef.RecipientInfo{Email: caller.SubjectEmail},
		OrgName:            result.OrgName,
		RequestID:          requestID,
		TrustCenterID:      tcID,
		AttachmentFilename: "signed_nda_file.pdf",
		AttachmentData:     result.AttestedPDF,
	}); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("nda attestation listener: failed to send NDA signed email")

		return err
	}

	return nil
}
