package microsoftteams

import "github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"

// Config holds operator-level credentials for the Microsoft Teams definition
type Config struct {
	microsoft.AppConfig `koanf:",squash"`
}
