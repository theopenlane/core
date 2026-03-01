package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/samber/lo"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/workflowgenerated"
	"github.com/theopenlane/core/internal/mutations"
	wfworkflows "github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/pkg/celx"
)

const (
	defaultWebhookRetries           = 2
	defaultWebhookInitialBackoffMS  = 200
	defaultWebhookMaxBackoffSeconds = 2
	defaultWebhookFallbackBackoffMS = 100
	defaultWebhookTimeoutMS         = 10_000

	httpStatusClientErrorMin = 400
	httpStatusClientErrorMax = 500
)

// Execute performs the action and returns any error.
func (e *WorkflowEngine) Execute(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	actionType := enums.ToWorkflowActionType(action.Type)
	if actionType == nil {
		return fmt.Errorf("%w: %s", ErrInvalidActionType, action.Type)
	}

	switch *actionType {
	case enums.WorkflowActionTypeApproval:
		return e.executeApproval(ctx, action, instance, obj)
	case enums.WorkflowActionTypeReview:
		return e.executeReview(ctx, action, instance, obj)
	case enums.WorkflowActionTypeNotification:
		return e.executeNotification(ctx, action, instance, obj)
	case enums.WorkflowActionTypeFieldUpdate:
		return e.executeFieldUpdate(ctx, action, obj)
	case enums.WorkflowActionTypeIntegration:
		return e.executeIntegrationAction(ctx, action, instance, obj)
	case enums.WorkflowActionTypeWebhook:
		return e.executeWebhook(ctx, action, instance, obj)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidActionType, action.Type)
	}
}

// gatedActionConfig configures the common execution logic for approval and review actions
type gatedActionConfig struct {
	// ActionType is the workflow action type (approval or review)
	ActionType enums.WorkflowActionType
	// KeyPrefix is the assignment key prefix ("approval" or "review")
	KeyPrefix string
	// Role is an optional role to set on assignments (e.g., "REVIEWER")
	Role string
	// Targets are the target configurations to resolve
	Targets []wfworkflows.TargetConfig
	// Required indicates whether the action is required
	Required bool
	// RequiredCount is the quorum threshold
	RequiredCount int
	// Label is the optional display label
	Label string
	// ProposedHash is the approval-specific proposal hash (empty for reviews)
	ProposedHash string
	// NoTargetsError is the error to return when no targets are found
	NoTargetsError error
}

// resolveTargetUsers resolves target user IDs and logs warnings if no users are found
func (e *WorkflowEngine) resolveTargetUsers(ctx context.Context, target wfworkflows.TargetConfig, obj *wfworkflows.Object, actionType string, actionKey string) ([]string, error) {
	allowCtx := wfworkflows.AllowContext(ctx)
	userIDs, err := e.ResolveTargets(allowCtx, target, obj)
	if err != nil {
		return nil, err
	}

	normalized := mutations.NormalizeStrings(userIDs)
	if len(normalized) == 0 {
		observability.WarnEngine(ctx, observability.OpExecuteAction, actionType, observability.ActionFields(actionKey, observability.Fields{
			workflowassignmenttarget.FieldTargetType:  target.Type.String(),
			workflowassignmenttarget.FieldResolverKey: target.ResolverKey,
		}), nil)
	}

	return normalized, nil
}

// executeGatedAction creates workflow assignments for approval and review actions
func (e *WorkflowEngine) executeGatedAction(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object, cfg gatedActionConfig) error {
	allowCtx := wfworkflows.AllowContext(ctx)

	if len(cfg.Targets) == 0 {
		observability.WarnEngine(ctx, observability.OpExecuteAction, action.Type, observability.ActionFields(action.Key, nil), nil)
		return cfg.NoTargetsError
	}

	if obj == nil {
		return ErrObjectRefMissingID
	}

	ownerID := instance.OwnerID
	if ownerID == "" {
		caller, callerOk := auth.CallerFromContext(ctx)
		if !callerOk || caller == nil || caller.OrganizationID == "" {
			return auth.ErrNoAuthUser
		}

		ownerID = caller.OrganizationID
	}

	actionIndex := actionIndexForKey(instance.DefinitionSnapshot.Actions, action.Key)
	assignmentIDs := make([]string, 0)
	targetUserIDs := make([]string, 0)
	seenAssignmentIDs := make(map[string]struct{})
	seenTargetUserIDs := make(map[string]struct{})
	seenAssignments := make(map[string]struct{})

	for _, targetConfig := range cfg.Targets {
		userIDs, err := e.resolveTargetUsers(ctx, targetConfig, obj, action.Type, action.Key)
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrFailedToResolveTarget, targetConfig.Type.String(), err)
		}
		if len(userIDs) == 0 {
			continue
		}

		for _, userID := range userIDs {
			assignmentKey := fmt.Sprintf("%s_%s_%s", cfg.KeyPrefix, action.Key, userID)
			if _, ok := seenAssignments[assignmentKey]; ok {
				continue
			}
			seenAssignments[assignmentKey] = struct{}{}

			meta := models.WorkflowAssignmentApproval{
				ActionKey:     action.Key,
				Required:      cfg.Required,
				Label:         cfg.Label,
				RequiredCount: cfg.RequiredCount,
			}
			if cfg.ProposedHash != "" {
				meta.ProposedHash = cfg.ProposedHash
			}

			assignmentCreate := e.client.WorkflowAssignment.
				Create().
				SetWorkflowInstanceID(instance.ID).
				SetAssignmentKey(assignmentKey).
				SetStatus(enums.WorkflowAssignmentStatusPending).
				SetRequired(cfg.Required).
				SetLabel(cfg.Label).
				SetApprovalMetadata(meta)
			assignmentCreate.SetOwnerID(ownerID)
			if cfg.Role != "" {
				assignmentCreate.SetRole(cfg.Role)
			}

			assignmentCreated := true
			assignment, err := assignmentCreate.Save(allowCtx)
			if err != nil {
				if generated.IsConstraintError(err) {
					assignmentCreated = false
					assignment, err = e.client.WorkflowAssignment.
						Query().
						Where(
							workflowassignment.WorkflowInstanceIDEQ(instance.ID),
							workflowassignment.AssignmentKeyEQ(assignmentKey),
							workflowassignment.OwnerIDEQ(ownerID),
						).
						Only(allowCtx)
				}
			}

			if err != nil {
				return ErrAssignmentCreationFailed
			}

			if assignment != nil {
				if _, ok := seenAssignmentIDs[assignment.ID]; !ok {
					seenAssignmentIDs[assignment.ID] = struct{}{}
					assignmentIDs = append(assignmentIDs, assignment.ID)
				}
			}
			if _, ok := seenTargetUserIDs[userID]; !ok {
				seenTargetUserIDs[userID] = struct{}{}
				targetUserIDs = append(targetUserIDs, userID)
			}

			targetCreate := e.client.WorkflowAssignmentTarget.
				Create().
				SetWorkflowAssignmentID(assignment.ID).
				SetTargetType(targetConfig.Type).
				SetTargetUserID(userID)

			switch targetConfig.Type {
			case enums.WorkflowTargetTypeGroup:
				if targetConfig.ID != "" {
					targetCreate.SetTargetGroupID(targetConfig.ID)
				}
			case enums.WorkflowTargetTypeResolver:
				if targetConfig.ResolverKey != "" {
					targetCreate.SetResolverKey(targetConfig.ResolverKey)
				}
			}

			targetCreate.SetOwnerID(ownerID)

			if err := targetCreate.Exec(allowCtx); err != nil {
				if !generated.IsConstraintError(err) {
					return ErrFailedToCreateAssignmentTarget
				}
			}

			if assignmentCreated {
				e.emitAssignmentCreated(ctx, instance, obj, assignment.ID, userID, cfg.ActionType)
			}
		}
	}

	if len(assignmentIDs) > 0 {
		details := assignmentCreatedDetailsForAction(action, cfg.ActionType, actionIndex, obj, assignmentIDs, targetUserIDs, cfg.Required, cfg.RequiredCount, cfg.Label)
		e.recordAssignmentsCreated(ctx, instance, details)
	}

	if len(assignmentIDs) == 0 {
		return cfg.NoTargetsError
	}

	return nil
}

// executeFieldUpdate applies field updates to the target object
func (e *WorkflowEngine) executeFieldUpdate(ctx context.Context, action models.WorkflowAction, obj *wfworkflows.Object) error {
	var params wfworkflows.FieldUpdateActionParams

	if err := json.Unmarshal(action.Params, &params); err != nil {
		return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
	}

	if len(params.Updates) == 0 {
		return fmt.Errorf("%w: updates", ErrMissingRequiredField)
	}

	updates := params.Updates
	if obj != nil {
		replacements := wfworkflows.BuildObjectReplacements(obj)
		if len(replacements) > 0 {
			updates = applyStringTemplates(updates, replacements)
		}
	}

	// Use both workflow bypass and privacy bypass for internal field updates
	bypassCtx := wfworkflows.AllowBypassContext(ctx)
	return wfworkflows.ApplyObjectFieldUpdates(bypassCtx, e.client, obj.Type, obj.ID, updates)
}

// executeApproval creates workflow assignments for approval actions
func (e *WorkflowEngine) executeApproval(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	var params wfworkflows.ApprovalActionParams

	if action.Params != nil {
		if err := json.Unmarshal(action.Params, &params); err != nil {
			return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
		}
	}

	required := true
	if params.Required != nil {
		required = *params.Required
	}
	requiredCount := max(params.RequiredCount, 0)
	if !required && requiredCount == 0 {
		requiredCount = 1
	}

	// Compute proposal hash for approval domain tracking (approval-specific)
	var proposedHash string
	if obj != nil {
		domainKey := ""
		if len(params.Fields) > 0 {
			domain, err := workflowgenerated.NewWorkflowDomain(obj.Type, params.Fields)
			if err != nil {
				return fmt.Errorf("%w: %v", wfworkflows.ErrApprovalActionParamsInvalid, err)
			}
			domainKey = domain.Key()
		}
		hash, err := e.proposalManager.ComputeHash(ctx, instance, obj, domainKey)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToComputeProposalHash, err)
		}
		proposedHash = hash
	}

	return e.executeGatedAction(ctx, action, instance, obj, gatedActionConfig{
		ActionType:     enums.WorkflowActionTypeApproval,
		KeyPrefix:      "approval",
		Targets:        params.Targets,
		Required:       required,
		RequiredCount:  requiredCount,
		Label:          params.Label,
		ProposedHash:   proposedHash,
		NoTargetsError: ErrApprovalNoTargets,
	})
}

// executeNotification sends notifications to targets
func (e *WorkflowEngine) executeNotification(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	var params wfworkflows.NotificationActionParams

	if action.Params != nil {
		if err := json.Unmarshal(action.Params, &params); err != nil {
			return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
		}
	}

	if obj == nil {
		return ErrObjectRefMissingID
	}

	if len(params.Targets) == 0 {
		return nil
	}

	channels := params.Channels
	if len(channels) == 0 {
		channels = []enums.Channel{enums.ChannelInApp}
	}

	ownerID, err := wfworkflows.ResolveOwnerID(ctx, instance.OwnerID)
	if err != nil {
		return err
	}

	var rendered *renderedNotificationTemplate
	if params.TemplateID != "" || params.TemplateKey != "" {
		rendered, err = e.renderNotificationTemplate(ctx, instance, obj, action.Key, params, ownerID)
		if err != nil {
			return err
		}
	}

	defaultTitle := lo.CoalesceOrEmpty(params.Title, fmt.Sprintf("Workflow notification (%s)", action.Key))
	defaultBody := lo.CoalesceOrEmpty(params.Body, fmt.Sprintf("Workflow instance %s emitted a notification action (%s).", instance.ID, action.Key))

	var (
		title string
		body  string
		data  map[string]any
		vars  map[string]any
	)

	if rendered != nil {
		title = rendered.Title
		body = rendered.Body
		data = rendered.Data
		vars = rendered.Vars
	} else {
		var err error
		vars, data, err = e.buildNotificationTemplateVars(ctx, instance, obj, action.Key, params.Data)
		if err != nil {
			return err
		}
	}

	if title == "" {
		var err error
		title, err = renderTemplateText(ctx, e.celEvaluator, defaultTitle, vars)
		if err != nil {
			return err
		}
	}

	if body == "" {
		var err error
		body, err = renderTemplateText(ctx, e.celEvaluator, defaultBody, vars)
		if err != nil {
			return err
		}
	}

	templateID := ""
	if rendered != nil && rendered.Template != nil {
		templateID = rendered.Template.ID
	}

	userIDs, err := e.dispatchWorkflowNotifications(ctx, wfworkflows.AllowContext(ctx), obj, params, title, body, data, channels, ownerID, action.Type, action.Key, templateID)
	if err != nil {
		return err
	}

	if rendered != nil {
		if err := e.dispatchNotificationIntegrations(ctx, ownerID, channels, rendered, userIDs); err != nil {
			return err
		}
	}

	return nil
}

// dispatchWorkflowNotifications sends notification payloads to configured channels
func (e *WorkflowEngine) dispatchWorkflowNotifications(ctx context.Context, allowCtx context.Context, obj *wfworkflows.Object, params wfworkflows.NotificationActionParams, title, body string, data map[string]any, channels []enums.Channel, ownerID string, actionType string, actionKey string, templateID string) ([]string, error) {
	seenUsers := make(map[string]struct{})
	resolvedUserIDs := make([]string, 0)

	for _, targetConfig := range params.Targets {
		targetUserIDs, err := e.resolveTargetUsers(ctx, targetConfig, obj, actionType, actionKey)
		if err != nil {
			return nil, fmt.Errorf("%w %s: %w", ErrFailedToResolveNotificationTarget, targetConfig.Type.String(), err)
		}

		for _, userID := range targetUserIDs {
			if _, ok := seenUsers[userID]; ok {
				continue
			}
			seenUsers[userID] = struct{}{}
			resolvedUserIDs = append(resolvedUserIDs, userID)

			notificationData := make(map[string]any, len(data)+1)
			maps.Copy(notificationData, data)
			notificationData["user_id"] = userID

			builder := e.client.Notification.Create().
				SetOwnerID(ownerID).
				SetNotificationType(enums.NotificationTypeUser).
				SetObjectType("workflow.notification").
				SetTitle(title).
				SetBody(body).
				SetData(notificationData).
				SetChannels(channels).
				SetUserID(userID)

			if templateID != "" {
				builder.SetTemplateID(templateID)
			}

			if params.Topic != "" {
				builder.SetTopic(enums.NotificationTopic(params.Topic))
			}

			if err := builder.Exec(allowCtx); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrNotificationCreationFailed, err)
			}
		}
	}

	return resolvedUserIDs, nil
}

// executeWebhook sends a webhook to an external system.
func (e *WorkflowEngine) executeWebhook(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	var params wfworkflows.WebhookActionParams

	if action.Params != nil {
		var payloadCheck map[string]json.RawMessage
		if err := json.Unmarshal(action.Params, &payloadCheck); err != nil {
			return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
		}
		if _, exists := payloadCheck["payload"]; exists {
			return ErrWebhookPayloadUnsupported
		}

		if err := json.Unmarshal(action.Params, &params); err != nil {
			return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
		}
	}

	if params.URL == "" {
		return ErrWebhookURLRequired
	}

	method := params.Method
	if method == "" {
		method = "POST"
	}

	_, basePayload := wfworkflows.BuildWorkflowActionContext(instance, obj, action.Key)

	// Resolve user IDs to display names for human-readable webhook payloads
	allowCtx := wfworkflows.AllowContext(ctx)
	// Get initiator from the object that triggered the workflow (not the service that created the instance)
	initiatorID := wfworkflows.GetObjectUpdatedBy(obj)
	if initiatorID == "" {
		initiatorID = instance.CreatedBy
	}
	initiatorName := wfworkflows.ResolveUserDisplayName(allowCtx, e.client, initiatorID)
	approverName := wfworkflows.ResolveUserDisplayName(allowCtx, e.client, instance.UpdatedBy)

	basePayload["approved_by"] = approverName
	basePayload["initiator"] = initiatorName
	basePayload["object_type"] = obj.Type.String()

	// Enrich with object-specific details when possible using generated helper.
	// This adds fields like ref_code, title, name, status based on schema annotations.
	if err := e.client.EnrichWebhookPayload(allowCtx, obj.Type, obj.ID, basePayload); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToEnrichWebhookPayload, err)
	}

	if strings.TrimSpace(params.PayloadExpr) != "" {
		vars, err := e.buildActionCELVars(ctx, instance, obj)
		if err != nil {
			return err
		}

		exprPayload, err := e.celEvaluator.EvaluateJSONMap(ctx, params.PayloadExpr, vars)
		if err != nil {
			if errors.Is(err, celx.ErrJSONMapExpected) {
				return fmt.Errorf("%w: %w", ErrWebhookPayloadExpressionInvalid, err)
			}

			return fmt.Errorf("%w: %w", ErrWebhookPayloadExpressionFailed, err)
		}

		maps.Copy(basePayload, exprPayload)
	}

	payloadBytes, err := json.Marshal(basePayload)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalPayload, err)
	}

	timeoutMS := params.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = defaultWebhookTimeoutMS
	}

	payloadSum := sha256.Sum256(payloadBytes)
	idempotencyKey := params.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("wf_%s_%s_%s", instance.ID, action.Key, hex.EncodeToString(payloadSum[:]))
	}

	requestOpts := []httpsling.Option{
		httpsling.Method(method),
		httpsling.URL(params.URL),
		httpsling.Body(basePayload),
		httpsling.ContentType(httpsling.ContentTypeJSON),
		httpsling.Accept(httpsling.ContentTypeJSON),
		httpsling.Client(httpclient.Timeout(time.Duration(timeoutMS) * time.Millisecond)),
		httpsling.Header("Idempotency-Key", idempotencyKey),
		httpsling.Header("X-Workflow-Idempotency-Key", idempotencyKey),
	}

	if len(params.Headers) > 0 {
		requestOpts = append(requestOpts, httpsling.HeadersFromMap(params.Headers))
	}

	if params.Secret != "" {
		signature := computeHMACSignature(params.Secret, payloadBytes)
		requestOpts = append(requestOpts, httpsling.Header("X-Workflow-Signature", signature))
	}

	maxRetries := max(params.Retries, 0)
	if maxRetries == 0 {
		maxRetries = defaultWebhookRetries
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = defaultWebhookInitialBackoffMS * time.Millisecond
	policy.MaxInterval = defaultWebhookMaxBackoffSeconds * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
		resp, err := httpsling.ReceiveWithContext(attemptCtx, nil, append(requestOpts, httpsling.Header("X-Workflow-Delivery-Attempt", fmt.Sprintf("%d", attempt+1)))...)
		cancel()

		if err == nil && httpsling.IsSuccess(resp) {
			return nil
		}

		if err == nil && resp != nil && resp.StatusCode >= httpStatusClientErrorMin && resp.StatusCode < httpStatusClientErrorMax {
			return ErrWebhookFailed
		}

		if attempt == maxRetries {
			if err != nil {
				return fmt.Errorf("%w: %w", ErrWebhookFailed, err)
			}

			return ErrWebhookFailed
		}

		wait := policy.NextBackOff()
		if wait == backoff.Stop {
			wait = defaultWebhookFallbackBackoffMS * time.Millisecond
		}
		time.Sleep(wait)
	}

	return nil
}

// executeReview creates workflow assignments for review actions
func (e *WorkflowEngine) executeReview(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	var params wfworkflows.ReviewActionParams

	if action.Params != nil {
		if err := json.Unmarshal(action.Params, &params); err != nil {
			return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
		}
	}

	required := true
	if params.Required != nil {
		required = *params.Required
	}
	requiredCount := max(params.RequiredCount, 0)
	if !required && requiredCount == 0 {
		requiredCount = 1
	}

	return e.executeGatedAction(ctx, action, instance, obj, gatedActionConfig{
		ActionType:     enums.WorkflowActionTypeReview,
		KeyPrefix:      "review",
		Role:           "REVIEWER",
		Targets:        params.Targets,
		Required:       required,
		RequiredCount:  requiredCount,
		Label:          params.Label,
		NoTargetsError: ErrReviewNoTargets,
	})
}
