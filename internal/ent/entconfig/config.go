package entconfig

// Config holds the configuration for the ent server
type Config struct {
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entityTypes" koanf:"entityTypes" default:"" description:"entity types to create for the organization"`

	Summarizer Summarizer `json:"summarizer" koanf:"summarizer"`
}

type SummarizerType string

const (
	LexRank SummarizerType = "lexrank"
	LLM     SummarizerType = "llm"
)

type LLMProvider string

const (
	OpenAI    LLMProvider = "open_ai"
	Anthropic LLMProvider = "anthropic"
	Mistral   LLMProvider = "mistral"
)

type SummarizerLLM struct {
	Model    string      `json:"model" koanf:"model"`
	APIKey   string      `json:"apiKey" koanf:"apiKey"`
	Provider LLMProvider `json:"provider" koanf:"provider"`
}

type Summarizer struct {
	Type             SummarizerType `json:"type" koanf:"type" default:"lexrank"`
	LLM              SummarizerLLM  `json:"llm" koanf:"llm"`
	MaximumCharacter int            `json:"maximumCharacter" koanf:"maximumCharacter" default:50`
}
