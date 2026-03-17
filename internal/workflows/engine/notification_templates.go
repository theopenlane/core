package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notificationpreference"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	"github.com/theopenlane/core/internal/integrations/definitions/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/internal/integrations/operations"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
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

	if e.integrationRuntime == nil {
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
	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}

	for _, target := range targets {
		allowCtx := workflows.AllowContext(ctx)
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

		_, err = e.integrationRuntime.Dispatch(allowCtx, operations.DispatchRequest{
			InstallationID: installationIDForRecord(selection.Installation),
			Operation:      selection.Operation.Name,
			Config:         config,
			RunType:        enums.IntegrationRunTypeEvent,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// dispatchTemplateNotificationDestinations executes template-scoped provider destinations once per template run
func (e *WorkflowEngine) dispatchTemplateNotificationDestinations(ctx context.Context, ownerID string, rendered *renderedNotificationTemplate) error {
	if rendered == nil || rendered.Template == nil || len(rendered.Template.Destinations) == 0 {
		return nil
	}
	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}

	config, err := buildTemplateDestinationOperationConfig(rendered.Template.Channel, rendered)
	if err != nil || len(config) == 0 {
		return err
	}

	selection, err := e.resolveNotificationExecutionTarget(ctx, ownerID, rendered.Template, rendered.Template.Channel)
	if err != nil {
		return err
	}

	_, err = e.integrationRuntime.Dispatch(workflows.AllowContext(ctx), operations.DispatchRequest{
		InstallationID: installationIDForRecord(selection.Installation),
		Operation:      selection.Operation.Name,
		Config:         config,
		RunType:        enums.IntegrationRunTypeEvent,
	})

	return err
}

// dispatchChannelNotifications sends notifications to all users for a specific channel
func (e *WorkflowEngine) dispatchChannelNotifications(ctx context.Context, ownerID string, channel enums.Channel, rendered *renderedNotificationTemplate, userIDs []string) error {
	selection, err := e.resolveNotificationExecutionTarget(ctx, ownerID, rendered.Template, channel)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := e.dispatchUserNotification(ctx, ownerID, userID, channel, selection, rendered); err != nil {
			return err
		}
	}

	return nil
}

// dispatchUserNotification sends a notification to a single user via integration
func (e *WorkflowEngine) dispatchUserNotification(ctx context.Context, ownerID, userID string, channel enums.Channel, selection notificationExecutionTarget, rendered *renderedNotificationTemplate) error {
	allowCtx := workflows.AllowContext(ctx)
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

	_, err = e.integrationRuntime.Dispatch(allowCtx, operations.DispatchRequest{
		InstallationID: installationIDForRecord(selection.Installation),
		Operation:      selection.Operation.Name,
		Config:         config,
		RunType:        enums.IntegrationRunTypeEvent,
	})

	return err
}

// notificationExecutionTarget captures integration execution routing for notifications
type notificationExecutionTarget struct {
	// Installation is the resolved installation record
	Installation *generated.Integration
	// DefinitionID is the resolved definition identifier
	DefinitionID string
	// Operation is the resolved operation registration
	Operation types.OperationRegistration
}

// resolveNotificationExecutionTarget resolves the installation and operation for notification dispatch
func (e *WorkflowEngine) resolveNotificationExecutionTarget(ctx context.Context, ownerID string, template *generated.NotificationTemplate, channel enums.Channel) (notificationExecutionTarget, error) {
	if e.integrationRuntime == nil {
		return notificationExecutionTarget{}, ErrIntegrationRegistryRequired
	}

	if template != nil {
		templateChannel := template.Channel
		if templateChannel != "" && templateChannel != enums.ChannelInvalid && channel != templateChannel {
			return notificationExecutionTarget{}, ErrNotificationTemplateChannelMismatch
		}
	}

	definitionID, err := definitionIDForNotificationChannel(channel)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	operationName, err := operationNameForChannel(channel)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	def, ok := e.integrationRuntime.Registry().Definition(definitionID)
	if !ok {
		return notificationExecutionTarget{}, ErrIntegrationRegistryRequired
	}

	operation, err := e.integrationRuntime.Registry().Operation(def.ID, operationName)
	if err != nil {
		return notificationExecutionTarget{}, err
	}

	installationID := ""
	if template != nil {
		installationID = template.IntegrationID
	}

	installation, err := e.integrationRuntime.ResolveInstallation(workflows.AllowContext(ctx), ownerID, installationID, def.ID)
	if err != nil {
		switch {
		case errors.Is(err, integrationsruntime.ErrInstallationRequired),
			errors.Is(err, integrationsruntime.ErrDefinitionIDRequired):
			return notificationExecutionTarget{}, ErrInstallationRequired
		case errors.Is(err, integrationsruntime.ErrInstallationIDRequired):
			return notificationExecutionTarget{}, ErrInstallationIDRequired
		case errors.Is(err, integrationsruntime.ErrInstallationNotFound):
			return notificationExecutionTarget{}, ErrInstallationNotFound
		case errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch):
			return notificationExecutionTarget{}, ErrInstallationDefinitionMismatch
		default:
			return notificationExecutionTarget{}, err
		}
	}

	return notificationExecutionTarget{
		Installation: installation,
		DefinitionID: def.ID,
		Operation:    operation,
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

// buildTemplateDestinationOperationConfig builds operation config for template-scoped explicit destinations
func buildTemplateDestinationOperationConfig(channel enums.Channel, rendered *renderedNotificationTemplate) (json.RawMessage, error) {
	if rendered == nil || rendered.Template == nil || len(rendered.Template.Destinations) == 0 {
		return nil, nil
	}

	switch channel {
	case enums.ChannelSlack:
	default:
		return nil, nil
	}

	baseConfig, err := buildNotificationOperationConfig(channel, nil, rendered)
	if err != nil {
		return nil, err
	}

	config := map[string]any{}
	if len(baseConfig) > 0 {
		if err := json.Unmarshal(baseConfig, &config); err != nil {
			return nil, err
		}
	}

	config["destinations"] = append([]string(nil), rendered.Template.Destinations...)

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

// definitionIDForNotificationChannel maps notification channels to integration definition IDs
func definitionIDForNotificationChannel(channel enums.Channel) (string, error) {
	switch channel {
	case enums.ChannelSlack:
		return slack.DefinitionID.ID(), nil
	case enums.ChannelTeams:
		return microsoftteams.DefinitionID.ID(), nil
	default:
		return "", ErrNotificationChannelUnsupported
	}
}

// operationNameForChannel maps a notification channel to an operation name
func operationNameForChannel(channel enums.Channel) (string, error) {
	switch channel {
	case enums.ChannelSlack:
		return slack.MessageSendOperation.Name(), nil
	case enums.ChannelTeams:
		return microsoftteams.MessageSendOperation.Name(), nil
	default:
		return "", ErrNotificationChannelUnsupported
	}
}

// installationIDForRecord returns the installation id when a record is available
func installationIDForRecord(record *generated.Integration) string {
	if record == nil {
		return ""
	}

	return record.ID
}
