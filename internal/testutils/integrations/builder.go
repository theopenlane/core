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
			CredentialRegistrations: []types.CredentialRegistration{
				TokenCredential.Registration(types.CredentialRegistration{
					Name:        "Test Token",
					Description: "API token the test client is built from.",
				}),
				OAuthCredential.Registration(types.CredentialRegistration{
					Name:        "Test OAuth",
					Description: "Auth-managed credential slot filled by the OAuth fixture.",
				}),
				ServiceAccountCredential.Registration(types.CredentialRegistration{
					Name:        "Test Service Account",
					Description: "Strict-schema credential slot used by config flows.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: OAuthCredential.ID(),
					Name:          "Test OAuth",
					Description:   "Authenticate through the OAuth callback fixture.",
					Auth: &types.AuthRegistration{
						CredentialRef: OAuthCredential.ID(),
						Start:         oauthStart,
						Complete:      oauthComplete,
					},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: OAuthCredential.ID(),
						Description:   "Remove the persisted OAuth credential and disconnect this installation.",
					},
				},
				{
					CredentialRef: TokenCredential.ID(),
					Name:          "Test Token",
					Description:   "Connect with an API token validated by the health check.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: TokenCredential.ID(),
						Description:   "Remove the persisted token credential and disconnect this installation.",
					},
				},
				{
					CredentialRef: ServiceAccountCredential.ID(),
					Name:          "Test Service Account",
					Description:   "Connect with a service account validated by the health check.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: ServiceAccountCredential.ID(),
						Description:   "Remove the persisted service account credential and disconnect this installation.",
					},
				},
			},
			HealthCheck: types.CredentialHealthCheck(healthHandler),
			Clients: []types.ClientRegistration{
				testClient.Registration(tokenClient(TokenCredential, func(c tokenCred) string { return c.Token }), types.ClientRegistration{
					Description: "Test client built from the stored token credential.",
				}),
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name:  "inbound.events",
					Event: webhookInboundEvent,
					Events: []types.WebhookEventRegistration{
						WebhookAlertCreated.Registration(DefinitionID, types.WebhookEventRegistration{
							Handle: func(context.Context, types.WebhookHandleRequest) error { return nil },
						}),
					},
				},
			},
			Operations: []types.OperationRegistration{
				RepoSyncOp.Description("Async operation running with the built client").Registration(DefinitionID),
				ValidatedOp.Description("Inline operation with a required config field").Registration(DefinitionID),
				RecurringOp.Description("Healthy idle reconcile loop").Registration(DefinitionID),
				ExhaustingOp.Description("Always-failing reconcile loop for exhaustion").Registration(DefinitionID),
				UnresolvableOp.Description("Reconcile loop whose client cannot resolve without a stored credential").Registration(DefinitionID),
			},
		}, nil
	})
}
