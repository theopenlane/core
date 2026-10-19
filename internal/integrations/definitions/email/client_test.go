package email

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestCustomerClientBuilder_Resend(t *testing.T) {
	client, err := customerClientBuilder(resendSender)(context.Background(), types.ConnectionRequest[ResendCredential]{
		Credential: ResendCredential{APIKey: "k"},
	})
	require.NoError(t, err)
	assert.NotNil(t, client.Sender)
}

func TestCustomerClientBuilder_Postmark(t *testing.T) {
	client, err := customerClientBuilder(postmarkSender)(context.Background(), types.ConnectionRequest[PostmarkCredential]{
		Credential: PostmarkCredential{ServerToken: "k", MessageStream: "broadcast", TrackOpens: true, TrackLinks: "HtmlAndText"},
	})
	require.NoError(t, err)
	assert.NotNil(t, client.Sender)
}

func TestCustomerClientBuilder_SenderError(t *testing.T) {
	client, err := customerClientBuilder(func(ResendCredential) (newman.EmailSender, error) {
		return nil, errors.New("boom")
	})(context.Background(), types.ConnectionRequest[ResendCredential]{})

	require.ErrorIs(t, err, ErrClientBuildFailed)
	assert.Nil(t, client)
}

func TestEmailScrubber_NotNil(t *testing.T) {
	assert.NotNil(t, Scrubber())
}

func TestDevModeRuntimeClient_SendsWelcomeEmail(t *testing.T) {
	t.Parallel()

	testDir := t.TempDir()

	cfg := RuntimeEmailConfig{
		Provider:    "resend",
		FromEmail:   "test@example.com",
		CompanyName: "TestCo",
		ProductURL:  "https://app.example.com",
		DocsURL:     "https://docs.example.com",
		TestDir:     testDir,
	}

	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := runtimeClientBuilder(true)(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*Client)
	require.True(t, ok)
	require.NotNil(t, client.Sender)

	welcome := testDispatcher[WelcomeRequest](t, "WelcomeRequest")

	err = welcome.SendByKey(context.Background(), types.OperationRequest{}, client, mustMarshal(t, WelcomeRequest{
		RecipientInfo: RecipientInfo{
			Email:     "dev@example.com",
			FirstName: "Dev",
		},
	}))
	require.NoError(t, err)

	recipientDir := filepath.Join(testDir, "dev@example.com")
	entries, err := os.ReadDir(recipientDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, filepath.Ext(entries[0].Name()) == ".mim")
}

func TestDevModeRuntimeClient_ProvisionedIgnoresDevMode(t *testing.T) {
	t.Parallel()

	cfg := RuntimeEmailConfig{
		Provider:    "resend",
		Resend:      ResendConfig{ResendCredential: ResendCredential{APIKey: "real-key"}},
		FromEmail:   "prod@example.com",
		CompanyName: "ProdCo",
		TestDir:     t.TempDir(),
	}

	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := runtimeClientBuilder(false)(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*Client)
	require.True(t, ok)
	require.NotNil(t, client.Sender)
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}

func TestRuntimeSender(t *testing.T) {
	tests := []struct {
		name    string
		cfg     RuntimeEmailConfig
		devMode bool
		want    any
		wantErr error
	}{
		{name: "resend", cfg: RuntimeEmailConfig{Provider: "resend", Resend: ResendConfig{ResendCredential: ResendCredential{APIKey: "key"}}}, want: nil},
		{name: "postmark", cfg: RuntimeEmailConfig{Provider: "postmark", Postmark: PostmarkConfig{PostmarkCredential: PostmarkCredential{ServerToken: "token", MessageStream: "outbound", TrackOpens: true}}}, want: nil},
		{name: "mock", cfg: RuntimeEmailConfig{Provider: ProviderMock}, want: &mock.EmailSender{}},
		{name: "unsupported", cfg: RuntimeEmailConfig{Provider: "sendgrid"}, wantErr: ErrProviderNotSupported},
		{name: "empty provider", cfg: RuntimeEmailConfig{}, wantErr: ErrProviderNotSupported},
		{name: "dev mode resend", cfg: RuntimeEmailConfig{Provider: "resend", TestDir: t.TempDir()}, devMode: true, want: &mock.EmailSender{}},
		{name: "dev mode postmark", cfg: RuntimeEmailConfig{Provider: "postmark", TestDir: t.TempDir()}, devMode: true, want: &mock.EmailSender{}},
		{name: "dev mode unsupported", cfg: RuntimeEmailConfig{Provider: "sendgrid", TestDir: t.TempDir()}, devMode: true, want: &mock.EmailSender{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sender, err := runtimeSender(tc.cfg, tc.devMode)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, sender)

			if tc.want != nil {
				assert.IsType(t, tc.want, sender)
			}
		})
	}
}
