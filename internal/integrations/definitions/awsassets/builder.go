package awsassets

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// operatorConfig holds operator-owned defaults that apply across all AWS Assets installations
type operatorConfig struct {
	DefaultSessionDuration string   `json:"defaultSessionDuration,omitempty" jsonschema:"title=Default Session Duration"`
	DefaultExternalID      string   `json:"defaultExternalId,omitempty"      jsonschema:"title=Default External ID"`
	DefaultRegions         []string `json:"defaultRegions,omitempty"         jsonschema:"title=Default Regions"`
}

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label            string   `json:"label,omitempty"            jsonschema:"title=Installation Label"`
	AccountSelectors []string `json:"accountSelectors,omitempty" jsonschema:"title=Account Selectors"`
	RegionSelectors  []string `json:"regionSelectors,omitempty"  jsonschema:"title=Region Selectors"`
	FilterExpr       string   `json:"filterExpr,omitempty"       jsonschema:"title=Asset Filter Expression"`
}

// credential holds the AWS role and optional static key material for one installation
type credential struct {
	RoleARN         string `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN"`
	ExternalID      string `json:"externalId"                jsonschema:"required,title=External ID"`
	HomeRegion      string `json:"homeRegion"                jsonschema:"required,title=Home Region"`
	AccountID       string `json:"accountId,omitempty"       jsonschema:"title=Account ID"`
	AccessKeyID     string `json:"accessKeyId,omitempty"     jsonschema:"title=Access Key ID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" jsonschema:"title=Secret Access Key"`
	SessionToken    string `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}

// Builder returns the AWS Assets definition builder
func Builder() definition.Builder {
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0AWSAST00000000000000001",
				Slug:        "aws_assets",
				Version:     "v1",
				Family:      "aws",
				DisplayName: "AWS Assets",
				Description: "Collect AWS asset inventory and account-scoped signals",
				Category:    "cloud",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
				Labels:      map[string]string{"vendor": "aws"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[operatorConfig](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema:  providerkit.SchemaFrom[credential](),
				Persist: types.CredentialPersistModeKeystore,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "aws",
					Description: "AWS API client",
					Build:       buildAWSClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Validate AWS credentials and installation scope",
					Topic:       gala.TopicName("integration.aws_assets.health.default"),
					Client:      "aws",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "asset.collect",
					Kind:        types.OperationKindCollect,
					Description: "Collect asset inventory from AWS",
					Topic:       gala.TopicName("integration.aws_assets.asset.collect"),
					Client:      "aws",
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runAssetCollectionOperation,
				},
			},
			Mappings: []types.MappingRegistration{
				{Schema: "asset", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
			},
		}, nil
	})
}
