package serveropts

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/integrations/definitions/email"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	runtime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/logx"
)

// WithIntegrationsRuntime builds the integration runtime from server settings and wires it
// into the handler. When a workflow engine is present it also injects integration dependencies.
// Initialization is skipped if the database client or Gala runtime is nil.
func WithIntegrationsRuntime(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.Config.Handler.IntegrationsConfig = s.Config.Settings.Integrations

		if dbClient == nil {
			return
		}

		galaInstance := s.Config.Handler.Gala
		if galaInstance == nil {
			log.Warn().Msg("gala runtime not available; integration runtime will not be initialized")
			return
		}

		credStore, err := keystore.NewStore(dbClient)
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize keystore for integrations")
		}

		wf := s.Config.Handler.WorkflowEngine
		rt, err := runtime.New(runtime.Config{
			DB:            dbClient,
			Gala:          galaInstance,
			Keystore:      credStore,
			RedisClient:   s.Config.Handler.RedisClient,
			CatalogConfig: s.Config.Settings.Integrations,
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration runtime")
		}

		s.Config.Handler.IntegrationsRuntime = rt

		email.SetClientResolver(buildEmailClientResolver(dbClient, rt))

		if wf == nil {
			return
		}

		if err := wf.SetIntegrationDeps(engine.IntegrationDeps{
			Runtime: rt,
		}); err != nil {
			log.Panic().Err(err).Msg("failed to wire integration deps into workflow engine")
		}
	})
}

// buildEmailClientResolver creates an email.ClientResolver that checks for a customer
// integration installation and falls back to the runtime system client
func buildEmailClientResolver(db *ent.Client, rt *runtime.Runtime) email.ClientResolver {
	return func(ctx context.Context, ownerID string) (*email.EmailClient, error) {
		systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

		inst, err := db.Integration.Query().
			Where(
				integration.OwnerIDEQ(ownerID),
				integration.DefinitionIDEQ(email.DefinitionID()),
			).
			Only(systemCtx)

		switch {
		case err == nil:
			raw, buildErr := rt.BuildClientForIntegration(ctx, inst, email.CustomerClientID())
			if buildErr != nil {
				logx.FromContext(ctx).Error().Err(buildErr).Str("owner_id", ownerID).Msg("failed building customer email client, falling back to system")

				return resolveSystemEmailClient(rt)
			}

			client, ok := raw.(*email.EmailClient)
			if !ok {
				return nil, fmt.Errorf("%w: unexpected type %T", email.ErrInvalidOperationClient, raw)
			}

			return client, nil
		case ent.IsNotFound(err):
			return resolveSystemEmailClient(rt)
		default:
			logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Msg("failed querying customer email integration")

			return resolveSystemEmailClient(rt)
		}
	}
}

// resolveSystemEmailClient extracts the system email client from the runtime registry
func resolveSystemEmailClient(rt *runtime.Runtime) (*email.EmailClient, error) {
	raw, ok := rt.Registry().RuntimeClient(email.DefinitionID())
	if !ok {
		return nil, email.ErrSenderNotConfigured
	}

	client, ok := raw.(*email.EmailClient)
	if !ok {
		return nil, fmt.Errorf("%w: unexpected type %T", email.ErrInvalidOperationClient, raw)
	}

	return client, nil
}
