package hooks

import (
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/gala"
)

var emailGala *gala.Gala

// emailConfig holds the runtime email configuration used by hooks that need
// to construct URLs (e.g. questionnaire auth links). Set during server startup
var emailConfig email.RuntimeEmailConfig

// SetEmailGala configures the gala runtime used by hooks to dispatch transactional
// email events. Must be called during server startup before any hooks fire
func SetEmailGala(g *gala.Gala) {
	emailGala = g
}

// SetEmailConfig configures the runtime email config used by hooks that need
// access to email-related URLs. Must be called during server startup before any hooks fire
func SetEmailConfig(cfg email.RuntimeEmailConfig) {
	emailConfig = cfg
}
