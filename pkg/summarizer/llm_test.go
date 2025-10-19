package summarizer

import (
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestLLM_Summarize(t *testing.T) {

	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		t.Skip("Skipping llm tests as open ai key not present")
	}

	tt := []struct {
		name     string
		sentence string
		hasError bool
	}{
		{
			name:     "empty string",
			hasError: true,
		},
		{
			name:     "short string",
			sentence: gofakeit.LoremIpsumSentence(200),
		},
		{
			name:     "long string",
			sentence: gofakeit.LoremIpsumSentence(1000),
		},
		{
			name:     "really long string",
			sentence: gofakeit.LoremIpsumSentence(10000),
		},
	}

	summarizer, err := newLLMSummarizer(Config{
		Type: TypeLlm,
		LLM: LLM{
			Provider: LLMProviderOpenai,
			OpenAI: OpenAIConfig{
				GenericLLMConfig: GenericLLMConfig{
					Model:  "gpt-4",
					APIKey: apiKey,
				},
			},
		},
	})

	require.NoError(t, err)

	for _, v := range tt {
		t.Run("Test "+v.name, func(t *testing.T) {

			summarized, err := summarizer.Summarize(t.Context(), v.sentence)
			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, summarized)
			require.NotEqual(t, summarized, v.sentence)
		})
	}
}

func TestNewLLMSummarizer(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "anthropic with all options",
			cfg: Config{
				LLM: LLM{
					Provider: LLMProviderAnthropic,
					Anthropic: AnthropicConfig{
						GenericLLMConfig: GenericLLMConfig{
							Model:  "claude-2",
							APIKey: "test-key",
						},
						BetaHeader:           "beta-header",
						LegacyTextCompletion: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "cloudflare with all options",
			cfg: Config{
				LLM: LLM{
					Provider: LLMProviderCloudflare,
					Cloudflare: CloudflareConfig{
						GenericLLMConfig: GenericLLMConfig{
							Model:  "@cf/meta/llama-2-7b-chat-int8",
							APIKey: "test-key",
						},
						AccountID: "account-id",
						ServerURL: "https://api.cloudflare.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "openai with all options",
			cfg: Config{
				LLM: LLM{
					Provider: LLMProviderOpenai,
					OpenAI: OpenAIConfig{
						GenericLLMConfig: GenericLLMConfig{
							Model:  "gpt-4",
							APIKey: "test-key",
						},
						OrganizationID: "org-123",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported provider",
			cfg: Config{
				LLM: LLM{
					Provider: "unsupported",
				},
			},
			wantErr: true,
		},
		{
			name: "missing required api key for anthropic",
			cfg: Config{
				LLM: LLM{
					Provider: LLMProviderAnthropic,
					Anthropic: AnthropicConfig{
						GenericLLMConfig: GenericLLMConfig{
							Model: "claude-2",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summarizer, err := newLLMSummarizer(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, summarizer)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, summarizer)
		})
	}
}
