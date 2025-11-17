package summarizer

// Config holds configuration for the text summarization functionality
type Config struct {
	// Type specifies the summarization algorithm to use
	Type Type `json:"type" koanf:"type" default:"lexrank"`
	// LLM contains configuration for large language model based summarization
	LLM LLM `json:"llm" koanf:"llm"`
	// MaximumSentences specifies the maximum number of sentences in the summary
	MaximumSentences int `json:"maximumsentences" koanf:"maximumsentences" default:"10"`
}

// Type defines the type of summarization algorithm
// ENUM(lexrank,llm)
type Type string

// LLMProvider defines supported LLM service providers
// ENUM(openai,anthropic,mistral,gemini,cloudflare,huggingface,ollama)
type LLMProvider string

// LLM contains configuration for multiple LLM providers
type LLM struct {
	// Provider specifies which LLM service to use
	Provider LLMProvider `json:"provider" koanf:"provider" default:"none"`

	// Anthropic contains configuration for Anthropic's API
	Anthropic AnthropicConfig `json:"anthropic" koanf:"anthropic"`

	// Cloudflare contains configuration for Cloudflare's API
	Cloudflare CloudflareConfig `json:"cloudflare" koanf:"cloudflare"`

	// OpenAI contains configuration for OpenAI's API
	OpenAI OpenAIConfig `json:"openai" koanf:"openai"`
}

// GenericLLMConfig contains common and reusable fields for LLM providers
type GenericLLMConfig struct {
	// Model specifies the model name to use
	Model string `json:"model" koanf:"model"`

	// APIKey contains the authentication key for the service
	APIKey string `json:"apiKey" koanf:"apikey" sensitive:"true"`
}

// AnthropicConfig contains Anthropic specific configuration
type AnthropicConfig struct {
	// BetaHeader specifies the beta API features to enable
	BetaHeader string `json:"betaheader" koanf:"betaheader"`

	// LegacyTextCompletion enables legacy text completion API
	LegacyTextCompletion bool `json:"legacytextcompletion" koanf:"legacytextcompletion"`

	// BaseURL specifies the API endpoint
	BaseURL string `json:"baseurl" koanf:"baseurl"`

	GenericLLMConfig
}

// CloudflareConfig contains Cloudflare specific configuration
type CloudflareConfig struct {
	GenericLLMConfig

	// AccountID specifies the Cloudflare account ID
	AccountID string `json:"accountid" koanf:"accountid"`

	// ServerURL specifies the API endpoint
	ServerURL string `json:"serverurl" koanf:"serverurl"`
}

// OpenAIConfig contains OpenAI specific configuration
type OpenAIConfig struct {
	GenericLLMConfig

	// URL specifies the API endpoint
	URL string `json:"url" koanf:"url"`

	// OrganizationID specifies the OpenAI organization ID
	OrganizationID string `json:"organizationid" koanf:"organizationid"`
}
