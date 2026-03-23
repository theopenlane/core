package catalog

import (
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/definitions/awssecurityhub"
	"github.com/theopenlane/core/internal/integrations/definitions/azureentraid"
	"github.com/theopenlane/core/internal/integrations/definitions/azuresecuritycenter"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrations/definitions/gcpscc"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/definitions/googleworkspace"
	"github.com/theopenlane/core/internal/integrations/definitions/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/definitions/okta"
	"github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/internal/integrations/definitions/slack"
)

// Builders returns the built-in reference definition builders
func Builders(cfg Config) []registry.Builder {
	return []registry.Builder{
		githubapp.Builder(cfg.GitHubApp),
		awssecurityhub.Builder(),
		scim.Builder(),
		slack.Builder(cfg.Slack),
		okta.Builder(),
		cloudflare.Builder(),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(),
		gcpscc.Builder(),
		microsoftteams.Builder(cfg.MicrosoftTeams),
	}
}
