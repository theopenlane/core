package zitadel

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the Zitadel definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Zitadel",
			DisplayName:  "Zitadel",
			Description:  "Collect Zitadel directory users for identity posture and access governance.",
			Category:     "identity",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/zitadel/overview",
			Tags:         []string{"directory"},
			Active:       false,
			Visible:      true,
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				patConnection.
					Name("Zitadel Personal Access Token").
					Description("Configure Zitadel access using a Personal Access Token from your instance.").
					Provides(patClient).
					Verified(verify[CredentialSchema]).
					Disconnects("Removes the stored Personal Access Token from Openlane. If the token is no longer needed, revoke it in your Zitadel admin console under Personal Access Tokens.", nil),
				oauthConnection.
					Name("Zitadel OAuth (Client Credentials)").
					Description("Configure Zitadel access using a service user Client ID and Client Secret.").
					Recommended().
					Provides(oauthClient).
					Verified(verify[OAuthCredentialSchema]).
					Disconnects("Removes the stored Client ID and Client Secret from Openlane. If the service user is no longer needed, delete it in your Zitadel admin console.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					SkipDefaultLookback().
					Ingest(types.IngestContract{Schema: entityops.SchemaDirectoryAccount.Name}).
					Description("Collect Zitadel directory users as directory accounts").
					Registration(),
			},
			Mappings: zitadelMappings(),
		}, nil
	})
}
