package catalog

import (
	"github.com/theopenlane/core/internal/integrations/definitions/authentik"
	"github.com/theopenlane/core/internal/integrations/definitions/awssecurityhub"
	"github.com/theopenlane/core/internal/integrations/definitions/azureentraid"
	"github.com/theopenlane/core/internal/integrations/definitions/azuresecuritycenter"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/definitions/gcpscc"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/internal/integrations/definitions/googleworkspace"
	"github.com/theopenlane/core/internal/integrations/definitions/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/definitions/oidclocal"
	"github.com/theopenlane/core/internal/integrations/definitions/okta"
	"github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/internal/integrations/registry"
)

// Builders returns the built-in reference definition builders. devMode is the
// server-level development flag; when true, integrations that support it (e.g.
// email) use local file-based senders instead of calling provider APIs
func Builders(cfg Config, devMode bool) []registry.Builder {
	return []registry.Builder{
		authentik.Builder(),
		awssecurityhub.Builder(cfg.AWSSecurityHub),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(),
		cloudflare.Builder(),
		email.Builder(&cfg.Email, devMode),
		gcpscc.Builder(),
		githubapp.Builder(cfg.GitHubApp),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		microsoftteams.Builder(cfg.MicrosoftTeams),
		oidclocal.Builder(cfg.OIDCLocal),
		okta.Builder(),
		scim.Builder(),
		slack.Builder(cfg.Slack, &cfg.SlackRuntime, devMode),
	}
}
