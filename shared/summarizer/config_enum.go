package summarizer

import (
	"errors"
	"fmt"
)

const (
	// LLMProviderOpenai is a LLMProvider of type openai.
	LLMProviderOpenai LLMProvider = "openai"
	// LLMProviderAnthropic is a LLMProvider of type anthropic.
	LLMProviderAnthropic LLMProvider = "anthropic"
	// LLMProviderMistral is a LLMProvider of type mistral.
	LLMProviderMistral LLMProvider = "mistral"
	// LLMProviderGemini is a LLMProvider of type gemini.
	LLMProviderGemini LLMProvider = "gemini"
	// LLMProviderCloudflare is a LLMProvider of type cloudflare.
	LLMProviderCloudflare LLMProvider = "cloudflare"
	// LLMProviderHuggingface is a LLMProvider of type huggingface.
	LLMProviderHuggingface LLMProvider = "huggingface"
	// LLMProviderOllama is a LLMProvider of type ollama.
	LLMProviderOllama LLMProvider = "ollama"
)

var ErrInvalidLLMProvider = errors.New("not a valid LLMProvider")

// String implements the Stringer interface.
func (x LLMProvider) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x LLMProvider) IsValid() bool {
	_, err := ParseLLMProvider(string(x))
	return err == nil
}

var _LLMProviderValue = map[string]LLMProvider{
	"openai":      LLMProviderOpenai,
	"anthropic":   LLMProviderAnthropic,
	"mistral":     LLMProviderMistral,
	"gemini":      LLMProviderGemini,
	"cloudflare":  LLMProviderCloudflare,
	"huggingface": LLMProviderHuggingface,
	"ollama":      LLMProviderOllama,
}

// ParseLLMProvider attempts to convert a string to a LLMProvider.
func ParseLLMProvider(name string) (LLMProvider, error) {
	if x, ok := _LLMProviderValue[name]; ok {
		return x, nil
	}

	return LLMProvider(""), fmt.Errorf("%s is %w", name, ErrInvalidLLMProvider)
}

const (
	// TypeLexrank is a Type of type lexrank.
	TypeLexrank Type = "lexrank"
	// TypeLlm is a Type of type llm.
	TypeLlm Type = "llm"
)

var ErrInvalidType = errors.New("not a valid Type")

// String implements the Stringer interface.
func (x Type) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x Type) IsValid() bool {
	_, err := ParseType(string(x))
	return err == nil
}

var _TypeValue = map[string]Type{
	"lexrank": TypeLexrank,
	"llm":     TypeLlm,
}

// ParseType attempts to convert a string to a Type.
func ParseType(name string) (Type, error) {
	if x, ok := _TypeValue[name]; ok {
		return x, nil
	}

	return Type(""), fmt.Errorf("%s is %w", name, ErrInvalidType)
}
