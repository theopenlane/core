package graphapi

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

// catalogVariables combines the shared template variables with one catalog entry's own variables
func catalogVariables(shared, entry []models.TemplateVariable) []*model.TemplateVariable {
	return lo.Map(append(append([]models.TemplateVariable{}, shared...), entry...), func(v models.TemplateVariable, _ int) *model.TemplateVariable {
		return &model.TemplateVariable{Name: v.Name, Description: v.Description}
	})
}

// emailRuntimeClient resolves the runtime email client from the process-wide integrations runtime
func (r *queryResolver) emailRuntimeClient() (*email.Client, error) {
	_, client, ok := runtimeClient(email.DefinitionID.ID())
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	emailClient, ok := client.(*email.Client)
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	return emailClient, nil
}
