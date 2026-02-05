package slack

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
)

// TypeSlack identifies the Slack provider
const TypeSlack = types.ProviderType("slack")

// Builder returns the Slack provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeSlack, oauth.WithOperations(slackOperations()))
}
