package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/documentdata"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/xeipuuv/gojsonschema"
)

var (
	errMissingTemplate                      = errors.New("missing template")
	errDocInfoDoesNotMatchAuthenticatedUser = errors.New("NDA submission does not match authenticated user")
	errUserHasAlreadySignedNDA              = errors.New("user has already signed the NDA")
	errValidationFailed                     = errors.New("validation failed")
	errMustBeAnonymousUser                  = errors.New("must be an anonymous user")
	errMissingResponse                      = errors.New("missing response")
)

// HookDocumentDataTrustCenterNDA runs on document data create mutations to ensure trust center NDA document submissions are valid
func HookDocumentDataTrustCenterNDA() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DocumentDataFunc(func(ctx context.Context, m *generated.DocumentDataMutation) (generated.Value, error) {
			templateID, _ := m.TemplateID()
			if templateID == "" {
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

			previousDocs, err := m.Client().DocumentData.Query().Where(
				documentdata.And(
					documentdata.TemplateIDEQ(docTemplate.ID),
					func(s *sql.Selector) {
						s.Where(
							sqljson.ValueEQ(documentdata.FieldData, anon.SubjectEmail, sqljson.DotPath("signatory_info.email")),
						)
					},
				),
			).Count(ctx)
			if err != nil {
				return nil, err
			}

			if previousDocs > 0 {
				return nil, errUserHasAlreadySignedNDA
			}

			if err = validateTrustCenterNDAJSON(docTemplate.Jsonconfig, response, anon); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// validateTrustCenterNDAJSON validates the JSON against the schema and checks the trust center id, email, and user id match the authenticated user
func validateTrustCenterNDAJSON(schema interface{}, document map[string]interface{}, anon *auth.AnonymousTrustCenterUser) (err error) {
	if err = validateJSON(schema, document); err != nil {
		return err
	}

	if document["trust_center_id"] != anon.TrustCenterID ||
		document["signatory_info"].(map[string]any)["email"] != anon.SubjectEmail ||
		document["signature_metadata"].(map[string]any)["user_id"] != anon.SubjectID {
		return errDocInfoDoesNotMatchAuthenticatedUser
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
