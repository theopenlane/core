package azureentraid

// Config holds operator-level credentials for the Azure Entra ID definition
type Config struct {
	// ClientID is the Azure application (client) identifier registered for this integration
	ClientID string `json:"clientid" koanf:"clientid"`
	// ClientSecret is the Azure application client secret used for client credentials auth
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
}
