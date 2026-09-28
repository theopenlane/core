package keycloak

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Keycloak definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Keycloak",
				DisplayName: "Keycloak",
				Description: "Collect Keycloak realm users, groups, and memberships for identity posture and access governance.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/keycloak",
				Tags:        []string{"directory"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				keycloakCredential.Registration(types.CredentialRegistration{
					Name:        "Keycloak Credential",
					Description: "Client credentials used to access Keycloak realm data.",
					Schema:      keycloakCredential.Schema(),
				}),
			},
			Connections: []types.ConnectionRegistration{
				keycloakConnection.Registration(types.ConnectionRegistration{
					Name:           "Keycloak Client Credentials",
					Description:    "Configure Keycloak access using client credentials from your realm.",
					CredentialRefs: []types.CredentialSlotID{keycloakCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: keycloakClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: integration.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: keycloakCredential.ID(),
						Description:   "Removes the stored client credentials from Openlane. If the client is no longer needed, disable or delete it in your Keycloak admin console under Clients.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				keycloakClient.Registration(Client{}.Build, types.ClientRegistration{
					Description: "Keycloak API client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:         "Collect Keycloak realm users, groups, and memberships as directory accounts",
					Policy:              types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"view-realm", "view-users", "query-groups", "view-events"},
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
					IngestHandle: DirectorySync{}.IngestHandle(),
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
