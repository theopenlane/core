package catalog

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/definitions/awsassets"
	"github.com/theopenlane/core/internal/integrations/definitions/awsauditmanager"
	"github.com/theopenlane/core/internal/integrations/definitions/awssecurityhub"
	"github.com/theopenlane/core/internal/integrations/definitions/azureentraid"
	"github.com/theopenlane/core/internal/integrations/definitions/azuresecuritycenter"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrations/definitions/gcpscc"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/definitions/googleworkspace"
	"github.com/theopenlane/core/internal/integrations/definitions/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/definitions/oidcgeneric"
	"github.com/theopenlane/core/internal/integrations/definitions/okta"
	"github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/internal/integrations/definitions/vercel"
)

// Config aggregates operator-level configuration for all definitions that require
// credentials or secrets at deploy time. Providers that derive all configuration
// from user-supplied credentials at install time do not require fields here
type Config struct {
	// GitHubApp holds operator credentials for the GitHub App definition
	GitHubApp githubapp.Config `json:"githubapp" koanf:"githubapp"`
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
		awsassets.Builder(),
		awsauditmanager.Builder(),
		awssecurityhub.Builder(),
		scim.Builder(),
		slack.Builder(cfg.Slack),
		okta.Builder(),
		vercel.Builder(),
		cloudflare.Builder(),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(),
		gcpscc.Builder(cfg.GCPSCC),
		oidcgeneric.Builder(cfg.OIDCGeneric),
		microsoftteams.Builder(cfg.MicrosoftTeams),
	}
}
