package serveropts

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

// WithGala configures Gala mutation emission and, when enabled, starts Gala workers.
func WithGala(ctx context.Context, so *ServerOptions, dbClient *ent.Client) (*gala.Gala, error) {
	galaCfg := so.Config.Settings.Workflows.Gala
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
		MaxRetries:    galaCfg.MaxRetries,
	})
	if err != nil {
		return nil, err
	}

	dbClient.Use(hooks.EmitGalaEventHook(func() *gala.Gala { return galaApp }, galaCfg.FailOnEnqueueError))

	provideGalaDependencies(galaApp.Injector(), dbClient)

	if _, err := hooks.RegisterGalaSlackListeners(galaApp.Registry()); err != nil {
		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala after listener registration failure")
		}

		return nil, err
	}

	if err := galaApp.StartWorkers(ctx); err != nil {
		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala after worker start failure")
		}

		return nil, err
	}

	log.Info().Int("gala_worker_count", max(galaCfg.WorkerCount, 1)).Str("gala_queue", galaQueueName).Msg("gala worker client started")

	return galaApp, nil
}

// provideGalaDependencies registers explicit dependencies that gala listeners resolve via samber/do.
// Current listeners require:
//   - *ent.Client: used by entitlement and workflow listeners
//   - *engine.WorkflowEngine: used by workflow listeners (when workflows enabled)
func provideGalaDependencies(injector do.Injector, dbClient *ent.Client) {
	do.ProvideValue(injector, dbClient)

	if wfEngine, ok := dbClient.WorkflowEngine.(*engine.WorkflowEngine); ok && wfEngine != nil {
		do.ProvideValue(injector, wfEngine)
	}
}
