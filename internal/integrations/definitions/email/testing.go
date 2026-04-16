package email

import (
	"github.com/theopenlane/newman/providers/mock"
)

// MockRuntimeConfig returns a RuntimeEmailConfig backed by the mock provider
// for use in integration test suites
func MockRuntimeConfig() *RuntimeEmailConfig {
	return &RuntimeEmailConfig{
		APIKey:         "mock-api-key",
		Provider:       ProviderMock,
		FromEmail:      "test@example.com",
		CompanyName:    "MITB Company",
		CompanyAddress: "123 Paper street",
		Corporation:    "MITB Corp",
		SupportEmail:   "support@example.com",
		LogoURL:        "https://example.com/logo.png",
		RootURL:        "https://example.com",
		ProductURL:     "https://app.example.com",
		DocsURL:        "https://docs.example.com",
	}
}

// MockSenderFromClient extracts the *mock.EmailSender from an EmailClient.
// Returns nil if the sender is not a mock
func MockSenderFromClient(client any) *mock.EmailSender {
	ec, ok := client.(*EmailClient)
	if !ok {
		return nil
	}

	ms, ok := ec.Sender.(*mock.EmailSender)
	if !ok {
		return nil
	}

	return ms
}
