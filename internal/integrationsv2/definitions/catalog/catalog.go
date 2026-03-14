package catalog

import (
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/awsassets"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/azureentraid"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/azuresecuritycenter"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/buildkite"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/gcpscc"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/githuboauth"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/googleworkspace"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/microsoftteams"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/oidcgeneric"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/okta"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/scim"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/slack"
	"github.com/theopenlane/core/internal/integrationsv2/definitions/vercel"
)

// Config aggregates operator-level configuration for all definitions that require
// credentials or secrets at deploy time. Providers that derive all configuration
// from user-supplied credentials at install time are not included here.
type Config struct {
	// GitHubApp holds operator credentials for the GitHub App definition
	GitHubApp githubapp.Config `json:"githubapp" koanf:"githubapp"`
	// GitHubOAuth holds OAuth credentials for the GitHub OAuth definition
	GitHubOAuth githuboauth.Config `json:"githuboauth" koanf:"githuboauth"`
	// Slack holds OAuth credentials for the Slack definition
	Slack slack.Config `json:"slack" koanf:"slack"`
	// GoogleWorkspace holds OAuth credentials for the Google Workspace definition
	GoogleWorkspace googleworkspace.Config `json:"googleworkspace" koanf:"googleworkspace"`
	// AzureEntraID holds OAuth credentials for the Azure Entra ID definition
	AzureEntraID azureentraid.Config `json:"azureentraid" koanf:"azureentraid"`
	// GCPSCC holds workload identity configuration for the GCP Security Command Center definition
	GCPSCC gcpscc.Config `json:"gcpscc" koanf:"gcpscc"`
	// OIDCGeneric holds credentials for the generic OIDC definition
	OIDCGeneric oidcgeneric.Config `json:"oidcgeneric" koanf:"oidcgeneric"`
	// MicrosoftTeams holds OAuth credentials for the Microsoft Teams definition
	MicrosoftTeams microsoftteams.Config `json:"microsoftteams" koanf:"microsoftteams"`
}

// Builders returns the built-in reference definition builders
func Builders(cfg Config) []definition.Builder {
	return []definition.Builder{
		githubapp.Builder(cfg.GitHubApp),
		githuboauth.Builder(cfg.GitHubOAuth),
		awsassets.Builder(),
		scim.Builder(),
		slack.Builder(cfg.Slack),
		okta.Builder(),
		vercel.Builder(),
		cloudflare.Builder(),
		buildkite.Builder(),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(),
		gcpscc.Builder(cfg.GCPSCC),
		oidcgeneric.Builder(cfg.OIDCGeneric),
		microsoftteams.Builder(cfg.MicrosoftTeams),
	}
}
