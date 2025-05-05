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
	Summarize the following text in the line below. Be brief, concise and precise. 

	%s
	`

type llmSummarizer struct {
	llmClient llms.Model
}

func newLLMSummarizer(cfg entconfig.Config) (*llmSummarizer, error) {
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

func getClient(cfg entconfig.Config) (llms.Model, error) {
	switch cfg.Summarizer.LLM.Provider {
	case entconfig.LLMProviderAnthropic:
		return newAnthropicClient(cfg)
	case entconfig.LLMProviderCloudflare:
		return newCloudflareClient(cfg)
	case entconfig.LLMProviderMistral:
		return newMistralClient(cfg)
	case entconfig.LLMProviderGemini:
		return newGeminiClient(cfg)
	case entconfig.LLMProviderHuggingface:
		return newHuggingfaceClient(cfg)
	case entconfig.LLMProviderOllama:
		return newOllamaClient(cfg)
	case entconfig.LLMProviderOpenai:
		return newOpenAIClient(cfg)
	default:
		return nil, errors.New("unsupported llm model selected") // nolint:err113
	}
}

func newAnthropicClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []anthropic.Option{}

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
}

func newCloudflareClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []cloudflare.Option{}
	cfCfg := cfg.Summarizer.LLM.Cloudflare

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

func newMistralClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []mistral.Option{}
	mCfg := cfg.Summarizer.LLM.Mistral

	if mCfg.APIKey != "" {
		opts = append(opts, mistral.WithAPIKey(mCfg.APIKey))
	}

	if mCfg.Model != "" {
		opts = append(opts, mistral.WithModel(mCfg.Model))
	}

	if mCfg.URL != "" {
		opts = append(opts, mistral.WithEndpoint(mCfg.URL))
	}

	return mistral.New(opts...)
}

func newGeminiClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []googleai.Option{}
	gCfg := cfg.Summarizer.LLM.Gemini

	if gCfg.APIKey != "" {
		opts = append(opts, googleai.WithAPIKey(gCfg.APIKey))
	}

	if gCfg.Model != "" {
		opts = append(opts, googleai.WithDefaultModel(gCfg.Model))
	}

	if gCfg.MaxTokens != 0 {
		opts = append(opts, googleai.WithDefaultMaxTokens(gCfg.MaxTokens))
	}

	if gCfg.CredentialsJSON != "" {
		opts = append(opts, googleai.WithCredentialsJSON([]byte(gCfg.CredentialsJSON)))
	}

	if gCfg.CredentialsPath != "" {
		opts = append(opts, googleai.WithCredentialsFile(gCfg.CredentialsPath))
	}

	return googleai.New(context.Background(), opts...)
}

func newHuggingfaceClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []huggingface.Option{}
	hfCfg := cfg.Summarizer.LLM.HuggingFace

	if hfCfg.APIKey != "" {
		opts = append(opts, huggingface.WithToken(hfCfg.APIKey))
	}

	if hfCfg.Model != "" {
		opts = append(opts, huggingface.WithModel(hfCfg.Model))
	}

	if hfCfg.URL != "" {
		opts = append(opts, huggingface.WithURL(hfCfg.URL))
	}

	return huggingface.New(opts...)
}

func newOllamaClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []ollama.Option{}
	oCfg := cfg.Summarizer.LLM.Ollama

	if oCfg.Model != "" {
		opts = append(opts, ollama.WithModel(oCfg.Model))
	}

	if oCfg.URL != "" {
		opts = append(opts, ollama.WithServerURL(oCfg.URL))
	}

	return ollama.New(opts...)
}

func newOpenAIClient(cfg entconfig.Config) (llms.Model, error) {
	opts := []openai.Option{}
	oaiCfg := cfg.Summarizer.LLM.OpenAI

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
