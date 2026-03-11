package catalog

import (
	"github.com/theopenlane/core/internal/integrations/providers/azureentraid"
	"github.com/theopenlane/core/internal/integrations/providers/azuresecuritycenter"
	"github.com/theopenlane/core/internal/integrations/providers/gcpscc"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/providers/googleworkspace"
	"github.com/theopenlane/core/internal/integrations/providers/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/providers/oidcgeneric"
	"github.com/theopenlane/core/internal/integrations/providers/slack"
)

// Config aggregates operator-level configuration for all providers that require
// credentials or secrets at deploy time. Providers that derive all configuration
// from user-supplied credentials at mint time (AWS, SCIM, Buildkite, etc.) are
// not included here. Each field corresponds to one provider package's Config type,
// mirroring the pattern used by pkg/objects/storage.
type Config struct {
	// GitHub holds OAuth credentials for the GitHub provider.
	GitHub github.Config `json:"github" koanf:"github"`
	// GitHubApp holds operator credentials for the GitHub App provider.
	GitHubApp github.AppConfig `json:"githubapp" koanf:"githubapp"`
	// GoogleWorkspace holds OAuth credentials for the Google Workspace provider.
	GoogleWorkspace googleworkspace.Config `json:"googleworkspace" koanf:"googleworkspace"`
	// Slack holds OAuth credentials for the Slack provider.
	Slack slack.Config `json:"slack" koanf:"slack"`
	// MicrosoftTeams holds OAuth credentials for the Microsoft Teams provider.
	MicrosoftTeams microsoftteams.Config `json:"microsoftteams" koanf:"microsoftteams"`
	// AzureEntraID holds OAuth credentials for the Azure Entra ID provider.
	AzureEntraID azureentraid.Config `json:"azureentraid" koanf:"azureentraid"`
	// AzureSecurityCenter holds OAuth credentials for the Azure Security Center provider.
	AzureSecurityCenter azuresecuritycenter.Config `json:"azuresecuritycenter" koanf:"azuresecuritycenter"`
	// GCPSCC holds workload identity configuration for the GCP Security Command Center provider.
	GCPSCC gcpscc.Config `json:"gcpscc" koanf:"gcpscc"`
	// OIDCGeneric holds credentials for the generic OIDC provider.
	OIDCGeneric oidcgeneric.Config `json:"oidcgeneric" koanf:"oidcgeneric"`
}
