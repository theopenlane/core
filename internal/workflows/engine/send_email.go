package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/user"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	wfworkflows "github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// executeSendEmail dispatches a workflow send_email action through the integration
// framework. It resolves the email template, recipients, and routes through either
// a linked integration installation or the runtime email definition
func (e *WorkflowEngine) executeSendEmail(ctx context.Context, action models.WorkflowAction, instance *generated.WorkflowInstance, obj *wfworkflows.Object) error {
	if e.integrationRuntime == nil {
		if rt := intruntime.FromClient(ctx, e.client); rt != nil {
			e.integrationRuntime = rt
		}
	}

	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}

	var params wfworkflows.SendEmailActionParams
	if err := jsonx.RoundTrip(action.Params, &params); err != nil {
		return errors.Join(ErrUnmarshalParams, err)
	}

	hasID := params.EmailTemplateID != ""
	hasKey := params.EmailTemplateKey != ""

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

	emailTpl, err := e.loadSendEmailTemplate(ctx, ownerID, params.EmailTemplateID, params.EmailTemplateKey)
	if err != nil {
		return err
	}

	recipients, err := e.resolveSendEmailRecipients(ctx, obj, action, params)
	if err != nil {
		return err
	}

	if len(recipients) == 0 {
		return ErrSendEmailNoRecipients
	}

	for _, recipient := range recipients {
		if err := e.dispatchSendEmail(ctx, instance, action, obj, emailTpl, ownerID, recipient, params); err != nil {
			return err
		}
	}

	return nil
}

// loadSendEmailTemplate resolves an active email template by ID or key for the given owner
func (e *WorkflowEngine) loadSendEmailTemplate(ctx context.Context, ownerID, templateID, templateKey string) (*generated.EmailTemplate, error) {
	allowCtx := wfworkflows.AllowContext(ctx)

	query := e.client.EmailTemplate.Query().Where(
		emailtemplate.ActiveEQ(true),
		emailtemplate.OwnerIDEQ(ownerID),
	)

	switch {
	case templateID != "":
		query = query.Where(emailtemplate.IDEQ(templateID))
	case templateKey != "":
		query = query.Where(emailtemplate.KeyEQ(templateKey))
	}

	tpl, err := query.Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrSendEmailTemplateNotFound
		}

		return nil, fmt.Errorf("%w: %w", ErrSendEmailTemplateNotFound, err)
	}

	return tpl, nil
}

// resolveSendEmailRecipients resolves explicit and target-based recipients for send_email actions
func (e *WorkflowEngine) resolveSendEmailRecipients(ctx context.Context, obj *wfworkflows.Object, action models.WorkflowAction, params wfworkflows.SendEmailActionParams) ([]string, error) {
	recipients := make([]string, 0, len(params.To))

	for _, raw := range params.To {
		addr, err := mail.ParseAddress(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrSendEmailRecipientInvalid, raw)
		}

		recipients = append(recipients, addr.Address)
	}

	targetUserIDs := make([]string, 0)

	for _, targetConfig := range params.Targets {
		userIDs, err := e.resolveTargetUsers(ctx, targetConfig, obj, action.Type, action.Key)
		if err != nil {
			return nil, fmt.Errorf("%w %s: %w", ErrFailedToResolveNotificationTarget, targetConfig.Type.String(), err)
		}

		targetUserIDs = append(targetUserIDs, userIDs...)
	}

	if len(targetUserIDs) > 0 {
		userEmails, err := e.resolveTargetUserEmails(wfworkflows.AllowContext(ctx), lo.Uniq(targetUserIDs))
		if err != nil {
			return nil, err
		}

		recipients = append(recipients, userEmails...)
	}

	return lo.Uniq(recipients), nil
}

// resolveTargetUserEmails loads users by ID and returns unique email addresses
func (e *WorkflowEngine) resolveTargetUserEmails(ctx context.Context, userIDs []string) ([]string, error) {
	users, err := e.client.User.Query().Where(user.IDIn(userIDs...)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSendEmailUserLookupFailed, err)
	}

	return lo.FilterMap(users, func(u *generated.User, _ int) (string, bool) {
		return u.Email, u.Email != ""
	}), nil
}

// dispatchSendEmail marshals a SendEmailRequest and routes through either the
// template's linked integration or the runtime email definition
func (e *WorkflowEngine) dispatchSendEmail(ctx context.Context, instance *generated.WorkflowInstance, action models.WorkflowAction, obj *wfworkflows.Object, emailTpl *generated.EmailTemplate, ownerID, recipient string, params wfworkflows.SendEmailActionParams) error {
	configBytes, err := json.Marshal(emaildef.SendEmailRequest{
		TemplateID: emailTpl.ID,
		OwnerID:    ownerID,
		To:         recipient,
		From:       params.From,
		ReplyTo:    params.ReplyTo,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalPayload, err)
	}

	if emailTpl.IntegrationID != "" {
		meta := &types.WorkflowMeta{
			InstanceID:  instance.ID,
			ActionKey:   action.Key,
			ActionIndex: actionIndexForKey(instance.DefinitionSnapshot.Actions, action.Key),
		}

		if obj != nil {
			meta.ObjectID = obj.ID
			meta.ObjectType = obj.Type
		}

		_, err := e.QueueIntegrationOperation(ctx, IntegrationQueueRequest{
			OrgID:          ownerID,
			InstallationID: emailTpl.IntegrationID,
			Operation:      emaildef.SendEmailOp.Name(),
			Config:         configBytes,
			RunType:        enums.IntegrationRunTypeEvent,
			Workflow:       meta,
		})
		if err != nil {
			return err
		}

		markIntegrationQueued(ctx)

		return nil
	}

	if _, err := e.integrationRuntime.Dispatch(ctx, operations.DispatchRequest{
		DefinitionID: emaildef.DefinitionID.ID(),
		OwnerID:      ownerID,
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
