package cloudflare

import (
	cf "github.com/cloudflare/cloudflare-go/v6"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Cloudflare integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	// cloudflareCredential is the credential slot used by the Cloudflare client
	cloudflareCredential = types.NewCredentialRef(Slug)
	// CloudflareClient is the client ref for the Cloudflare API client used by this definition
	CloudflareClient = types.NewClientRef[*cf.Client]()
	// HealthDefaultOperation is the operation ref for the Cloudflare health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
)

// Slug is the unique identifier for the Cloudflare integration
const Slug = "cloudflare"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}

// CredentialSchema holds the Cloudflare API credentials for one installation
type CredentialSchema struct {
	// APIToken is the Cloudflare API token with permissions to read account and zone metadata
	APIToken string `json:"apiToken"          jsonschema:"required,title=API Token"`
}
