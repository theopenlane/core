package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
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
	wfworkflows "github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
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
	case enums.WorkflowActionTypeNotification:
		return e.executeNotification(ctx, action, instance, obj)
	case enums.WorkflowActionTypeFieldUpdate:
		return e.executeFieldUpdate(ctx, action, obj)
	case enums.WorkflowActionTypeWebhook:
		return e.executeWebhook(ctx, action, instance, obj)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidActionType, action.Type)
	}
}

// resolveTargetUsers resolves target user IDs and logs warnings if no users are found
func (e *WorkflowEngine) resolveTargetUsers(ctx context.Context, target wfworkflows.TargetConfig, obj *wfworkflows.Object, actionType string, actionKey string) ([]string, error) {
	allowCtx := wfworkflows.AllowContext(ctx)
	userIDs, err := e.ResolveTargets(allowCtx, target, obj)
	if err != nil {
		return nil, err
	}

	normalized := wfworkflows.NormalizeStrings(userIDs)
	if len(normalized) == 0 {
		observability.WarnEngine(ctx, observability.OpExecuteAction, actionType, observability.ActionFields(actionKey, observability.Fields{
			workflowassignmenttarget.FieldTargetType:  target.Type.String(),
			workflowassignmenttarget.FieldResolverKey: target.ResolverKey,
		}), nil)
	}

	return normalized, nil
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

	// Use both workflow bypass and privacy bypass for internal field updates
	bypassCtx := wfworkflows.AllowBypassContext(ctx)
	return wfworkflows.ApplyObjectFieldUpdates(bypassCtx, e.client, obj.Type, obj.ID, params.Updates)
}

// executeApproval creates workflow assignments for approval actions
func (e *WorkflowEngine) executeApproval(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	// Use allow context for internal workflow operations
	allowCtx := wfworkflows.AllowContext(ctx)

	// Parse approval params
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
		// Optional approvals without an explicit quorum should still allow forward progress
		requiredCount = 1
	}

	if len(params.Targets) == 0 {
		observability.WarnEngine(ctx, observability.OpExecuteAction, action.Type, observability.ActionFields(action.Key, nil), nil)
		return ErrApprovalNoTargets
	}

	if obj == nil {
		return ErrObjectRefMissingID
	}

	ownerID := instance.OwnerID
	if ownerID == "" {
		extracted, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return err
		}

		ownerID = extracted
	}

	// Capture a stable hash of the proposed changes so approvals are tied to the payload approvers saw
	domainKey := ""
	if len(params.Fields) > 0 {
		domain, err := workflowgenerated.NewWorkflowDomain(obj.Type, params.Fields)
		if err != nil {
			return fmt.Errorf("%w: %v", wfworkflows.ErrApprovalActionParamsInvalid, err)
		}
		domainKey = domain.Key()
	}
	proposedHash, err := e.proposalManager.ComputeHash(ctx, instance, obj, domainKey)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToComputeProposalHash, err)
	}

	actionIndex := actionIndexForKey(instance.DefinitionSnapshot.Actions, action.Key)
	assignmentIDs := make([]string, 0)
	targetUserIDs := make([]string, 0)
	seenAssignmentIDs := make(map[string]struct{})
	seenTargetUserIDs := make(map[string]struct{})
	seenAssignments := make(map[string]struct{})

	// Resolve each target and create one assignment per resolved user
	for _, targetConfig := range params.Targets {
		userIDs, err := e.resolveTargetUsers(ctx, targetConfig, obj, action.Type, action.Key)
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrFailedToResolveTarget, targetConfig.Type.String(), err)
		}
		if len(userIDs) == 0 {
			continue
		}

		for _, userID := range userIDs {
			assignmentKey := fmt.Sprintf("approval_%s_%s", action.Key, userID)
			if _, ok := seenAssignments[assignmentKey]; ok {
				continue
			}
			seenAssignments[assignmentKey] = struct{}{}

			approvalMeta := models.WorkflowAssignmentApproval{
				ActionKey:     action.Key,
				Required:      required,
				Label:         params.Label,
				ProposedHash:  proposedHash,
				RequiredCount: requiredCount,
			}

			assignmentCreate := e.client.WorkflowAssignment.
				Create().
				SetWorkflowInstanceID(instance.ID).
				SetAssignmentKey(assignmentKey).
				SetStatus(enums.WorkflowAssignmentStatusPending).
				SetRequired(required).
				SetLabel(params.Label).
				SetApprovalMetadata(approvalMeta)
			assignmentCreate.SetOwnerID(ownerID)

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
				e.emitAssignmentCreated(ctx, instance, obj, assignment.ID, userID)
			}
		}
	}

	if len(assignmentIDs) > 0 {
		details := assignmentCreatedDetailsForApproval(action, actionIndex, obj, assignmentIDs, targetUserIDs, required, requiredCount, params.Label)
		e.recordAssignmentsCreated(ctx, instance, details)
	}

	if len(assignmentIDs) == 0 {
		return ErrApprovalNoTargets
	}

	return nil
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

	title := lo.CoalesceOrEmpty(params.Title, fmt.Sprintf("Workflow notification (%s)", action.Key))
	body := lo.CoalesceOrEmpty(params.Body, fmt.Sprintf("Workflow instance %s emitted a notification action (%s).", instance.ID, action.Key))

	replacements, baseData := wfworkflows.BuildWorkflowActionContext(instance, obj, action.Key)

	title = replaceTokens(title, replacements)
	body = replaceTokens(body, replacements)

	data := applyStringTemplates(params.Data, replacements)
	for k, v := range baseData {
		data[k] = v
	}

	ownerID, err := wfworkflows.ResolveOwnerID(ctx, instance.OwnerID)
	if err != nil {
		return err
	}

	return e.dispatchWorkflowNotifications(ctx, wfworkflows.AllowContext(ctx), obj, params, title, body, data, channels, ownerID, action.Type, action.Key)
}

// dispatchWorkflowNotifications sends notification payloads to configured channels
func (e *WorkflowEngine) dispatchWorkflowNotifications(ctx context.Context, allowCtx context.Context, obj *wfworkflows.Object, params wfworkflows.NotificationActionParams, title, body string, data map[string]any, channels []enums.Channel, ownerID string, actionType string, actionKey string) error {
	seenUsers := make(map[string]struct{})

	for _, targetConfig := range params.Targets {
		userIDs, err := e.resolveTargetUsers(ctx, targetConfig, obj, actionType, actionKey)
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrFailedToResolveNotificationTarget, targetConfig.Type.String(), err)
		}

		for _, userID := range userIDs {
			if _, ok := seenUsers[userID]; ok {
				continue
			}
			seenUsers[userID] = struct{}{}

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

			if params.Topic != "" {
				builder.SetTopic(enums.NotificationTopic(params.Topic))
			}

			if err := builder.Exec(allowCtx); err != nil {
				return fmt.Errorf("%w: %w", ErrNotificationCreationFailed, err)
			}
		}
	}

	return nil
}

// executeWebhook sends a webhook to an external system.
func (e *WorkflowEngine) executeWebhook(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	var params wfworkflows.WebhookActionParams

	if action.Params != nil {
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

	replacements, basePayload := wfworkflows.BuildWorkflowActionContext(instance, obj, action.Key)

	// Resolve user IDs to display names for human-readable webhook payloads
	allowCtx := wfworkflows.AllowContext(ctx)
	initiatorName := wfworkflows.ResolveUserDisplayName(allowCtx, e.client, instance.CreatedBy)
	approverName := wfworkflows.ResolveUserDisplayName(allowCtx, e.client, instance.UpdatedBy)

	basePayload["approved_by"] = approverName
	basePayload["initiator"] = initiatorName
	basePayload["object_type"] = obj.Type.String()

	replacements["initiator"] = initiatorName
	replacements["approved_by"] = approverName
	replacements["approved_at"] = time.Now().UTC().Format(time.RFC3339)

	// Enrich with object-specific details when possible using generated helper.
	// This adds fields like ref_code, title, name, status based on schema annotations.
	if err := e.client.EnrichWebhookPayload(allowCtx, obj.Type, obj.ID, basePayload); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToEnrichWebhookPayload, err)
	}

	// Copy enriched fields to replacements for template substitution
	for key, value := range basePayload {
		if strVal, ok := value.(string); ok {
			if _, exists := replacements[key]; !exists {
				replacements[key] = strVal
			}
		}
	}

	templatedPayload := applyStringTemplates(params.Payload, replacements)

	// Merge caller-provided payload onto base payload after templating.
	maps.Copy(basePayload, templatedPayload)

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
