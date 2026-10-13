package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// TestStaticWebhooksFollowSelectedProvider registers the email definition under each provider and checks the registry exposes only that provider's inbound route
func TestStaticWebhooksFollowSelectedProvider(t *testing.T) {
	t.Parallel()

	resendRoute := types.StaticWebhookEntry{DefinitionID: DefinitionID.ID(), WebhookName: resendWebhookRef.Name(), StaticRoute: "/email/webhook"}
	postmarkRoute := types.StaticWebhookEntry{DefinitionID: DefinitionID.ID(), WebhookName: postmarkWebhookRef.Name(), StaticRoute: "/email/webhook/postmark"}

	tests := []struct {
		name string
		cfg  *RuntimeEmailConfig
		want []types.StaticWebhookEntry
	}{
		{name: "resend selected", cfg: MockProviderRuntimeConfig(providerResend), want: []types.StaticWebhookEntry{resendRoute}},
		{name: "postmark selected", cfg: MockProviderRuntimeConfig(providerPostmark), want: []types.StaticWebhookEntry{postmarkRoute}},
		{name: "mock provider registers no inbound route", cfg: MockRuntimeConfig(), want: nil},
		{name: "unset provider registers no inbound route", cfg: &RuntimeEmailConfig{FromEmail: "test@example.com"}, want: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			require.NoError(t, reg.RegisterAll(Builder(tc.cfg, false)))

			assert.Equal(t, tc.want, reg.StaticWebhooks())
		})
	}
}
