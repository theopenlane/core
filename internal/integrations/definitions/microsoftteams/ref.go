package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0MSTEAMS00000000000000001")
	TeamsClient            = types.NewClientRef[*providerkit.AuthenticatedClient]()
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	TeamsSampleOperation   = types.NewOperationRef[TeamsSample]("teams.sample")
	MessageSendOperation   = types.NewOperationRef[MessageSend]("message.send")
)

const Slug = "microsoft_teams"
