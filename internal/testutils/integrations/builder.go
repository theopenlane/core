//go:build test

package integrations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Builder returns the shared test integration definition
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "Openlane",
				DisplayName: "Test Integration",
				Description: "Shared test integration definition.",
				Category:    "system",
				Active:      true,
				Visible:     true,
			},
			Installation: installation.Registration(),
			Connections: []types.Connector{
				OAuth.
					Name("Test OAuth").
					Description("Authenticate through the OAuth callback fixture.").
					Authenticates(types.NewAuthFlow[oauthTokenCred](oauthStart, oauthComplete)).
					Provides(tokenClient(func(c oauthTokenCred) string { return c.AccessToken })).
					Verified(verifyAny[oauthTokenCred]).
					Disconnects("Remove the persisted OAuth credential and disconnect this installation.", nil),
				Token.
					Name("Test Token").
					Description("Connect with an API token validated by connection verification.").
					Provides(tokenClient(func(c tokenCred) string { return c.Token })).
					Verified(verifyAny[tokenCred]).
					Disconnects("Remove the persisted token credential and disconnect this installation.", nil),
				ServiceAccount.
					Name("Test Service Account").
					Description("Connect with a service account validated by connection verification.").
					Provides(tokenClient(func(serviceAccountCred) string { return "service-account" })).
					Verified(verifyAny[serviceAccountCred]).
					Disconnects("Remove the persisted service account credential and disconnect this installation.", nil),
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name:  "inbound.events",
					Event: webhookInboundEvent,
					Events: []types.WebhookEventRegistration{
						WebhookAlertCreated.Registration(types.WebhookEventRegistration{
							Handle: func(context.Context, types.WebhookHandleRequest) error { return nil },
						}),
					},
				},
			},
			Operations: []types.OperationRegistration{
				RepoSyncOp.Description("Async operation running with the built client").Registration(),
				ValidatedOp.Description("Inline operation with a required config field").Registration(),
				RecurringOp.Description("Healthy idle reconcile loop").Registration(),
				ExhaustingOp.Description("Always-failing reconcile loop for exhaustion").Registration(),
				UnresolvableOp.Description("Reconcile loop whose client cannot resolve without a stored credential").Registration(),
			},
		}, nil
	})
}
