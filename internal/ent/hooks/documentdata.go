package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

var (
	errMissingTemplate                      = errors.New("missing template")
	errDocInfoDoesNotMatchAuthenticatedUser = errors.New("NDA submission does not match authenticated user")
	errUserHasAlreadySignedNDA              = errors.New("user has already signed the NDA")
	errValidationFailed                     = errors.New("validation failed")
	errMustBeAnonymousUser                  = errors.New("must be an anonymous user")
	errMissingResponse                      = errors.New("missing response")
	errOnlyOneDocumentData                  = errors.New("you can only upload one document data file for an nda")
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
				_, ok := auth.AnonymousQuestionnaireUserFromContext(ctx)
				if ok {
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

			anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx)
			if !ok || anon.SubjectEmail == "" || anon.TrustCenterID == "" || anon.OrganizationID == "" {
				return nil, errMustBeAnonymousUser
			}

			response, ok := m.Data()
			if !ok {
				return nil, errMissingResponse
			}

			signedID, err := m.Client().TrustCenterNDARequest.Query().Where(
				trustcenterndarequest.EmailEqualFold(anon.SubjectEmail),
				trustcenterndarequest.TrustCenterID(anon.TrustCenterID),
				trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusSigned),
			).FirstID(ctx)
			if err == nil && signedID != "" {
				return nil, errUserHasAlreadySignedNDA
			}

			if err = validateTrustCenterNDAJSON(docTemplate.Jsonconfig, response, anon); err != nil {
				return nil, err
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the id of the created document data
			createdDocData, ok := v.(*generated.DocumentData)
			if !ok {
				logx.FromContext(ctx).Error().Msgf("unexpected type %T for created document data", v)

				return nil, fmt.Errorf("unexpected type %T: %w", v, ErrInternalServerError)
			}

			// update nda requests that it has been signed
			if err := m.Client().TrustCenterNDARequest.Update().Where(
				trustcenterndarequest.EmailEqualFold(anon.SubjectEmail),
				trustcenterndarequest.TrustCenterID(anon.TrustCenterID),
				trustcenterndarequest.StatusNEQ(enums.TrustCenterNDARequestStatusSigned),
			).SetStatus(enums.TrustCenterNDARequestStatusSigned).SetDocumentDataID(createdDocData.ID).Exec(ctx); err != nil {
				if !generated.IsNotFound(err) {
					logx.FromContext(ctx).Error().Err(err).Str("email", anon.SubjectEmail).Str("trust_center_id", anon.TrustCenterID).Msg("failed to mark nda request signed status")

					return nil, err
				}

				// this shouldn't happen, unless it was already marked as signed
				logx.FromContext(ctx).Error().Str("email", anon.SubjectEmail).Str("trust_center_id", anon.TrustCenterID).Msg("no existing nda request to mark signed status")
			}

			// add the nda_signed tuple to the anonymous user to allow file access
			tuple := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   anon.SubjectID,
				SubjectType: "user",
				ObjectID:    anon.TrustCenterID,
				ObjectType:  "trust_center",
				Relation:    "nda_signed",
			})

			if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{tuple}, nil); err != nil {
				return nil, err
			}

			ndaRequestID, err := m.Client().TrustCenterNDARequest.Query().Where(
				trustcenterndarequest.EmailEqualFold(anon.SubjectEmail),
				trustcenterndarequest.TrustCenterID(anon.TrustCenterID),
				trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusSigned),
			).FirstID(ctx)
			if err != nil {
				return nil, err
			}

			if err := enqueueJob(ctx, m.Job, jobspec.AttestNDARequestArgs{
				NDARequestID: ndaRequestID,
			}, nil); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("nda_request_id", ndaRequestID).
					Msg("failed to enqueue attest nda request job")

				return nil, err
			}

			return v, err
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

			if !auth.IsSystemAdminFromContext(ctx) {
				return nil, generated.ErrPermissionDenied
			}

			id, err := m.OldTemplateID(ctx)
			if err != nil || id == "" {
				return nil, errMissingTemplate
			}

			exists, err := m.Client().Template.Query().
				Select(template.FieldKind).
				Where(template.KindEQ(enums.TemplateKindTrustCenterNda)).
				Where(template.ID(id)).
				Exist(ctx)
			if err != nil {
				return nil, err
			}

			if !exists {
				return nil, generated.ErrPermissionDenied
			}

			adapter := objects.NewGenericMutationAdapter(m,
				func(mut *generated.DocumentDataMutation) (string, bool) { return mut.ID() },
				func(mut *generated.DocumentDataMutation) string { return mut.Type() },
			)

			ctx, err = objects.ProcessFilesForMutation(ctx, adapter, "documentDataFile")
			if err != nil {
				return nil, err
			}

			m.AddFileIDs(fileIDs...)

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne)
}

// validateTrustCenterNDAJSON validates the JSON against the schema and checks the trust center id, email, and user id match the authenticated user
func validateTrustCenterNDAJSON(schema interface{}, document map[string]interface{}, anon *auth.AnonymousTrustCenterUser) (err error) {
	if err = validateJSON(schema, document); err != nil {
		return err
	}

	signatoryInfo := document["signatory_info"].(map[string]any)

	if document["trust_center_id"] != anon.TrustCenterID ||
		signatoryInfo["email"] != anon.SubjectEmail ||
		document["signature_metadata"].(map[string]any)["user_id"] != anon.SubjectID {
		return errDocInfoDoesNotMatchAuthenticatedUser
	}

	firstName, _ := signatoryInfo["first_name"].(string)
	lastName, _ := signatoryInfo["last_name"].(string)
	companyName, _ := signatoryInfo["company_name"].(string)

	if firstName == "" || lastName == "" || companyName == "" {
		return fmt.Errorf("%w: first_name, last_name, and company_name are all required", errValidationFailed)
	}

	return nil
}

// validateJSON validates a document against a schema
func validateJSON(schema interface{}, document interface{}) error {
	// Convert to JSON first
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return err
	}

	// Create loaders
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	documentLoader := gojsonschema.NewBytesLoader(documentBytes)

	// Perform validation
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var errors []string
		for _, err := range result.Errors() {
			errors = append(errors, err.String())
		}

		return fmt.Errorf("%w: %v", errValidationFailed, errors)
	}

	return nil
}
