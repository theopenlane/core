package graphapi

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	intr "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

// integrationScopeVariableNames lists CEL variables exposed to integration scope expressions
var integrationScopeVariableNames = []string{
	"payload",
	"resource",
	"provider",
	"operation",
	"config",
	"integration_config",
	"provider_state",
	"org_id",
	"integration_id",
}

// integrationProviderExtensions captures workflow metadata for one provider
type integrationProviderExtensions struct {
	Provider          string                         `json:"provider"`
	DisplayName       string                         `json:"display_name"`
	Category          string                         `json:"category"`
	Available         bool                           `json:"available"`
	HasAuth           bool                           `json:"has_auth,omitempty"`
	Installations     []integrationInstallationEntry `json:"installations,omitempty"`
	CredentialSchemas []integrationCredentialEntry   `json:"credential_schemas,omitempty"`
	UserInputSchema   json.RawMessage                `json:"user_input_schema,omitempty"`
	Operations        []integrationOperationEntry    `json:"operations"`
}

// integrationInstallationEntry describes one connected integration installation
type integrationInstallationEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// integrationCredentialEntry captures workflow metadata for one provider credential slot
type integrationCredentialEntry struct {
	Ref         types.CredentialSlotID `json:"ref"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Schema      json.RawMessage        `json:"schema,omitempty"`
}

// integrationOperationEntry captures workflow metadata for one operation
type integrationOperationEntry struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description,omitempty"`
	ConfigSchema json.RawMessage            `json:"config_schema,omitempty"`
	Templates    []integrationTemplateEntry `json:"templates,omitempty"`
}

// integrationTemplateEntry describes one notification template available for workflow composition
type integrationTemplateEntry struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// workflowMetadataExtensions builds the extensions payload for the workflow metadata query
func workflowMetadataExtensions(ctx context.Context, rt *intr.Runtime, db *ent.Client) map[string]any {
	return map[string]any{
		"integrations": map[string]any{
			"action_contract": map[string]any{
				"target_selector":    []string{"installation_id", "definition_id"},
				"operation_selector": []string{"operation_name"},
				"scope_fields":       []string{"scope_expression", "scope_payload", "scope_resource"},
				"scope_variables":    integrationScopeVariableNames,
				"run_types":          enums.IntegrationRunTypes,
			},
			"providers": integrationWorkflowProviders(ctx, rt, db),
		},
	}
}

// integrationWorkflowProviders builds provider metadata for integration workflow extensions.
// When the caller is authenticated, org-scoped availability data (connected installations
// and matching notification templates) is included so the composition surface can determine
// which integration definitions are ready for use
func integrationWorkflowProviders(ctx context.Context, rt *intr.Runtime, db *ent.Client) []integrationProviderExtensions {
	if rt == nil {
		return []integrationProviderExtensions{}
	}

	orgAvailability := resolveOrgIntegrationAvailability(ctx, db)

	specs := rt.Catalog()
	entries := make([]integrationProviderExtensions, 0, len(specs))

	for _, spec := range specs {
		def, ok := rt.Definition(spec.ID)
		if !ok {
			continue
		}

		entry := integrationProviderExtensions{
			Provider:    spec.ID,
			DisplayName: spec.DisplayName,
			Category:    spec.Category,
		}

		if lo.ContainsBy(def.Connections, func(connection types.ConnectionRegistration) bool {
			return connection.Auth != nil && connection.Auth.Start != nil
		}) {
			entry.HasAuth = true
		}

		if len(def.CredentialRegistrations) > 0 {
			entry.CredentialSchemas = lo.Map(def.CredentialRegistrations, func(credential types.CredentialRegistration, _ int) integrationCredentialEntry {
				return integrationCredentialEntry{
					Ref:         credential.Ref,
					Name:        credential.Name,
					Description: credential.Description,
					Schema:      jsonx.CloneRawMessage(credential.Schema),
				}
			})
		}

		if def.UserInput != nil {
			entry.UserInputSchema = jsonx.CloneRawMessage(def.UserInput.Schema)
		}

		operations := buildOperationEntries(lo.Filter(def.Operations, func(op types.OperationRegistration, _ int) bool {
			return op.CustomerSelectable == nil || *op.CustomerSelectable
		}))
		slices.SortFunc(operations, func(a, b integrationOperationEntry) int {
			if a.Name < b.Name {
				return -1
			}
			if a.Name > b.Name {
				return 1
			}
			return 0
		})

		if orgAvailability != nil {
			operations = applyProviderAvailability(&entry, spec.ID, operations, orgAvailability)
		}

		entry.Operations = operations
		entries = append(entries, entry)
	}

	return entries
}

// applyProviderAvailability enriches a provider entry with org-scoped installation
// and template availability data, setting Available to true only when at least one
// connected installation and one matching template both exist
func applyProviderAvailability(entry *integrationProviderExtensions, definitionID string, operations []integrationOperationEntry, avail *orgIntegrationAvailability) []integrationOperationEntry {
	installations := avail.installationsByDefinition[definitionID]
	entry.Installations = lo.Map(installations, func(inst *ent.Integration, _ int) integrationInstallationEntry {
		return integrationInstallationEntry{ID: inst.ID, Name: inst.Name}
	})

	hasTemplates := false
	enriched := make([]integrationOperationEntry, len(operations))
	copy(enriched, operations)

	for i, op := range enriched {
		tpls := avail.templatesByTopicPattern[op.Name]
		if len(tpls) > 0 {
			hasTemplates = true

			enriched[i].Templates = lo.Map(tpls, func(t *ent.NotificationTemplate, _ int) integrationTemplateEntry {
				return integrationTemplateEntry{ID: t.ID, Key: t.Key, Name: t.Name}
			})
		}
	}

	entry.Available = len(installations) > 0 && hasTemplates

	return enriched
}

// orgIntegrationAvailability holds pre-fetched org-scoped integration and template data
type orgIntegrationAvailability struct {
	installationsByDefinition map[string][]*ent.Integration
	templatesByTopicPattern   map[string][]*ent.NotificationTemplate
}

// resolveOrgIntegrationAvailability queries connected integrations and active workflow
// action templates for the caller's organization. Returns nil when no caller context
// is available, preserving catalog-only behavior for unauthenticated requests
func resolveOrgIntegrationAvailability(ctx context.Context, db *ent.Client) *orgIntegrationAvailability {
	if db == nil {
		return nil
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return nil
	}

	ownerID := caller.OrganizationID

	installations, err := db.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.StatusEQ(enums.IntegrationStatusConnected),
		).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed querying integrations for workflow metadata availability")
		return nil
	}

	templates, err := db.NotificationTemplate.Query().
		Where(
			notificationtemplate.OwnerIDEQ(ownerID),
			notificationtemplate.ActiveEQ(true),
			notificationtemplate.TemplateContextEQ(enums.TemplateContextWorkflowAction),
		).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed querying notification templates for workflow metadata availability")
		return nil
	}

	return &orgIntegrationAvailability{
		installationsByDefinition: lo.GroupBy(installations, func(inst *ent.Integration) string {
			return inst.DefinitionID
		}),
		templatesByTopicPattern: lo.GroupBy(templates, func(t *ent.NotificationTemplate) string {
			return t.TopicPattern
		}),
	}
}

// buildOperationEntries converts operation registrations to extension entries
func buildOperationEntries(ops []types.OperationRegistration) []integrationOperationEntry {
	return lo.Map(ops, func(op types.OperationRegistration, _ int) integrationOperationEntry {
		return integrationOperationEntry{
			Name:         op.Name,
			Description:  op.Description,
			ConfigSchema: jsonx.CloneRawMessage(op.ConfigSchema),
		}
	})
}
