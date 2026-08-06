package serveropts

import (
	"context"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/notifications"
	"github.com/theopenlane/core/pkg/gala"
)

// NewGalaRuntimes creates the main and notification gala runtimes from configuration.
// It does not wire the runtimes to the database client; call ConfigureGala for that.
func NewGalaRuntimes(ctx context.Context, so *ServerOptions) (*gala.Gala, *gala.Gala, error) {
	galaCfg := so.Config.Settings.Workflows.Gala
	if !galaCfg.Enabled {
		return nil, nil, nil
	}

	galaQueueName := galaCfg.QueueName
	if galaQueueName == "" {
		galaQueueName = gala.DefaultQueueName
	}

	galaApp, err := gala.NewGala(ctx, gala.Config{
		ConnectionURI: so.Config.Settings.JobQueue.ConnectionURI,
		QueueName:     galaQueueName,
		WorkerCount:   max(galaCfg.WorkerCount, 1),
		MaxRetries:    galaCfg.MaxRetries,
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

// ConfigureGala wires the gala runtimes to the database client, registers all listeners,
// and starts workers. It must be called after the database client is created.
func ConfigureGala(ctx context.Context, galaApp, notificationGala *gala.Gala, dbClient *ent.Client, so *ServerOptions) error {
	if galaApp == nil {
		return nil
	}

	galaCfg := so.Config.Settings.Workflows.Gala

	closeRuntimes := func() {
		if closeErr := notificationGala.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close in-memory gala runtime")
		}

		if closeErr := galaApp.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed to close gala runtime")
		}
	}

	dbClient.Use(hooks.EmitGalaEventHook(func() *gala.Gala {
		return galaApp
	}, func() *gala.Gala {
		return notificationGala
	}))

	if err := provideGalaDependencies(galaApp, dbClient); err != nil {
		closeRuntimes()

		return err
	}

	if err := provideGalaDependencies(notificationGala, dbClient); err != nil {
		closeRuntimes()

		return err
	}

	for _, register := range []func(*gala.Gala) ([]gala.ListenerID, error){
		hooks.RegisterGalaOrganizationAvatarListeners,
		hooks.RegisterGalaTaskRuleListeners,
		hooks.RegisterGalaEntitlementListeners,
		hooks.RegisterGalaOrganizationCleanupListeners,
		hooks.RegisterGalaTrustCenterCacheListeners,
		hooks.RegisterGalaTrustCenterWatermarkListeners,
		hooks.RegisterGalaWorkflowListeners,
		hooks.RegisterGalaVendorScoringListeners,
		hooks.RegisterGalaIdentityResolutionListeners,
		hooks.RegisterGalaDocumentAssociationListeners,
		hooks.RegisterGalaQuestionnaireTransformListeners,
		hooks.RegisterGalaCampaignRecurringListeners,
		hooks.RegisterGalaSubscriberLinkListeners,
		hooks.RegisterGalaNDAAttestationListeners,
		hooks.RegisterGalaDomainScanSubmitListeners,
		hooks.RegisterGalaDomainScanUpdateListener,
		hooks.RegisterGalaIntegrationCleanupListeners,
	} {
		if _, err := register(galaApp); err != nil {
			closeRuntimes()

			return err
		}
	}

	if _, err := notifications.RegisterGalaListeners(notificationGala); err != nil {
		closeRuntimes()

		return err
	}

	if err := galaApp.StartWorkers(ctx); err != nil {
		closeRuntimes()

		return err
	}

	log.Info().Int("gala_worker_count", max(galaCfg.WorkerCount, 1)).Str("gala_queue", galaCfg.QueueName).Msg("gala worker client started")

	return nil
}

// provideGalaDependencies registers explicit dependencies that gala listeners resolve via samber/do
// and the durable context codec that restores the ent client onto handler contexts; codec
// registration failure is a wiring error and fails startup
func provideGalaDependencies(galaApp *gala.Gala, dbClient *ent.Client) error {
	return galaApp.Attach(
		gala.WithValue(galaApp),
		gala.WithValue(dbClient),
		gala.WithRestoredValue("ent_client", ent.NewContext),
	)
}
