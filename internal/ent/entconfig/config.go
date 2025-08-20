package entconfig

// Config holds the configuration for the ent server
type Config struct {
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entityTypes" koanf:"entityTypes" default:"" description:"entity types to create for the organization"`
	// Summarizer contains configuration for text summarization
	Summarizer Summarizer `json:"summarizer" koanf:"summarizer"`
	// Windmill contains configuration for Windmill workflow automation
	Windmill Windmill `json:"windmill" koanf:"windmill"`
	// MaxPoolSize is the max pond pool workers that can be used by the ent client
	MaxPoolSize int `json:"maxPoolSize" koanf:"maxPoolSize" default:"100"`
	// Modules contains the configuration for the module system
	Modules Modules `json:"modules" koanf:"modules"`
}

// Summarizer holds configuration for the text summarization functionality
type Summarizer struct {
	// Type specifies the summarization algorithm to use
	Type SummarizerType `json:"type" koanf:"type" default:"lexrank"`
	// LLM contains configuration for large language model based summarization
	LLM SummarizerLLM `json:"llm" koanf:"llm"`
	// MaximumSentences specifies the maximum number of sentences in the summary
	MaximumSentences int `json:"maximumSentences" koanf:"maximumSentences" default:"60"`
}

// SummarizerType defines the type of summarization algorithm
// ENUM(lexrank,llm)
type SummarizerType string

// LLMProvider defines supported LLM service providers
// ENUM(openai,anthropic,mistral,gemini,cloudflare,huggingface,ollama)
type LLMProvider string

// SummarizerLLM contains configuration for multiple LLM providers
type SummarizerLLM struct {
	// Provider specifies which LLM service to use
	Provider LLMProvider `json:"provider" koanf:"provider"`

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
	APIKey string `json:"apiKey" koanf:"apiKey" sensitive:"true"`
}

// AnthropicConfig contains Anthropic specific configuration
type AnthropicConfig struct {
	// BetaHeader specifies the beta API features to enable
	BetaHeader string `json:"betaHeader" koanf:"betaHeader"`

	// LegacyTextCompletion enables legacy text completion API
	LegacyTextCompletion bool `json:"legacyTextCompletion" koanf:"legacyTextCompletion"`

	// BaseURL specifies the API endpoint
	BaseURL string `json:"baseURL" koanf:"baseURL"`

	GenericLLMConfig
}

// CloudflareConfig contains Cloudflare specific configuration
type CloudflareConfig struct {
	GenericLLMConfig

	// AccountID specifies the Cloudflare account ID
	AccountID string `json:"accountID" koanf:"accountID"`

	// ServerURL specifies the API endpoint
	ServerURL string `json:"serverURL" koanf:"serverURL"`
}

// OpenAIConfig contains OpenAI specific configuration
type OpenAIConfig struct {
	GenericLLMConfig

	// URL specifies the API endpoint
	URL string `json:"url" koanf:"url"`

	// OrganizationID specifies the OpenAI organization ID
	OrganizationID string `json:"organizationID" koanf:"organizationID"`
}

// Windmill holds configuration for the Windmill workflow automation platform
type Windmill struct {
	// Enabled specifies whether Windmill integration is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`

	// BaseURL is the base URL of the Windmill instance
	BaseURL string `json:"baseURL" koanf:"baseURL" default:"https://app.windmill.dev"`

	// Workspace is the Windmill workspace to use
	Workspace string `json:"workspace" koanf:"workspace"`

	// Token is the API token for authentication with Windmill
	Token string `json:"token" koanf:"token" sensitive:"true"`

	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout string `json:"defaultTimeout" koanf:"defaultTimeout" default:"30s"`

	// Timezone for scheduled jobs
	Timezone string `json:"timezone" koanf:"timezone" default:"UTC"`

	// OnFailureScript script to run when a scheduled job fails
	OnFailureScript string `json:"onFailureScript" koanf:"onFailureScript"`

	// OnSuccessScript script to run when a scheduled job succeeds
	OnSuccessScript string `json:"onSuccessScript," koanf:"onSuccessScript"`
}

// Modules settings for features access
type Modules struct {
	// Enabled indicates whether to check and verify module access
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
}
