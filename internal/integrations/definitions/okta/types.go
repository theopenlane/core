package okta

import (
	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Okta integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0OKTA0000000000000000001")
	// OktaClient is the client ref for the Okta API client used by this definition
	OktaClient = types.NewClientRef[*oktagosdk.APIClient]()
	// HealthDefaultOperation is the operation ref for the Okta health check
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	// PoliciesCollectOperation is the operation ref for the Okta policies collection operation
	PoliciesCollectOperation = types.NewOperationRef[struct{}]("policies.collect")
)

// Slug is the unique identifier for the Okta integration
const Slug = "okta"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// OrgURL is the Okta organization URL
	OrgURL string `json:"orgUrl,omitempty" jsonschema:"title=Org URL"`
}

// CredentialSchema holds the Okta tenant credentials for one installation
type CredentialSchema struct {
	// OrgURL is the Okta organization URL
	OrgURL string `json:"orgUrl"   jsonschema:"required,title=Org URL"`
	// APIToken is the Okta API token with permissions to read tenant and policy metadata
	APIToken string `json:"apiToken" jsonschema:"required,title=API Token"`
}
