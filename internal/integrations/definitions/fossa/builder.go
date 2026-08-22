package fossa

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Builder returns the FOSSA definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "FOSSA",
				DisplayName: "FOSSA",
				Description: "Collect FOSSA security vulnerabilities and OSS license compliance issues from scanned projects.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/fossa",
				Tags:        []string{"vulnerabilities", "findings", "licensing", "sbom"},
				Active:      false,
				Visible:     false,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         fossaCredential.ID(),
					Name:        "FOSSA API Token",
					Description: "FOSSA API token with full access, used to read issues and organization details.",
					Schema:      fossaSchema,
					Recommended: true,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       fossaCredential.ID(),
					Name:                "FOSSA API Token",
					Description:         "Configure FOSSA access using an API token generated from Account Settings, Integrations, API.",
					CredentialRefs:      []types.CredentialSlotID{fossaCredential.ID()},
					ClientRefs:          []types.ClientID{fossaClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: fossaCredential.ID(),
						Description:   "Removes the stored FOSSA API token from Openlane. If the token is no longer needed, revoke it from your FOSSA account settings.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            fossaClient.ID(),
					CredentialRefs: []types.CredentialSlotID{fossaCredential.ID()},
					Description:    "FOSSA REST API client",
					Build:          ClientBuilder{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Validate FOSSA access",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    fossaClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         vulnerabilitySyncOperation.Name(),
					Description:  "Collect FOSSA security vulnerabilities, and optionally OSS license compliance findings",
					Topic:        definitionID.OperationTopic(vulnerabilitySyncOperation.Name()),
					ClientRef:    fossaClient.ID(),
					ConfigSchema: vulnerabilitySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					// no Disabled resolver, security vulnerability collection is always on
					ConfigResolver: providerkit.ConfigFrom(func(u UserInput) VulnerabilitySync { return u.VulnerabilitySync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaVulnerability.Name,
						},
						{
							Schema: entityops.SchemaFinding.Name,
						},
					},
					IngestHandle:        VulnerabilityCollect{}.IngestHandle(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"FOSSA API token with full access"},
					Schedule:            gala.NewFullFetchSchedule(),
				},
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaVulnerability.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprVulnerability,
					},
				},
				{
					Schema: entityops.SchemaFinding.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprFinding,
					},
				},
			},
		}, nil
	})
}
