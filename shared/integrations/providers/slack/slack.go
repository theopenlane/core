package slack

import (
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/providers/oauth"
	"github.com/theopenlane/shared/integrations/types"
)

// TypeSlack identifies the Slack provider
const TypeSlack = types.ProviderType("slack")

// Builder returns the Slack provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeSlack, oauth.WithOperations(slackOperations()))
}
