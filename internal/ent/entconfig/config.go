package entconfig

// Config holds the configuration for the ent server
type Config struct {
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entityTypes" koanf:"entityTypes" default:"" description:"entity types to create for the organization"`

	// Summarizer contains configuration for text summarization
	Summarizer Summarizer `json:"summarizer" koanf:"summarizer"`
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

	// Mistral contains configuration for Mistral's API
	Mistral MistralConfig `json:"mistral" koanf:"mistral"`

	// Gemini contains configuration for Google's Gemini API
	Gemini GeminiConfig `json:"gemini" koanf:"gemini"`

	// HuggingFace contains configuration for HuggingFace's API
	HuggingFace HuggingFaceConfig `json:"huggingFace" koanf:"huggingFace"`

	// Ollama contains configuration for Ollama's API
	Ollama OllamaConfig `json:"ollama" koanf:"ollama"`

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
	APIKey string `json:"apiKey" koanf:"apiKey"`
}

// GeminiConfig contains Google Gemini specific configuration
type GeminiConfig struct {
	GenericLLMConfig

	// CredentialsPath is the path to Google Cloud credentials file
	CredentialsPath string `json:"credentialsPath" koanf:"credentialsPath"`

	// CredentialsJSON contains Google Cloud credentials as JSON string
	CredentialsJSON string `json:"credentialsJSON" koanf:"credentialsJSON"`

	// MaxTokens specifies the maximum tokens for response
	MaxTokens int `json:"maxTokens" koanf:"maxTokens"`
}

// HuggingFaceConfig contains HuggingFace specific configuration
type HuggingFaceConfig struct {
	GenericLLMConfig

	// URL specifies the API endpoint
	URL string `json:"url" koanf:"url"`
}

// MistralConfig contains Mistral specific configuration
type MistralConfig struct {
	GenericLLMConfig

	// URL specifies the API endpoint
	URL string `json:"url" koanf:"url"`
}

// OllamaConfig contains Ollama specific configuration
type OllamaConfig struct {
	// Model specifies the model to use
	Model string `json:"model" koanf:"model"`

	// URL specifies the API endpoint
	URL string `json:"url" koanf:"url"`
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
