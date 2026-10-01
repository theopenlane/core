package docextract

// Client runs extractions against a Provider using the configured system instruction
type Client struct {
	provider          Provider
	systemInstruction string
}

// Config holds the configuration options for the Client
type Config struct {
	// Provider is the model backend extractions run against
	Provider Provider
	// SystemInstruction is the system prompt applied to every extraction request
	SystemInstruction string
}

// NewClient creates a new Client from the supplied configuration
func NewClient(config Config) (*Client, error) {
	if config.Provider == nil {
		return nil, ErrProviderRequired
	}

	return &Client{provider: config.Provider, systemInstruction: config.SystemInstruction}, nil
}
