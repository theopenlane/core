package serveropts

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/notifications"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// NewGalaRuntimes creates the main and notification gala runtimes from configuration.
// It does not wire the runtimes to the database client; call ConfigureGala for that.
func NewGalaRuntimes(ctx context.Context, so *ServerOptions) (*gala.Gala, *gala.Gala, error) {
	galaCfg := so.Config.Settings.Workflows.Gala
	if !galaCfg.Enabled {
		if so.Config.Settings.Backfill.Enabled {
			return nil, nil, ErrBackfillRequiresGala
		}

		return nil, nil, nil
	}

	galaQueueName := galaCfg.QueueName
	if galaQueueName == "" {
		galaQueueName = gala.DefaultQueueName
	}

	galaApp, err := gala.NewGala(ctx, gala.Config{
		ConnectionURI:    so.Config.Settings.JobQueue.ConnectionURI,
		QueueName:        galaQueueName,
		WorkerCount:      max(galaCfg.WorkerCount, 1),
		MaxRetries:       galaCfg.MaxRetries,
		TopicRenames:     jobTopicRenames(so),
		OperationRenames: jobOperationRenames(),
	})
	if err != nil {
		return nil, nil, err
	}

	notificationGala, err := gala.NewGala(ctx, gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
		WorkerCount:  max(galaCfg.WorkerCount, 1),
	})
	if err != nil {
		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala after in-memory runtime creation failure")
		}

		return nil, nil, err
	}

	return galaApp, notificationGala, nil
}

// ConfigureGala wires the gala runtimes to the database client and registers all listeners;
// it must be called after the database client is created
func ConfigureGala(galaApp, notificationGala *gala.Gala, dbClient *ent.Client, so *ServerOptions) error {
	if galaApp == nil {
		return nil
	}

	closeRuntimes := func() {
		if closeErr := notificationGala.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close in-memory gala runtime")
		}

		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala runtime")
		}
	}

	dbClient.Use(hooks.EmitGalaEventHook(galaApp, notificationGala))

	if err := provideGalaDependencies(galaApp, dbClient); err != nil {
		closeRuntimes()

		return err
	}

	if err := provideGalaDependencies(notificationGala, dbClient); err != nil {
		closeRuntimes()

		return err
	}

	registrations := lo.Flatten([][]gala.Registration{
		hooks.OrganizationAvatarListeners(),
		hooks.TaskRuleListeners(),
		hooks.EntitlementListeners(),
		hooks.OrganizationCleanupListeners(),
		hooks.TrustCenterCacheListeners(),
		hooks.TrustCenterWatermarkListeners(),
		hooks.WorkflowListeners(),
		hooks.VendorScoringListeners(),
		hooks.IdentityResolutionListeners(),
		hooks.DocumentAssociationListeners(),
		hooks.QuestionnaireTransformListeners(),
		hooks.CampaignRecurringListeners(),
		hooks.SubscriberLinkListeners(),
		hooks.NDAAttestationListeners(),
		hooks.DomainScanListeners(),
		hooks.IntegrationCleanupListeners(),
	})

	if _, err := gala.Register(galaApp, registrations...); err != nil {
		closeRuntimes()

		return err
	}

	// Notification listeners perform durable user-visible side effects; register them on
	// the persistent runtime so a process restart cannot lose queued mutation events.
	if _, err := gala.Register(galaApp, notifications.Listeners()...); err != nil {
		closeRuntimes()

		return err
	}

	return nil
}

// StartGalaWorkers begins job processing on the durable gala runtime; call it only after all
// injector provisioning completes so a dequeued job never resolves a missing dependency
func StartGalaWorkers(ctx context.Context, galaApp *gala.Gala, so *ServerOptions) error {
	if galaApp == nil {
		return nil
	}

	if err := galaApp.StartWorkers(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to start gala workers")
		return err
	}

	return nil
}

// provideGalaDependencies registers explicit dependencies that gala listeners resolve via samber/do
// and the durable context codec that restores the ent client onto handler contexts; codec
// registration failure is a wiring error and fails startup. Subsystems riding on the client
// (workflow engine, entitlement manager) register here; the integrations runtime attaches
// itself when it is constructed
func provideGalaDependencies(galaApp *gala.Gala, dbClient *ent.Client) error {
	opts := []gala.AttachOption{
		gala.WithValue(galaApp),
		gala.WithValue(dbClient),
		gala.WithRestoredValue("ent_client", ent.NewContext),
	}

	if wfEngine, ok := dbClient.WorkflowEngine.(*engine.WorkflowEngine); ok && wfEngine != nil {
		opts = append(opts, gala.WithValue(wfEngine))
	}

	if dbClient.EntitlementManager != nil {
		opts = append(opts, gala.WithValue(dbClient.EntitlementManager))
	}

	return galaApp.Attach(opts...)
}
