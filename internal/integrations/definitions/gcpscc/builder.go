package gcpscc

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label          string   `json:"label,omitempty"          jsonschema:"title=Installation Label"`
	OrganizationID string   `json:"organizationId,omitempty" jsonschema:"title=Organization ID"`
	ProjectIDs     []string `json:"projectIds,omitempty"     jsonschema:"title=Project IDs"`
	ProjectScope   string   `json:"projectScope,omitempty"   jsonschema:"title=Project Scope"`
	SourceID       string   `json:"sourceId,omitempty"       jsonschema:"title=SCC Source ID"`
}

// credential holds the GCP SCC credentials for one installation
type credential struct {
	OrganizationID           string   `json:"organizationId,omitempty"           jsonschema:"title=Organization ID"`
	ProjectID                string   `json:"projectId,omitempty"                jsonschema:"title=Project ID"`
	ProjectScope             string   `json:"projectScope,omitempty"             jsonschema:"title=Project Scope"`
	ProjectIDs               []string `json:"projectIds,omitempty"               jsonschema:"title=Project IDs"`
	SourceIDs                []string `json:"sourceIds,omitempty"                jsonschema:"title=SCC Source IDs"`
	ServiceAccountKey        string   `json:"serviceAccountKey,omitempty"        jsonschema:"title=Service Account Key JSON"`
	WorkloadIdentityProvider string   `json:"workloadIdentityProvider,omitempty" jsonschema:"title=Workload Identity Provider"`
	SubjectToken             string   `json:"subjectToken,omitempty"             jsonschema:"title=Subject Token"`
	FindingFilter            string   `json:"findingFilter,omitempty"            jsonschema:"title=Findings Filter"`
}

// def holds operator config for the GCP Security Command Center integration
type def struct {
	cfg Config
}

// buildClient builds the GCP SCC client for one installation
func (d *def) buildClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	return buildSCCClient(ctx, req)
}

// Builder returns the GCP SCC definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0GCPSCC00000000000000001",
				Slug:        "gcp_scc",
				Version:     "v1",
				Family:      "gcp",
				DisplayName: "GCP Security Command Center",
				Description: "Collect Google Cloud Security Command Center findings and settings for security posture reporting.",
				Category:    "cloud",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/gcp_scc/overview",
				Labels:      map[string]string{"vendor": "google", "product": "security-command-center"},
				Active:      false,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema:  providerkit.SchemaFrom[credential](),
				Persist: types.CredentialPersistModeKeystore,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "securitycenter.v2",
					Description: "Google Cloud Security Command Center v2 client",
					Build:       d.buildClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Verify GCP SCC access by listing findings with a minimal query",
					Topic:       gala.TopicName("integration.gcp_scc.health.default"),
					Client:      "securitycenter.v2",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "findings.collect",
					Kind:        types.OperationKindCollect,
					Description: "Collect GCP Security Command Center findings",
					Topic:       gala.TopicName("integration.gcp_scc.findings.collect"),
					Client:      "securitycenter.v2",
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runFindingsCollectOperation,
				},
				{
					Name:        "settings.scan",
					Kind:        types.OperationKindCollect,
					Description: "Scan GCP Security Command Center source and notification settings",
					Topic:       gala.TopicName("integration.gcp_scc.settings.scan"),
					Client:      "securitycenter.v2",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runSettingsScanOperation,
				},
			},
		}, nil
	})
}
