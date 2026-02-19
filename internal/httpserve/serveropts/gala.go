package serveropts

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/ingest"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

// WithGala configures Gala mutation emission and, when enabled, starts Gala workers.
func WithGala(ctx context.Context, so *ServerOptions, dbClient *ent.Client) (*gala.Gala, error) {
	workflowCfg := so.Config.Settings.Workflows
	galaCfg := workflowCfg.Gala
	if !galaCfg.Enabled {
		return nil, nil
	}

	galaQueueName := galaCfg.QueueName
	if galaQueueName == "" {
		galaQueueName = gala.DefaultQueueName
	}

	galaApp, err := gala.NewGala(ctx, gala.Config{
		Enabled:       galaCfg.Enabled,
		ConnectionURI: so.Config.Settings.JobQueue.ConnectionURI,
		QueueName:     galaQueueName,
		WorkerCount:   max(galaCfg.WorkerCount, 1),
		QueueWorkers: map[string]int{
			integrations.IntegrationQueueName: max(galaCfg.WorkerCount, 1),
		},
		MaxRetries: galaCfg.MaxRetries,
	})
	if err != nil {
		return nil, err
	}

	notificationGala, err := gala.NewGala(ctx, gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
		WorkerCount:  max(galaCfg.WorkerCount, 1),
	})
	if err != nil {
		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala after in-memory runtime creation failure")
		}

		return nil, err
	}

	closeRuntimes := func() {
		if closeErr := notificationGala.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close in-memory gala runtime")
		}

		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala runtime")
		}
	}

	// slightly weird signature but same registration of a new ent hook
	dbClient.Use(hooks.EmitGalaEventHook(func() *gala.Gala {
		return galaApp
	}, func() *gala.Gala {
		return notificationGala
	}))
	so.Config.Handler.Gala = galaApp

	provideGalaDependencies(galaApp.Injector(), galaApp, dbClient, true)
	provideGalaDependencies(notificationGala.Injector(), notificationGala, dbClient, false)

	register := func(runtime *gala.Gala, registerListeners func(*gala.Registry) ([]gala.ListenerID, error)) error {
		if _, err := registerListeners(runtime.Registry()); err != nil {
			closeRuntimes()

			return err
		}

		return nil
	}

	if err := register(galaApp, hooks.RegisterGalaEntitlementListeners); err != nil {
		return nil, err
	}

	if err := register(galaApp, hooks.RegisterGalaTrustCenterCacheListeners); err != nil {
		return nil, err
	}

	if err := register(galaApp, hooks.RegisterGalaWorkflowListeners); err != nil {
		return nil, err
	}

	if err := register(galaApp, hooks.RegisterGalaSlackListeners); err != nil {
		return nil, err
	}

	if err := register(notificationGala, hooks.RegisterGalaNotificationListeners); err != nil {
		return nil, err
	}

	if _, err := ingest.RegisterIngestListeners(galaApp.Registry(), dbClient); err != nil {
		closeRuntimes()

		return nil, err
	}

	if err := galaApp.StartWorkers(ctx); err != nil {
		closeRuntimes()

		return nil, err
	}

	log.Info().Int("gala_worker_count", max(galaCfg.WorkerCount, 1)).Str("gala_queue", galaQueueName).Msg("gala worker client started")

	return galaApp, nil
}

// provideGalaDependencies registers explicit dependencies that gala listeners resolve via samber/do.
// Current listeners require:
//   - *ent.Client: used by entitlement and workflow listeners
//   - *engine.WorkflowEngine: used by workflow listeners (when workflows enabled)
func provideGalaDependencies(injector do.Injector, galaApp *gala.Gala, dbClient *ent.Client, setWorkflowRuntime bool) {
	if galaApp != nil {
		do.ProvideValue(injector, galaApp)
	}

	do.ProvideValue(injector, dbClient)

	if !setWorkflowRuntime {
		return
	}

	if wfEngine, ok := dbClient.WorkflowEngine.(*engine.WorkflowEngine); ok && wfEngine != nil {
		wfEngine.SetGala(galaApp)
		do.ProvideValue(injector, wfEngine)
	}
}
