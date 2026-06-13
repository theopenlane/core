package onedrive

// Config holds operator-level credentials for the OneDrive definition
type Config struct {
	// ClientID is the Azure OAuth application client identifier
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Azure OAuth application client secret
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Azure application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
	// DefaultTenant pins the OAuth flow to a specific tenant ID or domain (e.g. for local dev/testing);
	// when empty the generic /common endpoint is used
	DefaultTenant string `json:"defaulttenant" koanf:"defaulttenant"`
	// DocumentIntelligenceEndpoint is the Azure Document Intelligence resource endpoint
	// (e.g. https://my-resource.cognitiveservices.azure.com). When set, document content
	// is extracted via the prebuilt-layout model instead of local XML parsing.
	DocumentIntelligenceEndpoint string `json:"documentintelligenceendpoint" koanf:"documentintelligenceendpoint"`
	// DocumentIntelligenceKey is the API key for the Document Intelligence resource
	DocumentIntelligenceKey string `json:"documentintelligencekey" koanf:"documentintelligencekey" sensitive:"true"`
	// ContentMode controls how document content is returned for live external content queries.
	// Valid values: "html" (default, parses DOCX locally) and "iframe" (returns an embeddable preview iframe).
	ContentMode string `json:"contentmode" koanf:"contentmode" default:"html"`
}
