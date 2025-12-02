package catalog

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/awsauditmanager"
	"github.com/theopenlane/shared/integrations/providers/azureentraid"
	"github.com/theopenlane/shared/integrations/providers/azuresecuritycenter"
	"github.com/theopenlane/shared/integrations/providers/buildkite"
	"github.com/theopenlane/shared/integrations/providers/cloudflare"
	"github.com/theopenlane/shared/integrations/providers/gcpscc"
	"github.com/theopenlane/shared/integrations/providers/github"
	"github.com/theopenlane/shared/integrations/providers/googleworkspace"
	"github.com/theopenlane/shared/integrations/providers/microsoftteams"
	"github.com/theopenlane/shared/integrations/providers/oidcgeneric"
	"github.com/theopenlane/shared/integrations/providers/okta"
	"github.com/theopenlane/shared/integrations/providers/slack"
	"github.com/theopenlane/shared/integrations/providers/vercel"
)

// Builders returns the default provider builders supported by the system
func Builders() []providers.Builder {
	return []providers.Builder{
		awsauditmanager.Builder(),
		azureentraid.Builder(),
		azuresecuritycenter.Builder(),
		buildkite.Builder(),
		cloudflare.Builder(),
		gcpscc.Builder(),
		github.Builder(),
		googleworkspace.Builder(),
		microsoftteams.Builder(),
		oidcgeneric.Builder(),
		okta.Builder(),
		slack.Builder(),
		vercel.Builder(),
	}
}
