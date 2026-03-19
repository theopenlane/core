package okta

import (
	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Okta integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0OKTA0000000000000000001")
	// oktaCredential is the credential slot used by the Okta client
	oktaCredential = types.NewCredentialRef(Slug)
	// OktaClient is the client ref for the Okta API client used by this definition
	OktaClient = types.NewClientRef[*oktagosdk.APIClient]()
	// HealthDefaultOperation is the operation ref for the Okta health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// PoliciesCollectOperation is the operation ref for the Okta policies collection operation
	PoliciesCollectOperation = types.NewOperationRef[PoliciesCollect]("policies.collect")
	// DirectorySyncOperation is the operation ref for the Okta directory sync operation
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

// Slug is the unique identifier for the Okta integration
const Slug = "okta"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// Search is an optional Okta search expression applied server-side when listing users
	Search string `json:"search,omitempty" jsonschema:"title=User Search Expression,description=Optional Okta search expression for filtering users (e.g. profile.department eq \"Engineering\")."`
	// EnableGroupSync controls whether group and membership records are collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
}

// CredentialSchema holds the Okta tenant credentials for one installation
type CredentialSchema struct {
	// OrgURL is the Okta organization URL
	OrgURL string `json:"orgUrl"   jsonschema:"required,title=Org URL"`
	// APIToken is the Okta API token with permissions to read tenant and policy metadata
	APIToken string `json:"apiToken" jsonschema:"required,title=API Token"`
}
