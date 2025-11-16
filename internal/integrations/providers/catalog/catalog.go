package catalog

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awsauditmanager"
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
	"github.com/theopenlane/core/internal/integrations/providers/slack"
	"github.com/theopenlane/core/internal/integrations/providers/vercel"
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
