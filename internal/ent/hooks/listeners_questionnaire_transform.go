package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/stoewer/go-strcase"
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
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// QuestionnaireTransformListeners returns the listeners that transform completed
// questionnaire document data into the target schemas from `TemplateProjectionTarget`
func QuestionnaireTransformListeners() []gala.Registration {
	return []gala.Registration{entityops.MutationListener{
		Schema:     entityops.SchemaAssessmentResponse,
		Label:      "transform",
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
func handleAssessmentResponse(inv entityops.Invocation, payload entityops.MutationPayload) error {
	ctx := inv.Context

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

	// responses are submit-once, so the linked entity is the durable transform-complete
	// marker: a set entity_id skips redelivered events, and failed attempts leave it
	// empty for River's retry to re-attempt the transform
	if response.EntityID != "" {
		logx.FromContext(ctx).Debug().Str("linked_entity_id", response.EntityID).Msg("assessment response already transformed")

		return nil
	}

	organizationID := response.OwnerID
	if organizationID == "" {
		organizationID = document.OwnerID
	}
	if organizationID == "" {
		organizationID = assessment.OwnerID
	}

	req := questionnaireTransformRequest{
		OrganizationID:       organizationID,
		TemplateID:           assessment.TemplateID,
		TemplateKind:         assessment.Edges.Template.Kind,
		AssessmentID:         assessment.ID,
		AssessmentResponseID: response.ID,
		DocumentDataID:       response.DocumentDataID,
		Email:                response.Email,
		Data:                 document.Data,
		Config:               config,
	}

	ctx = logx.WithFields(ctx, map[string]any{
		"assessment_id":    assessment.ID,
		"template_id":      assessment.TemplateID,
		"document_data_id": response.DocumentDataID,
	})

	if err := transformQuestionnaire(ctx, inv.Client, req); err != nil {
		// validation errors are permanent configuration or data problems: log and drop
		// the event rather than burning River retries on an outcome that cannot change
		if isQuestionnaireValidationError(err) {
			logx.FromContext(ctx).Error().Err(err).Msg("questionnaire transform skipped due to invalid transform data")

			return nil
		}

		logx.FromContext(ctx).Error().Err(err).Msg("questionnaire transform failed")

		// deterministic persistence failures cannot succeed on redelivery: decode and
		// ent validation failures are data problems, an upsert conflict means the
		// entity already exists
		if errors.Is(err, entityops.ErrUpsertConflict) || errors.Is(err, entityops.ErrDecodeFailed) || entgen.IsValidationError(err) {
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
const entityTransformFieldEntityTypeID = "entityTypeID"

type questionnaireTransformRequest struct {
	OrganizationID       string
	TemplateID           string
	TemplateKind         enums.TemplateKind
	AssessmentID         string
	AssessmentResponseID string
	DocumentDataID       string
	Email                string
	Data                 map[string]any
	Config               models.TemplateProjectionConfig
}

type mappedTransform struct {
	Payload    map[string]any
	Notes      string
	ExternalID string
}

func transformQuestionnaire(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest) error {
	// if the transformation config is enabled
	if !req.Config.Enabled {
		return nil
	}

	if req.OrganizationID == "" {
		return &questionnaireValidationError{Message: "missing transform organization id"}
	}

	switch req.Config.Target {
	case enums.TemplateProjectionTargetEntity:
		return handleEntityTransform(ctx, client, req)
	default:
		return &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform target %q", req.Config.Target)}
	}
}

func handleEntityTransform(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest) error {
	values, err := resolveTransformMappings(ctx, client, entityops.SchemaEntity, req)
	if err != nil {
		return err
	}

	mapped, err := buildMappedTransformPayload(entityops.SchemaEntity.Name, values, req)
	if err != nil {
		return err
	}

	if err := setVendorEntityType(ctx, client, req, mapped.Payload); err != nil {
		return err
	}

	record, err := persistTransformPayload(ctx, client, entityops.SchemaEntity, req.OrganizationID, mapped)
	if err != nil {
		return err
	}

	claimed, err := connectEntitySources(ctx, client, req, record)
	if err != nil {
		return err
	}

	// a lost claim means a concurrent delivery of the same response already owns linking
	// and note creation
	if !claimed {
		return nil
	}

	return createEntityNote(ctx, client, req, record.ID, mapped.Notes)
}

func resolveTransformMappings(ctx context.Context, client *entgen.Client, schema *entityops.Schema, req questionnaireTransformRequest) (map[string]any, error) {
	if len(req.Config.Mappings) == 0 {
		return nil, &questionnaireValidationError{Message: "transform configuration has no mappings"}
	}

	values := map[string]any{}

	for _, mapping := range req.Config.Mappings {
		rawValue, ok := mapx.ValueAtPath(req.Data, mapping.From)
		if !ok || isEmptyValue(rawValue) {
			if mapping.Resolver == models.TemplateProjectionResolverInternalOwner && req.Email != "" {
				if err := resolveInternalOwner(ctx, client, req.OrganizationID, req.Email, values); err != nil {
					return nil, err
				}

				continue
			}

			if mapping.Required {
				return nil, &questionnaireValidationError{Message: fmt.Sprintf("missing required transform field %q", mapping.From)}
			}

			continue
		}

		if mapping.Resolver == models.TemplateProjectionResolverInternalOwner {
			if err := resolveInternalOwner(ctx, client, req.OrganizationID, rawValue, values); err != nil {
				return nil, err
			}

			continue
		}

		if mapping.Resolver == models.TemplateProjectionResolverEnvironment {
			if err := resolveEnvironment(ctx, client, req.OrganizationID, rawValue, values); err != nil {
				return nil, err
			}

			continue
		}

		if mapping.To == "" {
			return nil, &questionnaireValidationError{Message: fmt.Sprintf("transform mapping for %q is missing target field", mapping.From)}
		}

		value, err := applyTransform(schema, mapping.To, rawValue, mapping.Transform)
		if err != nil {
			return nil, err
		}

		values[mapping.To] = value
	}

	return values, nil
}

func resolveInternalOwner(ctx context.Context, client *entgen.Client, organizationID string, rawValue any, values map[string]any) error {
	ownerValue := strings.TrimSpace(getStringValue(rawValue))
	if ownerValue == "" {
		return nil
	}

	if _, err := mail.ParseAddress(ownerValue); err == nil {
		userID, err := orgUserIDByEmail(ctx, client, organizationID, ownerValue)
		if err != nil {
			return fmt.Errorf("resolve internal owner user: %w", err)
		}

		if userID != "" {
			values[entity.FieldInternalOwnerUserID] = userID
			delete(values, entity.FieldInternalOwnerGroupID)
			delete(values, entity.FieldInternalOwner)

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
		return fmt.Errorf("resolve internal owner group: %w", err)
	}

	if groupID != "" {
		values[entity.FieldInternalOwnerGroupID] = groupID
		delete(values, entity.FieldInternalOwnerUserID)
		delete(values, entity.FieldInternalOwner)

		return nil
	}

	values[entity.FieldInternalOwner] = ownerValue
	delete(values, entity.FieldInternalOwnerUserID)
	delete(values, entity.FieldInternalOwnerGroupID)

	return nil
}

func resolveEnvironment(ctx context.Context, client *entgen.Client, organizationID string, value any, values map[string]any) error {
	environment := strings.TrimSpace(getStringValue(value))
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
		return fmt.Errorf("resolve environment: %w", err)
	}

	if entgen.IsNotFound(err) {
		enum, err = client.CustomTypeEnum.Create().
			SetName(environment).
			SetField("environment").
			SetObjectType("").
			SetOwnerID(organizationID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create environment enum: %w", err)
		}
	}

	values[entity.FieldEnvironmentID] = enum.ID
	values[entity.FieldEnvironmentName] = enum.Name

	return nil
}

func setVendorEntityType(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, payload map[string]any) error {
	// the payload carries resolved create-input keys, so the guard and the write must use
	// the catalog's key for the field or the value is dropped at decode time
	inputKey, ok := entityops.SchemaEntity.ResolveInputKey(entityTransformFieldEntityTypeID)
	if !ok {
		return &questionnaireValidationError{Message: fmt.Sprintf("unsupported %s transform field %q", entityops.SchemaEntity.Name, entityTransformFieldEntityTypeID)}
	}

	if req.TemplateKind != enums.TemplateKindExternalIntake || payload[inputKey] != nil {
		return nil
	}

	id, err := client.EntityType.Query().
		Where(
			entitytype.OwnerIDEQ(req.OrganizationID),
			entitytype.NameEqualFold("vendor"),
		).
		FirstID(ctx)

	if entgen.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("resolve external intake entity type: %w", err)
	}

	payload[inputKey] = id
	return nil
}

// applyTransform applies the configured transform to a raw document value. Slugification keeps its
// bespoke string handling; every other transform delegates type coercion to the schema catalog,
// passing values through for fields the catalog does not describe (such as the notes pseudo-field)
// so later key resolution remains the arbiter of unknown fields
func applyTransform(schema *entityops.Schema, targetField string, value any, transform enums.TemplateProjectionTransform) (any, error) {
	switch transform {
	case "":
		return value, nil
	case enums.TemplateProjectionTransformSlugify:
		return strcase.KebabCase(strings.TrimSpace(getStringValue(value))), nil
	}

	// the catalog's field type is the coercion arbiter, but the declared transform must
	// still be a known enum member so misconfigured mappings fail loudly
	if parsed := enums.ToTemplateProjectionTransform(transform.String()); parsed == nil || *parsed == enums.TemplateProjectionTransformInvalid {
		return nil, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform %q", transform)}
	}

	coerced, err := schema.CoerceValue(targetField, value)

	switch {
	case errors.Is(err, entityops.ErrFieldNotFound):
		return value, nil
	case err != nil:
		return nil, &questionnaireValidationError{Message: fmt.Sprintf("invalid transform value for %q: %v", targetField, err)}
	}

	return coerced, nil
}

func buildMappedTransformPayload(schemaName string, values map[string]any, req questionnaireTransformRequest) (mappedTransform, error) {
	schema, ok := entityops.LookupSchema(schemaName)
	if !ok {
		return mappedTransform{}, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform schema %q", schemaName)}
	}

	mapped := mappedTransform{
		Payload: map[string]any{},
	}

	for field, value := range values {
		if field == entityTransformFieldNotes {
			mapped.Notes = strings.TrimSpace(getStringValue(value))

			continue
		}

		inputKey, ok := schema.ResolveInputKey(field)
		if !ok {
			return mappedTransform{}, &questionnaireValidationError{Message: fmt.Sprintf("unsupported %s transform field %q", schemaName, field)}
		}

		mapped.Payload[inputKey] = value
	}

	mapped.Payload[entityops.InputKeyEntityOwnerID] = req.OrganizationID
	mapped.Payload[entityops.InputKeyEntityVendorMetadata] = transformMetadata(req)

	if _, ok := mapped.Payload[entityops.InputKeyEntityExternalID]; !ok {
		mapped.Payload[entityops.InputKeyEntityExternalID] = mapped.Payload[entityops.InputKeyEntityName]
	}

	mapped.ExternalID = strings.TrimSpace(getStringValue(mapped.Payload[entityops.InputKeyEntityExternalID]))
	if mapped.ExternalID == "" {
		return mappedTransform{}, &questionnaireValidationError{Message: "entity transform requires external_id or name"}
	}

	return mapped, nil
}

// persistTransformPayload upserts the mapped payload through the schema catalog's
// lookup-key upsert and returns the persisted entity
func persistTransformPayload(ctx context.Context, client *entgen.Client, schema *entityops.Schema, ownerID string, mapped mappedTransform) (*entgen.Entity, error) {
	payload, err := json.Marshal(mapped.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal transformed input: %w", err)
	}

	entityID, err := schema.Upsert(ctx, client, ownerID, payload)
	if err != nil {
		return nil, fmt.Errorf("persist transformed input: %w", err)
	}

	record, err := client.Entity.Get(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("query transformed entity: %w", err)
	}

	return record, nil
}

// connectEntitySources atomically claims the assessment response by flipping its empty
// entity_id — the claim is the transform's mutex, so concurrent deliveries of the same
// response cannot double-link or duplicate notes — then links the document data
func connectEntitySources(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, record *entgen.Entity) (bool, error) {
	if record == nil || req.AssessmentResponseID == "" {
		return false, nil
	}

	update := client.AssessmentResponse.Update().
		Where(
			assessmentresponse.ID(req.AssessmentResponseID),
			assessmentresponse.Or(assessmentresponse.EntityIDIsNil(), assessmentresponse.EntityIDEQ("")),
		).
		SetEntityID(record.ID)

	displayName := strings.TrimSpace(record.DisplayName)
	if displayName == "" {
		displayName = strings.TrimSpace(record.Name)
	}

	if displayName != "" {
		update.SetDisplayName(displayName)
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return false, fmt.Errorf("link transformed entity to assessment response: %w", err)
	}

	if affected == 0 {
		return false, nil
	}

	if req.DocumentDataID != "" {
		if err := client.DocumentData.UpdateOneID(req.DocumentDataID).
			AddEntityIDs(record.ID).
			Exec(ctx); err != nil && !entgen.IsConstraintError(err) {
			return false, fmt.Errorf("link transformed entity to document data: %w", err)
		}
	}

	return true, nil
}

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
		createdEnum, err := client.Note.Create().
			SetOwnerID(req.OrganizationID).
			SetText(text).
			SetNoteRef(reference).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create transformed entity note: %w", err)
		}

		id = createdEnum.ID
	}

	if err := client.Entity.UpdateOneID(entityID).
		AddNoteIDs(id).
		Exec(ctx); err != nil && !entgen.IsConstraintError(err) {
		return fmt.Errorf("link transformed entity note: %w", err)
	}

	return nil
}

func transformMetadata(req questionnaireTransformRequest) map[string]any {
	return map[string]any{
		transformMetadataKey: map[string]any{
			"source":                 "questionnaire_transform",
			"template_id":            req.TemplateID,
			"assessment_id":          req.AssessmentID,
			"assessment_response_id": req.AssessmentResponseID,
			"document_data_id":       req.DocumentDataID,
		},
	}
}

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
