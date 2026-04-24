package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/user"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	wfworkflows "github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// executeSendEmail dispatches a workflow send_email action through the integration
// framework via QueueIntegrationOperation. It resolves the notification template,
// recipients, and sender address, then routes through either a linked integration
// installation or the runtime email definition
func (e *WorkflowEngine) executeSendEmail(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}

	var params wfworkflows.SendEmailActionParams
	if err := jsonx.RoundTrip(action.Params, &params); err != nil {
		return errors.Join(ErrUnmarshalParams, err)
	}

	// validate template reference
	hasID := params.TemplateID != ""
	hasKey := params.TemplateKey != ""

	switch {
	case hasID && hasKey:
		return ErrSendEmailTemplateReferenceConflict
	case !hasID && !hasKey:
		return ErrSendEmailTemplateRequired
	}

	ownerID := instance.OwnerID
	if ownerID == "" {
		caller, callerOk := auth.CallerFromContext(ctx)
		if !callerOk || caller == nil || caller.OrganizationID == "" {
			return ErrIntegrationOwnerRequired
		}

		ownerID = caller.OrganizationID
	}

	// render the notification template
	rendered, err := e.renderNotificationTemplate(ctx, instance, obj, action.Key, wfworkflows.NotificationActionParams{
		TemplateID:  params.TemplateID,
		TemplateKey: params.TemplateKey,
		Data:        params.Data,
	}, ownerID)
	if err != nil {
		return err
	}
	if rendered == nil || rendered.Template == nil {
		return ErrNotificationTemplateNotFound
	}

	// resolve recipients
	recipients, err := e.resolveSendEmailRecipients(ctx, obj, action, params, rendered.Vars)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return ErrSendEmailNoRecipients
	}

	// resolve from address: explicit param → runtime config default
	fromAddress, err := e.resolveSendEmailFromAddress(ctx, params.From, "", rendered.Vars)
	if err != nil {
		return err
	}

	// build the operation config from the rendered template
	config := buildRenderedTemplateConfig(rendered)
	config["to"] = recipients
	if fromAddress != "" {
		config["from"] = fromAddress
	}
	if params.ReplyTo != "" {
		config["reply_to"] = params.ReplyTo
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalPayload, err)
	}

	// dispatch through the notification template's linked integration when available
	if rendered.Template.IntegrationID != "" {
		meta := &types.WorkflowMeta{
			InstanceID:  instance.ID,
			ActionKey:   action.Key,
			ActionIndex: actionIndexForKey(instance.DefinitionSnapshot.Actions, action.Key),
		}
		if obj != nil {
			meta.ObjectID = obj.ID
			meta.ObjectType = obj.Type
		}

		_, queueErr := e.QueueIntegrationOperation(ctx, IntegrationQueueRequest{
			OrgID:          ownerID,
			InstallationID: rendered.Template.IntegrationID,
			Operation:      emaildef.SendEmailOp.Name(), // reuse the send operation
			Config:         configBytes,
			RunType:        enums.IntegrationRunTypeEvent,
			Workflow:       meta,
		})

		if queueErr != nil {
			return queueErr
		}

		markIntegrationQueued(ctx)

		return nil
	}

	if _, err := e.integrationRuntime.Dispatch(ctx, operations.DispatchRequest{
		DefinitionID: emaildef.DefinitionID.ID(),
		Operation:    emaildef.SendEmailOp.Name(),
		Config:       configBytes,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("send_email runtime execution failed")
		return err
	}

	return nil
}

// resolveSendEmailFromAddress resolves a sender address from explicit params or falls back to defaultFrom
func (e *WorkflowEngine) resolveSendEmailFromAddress(ctx context.Context, rawFrom string, defaultFrom string, vars map[string]any) (string, error) {
	if rawFrom != "" {
		rendered, err := renderTemplateText(ctx, e.celEvaluator, rawFrom, vars)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, err)
		}

		parsed, parseErr := mail.ParseAddress(strings.TrimSpace(rendered))
		if parseErr != nil {
			return "", fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, parseErr)
		}

		return parsed.Address, nil
	}

	if defaultFrom != "" {
		parsed, parseErr := mail.ParseAddress(defaultFrom)
		if parseErr != nil {
			return "", fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, parseErr)
		}

		return parsed.Address, nil
	}

	return "", nil
}

// resolveSendEmailRecipients resolves explicit and target-based recipients for send_email actions
func (e *WorkflowEngine) resolveSendEmailRecipients(ctx context.Context, obj *wfworkflows.Object, action models.WorkflowAction, params wfworkflows.SendEmailActionParams, vars map[string]any) ([]string, error) {
	recipients := make([]string, 0, len(params.To))

	for _, raw := range params.To {
		rendered, err := renderTemplateString(ctx, e.celEvaluator, raw, vars)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, err)
		}

		flattened, err := flattenSendEmailRecipients(rendered)
		if err != nil {
			return nil, err
		}

		recipients = append(recipients, flattened...)
	}

	targetUserIDs := make([]string, 0)
	for _, targetConfig := range params.Targets {
		userIDs, err := e.resolveTargetUsers(ctx, targetConfig, obj, action.Type, action.Key)
		if err != nil {
			return nil, fmt.Errorf("%w %s: %w", ErrFailedToResolveNotificationTarget, targetConfig.Type.String(), err)
		}
		targetUserIDs = append(targetUserIDs, userIDs...)
	}

	targetUserIDs = lo.Uniq(targetUserIDs)
	if len(targetUserIDs) > 0 {
		userEmails, err := e.resolveSendEmailTargetUserEmails(wfworkflows.AllowContext(ctx), targetUserIDs)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, userEmails...)
	}

	normalized, err := normalizeSendEmailAddresses(recipients)
	if err != nil {
		return nil, err
	}

	return normalized, nil
}

// resolveSendEmailTargetUserEmails loads users by ID and returns unique email addresses
func (e *WorkflowEngine) resolveSendEmailTargetUserEmails(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	users, err := e.client.User.Query().Where(user.IDIn(userIDs...)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSendEmailUserLookupFailed, err)
	}

	emails := make([]string, 0, len(users))
	for _, item := range users {
		if item.Email != "" {
			emails = append(emails, item.Email)
		}
	}

	return lo.Uniq(emails), nil
}

// flattenSendEmailRecipients normalizes rendered recipient values into a string slice
func flattenSendEmailRecipients(rendered any) ([]string, error) {
	switch typed := rendered.(type) {
	case nil:
		return nil, nil
	case string:
		value := strings.TrimSpace(typed)
		if value == "" {
			return nil, nil
		}

		return []string{value}, nil
	case []string:
		return typed, nil
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			flattened, err := flattenSendEmailRecipients(item)
			if err != nil {
				return nil, err
			}
			values = append(values, flattened...)
		}

		return values, nil
	default:
		value := strings.TrimSpace(formatTemplateValue(typed))
		if value == "" {
			return nil, nil
		}

		return []string{value}, nil
	}
}

// normalizeSendEmailAddresses validates and deduplicates email addresses
func normalizeSendEmailAddresses(addresses []string) ([]string, error) {
	normalized := make([]string, 0, len(addresses))

	for _, raw := range addresses {
		if raw == "" {
			continue
		}

		parsedList, err := mail.ParseAddressList(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, err)
		}

		for _, parsed := range parsedList {
			normalized = append(normalized, parsed.Address)
		}
	}

	return lo.Uniq(normalized), nil
}
