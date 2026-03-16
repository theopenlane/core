package awsassets

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// OperatorConfig holds operator-owned defaults that apply across all AWS Assets installations
type OperatorConfig struct {
	DefaultSessionDuration string   `json:"defaultSessionDuration,omitempty" jsonschema:"title=Default Session Duration"`
	DefaultExternalID      string   `json:"defaultExternalId,omitempty"      jsonschema:"title=Default External ID"`
	DefaultRegions         []string `json:"defaultRegions,omitempty"         jsonschema:"title=Default Regions"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
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
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
				Schema: providerkit.SchemaFrom[OperatorConfig](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         AWSAssetsClient.ID(),
					Description: "AWS API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate AWS credentials and installation scope",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   AWSAssetsClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        AssetCollectOperation.Name(),
					Description: "Collect asset inventory from AWS",
					Topic:       AssetCollectOperation.Topic(Slug),
					ClientRef:   AWSAssetsClient.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      AssetCollect{}.Handle(Client{}),
				},
			},
			Mappings: []types.MappingRegistration{
				{Schema: "asset", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
			},
		}, nil
	})
}
