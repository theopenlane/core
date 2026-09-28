package oci

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
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
			UserInput:    userInput.Registration(),
			HealthCheck:  identityClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				ociCredential.Registration(types.CredentialRegistration{
					Name:        "OCI API Key Credential",
					Description: "OCI API signing key used to authenticate against the tenancy.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				ociConnection.Registration(types.ConnectionRegistration{
					Name:        "OCI API Key",
					Description: "Configure Oracle Cloud Infrastructure access using an API signing key registered to a tenancy user.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored API signing key from Openlane. If the key is no longer needed, delete it from the user's API keys in the OCI console.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				identityClient.Registration(IdentityClientBuilder{}.Build, types.ClientRegistration{
					Description: "Oracle Cloud Infrastructure Identity client",
				}),
				cloudGuardClient.Registration(CloudGuardClientBuilder{}.Build, types.ClientRegistration{
					Description: "Oracle Cloud Infrastructure Cloud Guard client",
				}),
			},
			Operations: []types.OperationRegistration{
				findingsSyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Collect OCI Cloud Guard problems as findings",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaFinding.Name,
						},
					},
					RequiredPermissions: []string{"read cloud-guard-problems in tenancy"},
				}),
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
