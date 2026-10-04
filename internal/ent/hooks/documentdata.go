package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/template"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects"
)

// HookDocumentDataTrustCenterNDA runs on document data create mutations to ensure trust center NDA document submissions are valid
func HookDocumentDataTrustCenterNDA() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DocumentDataFunc(func(ctx context.Context, m *generated.DocumentDataMutation) (generated.Value, error) {
			templateID, _ := m.TemplateID()
			if templateID == "" {
				// verify anonymous questionnaire user context
				// assessments do not require a template id to be there
				// because not all assessments are tied to a template,
				// some are created from scratch
				if _, ok := auth.ActiveAssessmentIDKey.Get(ctx); ok {
					return next.Mutate(ctx, m)
				}

				return nil, errMissingTemplate
			}

			docTemplate, err := m.Client().Template.Query().Where(template.ID(templateID)).Only(ctx)
			if err != nil {
				return nil, err
			}

			if docTemplate.Kind != enums.TemplateKindTrustCenterNda {
				return next.Mutate(ctx, m)
			}

			caller, ok := auth.GetVerifiedTrustCenterUserCaller(ctx, docTemplate.TrustCenterID)
			if !ok {
				return nil, errMustBeAnonymousUser
			}

			// the above verified the trust center ID matched so we can continue with this id
			verifiedTCID := docTemplate.TrustCenterID

			// add the validated fields to the log context now
			ctx = logx.WithFields(ctx, map[string]any{"trust_center_id": verifiedTCID, "email": caller.SubjectEmail})

			response, ok := m.Data()
			if !ok {
				return nil, errMissingResponse
			}

			signedID, err := m.Client().TrustCenterNDARequest.Query().Where(
				trustcenterndarequest.EmailEqualFold(caller.SubjectEmail),
				trustcenterndarequest.TrustCenterID(verifiedTCID),
				trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusSigned),
			).FirstID(ctx)
			if err == nil && signedID != "" {
				return nil, errUserHasAlreadySignedNDA
			}

			f, err := fetchNDATemplateFile(ctx, m.Client(), templateID)
			if err != nil {
				return nil, err
			}

			if err = validateTrustCenterNDAJSON(ctx, response, verifiedTCID, caller.SubjectEmail, caller.SubjectID, f); err != nil {
				return nil, err
			}

			response["trust_center_id"] = verifiedTCID
			response["pdf_file_id"] = f.ID

			if metadata, ok := response["signature_metadata"].(map[string]any); ok {
				metadata["pdf_hash"] = f.Md5Hash
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			createdDocData, ok := v.(*generated.DocumentData)
			if !ok {
				logx.FromContext(ctx).Error().Msgf("unexpected type %T for created document data", v)
				return nil, fmt.Errorf("unexpected type %T: %w", v, ErrInternalServerError)
			}

			if err := m.Client().TrustCenterNDARequest.Update().Where(
				trustcenterndarequest.EmailEqualFold(caller.SubjectEmail),
				trustcenterndarequest.TrustCenterID(verifiedTCID),
				trustcenterndarequest.StatusNEQ(enums.TrustCenterNDARequestStatusSigned),
			).SetStatus(enums.TrustCenterNDARequestStatusSigned).SetDocumentDataID(createdDocData.ID).Exec(ctx); err != nil {
				if !generated.IsNotFound(err) {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to mark nda request signed status")
					return nil, err
				}

				logx.FromContext(ctx).Error().Msg("no existing nda request to mark signed status")
			}

			return v, nil
		})
	}, ent.OpCreate)
}

// HookDocumentDataFile handles file uploads and attaches them to document data.
// restricted to system admins updating NDA documents only for now in riverqueue.
// the old/regular case of adding FileIDs to mutations will still be accepted for non admins.
func HookDocumentDataFile() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DocumentDataFunc(func(ctx context.Context, m *generated.DocumentDataMutation) (generated.Value, error) {
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) == 0 {
				return next.Mutate(ctx, m)
			}

			if len(fileIDs) > 1 {
				return nil, errOnlyOneDocumentData
			}

			if !auth.HasInLineageContextCaller(ctx, auth.CapSystemAdmin) {
				return nil, generated.ErrPermissionDenied
			}

			id, err := m.OldTemplateID(ctx)
			if err != nil || id == "" {
				return nil, errMissingTemplate
			}

			exists, err := m.Client().Template.Query().
				Where(template.KindEQ(enums.TemplateKindTrustCenterNda)).
				Where(template.ID(id)).
				Exist(ctx)
			if err != nil {
				return nil, err
			}

			if !exists {
				return nil, generated.ErrPermissionDenied
			}

			ctx, err = objects.ProcessFilesForMutation(ctx, m, "documentDataFile")
			if err != nil {
				return nil, err
			}

			m.AddFileIDs(fileIDs...)

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne)
}

// validateTrustCenterNDAJSON validates the document against the struct-derived schema
// and checks the trust center id, email, user id, PDF file attached to the response
func validateTrustCenterNDAJSON(ctx context.Context, document map[string]any, trustCenterID, subjectEmail, subjectID string, templateFile *generated.File) error {
	schema := jsonx.SchemaFrom[signedNDADocumentData]()

	result, err := jsonx.ValidateSchema(schema, document)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).
			Msg("failed to validate trust center nda json schema")

		return err
	}

	if !result.Valid() {
		errs := jsonx.ValidationErrorStrings(result)
		logx.FromContext(ctx).Error().
			Strs("validation_errors", errs).
			Msg("trust center nda json failed schema validation")

		return fmt.Errorf("%w: %v", errValidationFailed, errs)
	}

	var doc signedNDADocumentData
	if err := jsonx.RoundTrip(document, &doc); err != nil {
		logx.FromContext(ctx).Error().Err(err).
			Msg("failed to decode trust center nda json")

		return fmt.Errorf("%w: %v", errValidationFailed, err)
	}

	if doc.TrustCenterID != trustCenterID ||
		doc.SignatoryInfo.Email != subjectEmail ||
		doc.SignatureMetadata.UserID != subjectID {

		logx.FromContext(ctx).Error().
			Str("expected_trust_center_id", trustCenterID).
			Str("document_trust_center_id", doc.TrustCenterID).
			Str("expected_subject_id", subjectID).
			Str("document_subject_id", doc.SignatureMetadata.UserID).
			Msg("trust center nda json does not match caller")

		return errDocInfoDoesNotMatchCaller
	}

	if templateFile == nil || templateFile.ID == "" {
		logx.FromContext(ctx).Error().
			Msg("trust center nda template file is missing")

		return ErrMissingNDATemplateFile
	}

	if doc.PDFFileID != templateFile.ID {
		logx.FromContext(ctx).Error().
			Str("document_pdf_file_id", doc.PDFFileID).
			Str("template_file_id", templateFile.ID).
			Msg("trust center nda pdf file does not match template")

		return errNDAPDFFileDoesNotMatchTemplate
	}

	if templateFile.Md5Hash == "" {
		logx.FromContext(ctx).Error().
			Str("template_file_id", templateFile.ID).
			Msg("trust center nda template file is missing md5 hash")

		return errNDATemplateFileMissingHash
	}

	if !strings.EqualFold(doc.SignatureMetadata.PDFHash, templateFile.Md5Hash) {
		logx.FromContext(ctx).Error().
			Str("template_file_id", templateFile.ID).
			Str("document_pdf_hash", doc.SignatureMetadata.PDFHash).
			Str("template_pdf_hash", templateFile.Md5Hash).
			Msg("trust center nda pdf hash does not match template")

		return errNDAPDFHashDoesNotMatchTemplate
	}

	return nil
}
