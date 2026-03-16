package gcpscc

import (
	"context"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID             = types.NewDefinitionRef("def_01K0GCPSCC00000000000000001")
	HealthDefaultOperation   = types.NewOperationRef[struct{}]("health.default")
	FindingsCollectOperation = types.NewOperationRef[struct{}]("findings.collect")
	SettingsScanOperation    = types.NewOperationRef[struct{}]("settings.scan")
)

const Slug = "gcp_scc"

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

// Builder returns the GCP SCC definition builder with the supplied operator config applied
func Builder(_ Config) definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "Google Cloud Security Command Center v2 client",
					Build:       buildSCCClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify GCP SCC access by listing findings with a minimal query",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:         FindingsCollectOperation.Name(),
					Description:  "Collect GCP Security Command Center findings for vulnerability ingestion",
					Topic:        FindingsCollectOperation.Topic(Slug),
					ClientRef:    clientRef.ID(),
					ConfigSchema: providerkit.SchemaFrom[sccFindingsConfig](),
					Policy:       types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaVulnerability,
							EnsurePayloads: true,
						},
					},
					Handle: runFindingsCollectOperation,
				},
				{
					Name:        SettingsScanOperation.Name(),
					Description: "Scan GCP Security Command Center source and notification settings",
					Topic:       SettingsScanOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runSettingsScanOperation,
				},
			},
			Mappings: gcpsccMappings(),
		}, nil
	})
}
