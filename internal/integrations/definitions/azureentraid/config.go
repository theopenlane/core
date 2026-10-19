package azureentraid

import "github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"

// Config holds operator-level credentials for the Azure Entra ID definition
type Config struct {
	microsoft.AppConfig `koanf:",squash"`
	// Legacy is the retired app registration that installations created before the tenant move still use
	Legacy microsoft.LegacyApp `json:"legacy" koanf:"legacy"`
}
