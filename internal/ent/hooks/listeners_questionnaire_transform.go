package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaQuestionnaireTransformListeners registers listeners that transform
// completed questionnaire document data into configured target schemas.
// Supported types are defined in `TemplateProjectionTarget` enums
func RegisterGalaQuestionnaireTransformListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeAssessmentResponse),
			Name:       "questionnaire.transform.assessment",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String()},
			Handle:     handleAssessmentResponse,
		},
	)
}

func handleAssessmentResponse(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !questionnaireTransformFieldChanged(payload) {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	id, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || id == "" {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx.Context)

	response, err := client.AssessmentResponse.Query().
		Where(assessmentresponse.IDEQ(id)).
		WithDocument().
		WithAssessment(func(query *entgen.AssessmentQuery) {
			query.WithTemplate()
		}).
		Only(allowCtx)
	if err != nil {
		if entgen.IsNotFound(err) {
			logx.FromContext(ctx.Context).Error().
				Err(err).
				Str("assessment_response_id", id).
				Msg("assessment response not found for questionnaire transform")

			return nil
		}

		logx.FromContext(ctx.Context).Error().
			Err(err).
			Str("assessment_response_id", id).
			Msg("failed to load assessment response for questionnaire transform")

		return err
	}

	assessment, document, config, ok := validateQuestionnaire(response)
	if !ok {
		return nil
	}

	organizationID := response.OwnerID
	if organizationID == "" {
		organizationID = document.OwnerID
	}
	if organizationID == "" {
		organizationID = assessment.OwnerID
	}

	err = transformQuestionnaire(allowCtx, client, questionnaireTransformRequest{
		OrganizationID:       organizationID,
		TemplateID:           assessment.TemplateID,
		AssessmentID:         assessment.ID,
		AssessmentResponseID: response.ID,
		DocumentDataID:       response.DocumentDataID,
		Email:                response.Email,
		Data:                 document.Data,
		Config:               config,
	})
	if err == nil {
		return nil
	}

	logger := logx.FromContext(ctx.Context).Error().
		Err(err).
		Str("assessment_response_id", response.ID).
		Str("assessment_id", assessment.ID).
		Str("template_id", assessment.TemplateID).
		Str("document_data_id", response.DocumentDataID)

	if isQuestionnaireValidationError(err) {
		logger.Msg("questionnaire transform skipped due to invalid transform data")

		return nil
	}

	logger.Msg("questionnaire transform failed")

	return err
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
const questionnaireTransformDefinitionID = "questionnaire_transform"

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

var questionnaireTransformInputTypes = map[string]reflect.Type{
	integrationgenerated.IntegrationMappingSchemaEntity: reflect.TypeOf(entgen.CreateEntityInput{}),
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
	values, err := resolveTransformMappings(ctx, client, req)
	if err != nil {
		return err
	}

	mapped, err := buildMappedTransformPayload(integrationgenerated.IntegrationMappingSchemaEntity, values, req)
	if err != nil {
		return err
	}

	record, err := persistTransformPayload(ctx, client, req, integrationgenerated.IntegrationMappingSchemaEntity, mapped)
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

func resolveTransformMappings(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest) (map[string]any, error) {
	if len(req.Config.Mappings) == 0 {
		return nil, &questionnaireValidationError{Message: "transform configuration has no mappings"}
	}

	values := map[string]any{}

	for _, mapping := range req.Config.Mappings {
		rawValue, ok := valueAtPath(req.Data, mapping.From)
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

		value, err := applyTransform(rawValue, mapping.Transform)
		if err != nil {
			return nil, err
		}

		if mapping.To == "" {
			return nil, &questionnaireValidationError{Message: fmt.Sprintf("transform mapping for %q is missing target field", mapping.From)}
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

func applyTransform(value any, transform enums.TemplateProjectionTransform) (any, error) {
	if transform == "" {
		return value, nil
	}

	switch transform {
	case enums.TemplateProjectionTransformString:
		return getStringValue(value), nil

	case enums.TemplateProjectionTransformSlugify:
		return strcase.KebabCase(strings.TrimSpace(getStringValue(value))), nil

	case enums.TemplateProjectionTransformDate:
		return getDatetimeValue(value)

	case enums.TemplateProjectionTransformBool:
		return getBoolValue(value)

	case enums.TemplateProjectionTransformFloat:
		return getFloatValue(value)

	case enums.TemplateProjectionTransformStringArray:

		return getStringArrayValue(value), nil
	default:
		return nil, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform %q", transform)}
	}
}

func buildMappedTransformPayload(schemaName string, values map[string]any, req questionnaireTransformRequest) (mappedTransform, error) {
	schema, ok := integrationgenerated.IntegrationMappingSchemas[schemaName]
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

		inputKey, ok := getIntegrationKey(schemaName, schema, field)
		if !ok {
			return mappedTransform{}, &questionnaireValidationError{Message: fmt.Sprintf("unsupported %s transform field %q", schemaName, field)}
		}

		mapped.Payload[inputKey] = value
	}

	mapped.Payload[integrationgenerated.IntegrationMappingEntityOwnerID] = req.OrganizationID
	mapped.Payload[integrationgenerated.IntegrationMappingEntityVendorMetadata] = transformMetadata(req)

	if _, ok := mapped.Payload[integrationgenerated.IntegrationMappingEntityExternalID]; !ok {
		mapped.Payload[integrationgenerated.IntegrationMappingEntityExternalID] = mapped.Payload[integrationgenerated.IntegrationMappingEntityName]
	}

	mapped.ExternalID = strings.TrimSpace(getStringValue(mapped.Payload[integrationgenerated.IntegrationMappingEntityExternalID]))
	if mapped.ExternalID == "" {
		return mappedTransform{}, &questionnaireValidationError{Message: "entity transform requires external_id or name"}
	}

	return mapped, nil
}

func getIntegrationKey(schemaName string, schema integrationgenerated.IntegrationMappingSchema, key string) (string, bool) {
	if _, ok := schema.AllowedKeys[key]; ok {
		return key, true
	}

	normalizedKey := strcase.LowerCamelCase(key)
	if _, ok := schema.AllowedKeys[normalizedKey]; ok {
		return normalizedKey, true
	}

	for _, field := range schema.Fields {
		if key == field.EntField || strings.EqualFold(key, field.InputKey) || normalizedKey == field.InputKey {
			return field.InputKey, true
		}
	}

	return directInputKey(schemaName, normalizedKey)
}

func directInputKey(schemaName string, key string) (string, bool) {
	inputType, ok := questionnaireTransformInputTypes[schemaName]
	if !ok || key == "" {
		return "", false
	}

	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		fieldKey := strcase.LowerCamelCase(field.Name)
		if key == fieldKey || strings.EqualFold(key, field.Name) {
			return fieldKey, true
		}
	}

	return "", false
}

func persistTransformPayload(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, schema string, mapped mappedTransform) (*entgen.Entity, error) {
	payload, err := json.Marshal(mapped.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal transformed input: %w", err)
	}

	ingestRegistry := registry.New()
	if err := ingestRegistry.Register(questionnaireTransformDefinition()); err != nil {
		return nil, fmt.Errorf("register questionnaire transform ingest definition: %w", err)
	}

	integration := &entgen.Integration{
		DefinitionID: questionnaireTransformDefinitionID,
		OwnerID:      req.OrganizationID,
	}

	sets := []integrationtypes.IngestPayloadSet{
		{
			Schema: schema,
			Envelopes: []integrationtypes.MappingEnvelope{
				{
					Resource: "questionnaire_document_data",
					Action:   "completed",
					Payload:  payload,
				},
			},
		},
	}

	if err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    ingestRegistry,
		DB:          client,
		Integration: integration,
	}, "", []integrationtypes.IngestContract{
		{Schema: schema},
	}, sets, operations.IngestOptions{
		Source: integrationgenerated.IntegrationIngestSourceDirect,
	}); err != nil {
		return nil, fmt.Errorf("persist transformed input: %w", err)
	}

	record, err := client.Entity.Query().
		Where(
			entity.OwnerIDEQ(req.OrganizationID),
			entity.ExternalIDEQ(mapped.ExternalID),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("query transformed entity: %w", err)
	}

	return record, nil
}

func questionnaireTransformDefinition() integrationtypes.Definition {
	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          questionnaireTransformDefinitionID,
			Family:      questionnaireTransformDefinitionID,
			DisplayName: "Questionnaire Transform",
			Active:      true,
			Visible:     false,
		},
		Mappings: []integrationtypes.MappingRegistration{
			{
				Schema: integrationgenerated.IntegrationMappingSchemaEntity,
				Spec: integrationtypes.MappingOverride{
					MapExpr: "",
				},
			},
		},
	}
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

// valueAtPath lets us read keys from a json map. this also supports the
// "outerkey.innerkey" format
func valueAtPath(data map[string]any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, false
	}

	current := any(data)

	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return nil, false
		}

		currentMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		value, ok := currentMap[part]
		if !ok {
			return nil, false
		}

		current = value
	}

	return current, true
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

func getStringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func getBoolValue(value any) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {

		case "true", "yes", "y", "1":

			return true, nil

		case "false", "no", "n", "0":

			return false, nil
		}
	}

	return false, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform bool value %T", value)}
}

func getFloatValue(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, &questionnaireValidationError{Message: fmt.Sprintf("invalid transform float %q", v)}
		}

		return parsed, nil
	default:
		return 0, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform float value %T", value)}
	}
}

func getStringArrayValue(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		values := make([]string, 0, len(v))

		for _, i := range v {
			value := strings.TrimSpace(getStringValue(i))
			if value != "" {
				values = append(values, value)
			}
		}

		return values
	default:

		value := strings.TrimSpace(getStringValue(value))
		if value == "" {
			return nil
		}

		return []string{value}
	}
}

func getDatetimeValue(value any) (models.DateTime, error) {
	switch v := value.(type) {
	case models.DateTime:
		return v, nil
	case time.Time:
		return models.DateTime(v), nil
	case string:
		parsed, err := models.ToDateTime(strings.TrimSpace(v))
		if err != nil {
			return models.DateTime{}, &questionnaireValidationError{Message: fmt.Sprintf("invalid transform date %q", v)}
		}

		return *parsed, nil
	default:
		return models.DateTime{}, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform date value %T", value)}
	}
}
