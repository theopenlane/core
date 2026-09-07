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
				{
					Ref:         TokenCredential.ID(),
					Name:        "Test Token",
					Description: "API token the test client is built from.",
					Schema:      tokenSchema,
				},
				{
					Ref:         OAuthCredential.ID(),
					Name:        "Test OAuth",
					Description: "Auth-managed credential slot filled by the OAuth fixture.",
				},
				{
					Ref:         ServiceAccountCredential.ID(),
					Name:        "Test Service Account",
					Description: "Strict-schema credential slot used by config flows.",
					Schema:      serviceAccountSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  OAuthCredential.ID(),
					Name:           "Test OAuth",
					Description:    "Authenticate through the OAuth callback fixture.",
					CredentialRefs: []types.CredentialSlotID{OAuthCredential.ID()},
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
					CredentialRef:  TokenCredential.ID(),
					Name:           "Test Token",
					Description:    "Connect with an API token validated by the health check.",
					CredentialRefs: []types.CredentialSlotID{TokenCredential.ID()},
					HealthCheck:    &types.HealthCheckRegistration{Handle: healthHandler},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: TokenCredential.ID(),
						Description:   "Remove the persisted token credential and disconnect this installation.",
					},
				},
				{
					CredentialRef:  ServiceAccountCredential.ID(),
					Name:           "Test Service Account",
					Description:    "Connect with a service account validated by the health check.",
					CredentialRefs: []types.CredentialSlotID{ServiceAccountCredential.ID()},
					HealthCheck:    &types.HealthCheckRegistration{Handle: healthHandler},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: ServiceAccountCredential.ID(),
						Description:   "Remove the persisted service account credential and disconnect this installation.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            testClient.ID(),
					CredentialRefs: []types.CredentialSlotID{TokenCredential.ID()},
					Description:    "Test client built from the stored token credential.",
					Build:          buildClient,
				},
			},
			Webhooks: []types.WebhookRegistration{
				{
					Name:  "inbound.events",
					Event: webhookInboundEvent,
					Events: []types.WebhookEventRegistration{
						{
							Name:   WebhookAlertCreated.Name(),
							Topic:  DefinitionID.WebhookEventTopic(WebhookAlertCreated.Name()),
							Handle: func(context.Context, types.WebhookHandleRequest) error { return nil },
						},
					},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         RepoSyncOp.Name(),
					Description:  "Async operation running with the built client",
					Topic:        DefinitionID.OperationTopic(RepoSyncOp.Name()),
					ClientRef:    testClient.ID(),
					ConfigSchema: repoSyncSchema,
					Policy:       types.ExecutionPolicy{},
					Handle:       repoSyncHandler,
				},
				{
					Name:         ValidatedOp.Name(),
					Description:  "Inline operation with a required config field",
					Topic:        DefinitionID.OperationTopic(ValidatedOp.Name()),
					ConfigSchema: validatedSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       validatedHandler,
				},
				{
					Name:         RecurringOp.Name(),
					Description:  "Healthy idle reconcile loop",
					Topic:        DefinitionID.OperationTopic(RecurringOp.Name()),
					ConfigSchema: recurringSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Schedule:     &gala.Schedule{MinInterval: recurringInterval},
					Handle:       idleCycle,
					Disabled:     disabledUnlessMode(ModeRecurring),
				},
				{
					Name:         ExhaustingOp.Name(),
					Description:  "Always-failing reconcile loop for exhaustion",
					Topic:        DefinitionID.OperationTopic(ExhaustingOp.Name()),
					ConfigSchema: exhaustingSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Schedule:     &gala.Schedule{MinInterval: exhaustingInterval, MaxErrorStreak: exhaustingMaxErrorStreak},
					Handle:       failingCycle,
					Disabled:     disabledUnlessMode(ModeExhausting),
				},
				{
					Name:         UnresolvableOp.Name(),
					Description:  "Reconcile loop whose client cannot resolve without a stored credential",
					Topic:        DefinitionID.OperationTopic(UnresolvableOp.Name()),
					ClientRef:    testClient.ID(),
					ConfigSchema: unresolvableSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Schedule:     &gala.Schedule{MinInterval: recurringInterval},
					Handle:       idleCycle,
					Disabled:     disabledUnlessMode(ModeUnresolvable),
				},
			},
		}, nil
	})
}
