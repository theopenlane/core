package email

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/mock"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestCustomerClientBuilder_Resend(t *testing.T) {
	client, err := customerClientBuilder(resendSender)(context.Background(), types.ConnectionRequest[ResendCredential]{
		Credential: ResendCredential{APIKey: "k"},
	})
	assert.NilError(t, err)
	assert.Check(t, client.Sender != nil)
}

func TestCustomerClientBuilder_Postmark(t *testing.T) {
	client, err := customerClientBuilder(postmarkSender)(context.Background(), types.ConnectionRequest[PostmarkCredential]{
		Credential: PostmarkCredential{ServerToken: "k", MessageStream: "broadcast", TrackOpens: true, TrackLinks: "HtmlAndText"},
	})
	assert.NilError(t, err)
	assert.Check(t, client.Sender != nil)
}

func TestCustomerClientBuilder_SenderError(t *testing.T) {
	client, err := customerClientBuilder(func(ResendCredential) (newman.EmailSender, error) {
		return nil, errors.New("boom")
	})(context.Background(), types.ConnectionRequest[ResendCredential]{})

	assert.ErrorIs(t, err, ErrClientBuildFailed)
	assert.Check(t, is.Nil(client))
}

func TestEmailScrubber_NotNil(t *testing.T) {
	assert.Check(t, Scrubber() != nil)
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
	assert.NilError(t, err)

	built, err := runtimeClientBuilder(true)(context.Background(), raw)
	assert.NilError(t, err)

	client, ok := built.(*Client)
	assert.Assert(t, ok)
	assert.Assert(t, client.Sender != nil)

	welcome := testDispatcher[WelcomeRequest](t, "WelcomeRequest")

	err = welcome.SendByKey(context.Background(), types.OperationRequest{}, client, mustMarshal(t, WelcomeRequest{
		RecipientInfo: RecipientInfo{
			Email:     "dev@example.com",
			FirstName: "Dev",
		},
	}))
	assert.NilError(t, err)

	recipientDir := filepath.Join(testDir, "dev@example.com")
	entries, err := os.ReadDir(recipientDir)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(entries, 1))
	assert.Check(t, filepath.Ext(entries[0].Name()) == ".mim")
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
	assert.NilError(t, err)

	built, err := runtimeClientBuilder(false)(context.Background(), raw)
	assert.NilError(t, err)

	client, ok := built.(*Client)
	assert.Assert(t, ok)
	assert.Assert(t, client.Sender != nil)
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	assert.NilError(t, err)

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
				assert.ErrorIs(t, err, tc.wantErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, sender != nil)

			if tc.want != nil {
				assert.Check(t, is.Equal(reflect.TypeOf(tc.want), reflect.TypeOf(sender)))
			}
		})
	}
}
