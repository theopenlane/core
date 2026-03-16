package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Microsoft Teams integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0MSTEAMS00000000000000001")
	// TeamsClient is the client ref for the Microsoft Graph API client used by this definition
	TeamsClient = types.NewClientRef[*providerkit.AuthenticatedClient]()
	// HealthDefaultOperation is the operation ref for the Microsoft Teams health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	// TeamsSampleOperation is the operation ref for the Microsoft Teams sample operation
	TeamsSampleOperation = types.NewOperationRef[TeamsSample]("teams.sample")
	// MessageSendOperation is the operation ref for the Microsoft Teams message send operation
	MessageSendOperation = types.NewOperationRef[MessageSend]("message.send")
)

// Slug is the unique identifier for the Microsoft Teams integration
const Slug = "microsoft_teams"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// TenantID is the Microsoft tenant identifier
	TenantID string `json:"tenantId" jsonschema:"required,title=Tenant ID"`
}
