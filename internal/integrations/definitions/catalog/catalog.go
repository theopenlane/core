package catalog

import (
	"github.com/theopenlane/core/v2/internal/integrations/definitions/authentik"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/awssecurityhub"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/azureentraid"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/azuresecuritycenter"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gcpscc"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/githubapp"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/googledrive"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/googleworkspace"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/keycloak"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoftteams"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/oci"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/oidclocal"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/okta"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/onedrive"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/system"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/tailscale"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/zitadel"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
)

// Builders returns the built-in reference definition builders. federationIssuer is
// the issuer URI customer identity providers federate against. devMode is the
// server-level development flag; when true, integrations that support it (e.g.
// email) use local file-based senders instead of calling provider APIs
func Builders(cfg Config, federationIssuer string, devMode bool) []registry.Builder {
	return []registry.Builder{
		authentik.Builder(),
		awssecurityhub.Builder(cfg.AWSSecurityHub),
		azureentraid.Builder(cfg.AzureEntraID),
		azuresecuritycenter.Builder(),
		cloudflare.Builder(&cfg.CloudflareRuntime),
		email.Builder(&cfg.Email, devMode),
		gcpscc.Builder(federationIssuer),
		githubapp.Builder(cfg.GitHubApp),
		googledrive.Builder(cfg.GoogleDrive),
		googleworkspace.Builder(cfg.GoogleWorkspace),
		keycloak.Builder(),
		microsoftteams.Builder(cfg.MicrosoftTeams),
		oci.Builder(),
		onedrive.Builder(cfg.OneDrive),
		oidclocal.Builder(cfg.OIDCLocal),
		okta.Builder(),
		scim.Builder(),
		slack.Builder(cfg.Slack, &cfg.SlackRuntime, devMode),
		system.Builder(cfg.PaymentReminder, cfg.OrganizationDelete),
		tailscale.Builder(),
		zitadel.Builder(),
	}
}
