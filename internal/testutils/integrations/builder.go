//go:build test

package integrations

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the shared test integration definition; only the reconcile operations are
// input-gated, since seeding sweeps every reconcile-policy operation on a connected installation
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
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				TokenCredential.Registration(types.CredentialRegistration{
					Name:        "Test Token",
					Description: "API token the test client is built from.",
					Schema:      TokenCredential.Schema(),
				}),
				OAuthCredential.Registration(types.CredentialRegistration{
					Name:        "Test OAuth",
					Description: "Auth-managed credential slot filled by the OAuth fixture.",
				}),
				ServiceAccountCredential.Registration(types.CredentialRegistration{
					Name:        "Test Service Account",
					Description: "Strict-schema credential slot used by config flows.",
					Schema:      ServiceAccountCredential.Schema(),
				}),
			},
			Connections: []types.ConnectionRegistration{
				oauthConnection.Registration(types.ConnectionRegistration{
					Name:           "Test OAuth",
					Description:    "Authenticate through the OAuth callback fixture.",
					CredentialRefs: []types.CredentialSlotID{OAuthCredential.ID()},
					Auth: &types.AuthRegistration{
						CredentialRef: OAuthCredential.ID(),
						Schema:        OAuthCredential.Schema(),
						Start:         oauthStart,
						Complete:      oauthComplete,
					},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: OAuthCredential.ID(),
						Description:   "Remove the persisted OAuth credential and disconnect this installation.",
					},
				}),
				tokenConnection.Registration(types.ConnectionRegistration{
					Name:           "Test Token",
					Description:    "Connect with an API token validated by the health check.",
					CredentialRefs: []types.CredentialSlotID{TokenCredential.ID()},
					HealthCheck:    &types.HealthCheckRegistration{Handle: healthHandler},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: TokenCredential.ID(),
						Description:   "Remove the persisted token credential and disconnect this installation.",
					},
				}),
				serviceAccountConnection.Registration(types.ConnectionRegistration{
					Name:           "Test Service Account",
					Description:    "Connect with a service account validated by the health check.",
					CredentialRefs: []types.CredentialSlotID{ServiceAccountCredential.ID()},
					HealthCheck:    &types.HealthCheckRegistration{Handle: healthHandler},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: ServiceAccountCredential.ID(),
						Description:   "Remove the persisted service account credential and disconnect this installation.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				testClient.Registration(buildClient, types.ClientRegistration{
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
				RepoSyncOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Async operation running with the built client",
					Policy:      types.ExecutionPolicy{},
					Handle:      repoSyncHandler,
				}),
				ValidatedOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Inline operation with a required config field",
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle:      validatedHandler,
				}),
				RecurringOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Healthy idle reconcile loop",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Schedule:    &gala.Schedule{MinInterval: recurringInterval},
					Handle:      idleCycle,
					Disabled:    disabledUnlessMode(ModeRecurring),
				}),
				ExhaustingOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Always-failing reconcile loop for exhaustion",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Schedule:    &gala.Schedule{MinInterval: exhaustingInterval, MaxErrorStreak: exhaustingMaxErrorStreak},
					Handle:      failingCycle,
					Disabled:    disabledUnlessMode(ModeExhausting),
				}),
				UnresolvableOp.Registration(DefinitionID, types.OperationRegistration{
					Description: "Reconcile loop whose client cannot resolve without a stored credential",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Schedule:    &gala.Schedule{MinInterval: recurringInterval},
					Handle:      idleCycle,
					Disabled:    disabledUnlessMode(ModeUnresolvable),
				}),
			},
		}, nil
	})
}
