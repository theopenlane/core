package googleworkspace

import (
	admin "google.golang.org/api/admin/directory/v1"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Google Workspace integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GWKSP000000000000000001")
	// workspaceCredential is the auth-managed credential slot used by the Workspace client
	workspaceCredential = types.NewCredentialRef(Slug)
	// WorkspaceClient is the client ref for the Google Workspace Admin SDK client used by this definition
	WorkspaceClient = types.NewClientRef[*admin.Service]()
	// HealthDefaultOperation is the operation ref for the Google Workspace health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// DirectorySyncOperation is the operation ref for the Google Workspace directory sync operation
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

// Slug is the unique identifier for the Google Workspace integration
const Slug = "google_workspace"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// AdminEmail is the delegated admin email for impersonation
	AdminEmail string `json:"adminEmail,omitempty" jsonschema:"title=Admin Email"`
	// CustomerID is the Google Workspace customer identifier
	CustomerID string `json:"customerId,omitempty" jsonschema:"title=Customer ID"`
	// Domain scopes directory listing to a specific domain; if set, CustomerID is ignored for listing calls
	Domain string `json:"domain,omitempty" jsonschema:"title=Domain"`
	// Query is a server-side filter applied to user and group listing requests
	Query string `json:"query,omitempty" jsonschema:"title=Directory Query"`
	// OrganizationalUnit limits collection to a specific org unit path
	OrganizationalUnit string `json:"organizationalUnitPath,omitempty" jsonschema:"title=Organizational Unit Path"`
	// IncludeSuspended controls whether suspended users are included
	IncludeSuspended bool `json:"includeSuspendedUsers,omitempty" jsonschema:"title=Include Suspended Users"`
	// EnableGroupSync controls whether group membership is collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
}
