package engine

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/samber/lo"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/notificationpreference"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	teamsprovider "github.com/theopenlane/core/internal/integrations/providers/microsoftteams"
	slackprovider "github.com/theopenlane/core/internal/integrations/providers/slack"
	"github.com/theopenlane/core/internal/workflows"
)

// teamsDestinationParts is the expected number of parts when splitting a Teams destination
const teamsDestinationParts = 2

// renderedNotificationTemplate captures a rendered notification template snapshot
type renderedNotificationTemplate struct {
	// Template holds the source template record
	Template *generated.NotificationTemplate
	// Title holds the rendered title text
	Title string
	// Subject holds the rendered subject text
	Subject string
	// Body holds the rendered body text
	Body string
	// Blocks holds rendered structured blocks
	Blocks any
	// Data holds the merged template data payload
	Data map[string]any
	// Vars holds the CEL variable map used for rendering
	Vars map[string]any
}

// renderNotificationTemplate resolves and renders a notification template
func (e *WorkflowEngine) renderNotificationTemplate(ctx context.Context, instance *generated.WorkflowInstance, obj *workflows.Object, actionKey string, params workflows.NotificationActionParams, ownerID string) (*renderedNotificationTemplate, error) {
	templateID := params.TemplateID
	templateKey := params.TemplateKey
	if templateID == "" && templateKey == "" {
		return nil, nil
	}
	if templateID != "" && templateKey != "" {
		return nil, ErrNotificationTemplateReferenceConflict
	}

	template, err := e.loadNotificationTemplate(ctx, ownerID, templateID, templateKey)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrNotificationTemplateNotFound
	}

	vars, data, err := e.buildNotificationTemplateVars(ctx, instance, obj, actionKey, params.Data)
	if err != nil {
		return nil, err
	}
	if err := validateNotificationTemplateData(template, data); err != nil {
		return nil, err
	}

	var blocks any
	if template.Blocks != nil {
		blocks, err = renderTemplateValue(ctx, e.celEvaluator, template.Blocks, vars)
		if err != nil {
			return nil, err
		}
	}

	title, err := renderTemplateText(ctx, e.celEvaluator, template.TitleTemplate, vars)
	if err != nil {
		return nil, err
	}
	body, err := renderTemplateText(ctx, e.celEvaluator, template.BodyTemplate, vars)
	if err != nil {
		return nil, err
	}
	subject, err := renderTemplateText(ctx, e.celEvaluator, template.SubjectTemplate, vars)
	if err != nil {
		return nil, err
	}

	return &renderedNotificationTemplate{
		Template: template,
		Title:    title,
		Subject:  subject,
		Body:     body,
		Blocks:   blocks,
		Data:     data,
		Vars:     vars,
	}, nil
}

// buildNotificationTemplateVars builds CEL vars and merged data for template rendering
func (e *WorkflowEngine) buildNotificationTemplateVars(ctx context.Context, instance *generated.WorkflowInstance, obj *workflows.Object, actionKey string, paramsData map[string]any) (map[string]any, map[string]any, error) {
	vars, err := e.buildActionCELVars(ctx, instance, obj)
	if err != nil {
		return nil, nil, err
	}

	_, baseData := workflows.BuildWorkflowActionContext(instance, obj, actionKey)
	vars = maps.Clone(vars)
	maps.Copy(vars, baseData)

	data := map[string]any{}
	if paramsData != nil {
		rendered, err := renderTemplateValue(ctx, e.celEvaluator, paramsData, vars)
		if err != nil {
			return nil, nil, err
		}

		if renderedMap, ok := rendered.(map[string]any); ok {
			data = renderedMap
		}
	}

	maps.Copy(data, baseData)
	vars["data"] = data
	maps.Copy(vars, data)

	return vars, data, nil
}

// validateNotificationTemplateData validates template data against jsonschema
func validateNotificationTemplateData(template *generated.NotificationTemplate, data map[string]any) error {
	if template == nil || template.Jsonconfig == nil || len(template.Jsonconfig) == 0 {
		return nil
	}

	schemaLoader := gojsonschema.NewGoLoader(template.Jsonconfig)
	documentLoader := gojsonschema.NewGoLoader(data)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}

	if len(result.Errors()) > 0 {
		return ErrNotificationTemplateDataInvalid
	}

	return nil
}

// loadNotificationTemplate loads an active notification template by id or key
func (e *WorkflowEngine) loadNotificationTemplate(ctx context.Context, ownerID string, templateID string, templateKey string) (*generated.NotificationTemplate, error) {
	if ownerID == "" {
		return nil, ErrMissingRequiredField
	}

	allowCtx := workflows.AllowContext(ctx)
	query := e.client.NotificationTemplate.Query().
		Where(
			notificationtemplate.ActiveEQ(true),
			notificationtemplate.Or(
				notificationtemplate.OwnerIDEQ(ownerID),
				notificationtemplate.SystemOwnedEQ(true),
			),
		)

	if templateID != "" {
		template, err := query.Where(notificationtemplate.IDEQ(templateID)).Only(allowCtx)
		if generated.IsNotFound(err) {
			return nil, ErrNotificationTemplateNotFound
		}
		return template, err
	}
	if templateKey != "" {
		templates, err := query.Where(notificationtemplate.KeyEQ(templateKey)).All(allowCtx)
		if err != nil {
			return nil, err
		}

		if len(templates) == 0 {
			return nil, ErrNotificationTemplateNotFound
		}

		if found, ok := lo.Find(templates, func(t *generated.NotificationTemplate) bool {
			return t.OwnerID == ownerID
		}); ok {
			return found, nil
		}

		return templates[0], nil
	}

	return nil, nil
}

// dispatchNotificationIntegrations executes integration operations for notification channels
func (e *WorkflowEngine) dispatchNotificationIntegrations(ctx context.Context, ownerID string, channels []enums.Channel, rendered *renderedNotificationTemplate, userIDs []string) error {
	if rendered == nil || len(userIDs) == 0 {
		return nil
	}
	if e.integrationOperations == nil {
		return ErrIntegrationOperationsRequired
	}

	for _, channel := range channels {
		if channel == enums.ChannelInApp {
			continue
		}

		if err := e.dispatchChannelNotifications(ctx, ownerID, channel, rendered, userIDs); err != nil {
			return err
		}
	}

	return nil
}

// dispatchChannelNotifications sends notifications to all users for a specific channel
func (e *WorkflowEngine) dispatchChannelNotifications(ctx context.Context, ownerID string, channel enums.Channel, rendered *renderedNotificationTemplate, userIDs []string) error {
	operationName, err := operationNameForChannel(channel)
	if err != nil {
		return err
	}

	integrationRecord, provider, err := e.resolveNotificationIntegration(ctx, ownerID, rendered.Template, channel)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := e.dispatchUserNotification(ctx, ownerID, userID, channel, operationName, provider, integrationRecord, rendered); err != nil {
			return err
		}
	}

	return nil
}

// dispatchUserNotification sends a notification to a single user via integration
func (e *WorkflowEngine) dispatchUserNotification(ctx context.Context, ownerID, userID string, channel enums.Channel, operationName types.OperationName, provider types.ProviderType, integrationRecord *generated.Integration, rendered *renderedNotificationTemplate) error {
	preference, err := e.loadNotificationPreference(ctx, ownerID, userID, channel)
	if err != nil {
		return err
	}
	if preference == nil {
		return nil
	}

	config, err := buildNotificationOperationConfig(channel, preference, rendered)
	if err != nil {
		return err
	}

	if integrationRecord != nil {
		merged, err := operations.ResolveOperationConfig(&integrationRecord.Config, string(operationName), config)
		if err != nil {
			return err
		}
		if merged != nil {
			config = merged
		}
	}

	_, err = e.integrationOperations.Run(ctx, types.OperationRequest{
		OrgID:    ownerID,
		Provider: provider,
		Name:     operationName,
		Config:   config,
	})

	return err
}

// resolveNotificationIntegration resolves the integration record and provider for a channel
func (e *WorkflowEngine) resolveNotificationIntegration(ctx context.Context, ownerID string, template *generated.NotificationTemplate, channel enums.Channel) (*generated.Integration, types.ProviderType, error) {
	if template != nil {
		templateChannel := template.Channel
		if templateChannel != "" && templateChannel != enums.ChannelInvalid && channel != templateChannel {
			return nil, types.ProviderUnknown, ErrNotificationTemplateChannelMismatch
		}
	}

	if template != nil {
		integrationID := template.IntegrationID
		if integrationID != "" {
			allowCtx := workflows.AllowContext(ctx)
			record, err := e.client.Integration.Query().
				Where(
					integration.IDEQ(integrationID),
					integration.OwnerIDEQ(ownerID),
				).
				Only(allowCtx)
			if err != nil {
				return nil, types.ProviderUnknown, err
			}
			provider := types.ProviderTypeFromString(record.Kind)
			if provider == types.ProviderUnknown {
				return record, provider, ErrIntegrationProviderUnknown
			}
			return record, provider, nil
		}
	}

	switch channel {
	case enums.ChannelSlack:
		return nil, slackprovider.TypeSlack, nil
	case enums.ChannelTeams:
		return nil, teamsprovider.TypeMicrosoftTeams, nil
	default:
		return nil, types.ProviderUnknown, ErrNotificationChannelUnsupported
	}
}

// loadNotificationPreference loads an enabled notification preference for a user and channel
func (e *WorkflowEngine) loadNotificationPreference(ctx context.Context, ownerID string, userID string, channel enums.Channel) (*generated.NotificationPreference, error) {
	if userID == "" {
		return nil, nil
	}

	allowCtx := workflows.AllowContext(ctx)
	preference, err := e.client.NotificationPreference.Query().
		Where(
			notificationpreference.OwnerIDEQ(ownerID),
			notificationpreference.UserIDEQ(userID),
			notificationpreference.ChannelEQ(channel),
		).
		Only(allowCtx)
	if generated.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !preference.Enabled {
		return nil, nil
	}

	disabledStatuses := []enums.NotificationChannelStatus{
		enums.NotificationChannelStatusDisabled,
		enums.NotificationChannelStatusError,
		enums.NotificationChannelStatusPending,
	}
	if lo.Contains(disabledStatuses, preference.Status) {
		return nil, nil
	}

	return preference, nil
}

// buildNotificationOperationConfig builds the provider operation config for a channel
func buildNotificationOperationConfig(channel enums.Channel, preference *generated.NotificationPreference, rendered *renderedNotificationTemplate) (map[string]any, error) {
	var config map[string]any
	if preference != nil && len(preference.Config) > 0 {
		config = maps.Clone(preference.Config)
	}
	if config == nil {
		config = map[string]any{}
	}

	switch channel {
	case enums.ChannelSlack:
		if preference != nil && preference.Destination != "" {
			config["channel"] = preference.Destination
		}
		if rendered != nil {
			text := lo.CoalesceOrEmpty(rendered.Body, rendered.Title)
			if text != "" {
				config["text"] = text
			}
			if rendered.Blocks != nil {
				config["blocks"] = rendered.Blocks
			}
		}
		return config, nil
	case enums.ChannelTeams:
		teamID, channelID := resolveTeamsDestination(preference, config)
		if teamID != "" {
			config["team_id"] = teamID
		}
		if channelID != "" {
			config["channel_id"] = channelID
		}
		if rendered != nil {
			body := lo.CoalesceOrEmpty(rendered.Body, rendered.Title)
			if body != "" {
				config["body"] = body
			}
			if rendered.Subject != "" {
				config["subject"] = rendered.Subject
			}
			if rendered.Template != nil {
				if _, ok := config["body_format"]; !ok {
					switch rendered.Template.Format {
					case enums.NotificationTemplateFormatHTML:
						config["body_format"] = "html"
					default:
						config["body_format"] = "text"
					}
				}
			}
		}
		return config, nil
	default:
		return nil, ErrNotificationChannelUnsupported
	}
}

// resolveTeamsDestination resolves Teams team and channel identifiers
func resolveTeamsDestination(preference *generated.NotificationPreference, config map[string]any) (string, string) {
	teamID := readConfigString(config, "team_id")
	channelID := readConfigString(config, "channel_id")

	if preference == nil {
		return teamID, channelID
	}

	destination := preference.Destination
	if destination == "" {
		return teamID, channelID
	}

	if teamID != "" && channelID != "" {
		return teamID, channelID
	}

	parsedTeam, parsedChannel := splitTeamsDestination(destination)
	if teamID == "" {
		teamID = parsedTeam
	}
	if channelID == "" {
		if parsedChannel != "" {
			channelID = parsedChannel
		} else if teamID != "" {
			channelID = destination
		}
	}

	return teamID, channelID
}

// splitTeamsDestination splits a Teams destination into team and channel parts
func splitTeamsDestination(destination string) (string, string) {
	for _, sep := range []string{":", "/"} {
		parts := strings.SplitN(destination, sep, teamsDestinationParts)
		if len(parts) == teamsDestinationParts {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

// readConfigString reads a string value from a config map
func readConfigString(config map[string]any, key string) string {
	if config == nil {
		return ""
	}
	value, ok := config[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}

// operationNameForChannel maps a notification channel to an operation name
func operationNameForChannel(channel enums.Channel) (types.OperationName, error) {
	switch channel {
	case enums.ChannelSlack, enums.ChannelTeams:
		return types.OperationName("message.send"), nil
	default:
		return "", ErrNotificationChannelUnsupported
	}
}
