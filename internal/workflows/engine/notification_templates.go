package engine

import (
	"context"
	"maps"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/mapx"
)

// buildNotificationTemplateVars builds CEL vars and merged data for template rendering
func (e *WorkflowEngine) buildNotificationTemplateVars(ctx context.Context, instance *generated.WorkflowInstance, obj *workflows.Object, actionKey string, paramsData map[string]any) (map[string]any, map[string]any, error) {
	vars, err := e.buildActionCELVars(ctx, instance, obj)
	if err != nil {
		return nil, nil, err
	}

	_, baseData := workflows.BuildWorkflowActionContext(instance, obj, actionKey)
	vars = mapx.DeepCloneMapAny(vars)
	maps.Copy(vars, baseData)

	data := make(map[string]any, len(paramsData))
	maps.Copy(data, paramsData)

	maps.Copy(data, baseData)
	vars["data"] = data
	maps.Copy(vars, data)

	return vars, data, nil
}
