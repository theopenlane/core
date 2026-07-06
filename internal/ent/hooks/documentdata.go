package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
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

			tcID, hasTCID := auth.ActiveTrustCenterIDKey.Get(ctx)
			caller, hasCaller := auth.CallerFromContext(ctx)
			if !hasTCID || tcID == "" || !hasCaller || caller == nil || caller.SubjectEmail == "" || caller.OrganizationID == "" {
				return nil, errMustBeAnonymousUser
			}

			if docTemplate.TrustCenterID != tcID {
				return nil, errNDATemplateDoesNotMatchTrustCenter
			}

			response, ok := m.Data()
			if !ok {
				return nil, errMissingResponse
			}

			signedID, err := m.Client().TrustCenterNDARequest.Query().Where(
				trustcenterndarequest.EmailEqualFold(caller.SubjectEmail),
				trustcenterndarequest.TrustCenterID(tcID),
				trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusSigned),
			).FirstID(ctx)
			if err == nil && signedID != "" {
				return nil, errUserHasAlreadySignedNDA
			}

			f, err := fetchNDATemplateFile(ctx, m.Client(), templateID)
			if err != nil {
				return nil, err
			}

			if err = validateTrustCenterNDAJSON(response, tcID, caller.SubjectEmail, caller.SubjectID, f); err != nil {
				return nil, err
			}

			response["trust_center_id"] = tcID
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
				trustcenterndarequest.TrustCenterID(tcID),
				trustcenterndarequest.StatusNEQ(enums.TrustCenterNDARequestStatusSigned),
			).SetStatus(enums.TrustCenterNDARequestStatusSigned).SetDocumentDataID(createdDocData.ID).Exec(ctx); err != nil {
				if !generated.IsNotFound(err) {
					logx.FromContext(ctx).Error().Err(err).Str("email", caller.SubjectEmail).Str("trust_center_id", tcID).Msg("failed to mark nda request signed status")
					return nil, err
				}

				logx.FromContext(ctx).Error().Str("email", caller.SubjectEmail).Str("trust_center_id", tcID).Msg("no existing nda request to mark signed status")
			}

			tuple := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   caller.SubjectID,
				SubjectType: "user",
				ObjectID:    tcID,
				ObjectType:  "trust_center",
				Relation:    "nda_signed",
			})

			if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{tuple}, nil); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to create nda_signed relationship tuple")
				return nil, ErrInternalServerError
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

			caller, callerOK := auth.CallerFromContext(ctx)
			isSystemAdmin := callerOK && caller != nil && caller.HasInLineage(auth.CapSystemAdmin)

			if !isSystemAdmin {
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
func validateTrustCenterNDAJSON(document map[string]any, trustCenterID, subjectEmail, subjectID string, templateFile *generated.File) error {
	schema := jsonx.SchemaFrom[signedNDADocumentData]()

	result, err := jsonx.ValidateSchema(schema, document)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return fmt.Errorf("%w: %v", errValidationFailed, jsonx.ValidationErrorStrings(result))
	}

	var doc signedNDADocumentData
	if err := jsonx.RoundTrip(document, &doc); err != nil {
		return fmt.Errorf("%w: %v", errValidationFailed, err)
	}

	if doc.TrustCenterID != trustCenterID ||
		doc.SignatoryInfo.Email != subjectEmail ||
		doc.SignatureMetadata.UserID != subjectID {

		return errDocInfoDoesNotMatchCaller
	}

	if templateFile == nil || templateFile.ID == "" {
		return ErrMissingNDATemplateFile
	}

	if doc.PDFFileID != templateFile.ID {
		return errNDAPDFFileDoesNotMatchTemplate
	}

	if templateFile.Md5Hash == "" {
		return errNDATemplateFileMissingHash
	}

	if !strings.EqualFold(doc.SignatureMetadata.PDFHash, templateFile.Md5Hash) {
		return errNDAPDFHashDoesNotMatchTemplate
	}

	return nil
}
