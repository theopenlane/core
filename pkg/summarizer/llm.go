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

		if cfg.Summarizer.LLM.Anthropic.APIKey != "" {
			opts = append(opts, anthropic.WithToken(cfg.Summarizer.LLM.Anthropic.APIKey))
		}

		if cfg.Summarizer.LLM.Anthropic.Model != "" {
			opts = append(opts, anthropic.WithModel(cfg.Summarizer.LLM.Anthropic.Model))
		}

		if cfg.Summarizer.LLM.Anthropic.BaseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(cfg.Summarizer.LLM.Anthropic.BaseURL))
		}

		if cfg.Summarizer.LLM.Anthropic.BetaHeader != "" {
			opts = append(opts, anthropic.WithAnthropicBetaHeader(cfg.Summarizer.LLM.Anthropic.BetaHeader))
		}

		if cfg.Summarizer.LLM.Anthropic.LegacyTextCompletion {
			opts = append(opts, anthropic.WithLegacyTextCompletionsAPI())
		}

		return anthropic.New(opts...)

	case entconfig.LLMProviderCloudflare:
		opts := make([]cloudflare.Option, 0)

		if cfg.Summarizer.LLM.Cloudflare.APIKey != "" {
			opts = append(opts, cloudflare.WithToken(cfg.Summarizer.LLM.Cloudflare.APIKey))
		}

		if cfg.Summarizer.LLM.Cloudflare.AccountID != "" {
			opts = append(opts, cloudflare.WithAccountID(cfg.Summarizer.LLM.Cloudflare.AccountID))
		}

		if cfg.Summarizer.LLM.Cloudflare.Model != "" {
			opts = append(opts, cloudflare.WithModel(cfg.Summarizer.LLM.Cloudflare.Model))
		}

		if cfg.Summarizer.LLM.Cloudflare.ServerURL != "" {
			opts = append(opts, cloudflare.WithServerURL(cfg.Summarizer.LLM.Cloudflare.ServerURL))
		}

		return cloudflare.New(opts...)

	case entconfig.LLMProviderMistral:
		opts := make([]mistral.Option, 0)

		if cfg.Summarizer.LLM.Mistral.APIKey != "" {
			opts = append(opts, mistral.WithAPIKey(cfg.Summarizer.LLM.Mistral.APIKey))
		}

		if cfg.Summarizer.LLM.Mistral.Model != "" {
			opts = append(opts, mistral.WithModel(cfg.Summarizer.LLM.Mistral.Model))
		}

		if cfg.Summarizer.LLM.Mistral.URL != "" {
			opts = append(opts, mistral.WithEndpoint(cfg.Summarizer.LLM.Mistral.URL))
		}

		return mistral.New(opts...)

	case entconfig.LLMProviderGemini:
		opts := make([]googleai.Option, 0)
		if cfg.Summarizer.LLM.Gemini.APIKey != "" {
			opts = append(opts, googleai.WithAPIKey(cfg.Summarizer.LLM.Gemini.APIKey))
		}

		if cfg.Summarizer.LLM.Gemini.Model != "" {
			opts = append(opts, googleai.WithDefaultModel(cfg.Summarizer.LLM.Gemini.Model))
		}

		if cfg.Summarizer.LLM.Gemini.MaxTokens != 0 {
			opts = append(opts, googleai.WithDefaultMaxTokens(cfg.Summarizer.LLM.Gemini.MaxTokens))
		}

		if cfg.Summarizer.LLM.Gemini.CredentialsJSON != "" {
			opts = append(opts, googleai.WithCredentialsJSON([]byte(cfg.Summarizer.LLM.Gemini.CredentialsJSON)))
		}

		if cfg.Summarizer.LLM.Gemini.CredentialsPath != "" {
			opts = append(opts, googleai.WithCredentialsFile(cfg.Summarizer.LLM.Gemini.CredentialsPath))
		}

		return googleai.New(context.Background(), opts...)

	case entconfig.LLMProviderHuggingface:
		opts := make([]huggingface.Option, 0)

		if cfg.Summarizer.LLM.HuggingFace.APIKey != "" {
			opts = append(opts, huggingface.WithToken(cfg.Summarizer.LLM.HuggingFace.APIKey))
		}

		if cfg.Summarizer.LLM.HuggingFace.Model != "" {
			opts = append(opts, huggingface.WithModel(cfg.Summarizer.LLM.HuggingFace.Model))
		}

		if cfg.Summarizer.LLM.HuggingFace.URL != "" {
			opts = append(opts, huggingface.WithURL(cfg.Summarizer.LLM.HuggingFace.URL))
		}

		return huggingface.New(opts...)

	case entconfig.LLMProviderOllama:
		opts := make([]ollama.Option, 0)

		if cfg.Summarizer.LLM.Ollama.Model != "" {
			opts = append(opts, ollama.WithModel(cfg.Summarizer.LLM.Ollama.Model))
		}

		if cfg.Summarizer.LLM.Ollama.URL != "" {
			opts = append(opts, ollama.WithServerURL(cfg.Summarizer.LLM.Ollama.URL))
		}

		return ollama.New(opts...)

	case entconfig.LLMProviderOpenai:
		opts := make([]openai.Option, 0)

		if cfg.Summarizer.LLM.OpenAI.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.Summarizer.LLM.OpenAI.APIKey))
		}

		if cfg.Summarizer.LLM.OpenAI.Model != "" {
			opts = append(opts, openai.WithModel(cfg.Summarizer.LLM.OpenAI.Model))
		}

		if cfg.Summarizer.LLM.OpenAI.URL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.Summarizer.LLM.OpenAI.URL))
		}

		if cfg.Summarizer.LLM.OpenAI.OrganizationID != "" {
			opts = append(opts, openai.WithOrganization(cfg.Summarizer.LLM.OpenAI.OrganizationID))
		}

		return openai.New(opts...)

	default:
		return nil, errors.New("unsupported llm model selected")
	}
}
