package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeMicrosoftTeams identifies the Microsoft Teams provider
const TypeMicrosoftTeams = types.ProviderType("microsoft_teams")

// Builder returns the Microsoft Teams provider builder
func Builder() providers.Builder {
	return oauth.Builder(TypeMicrosoftTeams, oauth.WithOperations(teamsOperations()))
}
