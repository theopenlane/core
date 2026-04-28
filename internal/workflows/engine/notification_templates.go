package engine

import (
	"context"
	"maps"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

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
	if template == nil {
		return nil
	}
	if len(template.Jsonconfig) == 0 {
		return nil
	}

	result, err := jsonx.ValidateSchema(template.Jsonconfig, data)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	return ErrNotificationTemplateDataInvalid
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
			notificationtemplate.OwnerIDEQ(ownerID),
		)

	if templateID != "" {
		template, err := query.Where(notificationtemplate.IDEQ(templateID)).Only(allowCtx)
		if generated.IsNotFound(err) {
			return nil, ErrNotificationTemplateNotFound
		}
		return template, err
	}
	if templateKey != "" {
		template, err := query.Where(notificationtemplate.KeyEQ(templateKey)).First(allowCtx)
		if generated.IsNotFound(err) {
			return nil, ErrNotificationTemplateNotFound
		}
		if err != nil {
			return nil, err
		}

		return template, nil
	}

	return nil, nil
}

// dispatchTemplateIntegration dispatches the rendered template through its associated integration
func (e *WorkflowEngine) dispatchTemplateIntegration(ctx context.Context, ownerID string, rendered *renderedNotificationTemplate, operationName string) error {
	if rendered == nil || rendered.Template == nil || rendered.Template.IntegrationID == "" {
		return nil
	}
	if e.integrationRuntime == nil {
		return ErrIntegrationOperationsRequired
	}
	if operationName == "" {
		return ErrIntegrationOperationCriteriaRequired
	}

	configBytes, err := jsonx.ToRawMessage(buildRenderedTemplateConfig(rendered))
	if err != nil {
		return err
	}

	_, err = e.QueueIntegrationOperation(ctx, IntegrationQueueRequest{
		OrgID:          ownerID,
		InstallationID: rendered.Template.IntegrationID,
		Operation:      operationName,
		Config:         configBytes,
	})

	return err
}

// buildRenderedTemplateConfig assembles the operation config from a rendered template
func buildRenderedTemplateConfig(rendered *renderedNotificationTemplate) map[string]any {
	config := map[string]any{}

	if rendered.Title != "" {
		config["title"] = rendered.Title
	}
	if rendered.Subject != "" {
		config["subject"] = rendered.Subject
	}
	if rendered.Body != "" {
		config["body"] = rendered.Body
	}
	if len(rendered.Blocks) > 0 {
		config["blocks"] = rendered.Blocks
	}
	if len(rendered.Data) > 0 {
		config["data"] = rendered.Data
	}
	if rendered.Template != nil && len(rendered.Template.Destinations) > 0 {
		config["destinations"] = append([]string(nil), rendered.Template.Destinations...)
	}
	if rendered.Template != nil && rendered.Template.Format != "" {
		config["format"] = rendered.Template.Format.String()
	}
	if rendered.Template != nil && len(rendered.Template.Metadata) > 0 {
		config["metadata"] = rendered.Template.Metadata
	}

	return config
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
