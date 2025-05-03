package entconfig

// Config holds the configuration for the ent server
type Config struct {
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entityTypes" koanf:"entityTypes" default:"" description:"entity types to create for the organization"`

	Summarizer Summarizer `json:"summarizer" koanf:"summarizer"`
}

type Summarizer struct {
	Type             SummarizerType `json:"type" koanf:"type" default:"lexrank"`
	LLM              SummarizerLLM  `json:"llm" koanf:"llm"`
	MaximumCharacter int            `json:"maximumCharacter" koanf:"maximumCharacter" default:50`
}

// ENUM(lexrank,llm)
type SummarizerType string

// ENUM(openai,anthropic,mistral,gemini,cloudflare,huggingface,ollama)
type LLMProvider string

type SummarizerLLM struct {
	Provider LLMProvider `json:"provider" koanf:"provider"`
}

type genericLLMConfig struct {
	Model  string `json:"model" koanf:"model"`
	APIKey string `json:"apiKey" koanf:"apiKey"`
}

type GeminiConfig struct {
	genericLLMConfig
	CredentialsPath string
	CredentialsJSON string
	MaxTokens       int
}

type HuggingFaceConfig struct {
	genericLLMConfig
	URL string
}

type MistralConfig struct {
	genericLLMConfig
	URL string
}

type OllamaConfig struct {
	Model string `json:"model" koanf:"model"`
	URL   string
}

type AnthropicConfig struct {
	BetaHeader           string
	LegacyTextCompletion bool
	BaseURL              string
}

type CloudflareConfig struct {
	genericLLMConfig
	AccountID string
	ServerURL string
}

type OpenAIConfig struct {
	genericLLMConfig
	URL            string
	OrganizationID string
}
