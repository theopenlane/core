package graphapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/resolvers"
)

// allowedTriggerOperations defines the set of valid trigger operations
var allowedTriggerOperations = map[string]struct{}{
	"CREATE": {},
	"UPDATE": {},
	"DELETE": {},
}

// validateWorkflowDefinitionInput validates the workflow definition document against schema constraints
func validateWorkflowDefinitionInput(schemaType string, doc *models.WorkflowDefinitionDocument, celCfg *workflows.Config) error {
	if doc == nil {
		return ErrDefinitionRequired
	}

	celCfg = resolveCELConfig(celCfg)

	if err := validateSchemaTypeMatch(schemaType, doc); err != nil {
		return err
	}

	if err := validateTriggers(schemaType, doc.Triggers, celCfg); err != nil {
		return err
	}

	if err := validateConditions(doc.Conditions, celCfg); err != nil {
		return err
	}

	eligibleFields := resolveEligibleFields(schemaType, doc.Triggers)

	if err := validateActions(doc.Actions, eligibleFields, celCfg); err != nil {
		return err
	}

	if err := validateApprovalFieldSetUniqueness(doc); err != nil {
		return err
	}

	if err := validateApprovalSubmissionMode(doc.ApprovalSubmissionMode); err != nil {
		return err
	}

	if err := validateApprovalTiming(doc.ApprovalTiming); err != nil {
		return err
	}

	return nil
}

// validateSchemaTypeMatch ensures the definition schemaType matches the input schemaType
func validateSchemaTypeMatch(schemaType string, doc *models.WorkflowDefinitionDocument) error {
	if schemaType != "" && doc.SchemaType != "" && !strings.EqualFold(doc.SchemaType, schemaType) {
		return fmt.Errorf("%w: definition schemaType %q does not match input schemaType %q", ErrSchemaTypeMismatch, doc.SchemaType, schemaType)
	}

	return nil
}

// validateTriggers validates all triggers in the workflow definition
func validateTriggers(schemaType string, triggers []models.WorkflowTrigger, celCfg *workflows.Config) error {
	if len(triggers) == 0 {
		return ErrNoTriggers
	}

	for i, trig := range triggers {
		if err := validateTrigger(schemaType, trig, i, celCfg); err != nil {
			return err
		}
	}

	return nil
}

// validateTrigger validates a single workflow trigger
func validateTrigger(schemaType string, trig models.WorkflowTrigger, index int, celCfg *workflows.Config) error {
	op := strings.ToUpper(strings.TrimSpace(trig.Operation))
	if op == "" {
		return fmt.Errorf("%w: trigger %d", ErrTriggerMissingOperation, index)
	}

	if _, ok := allowedTriggerOperations[op]; !ok {
		return fmt.Errorf("%w: trigger %d has operation %q", ErrTriggerUnsupportedOperation, index, trig.Operation)
	}

	if trig.ObjectType == "" {
		return fmt.Errorf("%w: trigger %d", ErrTriggerMissingObjectType, index)
	}

	if schemaType != "" && !strings.EqualFold(trig.ObjectType.String(), schemaType) {
		return fmt.Errorf("%w: trigger %d has objectType %q, expected %q", ErrTriggerObjectTypeMismatch, index, trig.ObjectType.String(), schemaType)
	}

	if err := validateTriggerFields(trig.Fields, index); err != nil {
		return err
	}

	if err := validateTriggerEdges(trig.Edges, index); err != nil {
		return err
	}

	if trig.Expression != "" {
		if err := workflows.ValidateCELExpression(celCfg, workflows.CELScopeBase, trig.Expression); err != nil {
			return err
		}
	}

	return nil
}

// validateTriggerFields validates that trigger fields are non-empty
func validateTriggerFields(fields []string, triggerIndex int) error {
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("%w: trigger %d", ErrTriggerEmptyFieldName, triggerIndex)
		}
	}

	return nil
}

// validateTriggerEdges validates that trigger edges are non-empty
func validateTriggerEdges(edges []string, triggerIndex int) error {
	for _, edge := range edges {
		if strings.TrimSpace(edge) == "" {
			return fmt.Errorf("%w: trigger %d", ErrTriggerEmptyEdgeName, triggerIndex)
		}
	}

	return nil
}

// resolveEligibleFields determines the eligible workflow fields based on schema type or triggers
func resolveEligibleFields(schemaType string, triggers []models.WorkflowTrigger) map[string]struct{} {
	if len(triggers) > 0 {
		return workflowEligibleFieldsForObjectType(triggers[0].ObjectType)
	}

	if schemaType != "" {
		if objType := enums.ToWorkflowObjectType(schemaType); objType != nil {
			return workflowEligibleFieldsForObjectType(*objType)
		}
	}

	return nil
}

// validateActions validates all actions in the workflow definition
func validateActions(actions []models.WorkflowAction, eligibleFields map[string]struct{}, celCfg *workflows.Config) error {
	if len(actions) == 0 {
		return ErrNoActions
	}

	seenKeys := make(map[string]struct{}, len(actions))

	for i, action := range actions {
		if err := validateAction(action, i, seenKeys, eligibleFields, celCfg); err != nil {
			return err
		}
	}

	return nil
}

// validateAction validates a single workflow action
func validateAction(action models.WorkflowAction, index int, seenKeys map[string]struct{}, eligibleFields map[string]struct{}, celCfg *workflows.Config) error {
	if err := validateActionKey(action.Key, index, seenKeys); err != nil {
		return err
	}

	actionType, err := validateActionType(action.Type, index)
	if err != nil {
		return err
	}

	if err := validateActionParams(actionType, action.Params, eligibleFields, action.Key, index, celCfg); err != nil {
		return err
	}

	if action.When != "" {
		if err := workflows.ValidateCELExpression(celCfg, workflows.CELScopeAction, action.When); err != nil {
			return err
		}
	}

	return nil
}

func validateConditions(conditions []models.WorkflowCondition, celCfg *workflows.Config) error {
	if len(conditions) == 0 {
		return nil
	}

	for _, cond := range conditions {
		if cond.Expression == "" {
			continue
		}

		if err := workflows.ValidateCELExpression(celCfg, workflows.CELScopeBase, cond.Expression); err != nil {
			return err
		}
	}

	return nil
}

func resolveCELConfig(celCfg *workflows.Config) *workflows.Config {
	if celCfg == nil {
		return workflows.NewDefaultConfig()
	}

	return celCfg
}

// validateActionKey validates the action key is present and unique
func validateActionKey(key string, index int, seenKeys map[string]struct{}) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("%w: action %d", ErrActionMissingKey, index)
	}

	if _, ok := seenKeys[key]; ok {
		return fmt.Errorf("%w: %q", ErrActionDuplicateKey, key)
	}

	seenKeys[key] = struct{}{}

	return nil
}

// validateActionType validates the action type is present and supported
func validateActionType(actionType string, index int) (enums.WorkflowActionType, error) {
	if actionType == "" {
		return "", fmt.Errorf("%w: action %d", ErrActionMissingType, index)
	}

	parsed := enums.ToWorkflowActionType(actionType)
	if parsed == nil {
		return "", fmt.Errorf("%w: action %d has type %s", ErrActionUnsupportedType, index, actionType)
	}

	return *parsed, nil
}

// validateActionParams validates action parameters based on action type
func validateActionParams(actionType enums.WorkflowActionType, params json.RawMessage, eligibleFields map[string]struct{}, actionKey string, index int, celCfg *workflows.Config) error {
	var err error

	switch actionType {
	case enums.WorkflowActionTypeApproval:
		err = validateApprovalActionParams(params, eligibleFields)
	case enums.WorkflowActionTypeReview:
		err = validateReviewActionParams(params)
	case enums.WorkflowActionTypeWebhook:
		err = validateWebhookActionParams(params, celCfg)
	case enums.WorkflowActionTypeFieldUpdate:
		err = validateFieldUpdateActionParams(params)
	case enums.WorkflowActionTypeNotification:
		// Notification action params are optional; validation happens in executor
		return nil
	}

	if err != nil {
		return fmt.Errorf("%w: action %d (%s): %w", ErrActionInvalidParams, index, actionKey, err)
	}

	return nil
}

// validateApprovalFieldSetUniqueness ensures only one approval action per field set
func validateApprovalFieldSetUniqueness(doc *models.WorkflowDefinitionDocument) error {
	approvalActions := filterApprovalActions(doc.Actions)
	if len(approvalActions) <= 1 {
		return nil
	}

	seenFieldSets := make(map[string]string)

	for _, action := range approvalActions {
		fieldSetKey := extractFieldSetKey(action)
		if fieldSetKey == "" {
			continue
		}

		if existingAction, exists := seenFieldSets[fieldSetKey]; exists {
			return fmt.Errorf("%w: %q - actions %q and %q both target the same fields", ErrDuplicateApprovalFieldSet, fieldSetKey, existingAction, action.Key)
		}

		seenFieldSets[fieldSetKey] = action.Key
	}

	return nil
}

// filterApprovalActions returns only approval type actions from the list
func filterApprovalActions(actions []models.WorkflowAction) []models.WorkflowAction {
	result := make([]models.WorkflowAction, 0)

	for _, action := range actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType != nil && *actionType == enums.WorkflowActionTypeApproval {
			result = append(result, action)
		}
	}

	return result
}

// extractFieldSetKey extracts and normalizes the field set key from an approval action
func extractFieldSetKey(action models.WorkflowAction) string {
	var params struct {
		Fields []string `json:"fields"`
		Edges  []string `json:"edges"`
	}

	if action.Params == nil {
		return ""
	}

	if err := json.Unmarshal(action.Params, &params); err != nil {
		return ""
	}

	allFields := append([]string(nil), params.Fields...)
	allFields = append(allFields, params.Edges...)

	if len(allFields) == 0 {
		return ""
	}

	sort.Strings(allFields)

	return strings.Join(allFields, ",")
}

// validateApprovalSubmissionMode validates the approval submission mode if specified
func validateApprovalSubmissionMode(mode enums.WorkflowApprovalSubmissionMode) error {
	if mode == "" {
		return nil
	}

	if enums.ToWorkflowApprovalSubmissionMode(mode.String()) == nil {
		return fmt.Errorf("%w: %q", ErrApprovalSubmissionModeInvalid, mode)
	}

	if mode == enums.WorkflowApprovalSubmissionModeManualSubmit {
		return ErrManualSubmitModeNotSupported
	}

	return nil
}

// validateApprovalTiming validates the approval timing if specified
func validateApprovalTiming(timing enums.WorkflowApprovalTiming) error {
	if timing == "" {
		return nil
	}
	if enums.ToWorkflowApprovalTiming(timing.String()) == nil {
		return fmt.Errorf("%w: %q", ErrApprovalTimingInvalid, timing)
	}

	return nil
}

// validateWorkflowDefinitionConflicts checks for conflicting approval domains with existing definitions
func validateWorkflowDefinitionConflicts(ctx context.Context, client *generated.Client, schemaType string, ownerID string, currentID string, doc *models.WorkflowDefinitionDocument) error {
	objectType := enums.ToWorkflowObjectType(schemaType)
	if objectType == nil {
		return ErrInvalidWorkflowSchema
	}

	domainKeys, err := extractDomainKeys(doc, *objectType)
	if err != nil {
		return err
	}
	if len(domainKeys) == 0 {
		return nil
	}

	definitions, err := queryActiveDefinitions(ctx, client, schemaType, ownerID, currentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToQueryDefinitions, err)
	}

	return checkDomainConflicts(definitions, domainKeys)
}

// extractDomainKeys extracts all domain keys from a workflow definition document
func extractDomainKeys(doc *models.WorkflowDefinitionDocument, objectType enums.WorkflowObjectType) (map[string]struct{}, error) {
	domains, err := workflows.ApprovalDomains(*doc)
	if err != nil {
		return nil, err
	}
	if len(domains) == 0 {
		return nil, nil
	}

	domainKeys := make(map[string]struct{}, len(domains))

	for _, domain := range domains {
		key := workflows.DeriveDomainKey(objectType, domain)
		if key != "" {
			domainKeys[key] = struct{}{}
		}
	}

	return domainKeys, nil
}

// queryActiveDefinitions queries for active workflow definitions matching the criteria
func queryActiveDefinitions(ctx context.Context, client *generated.Client, schemaType string, ownerID string, currentID string) ([]*generated.WorkflowDefinition, error) {
	query := client.WorkflowDefinition.Query().
		Where(
			workflowdefinition.SchemaTypeEQ(schemaType),
			workflowdefinition.ActiveEQ(true),
			workflowdefinition.DraftEQ(false),
		)

	if ownerID != "" {
		query = query.Where(workflowdefinition.OwnerIDEQ(ownerID))
	}

	if currentID != "" {
		query = query.Where(workflowdefinition.IDNEQ(currentID))
	}

	return query.All(ctx)
}

// checkDomainConflicts checks if any existing definitions conflict with the given domain keys
func checkDomainConflicts(definitions []*generated.WorkflowDefinition, domainKeys map[string]struct{}) error {
	for _, def := range definitions {
		objectType := enums.ToWorkflowObjectType(def.SchemaType)
		if objectType == nil {
			continue
		}
		otherDomains, err := workflows.ApprovalDomains(def.DefinitionJSON)
		if err != nil {
			return err
		}
		if len(otherDomains) == 0 && len(def.ApprovalFields) > 0 {
			otherDomains = [][]string{def.ApprovalFields}
		}

		for _, domain := range otherDomains {
			key := workflows.DeriveDomainKey(*objectType, domain)
			if key == "" {
				continue
			}

			if _, ok := domainKeys[key]; ok {
				return fmt.Errorf("%w: domain %q already used by workflow definition %q", ErrConflictingApprovalDomain, key, def.ID)
			}
		}
	}

	return nil
}

// approvalActionParams holds the parsed parameters for an approval action
type approvalActionParams struct {
	Targets   []workflows.TargetConfig `json:"targets"`
	Assignees struct {
		Users     []string `json:"users"`
		Groups    []string `json:"groups"`
		Roles     []string `json:"roles"`
		Resolvers []string `json:"resolvers"`
	} `json:"assignees"`
	Required      any      `json:"required"`
	RequiredCount int      `json:"required_count"`
	Fields        []string `json:"fields"`
	Edges         []string `json:"edges"`
}

// validateApprovalActionParams validates approval action parameters
func validateApprovalActionParams(raw json.RawMessage, eligibleFields map[string]struct{}) error {
	if len(raw) == 0 {
		return ErrApprovalParamsRequired
	}

	var params approvalActionParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return err
	}

	if err := validateApprovalFields(params.Fields, params.Edges, eligibleFields); err != nil {
		return err
	}

	targets := resolveApprovalTargets(params)
	if len(targets) == 0 {
		return ErrApprovalTargetsRequired
	}

	if err := validateRequiredCount(params.RequiredCount); err != nil {
		return err
	}

	if err := validateRequiredField(params.Required); err != nil {
		return err
	}

	return validateTargets(targets)
}

// validateApprovalFields validates the fields and edges for an approval action
func validateApprovalFields(fields []string, edges []string, eligibleFields map[string]struct{}) error {
	if len(fields) == 0 && len(edges) == 0 {
		return ErrApprovalFieldRequired
	}

	if len(edges) > 0 {
		return ErrApprovalEdgesNotSupported
	}

	if len(eligibleFields) == 0 {
		return nil
	}

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		if _, ok := eligibleFields[field]; !ok {
			return fmt.Errorf("%w: %q", ErrApprovalFieldNotEligible, field)
		}
	}

	return nil
}

// resolveApprovalTargets resolves targets from both targets and legacy assignees fields
func resolveApprovalTargets(params approvalActionParams) []workflows.TargetConfig {
	if len(params.Targets) > 0 {
		return params.Targets
	}

	var targets []workflows.TargetConfig

	for _, id := range params.Assignees.Users {
		targets = append(targets, workflows.TargetConfig{Type: enums.WorkflowTargetTypeUser, ID: id})
	}

	for _, id := range params.Assignees.Groups {
		targets = append(targets, workflows.TargetConfig{Type: enums.WorkflowTargetTypeGroup, ID: id})
	}

	for _, id := range params.Assignees.Roles {
		targets = append(targets, workflows.TargetConfig{Type: enums.WorkflowTargetTypeRole, ID: id})
	}

	for _, key := range params.Assignees.Resolvers {
		targets = append(targets, workflows.TargetConfig{Type: enums.WorkflowTargetTypeResolver, ResolverKey: key})
	}

	return targets
}

// validateRequiredCount validates the required_count field
func validateRequiredCount(count int) error {
	if count < 0 {
		return ErrRequiredCountNegative
	}

	return nil
}

// validateRequiredField validates the required field value
func validateRequiredField(required any) error {
	switch v := required.(type) {
	case nil, bool:
		return nil
	case float64:
		if v < 0 {
			return ErrRequiredInvalid
		}
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}

		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "false":
			return nil
		default:
			if n, err := json.Number(strings.TrimSpace(v)).Int64(); err != nil || n < 0 {
				return ErrRequiredInvalid
			}
		}
	}

	return nil
}

// validateTargets validates all targets in the approval action
func validateTargets(targets []workflows.TargetConfig) error {
	for _, t := range targets {
		if err := validateTarget(t); err != nil {
			return err
		}
	}

	return nil
}

// validateTarget validates a single target configuration
func validateTarget(t workflows.TargetConfig) error {
	switch t.Type {
	case enums.WorkflowTargetTypeUser, enums.WorkflowTargetTypeGroup:
		if strings.TrimSpace(t.ID) == "" {
			return fmt.Errorf("%w: %s", ErrTargetMissingID, t.Type)
		}
	case enums.WorkflowTargetTypeRole:
		if strings.TrimSpace(t.ID) == "" {
			return ErrTargetMissingID
		}

		if role := enums.ToRole(t.ID); role == nil || *role == enums.RoleInvalid {
			return fmt.Errorf("%w: %q", ErrTargetInvalidRole, t.ID)
		}
	case enums.WorkflowTargetTypeResolver:
		if strings.TrimSpace(t.ResolverKey) == "" {
			return ErrTargetMissingResolverKey
		}

		if _, ok := resolvers.Get(t.ResolverKey); !ok {
			return fmt.Errorf("%w: %q", ErrTargetUnknownResolver, t.ResolverKey)
		}
	default:
		return fmt.Errorf("%w: %q", ErrTargetInvalidType, t.Type)
	}

	return nil
}

// validateWebhookActionParams validates webhook action parameters
func validateWebhookActionParams(raw json.RawMessage, celCfg *workflows.Config) error {
	if len(raw) == 0 {
		return ErrWebhookParamsRequired
	}

	var payloadCheck map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payloadCheck); err != nil {
		return err
	}
	if _, exists := payloadCheck["payload"]; exists {
		return ErrWebhookPayloadUnsupported
	}

	var params struct {
		URL         string `json:"url"`
		Method      string `json:"method"`
		PayloadExpr string `json:"payload_expr"`
	}

	if err := json.Unmarshal(raw, &params); err != nil {
		return err
	}

	if strings.TrimSpace(params.URL) == "" {
		return ErrWebhookURLRequired
	}

	parsed, err := url.Parse(params.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrWebhookURLInvalid
	}

	if strings.TrimSpace(params.PayloadExpr) != "" {
		if err := workflows.ValidateCELExpression(celCfg, workflows.CELScopeAction, params.PayloadExpr); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldUpdateActionParams validates field update action parameters
func validateFieldUpdateActionParams(raw json.RawMessage) error {
	if len(raw) == 0 {
		return ErrFieldUpdateParamsRequired
	}

	var params struct {
		Updates map[string]any `json:"updates"`
	}

	if err := json.Unmarshal(raw, &params); err != nil {
		return err
	}

	if len(params.Updates) == 0 {
		return ErrFieldUpdateUpdatesRequired
	}

	return nil
}

// validateReviewActionParams validates review action parameters
func validateReviewActionParams(raw json.RawMessage) error {
	if len(raw) == 0 {
		return ErrReviewParamsRequired
	}

	var params workflows.ReviewActionParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return err
	}

	if len(params.Targets) == 0 {
		return ErrReviewTargetsRequired
	}

	if err := validateRequiredCount(params.RequiredCount); err != nil {
		return err
	}

	if err := validateRequiredField(params.Required); err != nil {
		return err
	}

	return validateTargets(params.Targets)
}
