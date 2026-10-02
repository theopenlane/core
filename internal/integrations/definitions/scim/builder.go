package scim

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns a registry builder that constructs the SCIM directory sync definition
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "scim",
			DisplayName: "SCIM 2.0",
			Description: "Synchronize directory objects through SCIM",
			Category:    "identity",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/scim",
			Tags:        []string{"directory"},
			Active:      true,
			Visible:     false,
			UserInput:   userInput.Registration(),
			Webhooks: []types.WebhookRegistration{
				SCIMAuthWebhook.Registration(types.WebhookRegistration{
					EndpointURLTemplate: "/v1/integrations/scim/{endpointID}/v2",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(DefinitionID, types.OperationRegistration{
					Description: "Synchronize directory state through SCIM",
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
