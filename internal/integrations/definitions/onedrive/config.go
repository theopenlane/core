package onedrive

import "github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"

// Config holds operator-level credentials for the OneDrive definition
type Config struct {
	microsoft.AppConfig `koanf:",squash"`
	// ContentMode controls how document content is returned for live external content queries.
	// Valid values: "iframe" (default, returns an embeddable preview iframe) and "pdf" (exports the document as PDF bytes).
	ContentMode string `json:"contentmode" koanf:"contentmode" default:"iframe"`
	// Legacy is the retired app registration that installations created before the tenant move still use
	Legacy microsoft.AppConfig `json:"legacy" koanf:"legacy"`
}
