package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	wfworkflows "github.com/theopenlane/core/internal/workflows"
)

// executeSendEmail composes and queues an email from a notification/email template reference
func (e *WorkflowEngine) executeSendEmail(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	var params wfworkflows.SendEmailActionParams

	if len(action.Params) == 0 {
		return ErrSendEmailTemplateRequired
	}

	if err := json.Unmarshal(action.Params, &params); err != nil {
		return fmt.Errorf("%w: %w", ErrUnmarshalParams, err)
	}

	templateID := params.TemplateID
	templateKey := params.TemplateKey
	if templateID != "" && templateKey != "" {
		return ErrSendEmailTemplateReferenceConflict
	}
	if templateID == "" && templateKey == "" {
		return ErrSendEmailTemplateRequired
	}

	if obj == nil {
		return ErrObjectRefMissingID
	}
	if instance == nil {
		return ErrInstanceNotFound
	}

	ownerID, err := wfworkflows.ResolveOwnerID(ctx, instance.OwnerID)
	if err != nil {
		return err
	}

	def, err := e.client.WorkflowDefinition.Query().
		Where(workflowdefinition.IDEQ(instance.WorkflowDefinitionID)).
		Only(wfworkflows.AllowContext(ctx))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToLoadWorkflowDefinition, err)
	}

	// system-owned workflow definitions may reference system-owned templates;
	// org-owned workflow definitions are restricted to owner-scoped templates only
	ownerOnly := !def.SystemOwned

	vars, data, err := e.buildNotificationTemplateVars(ctx, instance, obj, action.Key, params.Data)
	if err != nil {
		return err
	}

	recipients, err := e.resolveSendEmailRecipients(ctx, obj, action, params, vars)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return ErrSendEmailNoRecipients
	}

	emailClient, err := email.ResolveClient(ctx, ownerID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendEmailTemplateComposeFailed, err)
	}

	fromAddress, err := e.resolveSendEmailFromAddress(ctx, params.From, emailClient.Config.FromEmail, vars)
	if err != nil {
		return err
	}
	if fromAddress == "" {
		return ErrSendEmailSenderMissing
	}

	replyTo := ""
	if params.ReplyTo != "" {
		replyTo, err = renderTemplateText(ctx, e.celEvaluator, params.ReplyTo, vars)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, err)
		}
		replyTo = strings.TrimSpace(replyTo)
		if replyTo != "" {
			parsed, parseErr := mail.ParseAddress(replyTo)
			if parseErr != nil {
				return fmt.Errorf("%w: %w", ErrSendEmailRecipientTemplateInvalid, parseErr)
			}
			replyTo = parsed.Address
		}
	}

	allowCtx := wfworkflows.AllowContext(ctx)
	_, err = email.ComposeAndQueueFromNotificationTemplate(allowCtx, e.client, emailClient, email.ComposeRequest{
		OwnerID: ownerID,
		Template: email.TemplateRef{
			ID:  templateID,
			Key: templateKey,
		},
		To:        recipients,
		From:      fromAddress,
		ReplyTo:   replyTo,
		Data:      data,
		Headers:   params.Headers,
		OwnerOnly: ownerOnly,
	}, e.client.Job)
	if err != nil {
		switch {
		case errors.Is(err, email.ErrJobClientRequired):
			return ErrSendEmailJobClientRequired
		case errors.Is(err, email.ErrQueueInsertFailed):
			return fmt.Errorf("%w: %w", ErrSendEmailQueueInsertFailed, err)
		default:
			return fmt.Errorf("%w: %w", ErrSendEmailTemplateComposeFailed, err)
		}
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
