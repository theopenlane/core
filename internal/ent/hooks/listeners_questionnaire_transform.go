package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// questionnaireTransformListenerName identifies the assessment response transform listener
const questionnaireTransformListenerName = "questionnaire.transform.assessment"

// RegisterGalaQuestionnaireTransformListeners registers listeners that transform
// completed questionnaire document data into configured target schemas.
// Supported types are defined in `TemplateProjectionTarget` enums
func RegisterGalaQuestionnaireTransformListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g, eventqueue.MutationListener{
		Schema:     entgen.TypeAssessmentResponse,
		Name:       questionnaireTransformListenerName,
		Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String()},
		Handle:     handleAssessmentResponse,
	})
}

func handleAssessmentResponse(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	if !questionnaireTransformFieldChanged(payload) {
		return nil
	}

	allowCtx := workflows.AllowContext(inv.Context)

	response, err := inv.Client.AssessmentResponse.Query().
		Where(assessmentresponse.IDEQ(inv.EntityID)).
		WithDocument().
		WithAssessment(func(query *entgen.AssessmentQuery) {
			query.WithTemplate()
		}).
		Only(allowCtx)
	if err != nil {
		if entgen.IsNotFound(err) {
			logx.FromContext(inv.Context).Error().
				Err(err).
				Str("assessment_response_id", inv.EntityID).
				Msg("assessment response not found for questionnaire transform")

			return nil
		}

		logx.FromContext(inv.Context).Error().
			Err(err).
			Str("assessment_response_id", inv.EntityID).
			Msg("failed to load assessment response for questionnaire transform")

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
		logx.FromContext(inv.Context).Debug().
			Str("assessment_response_id", response.ID).
			Str("entity_id", response.EntityID).
			Msg("assessment response already transformed")

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

	if err := transformQuestionnaire(allowCtx, inv.Client, req); err != nil {
		logger := logx.FromContext(inv.Context).Error().
			Err(err).
			Str("assessment_response_id", response.ID).
			Str("assessment_id", assessment.ID).
			Str("template_id", assessment.TemplateID).
			Str("document_data_id", response.DocumentDataID)

		// validation errors are permanent configuration or data problems: log and drop
		// the event rather than burning River retries on an outcome that cannot change
		if isQuestionnaireValidationError(err) {
			logger.Msg("questionnaire transform skipped due to invalid transform data")

			return nil
		}

		logger.Msg("questionnaire transform failed")

		if errors.Is(err, entityops.ErrUpsertConflict) {
			return nil
		}

		return err
	}

	return nil
}

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

func questionnaireTransformFieldChanged(payload eventqueue.MutationGalaPayload) bool {
	if payload.Operation == ent.OpCreate.String() {
		return true
	}

	return eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldStatus) ||
		eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldDocumentDataID) ||
		eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldCompletedAt) ||
		eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldIsDraft)
}

const transformMetadataKey = "questionnaire_transform"
const entityTransformFieldNotes = "notes"
const entityTransformFieldEntityTypeID = "entityTypeID"

type questionnaireValidationError struct {
	Message string
}

func (e *questionnaireValidationError) Error() string { return e.Message }

func isQuestionnaireValidationError(err error) bool {
	var validationErr *questionnaireValidationError
	return errors.As(err, &validationErr)
}

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

	if err := connectEntitySources(ctx, client, req, record); err != nil {
		return err
	}

	if err := createEntityNote(ctx, client, req, record.ID, mapped.Notes); err != nil {
		return err
	}

	return nil
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
		userID, err := client.User.Query().
			Where(
				user.EmailEqualFold(ownerValue),
				user.HasOrgMembershipsWith(orgmembership.OrganizationID(organizationID)),
			).
			OnlyID(ctx)
		if err != nil && !entgen.IsNotFound(err) {
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

func connectEntitySources(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, record *entgen.Entity) error {
	if record == nil {
		return nil
	}

	if req.DocumentDataID != "" {
		if err := client.DocumentData.UpdateOneID(req.DocumentDataID).
			AddEntityIDs(record.ID).
			Exec(ctx); err != nil && !entgen.IsConstraintError(err) {
			return fmt.Errorf("link transformed entity to document data: %w", err)
		}
	}

	if req.AssessmentResponseID != "" {
		update := client.AssessmentResponse.UpdateOneID(req.AssessmentResponseID).
			SetEntityID(record.ID)

		displayName := strings.TrimSpace(record.DisplayName)
		if displayName == "" {
			displayName = strings.TrimSpace(record.Name)
		}

		if displayName != "" {
			update.SetDisplayName(displayName)
		}

		if err := update.Exec(ctx); err != nil {
			return fmt.Errorf("link transformed entity to assessment response: %w", err)
		}
	}

	return nil
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

// getStringValue coerces payload values to strings through the shared eventqueue helper;
// unrepresentable values coerce to the empty string
func getStringValue(value any) string {
	coerced, _ := eventqueue.ValueAsString(value)

	return coerced
}
