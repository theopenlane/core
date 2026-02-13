package catalog

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/aws"
	"github.com/theopenlane/core/internal/integrations/providers/buildkite"
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
		aws.Builder(),
		buildkite.Builder(),
		gcpscc.Builder(),
		github.Builder(),
		github.AppBuilder(),
		googleworkspace.Builder(),
		microsoftteams.Builder(),
		oidcgeneric.Builder(),
		okta.Builder(),
		slack.Builder(),
		vercel.Builder(),
	}
}
