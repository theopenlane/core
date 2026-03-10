package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/mo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notificationpreference"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	"github.com/theopenlane/core/internal/integrations/operations"
	teamsprovider "github.com/theopenlane/core/internal/integrations/providers/microsoftteams"
	slackprovider "github.com/theopenlane/core/internal/integrations/providers/slack"
	"github.com/theopenlane/core/internal/integrations/targetresolver"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
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
	Blocks []map[string]any
	// Data holds the merged template data payload
	Data map[string]any
	// Vars holds the CEL variable map used for rendering
	Vars map[string]any
}

// notificationChannelTarget captures direct notification channel send targets
type notificationChannelTarget struct {
	// Channel identifies which notification channel integration to execute
	Channel enums.Channel
	// Destination identifies the provider destination identifier for the channel
	Destination string
	// Config carries optional channel-specific operation config overrides
	Config map[string]any
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

	var blocks []map[string]any
	if template.Blocks != nil {
		renderedBlocks, err := renderTemplateValue(ctx, e.celEvaluator, template.Blocks, vars)
		if err != nil {
			return nil, err
		}
		blocks, err = decodeRenderedNotificationBlocks(renderedBlocks)
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
	vars = mapx.DeepCloneMapAny(vars)
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

	result, err := jsonx.ValidateSchema(template.Jsonconfig, data)
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
	if rendered == nil || len(userIDs) == 0 || len(channels) == 0 {
		return nil
	}

	integrationChannels := lo.Filter(channels, func(channel enums.Channel, _ int) bool {
		return channel != enums.ChannelInApp
	})
	if len(integrationChannels) == 0 {
		return nil
	}

	if e.integrationOperations == nil {
		return ErrIntegrationOperationsRequired
	}

	for _, channel := range integrationChannels {
		if err := e.dispatchChannelNotifications(ctx, ownerID, channel, rendered, userIDs); err != nil {
			return err
		}
	}

	return nil
}

// dispatchNotificationChannelTargets executes integration notifications for direct channel targets
func (e *WorkflowEngine) dispatchNotificationChannelTargets(ctx context.Context, ownerID string, rendered *renderedNotificationTemplate, targets []notificationChannelTarget) error {
	if rendered == nil || len(targets) == 0 {
		return nil
	}
	if e.integrationOperations == nil {
		return ErrIntegrationOperationsRequired
	}

	for _, target := range targets {
		selection, err := e.resolveNotificationExecutionTarget(ctx, ownerID, rendered.Template, target.Channel)
		if err != nil {
			return err
		}

		preference := &generated.NotificationPreference{
			Destination: target.Destination,
			Config:      mapx.DeepCloneMapAny(target.Config),
		}

		config, err := buildNotificationOperationConfig(target.Channel, preference, rendered)
		if err != nil {
			return err
		}

		if selection.Integration != nil {
			merged, err := operations.ResolveOperationConfig(&selection.Integration.Config, string(selection.Operation.Name), config)
			if err != nil {
				return err
			}
			if merged != nil {
				config = merged
			}
		}

		_, err = e.integrationOperations.Run(ctx, types.OperationRequest{
			OrgID:         ownerID,
			IntegrationID: integrationIDForRecord(selection.Integration),
			Provider:      selection.Provider,
			Name:          selection.Operation.Name,
			Config:        config,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// dispatchChannelNotifications sends notifications to all users for a specific channel
func (e *WorkflowEngine) dispatchChannelNotifications(ctx context.Context, ownerID string, channel enums.Channel, rendered *renderedNotificationTemplate, userIDs []string) error {
	selection, err := e.resolveNotificationExecutionTarget(ctx, ownerID, rendered.Template, channel)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := e.dispatchUserNotification(ctx, ownerID, userID, channel, selection.Operation.Name, selection.Provider, selection.Integration, rendered); err != nil {
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
		OrgID:         ownerID,
		IntegrationID: integrationIDForRecord(integrationRecord),
		Provider:      provider,
		Name:          operationName,
		Config:        config,
	})

	return err
}

// notificationExecutionTarget captures integration execution routing for notifications
type notificationExecutionTarget struct {
	// Integration is the selected installed integration
	Integration *generated.Integration
	// Provider is the selected provider type
	Provider types.ProviderType
	// Operation is the selected provider operation descriptor
	Operation types.OperationDescriptor
}

// resolveNotificationExecutionTarget resolves provider integration and operation selection for notification dispatch
func (e *WorkflowEngine) resolveNotificationExecutionTarget(ctx context.Context, ownerID string, template *generated.NotificationTemplate, channel enums.Channel) (notificationExecutionTarget, error) {
	if e.integrationRegistry == nil {
		return notificationExecutionTarget{}, ErrIntegrationRegistryRequired
	}

	if template != nil {
		templateChannel := template.Channel
		if templateChannel != "" && templateChannel != enums.ChannelInvalid && channel != templateChannel {
			return notificationExecutionTarget{}, ErrNotificationTemplateChannelMismatch
		}
	}

	provider, err := providerForNotificationChannel(channel)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	operationName, err := operationNameForChannel(channel)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	source, err := targetresolver.NewEntSource(e.client)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	resolver, err := targetresolver.NewResolver(source)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	criteria := targetresolver.ResolveCriteria{
		OwnerID:  ownerID,
		Provider: mo.Some(provider),
	}

	if template != nil && template.IntegrationID != "" {
		criteria.IntegrationID = mo.Some(template.IntegrationID)
	}

	result, err := resolver.Resolve(workflows.AllowContext(ctx), criteria)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	operation, err := e.integrationRegistry.ResolveOperation(provider, operationName, types.OperationKindNotify)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	return notificationExecutionTarget{
		Integration: result.Integration,
		Provider:    result.Provider,
		Operation:   operation,
	}, nil
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
func buildNotificationOperationConfig(channel enums.Channel, preference *generated.NotificationPreference, rendered *renderedNotificationTemplate) (json.RawMessage, error) {
	var config map[string]any
	if preference != nil && len(preference.Config) > 0 {
		config = mapx.DeepCloneMapAny(preference.Config)
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
			if len(rendered.Blocks) > 0 {
				config["blocks"] = rendered.Blocks
			}
		}
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
	default:
		return nil, ErrNotificationChannelUnsupported
	}

	out, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// decodeRenderedNotificationBlocks converts rendered template blocks into a structured block list
func decodeRenderedNotificationBlocks(value any) ([]map[string]any, error) {
	if value == nil {
		return nil, nil
	}

	var blocks []map[string]any
	if err := jsonx.RoundTrip(value, &blocks); err != nil {
		return nil, ErrNotificationTemplateBlocksInvalid
	}

	return blocks, nil
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

// providerForNotificationChannel maps notification channels to provider types
func providerForNotificationChannel(channel enums.Channel) (types.ProviderType, error) {
	switch channel {
	case enums.ChannelSlack:
		return slackprovider.TypeSlack, nil
	case enums.ChannelTeams:
		return teamsprovider.TypeMicrosoftTeams, nil
	default:
		return types.ProviderUnknown, ErrNotificationChannelUnsupported
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

// integrationIDForRecord returns the integration id when a record is available
func integrationIDForRecord(record *generated.Integration) string {
	if record == nil {
		return ""
	}

	return record.ID
}
