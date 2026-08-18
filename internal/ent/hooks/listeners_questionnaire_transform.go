package hooks

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// QuestionnaireTransformListeners transforms completed questionnaire document data into
// entities using the template's transform configuration
func QuestionnaireTransformListeners() []gala.Registration {
	return []gala.Registration{entityops.MutationListener{
		Schema:     entityops.SchemaAssessmentResponse,
		Operations: []string{entityops.OpCreate, entityops.OpUpdate, entityops.OpUpdateOne},
		Fields: []string{
			assessmentresponse.FieldStatus,
			assessmentresponse.FieldDocumentDataID,
			assessmentresponse.FieldCompletedAt,
			assessmentresponse.FieldIsDraft,
		},
		Caller: internalOperationBypassCaller,
		Handle: handleAssessmentResponse,
	}}
}

// handleAssessmentResponse reacts to completed assessment responses and transforms the associated
// document data
func handleAssessmentResponse(inv entityops.Invocation, _ entityops.MutationPayload) error {
	response, err := inv.Client.AssessmentResponse.Query().
		Where(assessmentresponse.IDEQ(inv.EntityID)).
		WithDocument().
		WithAssessment(func(query *entgen.AssessmentQuery) {
			query.WithTemplate()
		}).
		Only(inv.Context)
	if err != nil {
		if entgen.IsNotFound(err) {
			logx.FromContext(inv.Context).Error().Err(err).Msg("assessment response not found for questionnaire transform")

			return nil
		}

		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to load assessment response for questionnaire transform")

		return err
	}

	assessment, document, config, ok := validateQuestionnaire(response)
	if !ok {
		return nil
	}

	req := questionnaireTransformRequest{
		OrganizationID:       lo.CoalesceOrEmpty(response.OwnerID, document.OwnerID, assessment.OwnerID),
		TemplateID:           assessment.TemplateID,
		TemplateKind:         assessment.Edges.Template.Kind,
		AssessmentID:         assessment.ID,
		AssessmentResponseID: response.ID,
		DocumentDataID:       response.DocumentDataID,
		Email:                response.Email,
		Data:                 document.Data,
		Config:               config,
	}

	ctx := logx.WithFields(inv.Context, map[string]any{
		"assessment_id":    assessment.ID,
		"template_id":      assessment.TemplateID,
		"document_data_id": response.DocumentDataID,
	})

	// already transformed; rerun the idempotent linking so partial deliveries heal
	if response.EntityID != "" {
		return linkEntitySources(ctx, inv.Client, req, response.EntityID, mappedNotes(req))
	}

	if err := transformQuestionnaire(ctx, inv.Client, req); err != nil {
		if errors.Is(err, ErrQuestionnaireTransformInvalid) {
			logx.FromContext(ctx).Error().Err(err).Msg("questionnaire transform failed")

			return nil
		}

		return err
	}

	return nil
}

// validateQuestionnaire returns the assessment, document, and transform config if the response is complete and the template has a valid transform configuration
func validateQuestionnaire(response *entgen.AssessmentResponse) (*entgen.Assessment, *entgen.DocumentData, models.TemplateProjectionConfig, bool) {
	if response == nil || response.Status != enums.AssessmentResponseStatusCompleted || response.DocumentDataID == "" {
		return nil, nil, models.TemplateProjectionConfig{}, false
	}

	assessment := response.Edges.Assessment
	if assessment == nil || assessment.Edges.Template == nil {
		return nil, nil, models.TemplateProjectionConfig{}, false
	}

	config := assessment.Edges.Template.TransformConfiguration
	if !config.Enabled {
		return nil, nil, models.TemplateProjectionConfig{}, false
	}

	document := response.Edges.Document
	if document == nil {
		return nil, nil, models.TemplateProjectionConfig{}, false
	}

	return assessment, document, config, true
}

const transformMetadataKey = "questionnaire_transform"
const entityTransformFieldNotes = "notes"

type questionnaireTransformRequest struct {
	OrganizationID       string                          `json:"-"`
	TemplateID           string                          `json:"template_id"`
	TemplateKind         enums.TemplateKind              `json:"-"`
	AssessmentID         string                          `json:"assessment_id"`
	AssessmentResponseID string                          `json:"assessment_response_id"`
	DocumentDataID       string                          `json:"document_data_id"`
	Email                string                          `json:"-"`
	Data                 map[string]any                  `json:"-"`
	Config               models.TemplateProjectionConfig `json:"-"`
}

// transformQuestionnaire maps the submitted document data into an entity, upserts it, and
// links the response, document, and note to the persisted record
func transformQuestionnaire(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest) error {
	if req.OrganizationID == "" {
		return fmt.Errorf("%w: missing organization id", ErrQuestionnaireTransformInvalid)
	}

	input, notes, err := mapEntityInput(ctx, client, req)
	if err != nil {
		return err
	}

	record, err := upsertEntity(ctx, client, input)
	if err != nil {
		return err
	}

	claimed, err := claimAssessmentResponse(ctx, client, req, record)
	if err != nil {
		return err
	}

	// a lost claim means a concurrent delivery of the same response already owns linking
	// and note creation
	if !claimed {
		return nil
	}

	return linkEntitySources(ctx, client, req, record.ID, notes)
}

// mapEntityInput builds the entity create input and note text from the template's mappings
func mapEntityInput(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest) (entgen.CreateEntityInput, string, error) {
	if len(req.Config.Mappings) == 0 {
		return entgen.CreateEntityInput{}, "", fmt.Errorf("%w: configuration has no mappings", ErrQuestionnaireTransformInvalid)
	}

	fields := map[string]any{}

	var notes, ownerValue, environmentValue string

	for _, mapping := range req.Config.Mappings {
		rawValue, ok := mapx.ValueAtPath(req.Data, mapping.From)
		if !ok || isEmptyValue(rawValue) {
			if mapping.Resolver == models.TemplateProjectionResolverInternalOwner && req.Email != "" {
				rawValue = req.Email
			} else if mapping.Required {
				return entgen.CreateEntityInput{}, "", fmt.Errorf("%w: missing required field %q", ErrQuestionnaireTransformInvalid, mapping.From)
			} else {
				continue
			}
		}

		switch {
		case mapping.Resolver == models.TemplateProjectionResolverInternalOwner:
			ownerValue = getStringValue(rawValue)
		case mapping.Resolver == models.TemplateProjectionResolverEnvironment:
			environmentValue = getStringValue(rawValue)
		case mapping.To == entityTransformFieldNotes:
			notes = getStringValue(rawValue)
		case mapping.To == "":
			return entgen.CreateEntityInput{}, "", fmt.Errorf("%w: mapping for %q is missing target field", ErrQuestionnaireTransformInvalid, mapping.From)
		default:
			fields[mapping.To] = rawValue
		}
	}

	var input entgen.CreateEntityInput
	if err := jsonx.RoundTrip(fields, &input); err != nil {
		return entgen.CreateEntityInput{}, "", fmt.Errorf("%w: %v", ErrQuestionnaireTransformInvalid, err)
	}

	metadata, err := transformMetadata(req)
	if err != nil {
		return entgen.CreateEntityInput{}, "", err
	}

	input.OwnerID = &req.OrganizationID
	input.VendorMetadata = metadata

	if lo.FromPtr(input.ExternalID) == "" {
		input.ExternalID = input.Name
	}

	if lo.FromPtr(input.ExternalID) == "" {
		return entgen.CreateEntityInput{}, "", fmt.Errorf("%w: requires external_id or name", ErrQuestionnaireTransformInvalid)
	}

	if err := applyInternalOwner(ctx, client, req.OrganizationID, ownerValue, &input); err != nil {
		return entgen.CreateEntityInput{}, "", err
	}

	if err := applyEnvironment(ctx, client, req.OrganizationID, environmentValue, &input); err != nil {
		return entgen.CreateEntityInput{}, "", err
	}

	// external intake submissions default to the org's vendor entity type
	if req.TemplateKind == enums.TemplateKindExternalIntake && input.EntityTypeID == nil {
		id, err := client.EntityType.Query().
			Where(
				entitytype.OwnerIDEQ(req.OrganizationID),
				entitytype.NameEqualFold("vendor"),
			).
			FirstID(ctx)

		switch {
		case entgen.IsNotFound(err):
		case err != nil:
			return entgen.CreateEntityInput{}, "", err
		default:
			input.EntityTypeID = &id
		}
	}

	return input, notes, nil
}

// applyInternalOwner resolves an owner value to a user, group, or free-text internal owner
func applyInternalOwner(ctx context.Context, client *entgen.Client, organizationID, ownerValue string, input *entgen.CreateEntityInput) error {
	ownerValue = strings.TrimSpace(ownerValue)
	if ownerValue == "" {
		return nil
	}

	if _, err := mail.ParseAddress(ownerValue); err == nil {
		userID, err := orgUserIDByEmail(ctx, client, organizationID, ownerValue)
		if err != nil {
			return err
		}

		if userID != "" {
			input.InternalOwnerUserID = &userID

			return nil
		}
	}

	groupID, err := client.Group.Query().
		Where(
			group.OwnerIDEQ(organizationID),
			group.Or(
				group.NameEqualFold(ownerValue),
				group.DisplayNameEqualFold(ownerValue),
			),
		).
		FirstID(ctx)
	if err != nil && !entgen.IsNotFound(err) {
		return err
	}

	if groupID != "" {
		input.InternalOwnerGroupID = &groupID

		return nil
	}

	input.InternalOwner = &ownerValue

	return nil
}

// applyEnvironment finds or creates the environment custom enum and sets it on the input
func applyEnvironment(ctx context.Context, client *entgen.Client, organizationID, environment string, input *entgen.CreateEntityInput) error {
	environment = strings.TrimSpace(environment)
	if environment == "" {
		return nil
	}

	predicates := []predicate.CustomTypeEnum{
		customtypeenum.NameEqualFold(environment),
		customtypeenum.FieldEQ("environment"),
		customtypeenum.DeletedAtIsNil(),
		customtypeenum.Or(
			customtypeenum.SystemOwned(true),
			customtypeenum.OwnerIDEQ(organizationID),
		),
	}

	enum, err := client.CustomTypeEnum.Query().
		Where(append(predicates, customtypeenum.ObjectTypeEQ(""))...).
		Only(ctx)
	if err != nil && entgen.IsNotFound(err) {
		enum, err = client.CustomTypeEnum.Query().
			Where(append(predicates, customtypeenum.ObjectTypeEQ("entity"))...).
			Only(ctx)
	}

	if err != nil && !entgen.IsNotFound(err) {
		return err
	}

	if entgen.IsNotFound(err) {
		enum, err = client.CustomTypeEnum.Create().
			SetName(environment).
			SetField("environment").
			SetObjectType("").
			SetOwnerID(organizationID).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	input.EnvironmentID = &enum.ID
	input.EnvironmentName = &enum.Name

	return nil
}

// upsertEntity creates the mapped entity, updating the existing row when the owner already
// has an entity with the same external id
func upsertEntity(ctx context.Context, client *entgen.Client, input entgen.CreateEntityInput) (*entgen.Entity, error) {
	existing, err := client.Entity.Query().
		Where(
			entity.OwnerID(lo.FromPtr(input.OwnerID)),
			entity.ExternalID(lo.FromPtr(input.ExternalID)),
		).
		First(ctx)

	switch {
	case entgen.IsNotFound(err):
		return client.Entity.Create().SetInput(input).Save(ctx)
	case err != nil:
		return nil, err
	}

	var update entgen.UpdateEntityInput
	if err := jsonx.RoundTrip(input, &update); err != nil {
		return nil, err
	}

	return client.Entity.UpdateOne(existing).SetInput(update).Save(ctx)
}

// claimAssessmentResponse atomically claims the assessment response by flipping its empty
// entity_id, the mutex that keeps concurrent deliveries from double-linking
func claimAssessmentResponse(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, record *entgen.Entity) (bool, error) {
	if record == nil || req.AssessmentResponseID == "" {
		return false, nil
	}

	update := client.AssessmentResponse.Update().
		Where(
			assessmentresponse.ID(req.AssessmentResponseID),
			assessmentresponse.Or(assessmentresponse.EntityIDIsNil(), assessmentresponse.EntityIDEQ("")),
		).
		SetEntityID(record.ID)

	displayName := lo.CoalesceOrEmpty(strings.TrimSpace(record.DisplayName), strings.TrimSpace(record.Name))
	if displayName != "" {
		update.SetDisplayName(displayName)
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return false, fmt.Errorf("link transformed entity to assessment response: %w", err)
	}

	return affected > 0, nil
}

// linkEntitySources idempotently links the document data and note to the transformed entity
func linkEntitySources(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, entityID string, notes string) error {
	if req.DocumentDataID != "" {
		if err := client.DocumentData.UpdateOneID(req.DocumentDataID).
			AddEntityIDs(entityID).
			Exec(ctx); err != nil && !entgen.IsConstraintError(err) {
			return fmt.Errorf("link transformed entity to document data: %w", err)
		}
	}

	return createEntityNote(ctx, client, req, entityID, notes)
}

// mappedNotes extracts the mapped note text from the submitted document data
func mappedNotes(req questionnaireTransformRequest) string {
	for _, mapping := range req.Config.Mappings {
		if mapping.Resolver == "" && mapping.To == entityTransformFieldNotes {
			if value, ok := mapx.ValueAtPath(req.Data, mapping.From); ok && !isEmptyValue(value) {
				return getStringValue(value)
			}
		}
	}

	return ""
}

// createEntityNote find-or-creates the note recording the submitted note text and links it
// to the transformed entity
func createEntityNote(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, entityID string, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	reference := fmt.Sprintf("%s:%s", transformMetadataKey, req.AssessmentResponseID)

	id, err := client.Note.Query().
		Where(
			note.OwnerIDEQ(req.OrganizationID),
			note.NoteRefEQ(reference),
		).
		OnlyID(ctx)
	if err != nil && !entgen.IsNotFound(err) {
		return fmt.Errorf("query transformed entity note: %w", err)
	}

	if id == "" {
		created, err := client.Note.Create().
			SetOwnerID(req.OrganizationID).
			SetText(text).
			SetNoteRef(reference).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create transformed entity note: %w", err)
		}

		id = created.ID
	}

	if err := client.Entity.UpdateOneID(entityID).
		AddNoteIDs(id).
		Exec(ctx); err != nil && !entgen.IsConstraintError(err) {
		return fmt.Errorf("link transformed entity note: %w", err)
	}

	return nil
}

// questionnaireTransformMetadata is the questionnaire provenance stored on the transformed entity
type questionnaireTransformMetadata struct {
	Source string `json:"source"`
	questionnaireTransformRequest
}

// transformMetadata records the questionnaire provenance stored on the transformed entity
func transformMetadata(req questionnaireTransformRequest) (map[string]any, error) {
	metadata, err := jsonx.ToMap(questionnaireTransformMetadata{
		Source:                        transformMetadataKey,
		questionnaireTransformRequest: req,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{transformMetadataKey: metadata}, nil
}

// isEmptyValue reports whether a document value is nil or blank text
func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}

	if value, ok := value.(string); ok {
		return strings.TrimSpace(value) == ""
	}

	return false
}

// getStringValue coerces payload values to strings through the shared entityops helper;
// unrepresentable values coerce to the empty string
func getStringValue(value any) string {
	coerced, _ := entityops.ValueAsString(value)

	return coerced
}
