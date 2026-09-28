package authentik

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Authentik definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Authentik",
				DisplayName: "Authentik",
				Description: "Collect Authentik directory users, groups, and memberships for identity posture and access governance.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/authentik",
				Tags:        []string{"directory"},
				Active:      true,
				Visible:     true,
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  authentikClient.HealthCheck(checkHealth),
			Installation: integration.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				authentikCredential.Registration(types.CredentialRegistration{
					Name:        "Authentik Credential",
					Description: "API token used to access Authentik instance data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				authentikConnection.Registration(types.ConnectionRegistration{
					Name:        "Authentik API Token",
					Description: "Configure Authentik access using an API token from your instance.",
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored API token from Openlane. If the token is no longer needed, revoke it in your Authentik admin panel under Directory > Tokens.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				authentikClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Authentik API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Collect Authentik directory users, groups, and memberships as directory accounts",
					Policy:      types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaDirectoryAccount.Name,
						},
						{
							Schema: entityops.SchemaDirectoryGroup.Name,
						},
						{
							Schema: entityops.SchemaDirectoryMembership.Name,
						},
					},
				}),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaDirectoryAccount.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryAccount,
					},
				},
				{
					Schema: entityops.SchemaDirectoryGroup.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryGroup,
					},
				},
				{
					Schema: entityops.SchemaDirectoryMembership.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryMembership,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaDirectoryAccount.Name,
								TargetField:  directoryaccount.FieldExternalID,
								SourceField:  entityops.DirectoryMembershipFields.DirectoryAccountID.InputKey,
							},
							{
								TargetSchema: entityops.SchemaDirectoryGroup.Name,
								TargetField:  directorygroup.FieldExternalID,
								SourceField:  entityops.DirectoryMembershipFields.DirectoryGroupID.InputKey,
							},
						},
					},
				},
			},
		}, nil
	})
}
