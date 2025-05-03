package summarizer

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/cloudflare"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/theopenlane/core/internal/ent/entconfig"
)

// maybe make a config value too?
const prompt = `
	Summarise the following text in the line below. Be brief, concise and precise. 

	%s
	`

type LLMSummarizer struct {
	llmClient llms.Model
}

func newLLMSummarizer(cfg entconfig.Config) (*LLMSummarizer, error) {
	client, err := getClient(cfg)
	if err != nil {
		return nil, err
	}

	return &LLMSummarizer{
		llmClient: client,
	}, nil
}

// Summarize returns a shortened version of the provided string using the selected llm
func (l *LLMSummarizer) Summarize(ctx context.Context, s string) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l.llmClient, fmt.Sprintf(prompt, s))
}

func getClient(cfg entconfig.Config) (llms.Model, error) {
	switch cfg.Summarizer.LLM.Provider {
	case entconfig.LLMProviderAnthropic:
		opts := make([]anthropic.Option, 0)

		if cfg.Summarizer.Anthropic.APIKey != "" {
			opts = append(opts, anthropic.WithToken(cfg.Summarizer.Anthropic.APIKey))
		}

		if cfg.Summarizer.Anthropic.Model != "" {
			opts = append(opts, anthropic.WithModel(cfg.Summarizer.Anthropic.Model))
		}

		if cfg.Summarizer.Anthropic.BaseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(cfg.Summarizer.Anthropic.BaseURL))
		}

		if cfg.Summarizer.Anthropic.BetaHeader != "" {
			opts = append(opts, anthropic.WithAnthropicBetaHeader(cfg.Summarizer.Anthropic.BetaHeader))
		}

		if cfg.Summarizer.Anthropic.LegacyTextCompletion {
			opts = append(opts, anthropic.WithLegacyTextCompletionsAPI())
		}

		return anthropic.New(opts...)

	case entconfig.LLMProviderCloudflare:
		opts := make([]cloudflare.Option, 0)

		if cfg.Summarizer.Cloudflare.APIKey != "" {
			opts = append(opts, cloudflare.WithToken(cfg.Summarizer.Cloudflare.APIKey))
		}

		if cfg.Summarizer.Cloudflare.AccountID != "" {
			opts = append(opts, cloudflare.WithAccountID(cfg.Summarizer.Cloudflare.AccountID))
		}

		if cfg.Summarizer.Cloudflare.Model != "" {
			opts = append(opts, cloudflare.WithModel(cfg.Summarizer.Cloudflare.Model))
		}

		if cfg.Summarizer.Cloudflare.ServerURL != "" {
			opts = append(opts, cloudflare.WithServerURL(cfg.Summarizer.Cloudflare.ServerURL))
		}

		return cloudflare.New(opts...)

	case entconfig.LLMProviderMistral:
		opts := make([]mistral.Option, 0)

		if cfg.Summarizer.Mistral.APIKey != "" {
			opts = append(opts, mistral.WithAPIKey(cfg.Summarizer.Mistral.APIKey))
		}

		if cfg.Summarizer.Mistral.Model != "" {
			opts = append(opts, mistral.WithModel(cfg.Summarizer.Mistral.Model))
		}

		if cfg.Summarizer.Mistral.URL != "" {
			opts = append(opts, mistral.WithEndpoint(cfg.Summarizer.Mistral.URL))
		}

		return mistral.New(opts...)

	case entconfig.LLMProviderGemini:
		opts := make([]googleai.Option, 0)
		if cfg.Summarizer.Gemini.APIKey != "" {
			opts = append(opts, googleai.WithAPIKey(cfg.Summarizer.Gemini.APIKey))
		}

		if cfg.Summarizer.Gemini.Model != "" {
			opts = append(opts, googleai.WithDefaultModel(cfg.Summarizer.Gemini.Model))
		}

		if cfg.Summarizer.Gemini.MaxTokens != 0 {
			opts = append(opts, googleai.WithDefaultMaxTokens(cfg.Summarizer.Gemini.MaxTokens))
		}

		if cfg.Summarizer.Gemini.CredentialsJSON != "" {
			opts = append(opts, googleai.WithCredentialsJSON([]byte(cfg.Summarizer.Gemini.CredentialsJSON)))
		}

		if cfg.Summarizer.Gemini.CredentialsPath != "" {
			opts = append(opts, googleai.WithCredentialsFile(cfg.Summarizer.Gemini.CredentialsPath))
		}

		return googleai.New(context.Background(), opts...)

	case entconfig.LLMProviderHuggingface:
		opts := make([]huggingface.Option, 0)

		if cfg.Summarizer.HuggingFace.APIKey != "" {
			opts = append(opts, huggingface.WithToken(cfg.Summarizer.HuggingFace.APIKey))
		}

		if cfg.Summarizer.HuggingFace.Model != "" {
			opts = append(opts, huggingface.WithModel(cfg.Summarizer.HuggingFace.Model))
		}

		if cfg.Summarizer.HuggingFace.URL != "" {
			opts = append(opts, huggingface.WithURL(cfg.Summarizer.HuggingFace.URL))
		}

		return huggingface.New(opts...)

	case entconfig.LLMProviderOllama:
		opts := make([]ollama.Option, 0)

		if cfg.Summarizer.Ollama.Model != "" {
			opts = append(opts, ollama.WithModel(cfg.Summarizer.Ollama.Model))
		}

		if cfg.Summarizer.Ollama.URL != "" {
			opts = append(opts, ollama.WithServerURL(cfg.Summarizer.Ollama.URL))
		}

		return ollama.New(opts...)

	case entconfig.LLMProviderOpenai:
		opts := make([]openai.Option, 0)

		if cfg.Summarizer.OpenAI.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.Summarizer.OpenAI.APIKey))
		}

		if cfg.Summarizer.OpenAI.Model != "" {
			opts = append(opts, openai.WithModel(cfg.Summarizer.OpenAI.Model))
		}

		if cfg.Summarizer.OpenAI.URL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.Summarizer.OpenAI.URL))
		}

		if cfg.Summarizer.OpenAI.OrganizationID != "" {
			opts = append(opts, openai.WithOrganization(cfg.Summarizer.OpenAI.OrganizationID))
		}

		return openai.New(opts...)

	default:
		return nil, errors.New("unsupported llm model selected")
	}
}
