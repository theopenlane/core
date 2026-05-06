package hooks

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
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
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaQuestionnaireTransformListeners registers listeners that transform
// completed questionnaire document data into configured target schemas.
// this supports only entities for now
func RegisterGalaQuestionnaireTransformListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeAssessmentResponse),
			Name:       "questionnaire.transform.assessment",
			Operations: []string{ent.OpUpdateOne.String()},
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
			return nil
		}

		return err
	}

	if response.Status != enums.AssessmentResponseStatusCompleted || response.DocumentDataID == "" {
		return nil
	}

	assessment := response.Edges.Assessment
	if assessment == nil {
		return nil
	}

	if assessment.Edges.Template == nil {
		return nil
	}

	config := assessment.Edges.Template.TransformConfiguration
	if !config.Enabled {
		return nil
	}

	document := response.Edges.Document
	if document == nil {
		return nil
	}

	err = transformQuestionnaire(allowCtx, client, questionnaireTransformRequest{
		OrganizationID:       assessment.OwnerID,
		TemplateID:           assessment.TemplateID,
		AssessmentID:         assessment.ID,
		AssessmentResponseID: response.ID,
		DocumentDataID:       response.DocumentDataID,
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

func questionnaireTransformFieldChanged(payload eventqueue.MutationGalaPayload) bool {
	return eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldStatus) ||
		eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldDocumentDataID) ||
		eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldCompletedAt) ||
		eventqueue.MutationFieldChanged(payload, assessmentresponse.FieldIsDraft)
}

const transformMetadataKey = "questionnaire_transform"

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
	Data                 map[string]any
	Config               models.TemplateProjectionConfig
}

type entityTransform struct {
	Name                 string
	DisplayName          string
	Description          string
	Status               enums.EntityStatus
	ContractStartDate    *models.DateTime
	ContractEndDate      *models.DateTime
	HasSoc2              *bool
	AnnualSpend          *float64
	BillingModel         string
	Links                []string
	EnvironmentName      string
	EnvironmentID        string
	InternalOwner        string
	InternalOwnerUserID  string
	InternalOwnerGroupID string
}

func transformQuestionnaire(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest) error {
	// if the transformation config is enabled
	if !req.Config.Enabled {
		return nil
	}

	if req.OrganizationID == "" {
		return &questionnaireValidationError{Message: "missing transform organization id"}
	}

	switch {
	case strings.EqualFold(req.Config.Target.String(), enums.TemplateProjectionTargetEntity.String()):
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

	input, err := entityFromMapping(values)
	if err != nil {
		return err
	}

	if input.Status == "" {
		input.Status = getEntityStatusFromContract(input.ContractStartDate)
	}

	metadata := transformMetadata(req)
	record, err := upsertEntity(ctx, client, req.OrganizationID, input, metadata)
	if err != nil {
		return err
	}

	if err := connectEntitySources(ctx, client, req, record.ID); err != nil {
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
			if mapping.Required {
				return nil, &questionnaireValidationError{Message: fmt.Sprintf("missing required transform field %q", mapping.From)}
			}

			continue
		}

		if strings.EqualFold(string(mapping.Resolver), string(models.TemplateProjectionResolverInternalOwner)) {
			if err := resolveInternalOwner(ctx, client, req.OrganizationID, rawValue, values); err != nil {
				return nil, err
			}

			continue
		}

		if strings.EqualFold(string(mapping.Resolver), string(models.TemplateProjectionResolverEnvironment)) {
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

func getEntityStatusFromContract(t *models.DateTime) enums.EntityStatus {
	if t == nil || t.IsZero() {
		return enums.EntityStatusUnderReview
	}

	if time.Time(*t).Before(time.Now()) {
		return enums.EntityStatusActive
	}

	return enums.EntityStatusUnderReview
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

	if err != nil {
		if entgen.IsNotFound(err) {
			return &questionnaireValidationError{Message: fmt.Sprintf("environment %q is not configured", environment)}
		}

		return fmt.Errorf("resolve environment: %w", err)
	}

	values[entity.FieldEnvironmentID] = enum.ID
	values[entity.FieldEnvironmentName] = enum.Name

	return nil
}

func applyTransform(value any, transform enums.TemplateProjectionTransform) (any, error) {
	if transform == "" {
		return value, nil
	}

	switch {
	case strings.EqualFold(transform.String(), enums.TemplateProjectionTransformString.String()):
		return getStringValue(value), nil

	case strings.EqualFold(transform.String(), enums.TemplateProjectionTransformSlugify.String()):
		return strcase.KebabCase(strings.TrimSpace(getStringValue(value))), nil

	case strings.EqualFold(transform.String(), enums.TemplateProjectionTransformDate.String()):
		return getDatetimeValue(value)

	case strings.EqualFold(transform.String(), enums.TemplateProjectionTransformBool.String()):
		return getBoolValue(value)

	case strings.EqualFold(transform.String(), enums.TemplateProjectionTransformFloat.String()):
		return getFloatValue(value)

	case strings.EqualFold(transform.String(), enums.TemplateProjectionTransformStringArray.String()):

		return getStringArrayValue(value), nil
	default:
		return nil, &questionnaireValidationError{Message: fmt.Sprintf("unsupported transform %q", transform)}
	}
}

func entityFromMapping(values map[string]any) (entityTransform, error) {
	var input entityTransform

	for field, value := range values {
		switch field {
		case entity.FieldName:

			input.Name = strings.TrimSpace(getStringValue(value))

		case entity.FieldDisplayName:

			input.DisplayName = strings.TrimSpace(getStringValue(value))

		case entity.FieldDescription:
			input.Description = strings.TrimSpace(getStringValue(value))

		case entity.FieldStatus:

			status, ok := value.(enums.EntityStatus)
			if !ok {
				statusValue := strings.TrimSpace(getStringValue(value))
				parsed := enums.ToEntityStatus(statusValue)
				if parsed == nil || *parsed == "" {
					return input, &questionnaireValidationError{Message: fmt.Sprintf("invalid entity status %q", statusValue)}
				}
				status = *parsed
			}

			input.Status = status
		case entity.FieldContractStartDate:

			t, ok := value.(models.DateTime)
			if !ok {
				return input, &questionnaireValidationError{Message: "contract_start_date must be a date"}
			}

			input.ContractStartDate = &t

		case entity.FieldContractEndDate:
			t, ok := value.(models.DateTime)
			if !ok {
				return input, &questionnaireValidationError{Message: "contract_end_date must be a date"}
			}

			input.ContractEndDate = &t

		case entity.FieldHasSoc2:
			hasSoc2, ok := value.(bool)
			if !ok {
				return input, &questionnaireValidationError{Message: "has_soc2 must be a bool"}
			}

			input.HasSoc2 = &hasSoc2

		case entity.FieldAnnualSpend:
			annualSpend, ok := value.(float64)
			if !ok {
				return input, &questionnaireValidationError{Message: "annual_spend must be a float"}
			}

			input.AnnualSpend = &annualSpend

		case entity.FieldBillingModel:
			input.BillingModel = strings.TrimSpace(getStringValue(value))

		case entity.FieldLinks:
			links, ok := value.([]string)
			if !ok {
				return input, &questionnaireValidationError{Message: "links must be a string array"}
			}

			input.Links = links

		case entity.FieldEnvironmentName:
			input.EnvironmentName = strings.TrimSpace(getStringValue(value))

		case entity.FieldEnvironmentID:
			input.EnvironmentID = strings.TrimSpace(getStringValue(value))

		case entity.FieldInternalOwner:

			input.InternalOwner = strings.TrimSpace(getStringValue(value))

		case entity.FieldInternalOwnerUserID:

			input.InternalOwnerUserID = strings.TrimSpace(getStringValue(value))

		case entity.FieldInternalOwnerGroupID:

			input.InternalOwnerGroupID = strings.TrimSpace(getStringValue(value))

		default:
			return input, &questionnaireValidationError{Message: fmt.Sprintf("unsupported entity transform field %q", field)}
		}
	}

	if input.DisplayName == "" {
		return input, &questionnaireValidationError{Message: "entity transform requires display_name"}
	}

	if input.Name == "" {
		input.Name = strcase.KebabCase(strings.TrimSpace(input.DisplayName))
	}

	if input.Name == "" {
		return input, &questionnaireValidationError{Message: "entity transform requires name"}
	}

	return input, nil
}

func upsertEntity(ctx context.Context, client *entgen.Client, organizationID string, input entityTransform, metadata map[string]any) (*entgen.Entity, error) {
	existing, err := client.Entity.Query().
		Where(
			entity.OwnerIDEQ(organizationID),
			entity.NameEqualFold(input.Name),
		).
		Only(ctx)
	if err != nil && !entgen.IsNotFound(err) {
		return nil, fmt.Errorf("query transformed entity: %w", err)
	}

	if existing != nil {
		return updateEntity(ctx, client, existing, input, metadata)
	}

	created, err := createEntity(ctx, client, organizationID, input, metadata)
	if err != nil {
		return nil, fmt.Errorf("create transformed entity: %w", err)
	}

	return created, nil
}

func createEntity(ctx context.Context, client *entgen.Client, organizationID string, input entityTransform, metadata map[string]any) (*entgen.Entity, error) {
	create := client.Entity.Create().
		SetOwnerID(organizationID).
		SetName(input.Name).
		SetDisplayName(input.DisplayName).
		SetStatus(input.Status).
		SetVendorMetadata(metadata)

	if input.Description != "" {
		create.SetDescription(input.Description)
	}

	if input.ContractStartDate != nil {
		create.SetContractStartDate(*input.ContractStartDate)
	}

	if input.ContractEndDate != nil {
		create.SetContractEndDate(*input.ContractEndDate)
	}

	if input.HasSoc2 != nil {
		create.SetHasSoc2(*input.HasSoc2)
	}

	if input.AnnualSpend != nil {
		create.SetAnnualSpend(*input.AnnualSpend)
	}

	if input.BillingModel != "" {
		create.SetBillingModel(input.BillingModel)
	}

	if len(input.Links) > 0 {
		create.SetLinks(input.Links)
	}

	if input.EnvironmentID != "" {
		create.SetEnvironmentID(input.EnvironmentID)
		create.SetEnvironmentName(input.EnvironmentName)
	}

	switch {
	case input.InternalOwnerUserID != "":
		create.SetInternalOwnerUserID(input.InternalOwnerUserID)
	case input.InternalOwnerGroupID != "":
		create.SetInternalOwnerGroupID(input.InternalOwnerGroupID)
	case input.InternalOwner != "":
		create.SetInternalOwner(input.InternalOwner)
	}

	return create.Save(ctx)
}

func updateEntity(ctx context.Context, client *entgen.Client, existing *entgen.Entity, input entityTransform, metadata map[string]any) (*entgen.Entity, error) {
	update := client.Entity.UpdateOneID(existing.ID).
		SetDisplayName(input.DisplayName).
		SetStatus(input.Status).
		SetVendorMetadata(mergeTransformMetadata(existing.VendorMetadata, metadata))

	if input.Description != "" {
		update.SetDescription(input.Description)
	}

	if input.ContractStartDate != nil {
		update.SetContractStartDate(*input.ContractStartDate)
	}

	if input.ContractEndDate != nil {
		update.SetContractEndDate(*input.ContractEndDate)
	}

	if input.HasSoc2 != nil {
		update.SetHasSoc2(*input.HasSoc2)
	}

	if input.AnnualSpend != nil {
		update.SetAnnualSpend(*input.AnnualSpend)
	}

	if input.BillingModel != "" {
		update.SetBillingModel(input.BillingModel)
	}

	if len(input.Links) > 0 {
		update.SetLinks(input.Links)
	}

	if input.EnvironmentID != "" {
		update.SetEnvironmentID(input.EnvironmentID)
		update.SetEnvironmentName(input.EnvironmentName)
	}

	switch {
	case input.InternalOwnerUserID != "":
		update.SetInternalOwnerUserID(input.InternalOwnerUserID).
			ClearInternalOwnerGroupID().
			ClearInternalOwner()
	case input.InternalOwnerGroupID != "":
		update.SetInternalOwnerGroupID(input.InternalOwnerGroupID).
			ClearInternalOwnerUserID().
			ClearInternalOwner()
	case input.InternalOwner != "":
		update.SetInternalOwner(input.InternalOwner).
			ClearInternalOwnerUserID().
			ClearInternalOwnerGroupID()
	}

	return update.Save(ctx)
}

func connectEntitySources(ctx context.Context, client *entgen.Client, req questionnaireTransformRequest, entityID string) error {
	if req.DocumentDataID != "" {
		if err := client.DocumentData.UpdateOneID(req.DocumentDataID).
			AddEntityIDs(entityID).
			Exec(ctx); err != nil && !entgen.IsConstraintError(err) {
			return fmt.Errorf("link transformed entity to document data: %w", err)
		}
	}

	if req.AssessmentResponseID != "" {
		if err := client.AssessmentResponse.UpdateOneID(req.AssessmentResponseID).
			SetEntityID(entityID).
			Exec(ctx); err != nil {
			return fmt.Errorf("link transformed entity to assessment response: %w", err)
		}
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

func mergeTransformMetadata(existing map[string]any, metadata map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range existing {
		merged[key] = value
	}
	for key, value := range metadata {
		merged[key] = value
	}

	return merged
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
