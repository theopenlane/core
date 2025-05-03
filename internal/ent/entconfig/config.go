package entconfig

// Config holds the configuration for the ent server
type Config struct {
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entityTypes" koanf:"entityTypes" default:"" description:"entity types to create for the organization"`

	Summarizer Summarizer `json:"summarizer" koanf:"summarizer"`
}

type Summarizer struct {
	Type             SummarizerType    `json:"type" koanf:"type" default:"lexrank"`
	LLM              SummarizerLLM     `json:"llm" koanf:"llm"`
	MaximumSentences int               `json:"maximumSentences" koanf:"maximumSentences" default:"50"`
	Anthropic        AnthropicConfig   `json:"anthropic" koanf:"anthropic"`
	Mistral          MistralConfig     `json:"mistral" koanf:"mistral"`
	Gemini           GeminiConfig      `json:"gemini" koanf:"gemini"`
	HuggingFace      HuggingFaceConfig `json:"huggingFace" koanf:"huggingFace"`
	Ollama           OllamaConfig      `json:"ollama" koanf:"ollama"`
	Cloudflare       CloudflareConfig  `json:"cloudflare" koanf:"cloudflare"`
	OpenAI           OpenAIConfig      `json:"openai" koanf:"openai"`
}

// ENUM(lexrank,llm)
type SummarizerType string

// ENUM(openai,anthropic,mistral,gemini,cloudflare,huggingface,ollama)
type LLMProvider string

type SummarizerLLM struct {
	Provider LLMProvider `json:"provider" koanf:"provider"`
}

type GenericLLMConfig struct {
	Model  string `json:"model" koanf:"model"`
	APIKey string `json:"apiKey" koanf:"apiKey"`
}

type GeminiConfig struct {
	GenericLLMConfig
	CredentialsPath string `json:"credentialsPath" koanf:"credentialsPath"`
	CredentialsJSON string `json:"credentialsJSON" koanf:"credentialsJSON"`
	MaxTokens       int    `json:"maxTokens" koanf:"maxTokens"`
}

type HuggingFaceConfig struct {
	GenericLLMConfig
	URL string `json:"url" koanf:"url"`
}

type MistralConfig struct {
	GenericLLMConfig
	URL string `json:"url" koanf:"url"`
}

type OllamaConfig struct {
	Model string `json:"model" koanf:"model"`
	URL   string `json:"url" koanf:"url"`
}

type AnthropicConfig struct {
	BetaHeader           string `json:"betaHeader" koanf:"betaHeader"`
	LegacyTextCompletion bool   `json:"legacyTextCompletion" koanf:"legacyTextCompletion"`
	BaseURL              string `json:"baseURL" koanf:"baseURL"`
	GenericLLMConfig
}

type CloudflareConfig struct {
	GenericLLMConfig
	AccountID string `json:"accountID" koanf:"accountID"`
	ServerURL string `json:"serverURL" koanf:"serverURL"`
}

type OpenAIConfig struct {
	GenericLLMConfig
	URL            string `json:"url" koanf:"url"`
	OrganizationID string `json:"organizationID" koanf:"organizationID"`
}
