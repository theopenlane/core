package onedrive

import (
	"fmt"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

const microsoftAuthBaseURL = "https://login.microsoftonline.com/%s/oauth2/v2.0"

// Builder returns the OneDrive definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:          definitionID.ID(),
			Family:      "Microsoft",
			DisplayName: "Microsoft OneDrive",
			Description: "Live, read-only integration with Microsoft OneDrive for document management and policy sync.",
			Category:    "document",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/onedrive/overview",
			Tags:        []string{"document"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				oauthConnection.
					Name("OneDrive OAuth").
					Description("Connect your Microsoft account using OAuth to access OneDrive documents.").
					Authenticates(auth.OAuthRegistration(auth.OAuthRegistrationOptions[oneDriveCred]{
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      fmt.Sprintf(microsoftAuthBaseURL, "common") + "/authorize",
							TokenURL:     fmt.Sprintf(microsoftAuthBaseURL, "common") + "/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"https://graph.microsoft.com/Files.Read",
								"https://graph.microsoft.com/User.Read",
								"offline_access",
							},
						},
						Material: func(material auth.OAuthMaterial) (oneDriveCred, error) {
							return oneDriveCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					})).
					Provides(clientBuilder(cfg)).
					Verified(verify).
					Disconnects("Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Microsoft account under Account settings > Privacy.", nil),
			},
			Operations: []types.OperationRegistration{
				documentExportOperation.Description("Download a OneDrive file and return its content").Registration(),
				types.OperationRefOf[FolderSync]().
					Ingests(runFolderSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaInternalPolicy.Name}).
					Schedule(gala.NewFullFetchSchedule()).
					Description("List document files in the configured OneDrive folder and emit policy ingest envelopes").
					Registration(),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaInternalPolicy.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprInternalPolicy,
					},
				},
			},
		}, nil
	})
}
