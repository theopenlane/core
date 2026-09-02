package oci

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Oracle Cloud Infrastructure definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Oracle Cloud Infrastructure",
				DisplayName: "Oracle Cloud Infrastructure",
				Description: "Collect Oracle Cloud Infrastructure Cloud Guard problems for security posture reporting.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/oci",
				Tags:        []string{"findings"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         ociCredential.ID(),
					Name:        "OCI API Key Credential",
					Description: "OCI API signing key used to authenticate against the tenancy.",
					Schema:      ociSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       ociCredential.ID(),
					Name:                "OCI API Key",
					Description:         "Configure Oracle Cloud Infrastructure access using an API signing key registered to a tenancy user.",
					CredentialRefs:      []types.CredentialSlotID{ociCredential.ID()},
					ClientRefs:          []types.ClientID{identityClient.ID(), cloudGuardClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: ociCredential.ID(),
						Description:   "Removes the stored API signing key from Openlane. If the key is no longer needed, delete it from the user's API keys in the OCI console.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            identityClient.ID(),
					CredentialRefs: []types.CredentialSlotID{ociCredential.ID()},
					Description:    "Oracle Cloud Infrastructure Identity client",
					Build:          IdentityClientBuilder{}.Build,
				},
				{
					Ref:            cloudGuardClient.ID(),
					CredentialRefs: []types.CredentialSlotID{ociCredential.ID()},
					Description:    "Oracle Cloud Infrastructure Cloud Guard client",
					Build:          CloudGuardClientBuilder{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:                healthCheckOperation.Name(),
					Description:         "Verify OCI API signing key access by reading the tenancy",
					Topic:               definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:           identityClient.ID(),
					Policy:              types.ExecutionPolicy{Inline: true},
					Handle:              HealthCheck{}.Handle(),
					ConfigSchema:        healthCheckSchema,
					RequiredPermissions: []string{"inspect tenancies in tenancy"},
				},
				{
					Name:           findingsSyncOperation.Name(),
					Description:    "Collect OCI Cloud Guard problems as findings",
					Topic:          definitionID.OperationTopic(findingsSyncOperation.Name()),
					ClientRef:      cloudGuardClient.ID(),
					ConfigSchema:   findingsSyncSchema,
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.FindingsSync.Disable }),
					ConfigResolver: providerkit.ConfigFrom(func(u UserInput) FindingsSync { return u.FindingsSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaFinding.Name,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"read cloud-guard-problems in tenancy"},
				},
			},
			Mappings: []types.MappingRegistration{
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
