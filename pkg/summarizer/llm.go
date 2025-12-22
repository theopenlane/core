package summarizer

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/cloudflare"
	"github.com/tmc/langchaingo/llms/openai"
)

// maybe make a config value too?
const prompt = `
	Summarize the following text in the line below. Be brief, concise and precise.

	%s
	`

type llmSummarizer struct {
	llmClient llms.Model
}

func newLLMSummarizer(cfg Config) (*llmSummarizer, error) {
	client, err := getClient(cfg)
	if err != nil {
		return nil, err
	}

	return &llmSummarizer{
		llmClient: client,
	}, nil
}

// Summarize returns a shortened version of the provided string using the selected llm
func (l *llmSummarizer) Summarize(ctx context.Context, s string) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l.llmClient, fmt.Sprintf(prompt, s))
}

func getClient(cfg Config) (llms.Model, error) {
	switch cfg.LLM.Provider {
	case LLMProviderAnthropic:
		return newAnthropicClient(cfg)
	case LLMProviderCloudflare:
		return newCloudflareClient(cfg)
	case LLMProviderOpenai:
		return newOpenAIClient(cfg)
	default:
		return nil, errors.New("unsupported llm model selected") //nolint:err113
	}
}

func newAnthropicClient(cfg Config) (llms.Model, error) {
	opts := []anthropic.Option{}
	aCfg := cfg.LLM.Anthropic

	if aCfg.APIKey != "" {
		opts = append(opts, anthropic.WithToken(aCfg.APIKey))
	}

	if aCfg.Model != "" {
		opts = append(opts, anthropic.WithModel(aCfg.Model))
	}

	if aCfg.BaseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(aCfg.BaseURL))
	}

	if aCfg.BetaHeader != "" {
		opts = append(opts, anthropic.WithAnthropicBetaHeader(aCfg.BetaHeader))
	}

	if aCfg.LegacyTextCompletion {
		opts = append(opts, anthropic.WithLegacyTextCompletionsAPI())
	}

	return anthropic.New(opts...)
}

func newCloudflareClient(cfg Config) (llms.Model, error) {
	opts := []cloudflare.Option{}
	cfCfg := cfg.LLM.Cloudflare

	if cfCfg.APIKey != "" {
		opts = append(opts, cloudflare.WithToken(cfCfg.APIKey))
	}

	if cfCfg.AccountID != "" {
		opts = append(opts, cloudflare.WithAccountID(cfCfg.AccountID))
	}

	if cfCfg.Model != "" {
		opts = append(opts, cloudflare.WithModel(cfCfg.Model))
	}

	if cfCfg.ServerURL != "" {
		opts = append(opts, cloudflare.WithServerURL(cfCfg.ServerURL))
	}

	return cloudflare.New(opts...)
}

func newOpenAIClient(cfg Config) (llms.Model, error) {
	opts := []openai.Option{}
	oaiCfg := cfg.LLM.OpenAI

	if oaiCfg.APIKey != "" {
		opts = append(opts, openai.WithToken(oaiCfg.APIKey))
	}

	if oaiCfg.Model != "" {
		opts = append(opts, openai.WithModel(oaiCfg.Model))
	}

	if oaiCfg.URL != "" {
		opts = append(opts, openai.WithBaseURL(oaiCfg.URL))
	}

	if oaiCfg.OrganizationID != "" {
		opts = append(opts, openai.WithOrganization(oaiCfg.OrganizationID))
	}

	return openai.New(opts...)
}
