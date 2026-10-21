package googledrive

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/utils/jsonx"
)

// Builder returns the Google Drive definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:          definitionID.ID(),
			Family:      "Google Drive",
			DisplayName: "Google Drive",
			Description: "Live, read-only integration with Google Drive for on-the-fly HTML export of Google Docs.",
			Category:    "document",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/google_drive",
			Tags:        []string{"document"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				driveConnection(oauthConnection.
					Name("Google Drive OAuth").
					Description("Connect your Google account using OAuth to access the Drive documents you select.").
					Recommended(), cfg.AppConfig, "https://www.googleapis.com/auth/drive.file", false),
				driveConnection(legacyOauthConnection.
					Name("Google Drive OAuth (Legacy)").
					Description("Read access to all Drive documents for installations connected before document selection, reconnect to move to the current connection.").
					Retired(), cfg.Legacy, "https://www.googleapis.com/auth/drive.readonly", true),
			},
			Operations: []types.OperationRegistration{
				documentExportOperation.Description("Export a Google Doc as HTML via the Drive files.export endpoint").Registration(),
				pickerSessionOperation(cfg).Description("Return a short-lived access token and Picker settings for selecting Drive files in the console").Registration(),
				operations.NewResourceSelectionOperation(storeSelectedDocuments, 0).Description("Store the Google Docs selected in the Picker as the documents to sync").Registration(),
				types.OperationRefOf[FolderSync]().
					Ingests(runFolderSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaInternalPolicy.Name}).
					Schedule(gala.NewFullFetchSchedule()).
					Description("List Google Docs in the configured folder and emit policy ingest envelopes").
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

// driveConnection completes a Drive OAuth connection that authenticates against app with scope, legacy marking the retired drive.readonly connection
func driveConnection(ref types.ConnectionRef[googleDriveCred], app AppConfig, scope string, legacy bool) types.ConnectionRef[googleDriveCred] {
	return ref.
		Authenticates(auth.OAuthRegistration(auth.OAuthRegistrationOptions[googleDriveCred]{
			Config: auth.OAuthConfig{ClientID: app.ClientID,
				ClientSecret: app.ClientSecret,
				AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL:     "https://oauth2.googleapis.com/token",
				RedirectURL:  app.RedirectURL,
				Scopes:       []string{scope},
				AuthParams: map[string]string{
					"access_type": "offline",
					"prompt":      "consent",
				},
			},
			Material:              newGoogleDriveCred,
			EncodeCredentialError: ErrCredentialEncode,
		})).
		Provides(clientBuilder(app, legacy)).
		Verified(verify).
		Disconnects("Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Google account under Security > Third-party access.", nil)
}

// newGoogleDriveCred maps completed OAuth material to the Drive credential
func newGoogleDriveCred(material auth.OAuthMaterial) (googleDriveCred, error) {
	return googleDriveCred{
		AccessToken:  material.AccessToken,
		RefreshToken: material.RefreshToken,
		Expiry:       material.Expiry,
	}, nil
}
