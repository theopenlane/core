package onedrive

// Config holds operator-level credentials for the OneDrive definition
type Config struct {
	// ApplicationID is the Application (client) ID from the Azure app registration Overview page, used as the OAuth client ID and in the well-known configuration for domain validation
	ApplicationID string `json:"applicationid" koanf:"applicationid"`
	// ClientSecret is the client secret Value (not the Secret ID) from the Azure app registration
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Azure application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
	// ContentMode controls how document content is returned for live external content queries.
	// Valid values: "iframe" (default, returns an embeddable preview iframe) and "pdf" (exports the document as PDF bytes).
	ContentMode string `json:"contentmode" koanf:"contentmode" default:"iframe"`
}
