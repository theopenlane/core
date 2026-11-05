package fossa

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// Builder returns the FOSSA definition builder, devMode allows non-public base urls for local development
func Builder(devMode bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "FOSSA",
			DisplayName:  "FOSSA",
			Description:  "Collect FOSSA security vulnerabilities and OSS license compliance issues from scanned projects.",
			Category:     "security-posture",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/fossa",
			Tags:         []string{"vulnerabilities", "findings", "licensing", "sbom"},
			Active:       false,
			Visible:      false,
			Installation: installation.Registration(),
			Connections: []types.Connector{
				fossaConnection.
					Name("FOSSA API Token").
					Description("Configure FOSSA access using a full API token from a service account (Account Settings, Integrations, API Service Accounts) whose role can view issues.").
					Provides(clientBuilder(devMode)).
					Verified(verify).
					Disconnects("Removes the stored FOSSA API token from Openlane. If the token is no longer needed, revoke it from your FOSSA account settings.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[VulnerabilitySync]().
					Ingests(runVulnerabilitySync).
					HealthCheck(probeAccess).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Ingest(
						types.IngestContract{Schema: entityops.SchemaVulnerability.Name},
						types.IngestContract{Schema: entityops.SchemaFinding.Name},
					).
					Permissions("Full API token (not push-only) on a FOSSA service account with a role that can view issues").
					Description("Collect FOSSA security vulnerabilities, and optionally OSS license compliance findings").
					Registration(),
			},
			Mappings: []types.MappingRegistration{
				providerkit.VulnerabilityMapping(mapExprVulnerability),
				providerkit.FindingMapping(mapExprFinding),
			},
		}, nil
	})
}
