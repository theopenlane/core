package catalog

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/aws"
	"github.com/theopenlane/core/internal/integrations/providers/awsauditmanager"
	"github.com/theopenlane/core/internal/integrations/providers/awssecurityhub"
	"github.com/theopenlane/core/internal/integrations/providers/azureentraid"
	"github.com/theopenlane/core/internal/integrations/providers/azuresecuritycenter"
	"github.com/theopenlane/core/internal/integrations/providers/buildkite"
	"github.com/theopenlane/core/internal/integrations/providers/cloudflare"
	"github.com/theopenlane/core/internal/integrations/providers/gcpscc"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/providers/googleworkspace"
	"github.com/theopenlane/core/internal/integrations/providers/microsoftteams"
	"github.com/theopenlane/core/internal/integrations/providers/oidcgeneric"
	"github.com/theopenlane/core/internal/integrations/providers/okta"
	"github.com/theopenlane/core/internal/integrations/providers/scim"
	"github.com/theopenlane/core/internal/integrations/providers/slack"
	"github.com/theopenlane/core/internal/integrations/providers/vercel"
)

// Builders returns the provider builders for all supported providers, applying
// operator-level configuration from cfg to providers that require it.
func Builders(cfg Config) []providers.Builder {
	return []providers.Builder{
		aws.Builder(),
		awsauditmanager.Builder(),
		awssecurityhub.Builder(),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(cfg.AzureSecurityCenter),
		buildkite.Builder(),
		cloudflare.Builder(),
		gcpscc.Builder(cfg.GCPSCC),
		github.Builder(cfg.GitHub),
		github.AppBuilder(cfg.GitHubApp),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		microsoftteams.Builder(cfg.MicrosoftTeams),
		oidcgeneric.Builder(cfg.OIDCGeneric),
		okta.Builder(),
		scim.Builder(),
		slack.Builder(cfg.Slack),
		vercel.Builder(),
	}
}
