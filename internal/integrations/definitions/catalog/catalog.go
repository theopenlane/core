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
)

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
		cloudflare.Builder(),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(),
		gcpscc.Builder(cfg.GCPSCC),
		oidcgeneric.Builder(cfg.OIDCGeneric),
		microsoftteams.Builder(cfg.MicrosoftTeams),
	}
}
