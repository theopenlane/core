package scim

import "github.com/theopenlane/core/internal/integrations/types"

var (
	// DefinitionID is the stable identifier for the SCIM Directory Sync integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	// HealthDefaultOperation is the operation ref for the SCIM health check
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	// DirectorySyncOperation is the operation ref for the SCIM directory sync operation
	DirectorySyncOperation = types.NewOperationRef[struct{}]("directory.sync")
)

// Slug is the unique identifier for the SCIM Directory Sync integration
const Slug = "scim_directory_sync"

// OperatorConfig holds operator-owned defaults that apply across all SCIM installations
type OperatorConfig struct {
	// DefaultProvisioningMode is the fallback provisioning mode for new installations
	DefaultProvisioningMode string `json:"defaultProvisioningMode,omitempty" jsonschema:"title=Default Provisioning Mode"`
	// DefaultBasePath is the fallback SCIM base path for new installations
	DefaultBasePath string `json:"defaultBasePath,omitempty" jsonschema:"title=Default Base Path"`
}

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
	// TenantKey is the SCIM tenant key
	TenantKey string `json:"tenantKey,omitempty" jsonschema:"title=Tenant Key"`
	// MappingMode controls how directory objects are mapped
	MappingMode string `json:"mappingMode,omitempty" jsonschema:"title=Mapping Mode"`
	// ProvisioningMode controls how directory objects are provisioned
	ProvisioningMode string `json:"provisioningMode,omitempty" jsonschema:"title=Provisioning Mode"`
}

// credential holds the inbound or outbound authentication material for one SCIM installation
type credential struct {
	// BaseURL is the base URL of the SCIM server for outbound requests
	BaseURL string `json:"baseUrl,omitempty"       jsonschema:"title=SCIM Base URL"`
	// Token is the bearer token used for outbound SCIM requests
	Token string `json:"token,omitempty"         jsonschema:"title=Bearer Token"`
	// InboundSecret is the shared secret used to authenticate inbound SCIM pushes
	InboundSecret string `json:"inboundSecret,omitempty" jsonschema:"title=Inbound Secret"`
}
