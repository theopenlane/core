package azuresecuritycenter

import "github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"

// Config holds operator-level credentials for the Azure Security Center definition
type Config struct {
	microsoft.AppConfig `koanf:",squash"`
	// Lighthouse holds the managing tenant settings for the Azure Lighthouse connection
	Lighthouse microsoft.LighthouseConfig `json:"lighthouse" koanf:"lighthouse"`
}
