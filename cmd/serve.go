package cmd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/theopenlane/beacon/otelx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/cache"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/historygenerated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gala"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the core api server",
	Run: func(cmd *cobra.Command, _ []string) {
		err := serve(cmd.Context())
		cobra.CheckErr(err)
	},
}

// init registers the serve command and its flags on the root command.
func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("config", "./config/.config.yaml", "config file location")
}

// serve initializes dependencies and starts the core API server.
func serve(ctx context.Context) error {
	// setup db connection for server
	var (
		fgaClient *fgax.Client
		err       error
	)

	// create ent dependency injection
	entOpts := []ent.Option{}

	serverOpts := []serveropts.ServerOption{}
	serverOpts = append(serverOpts,
		serveropts.WithConfigProvider(&config.ProviderWithRefresh{}),
		serveropts.WithHTTPS(),
		serveropts.WithEmailConfig(),
		serveropts.WithMiddleware(),
		serveropts.WithRateLimiter(),
		serveropts.WithSecureMW(),
		serveropts.WithCacheHeaders(),
		serveropts.WithCORS(),
		serveropts.WithObjectStorage(),
		serveropts.WithEntitlements(),
		serveropts.WithSummarizer(),
		serveropts.WithKeyDirOption(),
		serveropts.WithSecretManagerKeysOption(),
		serveropts.WithShortlinks(),
	)

	so := serveropts.NewServerOptions(serverOpts, k.String("config"))

	hooks.SetSlackConfig(hooks.SlackConfig{
		WebhookURL:               so.Config.Settings.Slack.WebhookURL,
		NewSubscriberMessageFile: so.Config.Settings.Slack.NewSubscriberMessageFile,
		NewUserMessageFile:       so.Config.Settings.Slack.NewUserMessageFile,
	})

	// Create keys for development when no external keys are supplied
	if so.Config.Settings.Auth.Token.GenerateKeys && len(so.Config.Settings.Auth.Token.Keys) == 0 {
		so.AddServerOptions(serveropts.WithGeneratedKeys())
	}

	// setup token manager
	so.AddServerOptions(
		serveropts.WithTokenManager(),
	)

	err = otelx.NewTracer(so.Config.Settings.Tracer, appName)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize tracer")
	}

	// setup Authz connection
	// this must come before the database setup because the FGA Client
	// is used as an ent dependency
	fgaClient, err = fgax.CreateFGAClientWithStore(ctx, so.Config.Settings.Authz)
	if err != nil {
		return err
	}

	// Setup Redis connection
	redisClient := cache.New(so.Config.Settings.Redis)

	go func() {
		<-ctx.Done()

		if err := redisClient.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error closing redis")
		}
	}()

	defer redisClient.Close()

	// Setup pool if max workers is greater than 0
	var pool *soiree.Pool
	if so.Config.Settings.EntConfig.MaxPoolSize > 0 {
		pool = soiree.NewPool(
			soiree.WithWorkers(so.Config.Settings.EntConfig.MaxPoolSize),
			soiree.WithPoolName("ent_client_pool"),
		)
	}

	// add session manager
	so.AddServerOptions(
		serveropts.WithSessionManager(redisClient),
	)

	// add otp manager, after redis is setup
	so.AddServerOptions(
		serveropts.WithOTP(),
	)

	// add email verifier
	verifier := so.Config.Settings.EntConfig.EmailValidation.NewVerifier()

	// Set trust center config for hooks
	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		PreviewZoneID:            so.Config.Settings.Server.TrustCenterPreviewZoneID,
		CnameTarget:              so.Config.Settings.Server.TrustCenterCnameTarget,
		DefaultTrustCenterDomain: so.Config.Settings.Server.DefaultTrustCenterDomain,
	})

	// create history client
	histOpts := []historygenerated.Option{
		historygenerated.Authz(*fgaClient),
		historygenerated.EntConfig(&so.Config.Settings.EntConfig),
	}

	historyClient, err := entdb.NewHistory(so.Config.Settings.DB, histOpts...)
	if err != nil {
		return err
	}

	// add additional ent dependencies
	entOpts = append(
		entOpts,
		ent.Authz(*fgaClient),
		ent.TOTP(so.Config.Handler.OTPManager),
		ent.TokenManager(so.Config.Handler.TokenManager),
		ent.SessionConfig(so.Config.Handler.SessionConfig),
		ent.EntConfig(&so.Config.Settings.EntConfig),
		ent.Emailer(&so.Config.Settings.Email),
		ent.EntitlementManager(so.Config.Handler.Entitlements),
		ent.ObjectManager(so.Config.StorageService),
		ent.Summarizer(so.Config.Handler.Summarizer),
		ent.Shortlinks(so.Config.Handler.ShortlinksClient),
		ent.Pool(pool),
		ent.EmailVerifier(verifier),
		ent.HistoryClient(historyClient),
	)

	// Setup DB connection
	log.Info().Interface("db", so.Config.Settings.DB.DatabaseName).Msg("connecting to database")

	mutationOutboxCfg := so.Config.Settings.Workflows.MutationOutbox
	galaCfg := so.Config.Settings.Workflows.Gala
	mutationOutboxWorkerCount := max(mutationOutboxCfg.WorkerCount, 1)
	galaWorkerCount := max(galaCfg.WorkerCount, 1)
	mutationOutboxWorkersEnabled := mutationOutboxCfg.Enabled
	galaRuntimeEnabled := galaCfg.Enabled || galaCfg.DualEmit
	galaWorkersEnabled := galaCfg.Enabled

	var galaRuntime *gala.Runtime

	eventer := hooks.NewEventer(
		hooks.WithWorkflowListenersEnabled(so.Config.Settings.Workflows.Enabled),
		hooks.WithMutationOutboxEnabled(mutationOutboxCfg.Enabled),
		hooks.WithMutationOutboxFailOnEnqueueError(mutationOutboxCfg.FailOnEnqueueError),
		hooks.WithMutationOutboxTopics(mutationOutboxCfg.Topics),
		hooks.WithGalaRuntimeProvider(func() *gala.Runtime { return galaRuntime }),
		hooks.WithGalaDualEmitEnabled(galaCfg.DualEmit),
		hooks.WithGalaFailOnEnqueueError(galaCfg.FailOnEnqueueError),
		hooks.WithGalaTopics(galaCfg.Topics),
		hooks.WithGalaTopicModes(galaCfg.TopicModes),
	)

	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(so.Config.Settings.JobQueue.ConnectionURI),
	}

	if mutationOutboxWorkersEnabled || galaWorkersEnabled {
		workers := river.NewWorkers()

		if mutationOutboxWorkersEnabled {
			if err := river.AddWorkerSafely(workers, eventqueue.NewMutationDispatchWorker(func() *soiree.EventBus {
				return eventer.Emitter
			})); err != nil {
				return err
			}
		}

		if galaWorkersEnabled {
			if err := river.AddWorkerSafely(workers, gala.NewRiverDispatchWorker(func() *gala.Runtime {
				return galaRuntime
			})); err != nil {
				return err
			}
		}

		defaultQueueWorkers := 1
		if mutationOutboxWorkersEnabled {
			defaultQueueWorkers = max(defaultQueueWorkers, mutationOutboxWorkerCount)
		}

		queueName := galaCfg.QueueName
		if queueName == "" {
			queueName = jobspec.QueueDefault
		}

		queueConfig := map[string]river.QueueConfig{
			jobspec.QueueDefault:      {MaxWorkers: defaultQueueWorkers},
			jobspec.QueueCompliance:   {MaxWorkers: 1},
			jobspec.QueueTrustcenter:  {MaxWorkers: 1},
			jobspec.QueueNotification: {MaxWorkers: 1},
		}

		if galaWorkersEnabled {
			if queueName == jobspec.QueueDefault {
				queueConfig[jobspec.QueueDefault] = river.QueueConfig{MaxWorkers: max(defaultQueueWorkers, galaWorkerCount)}
			} else {
				queueConfig[queueName] = river.QueueConfig{MaxWorkers: galaWorkerCount}
			}
		}

		jobOpts = append(jobOpts,
			riverqueue.WithWorkers(workers),
			riverqueue.WithQueues(queueConfig),
		)

		maxRetries := 0
		if mutationOutboxWorkersEnabled {
			maxRetries = max(maxRetries, mutationOutboxCfg.MaxRetries)
		}

		if galaWorkersEnabled {
			maxRetries = max(maxRetries, galaCfg.MaxRetries)
		}

		if maxRetries > 0 {
			jobOpts = append(jobOpts, riverqueue.WithMaxRetries(maxRetries))
		}
	}

	clientOpts := []entdb.Option{
		entdb.WithEventer(eventer, &so.Config.Settings.Workflows),
		entdb.WithModules(),
		entdb.WithMetricsHook(),
	}

	dbClient, err := entdb.New(ctx, so.Config.Settings.DB, jobOpts, clientOpts, entOpts...)
	if err != nil {
		return err
	}

	if galaRuntimeEnabled {
		jobClient, ok := dbClient.Job.(*riverqueue.Client)
		if !ok {
			log.Warn().Bool("dual_emit", galaCfg.DualEmit).Bool("worker_enabled", galaCfg.Enabled).Msg("gala enabled but job client is not riverqueue client")
		} else {
			queueName := galaCfg.QueueName
			if queueName == "" {
				queueName = jobspec.QueueDefault
			}

			dispatcher, err := gala.NewRiverDispatcher(gala.RiverDispatcherOptions{
				JobClient: jobClient,
				QueueByClass: map[gala.QueueClass]string{
					gala.QueueClassWorkflow:    queueName,
					gala.QueueClassIntegration: queueName,
					gala.QueueClassGeneral:     queueName,
				},
				DefaultQueue: queueName,
			})
			if err != nil {
				return err
			}

			gRuntime, err := gala.NewRuntime(gala.RuntimeOptions{
				DurableDispatcher: dispatcher,
			})
			if err != nil {
				return err
			}

			gala.ProvideValue(gRuntime.Injector(), dbClient)

			if galaWorkersEnabled {
				if _, err := hooks.RegisterGalaEntitlementListeners(gRuntime.Registry()); err != nil {
					return err
				}

				if so.Config.Settings.Workflows.Enabled {
					if _, err := hooks.RegisterGalaWorkflowListeners(gRuntime.Registry()); err != nil {
						return err
					}
				}
			}

			galaRuntime = gRuntime
		}
	}

	var (
		stopEventWorker     func()
		stopEventWorkerOnce sync.Once
	)

	stopEventWorkers := func() {
		if stopEventWorker == nil {
			return
		}

		stopEventWorkerOnce.Do(stopEventWorker)
	}
	defer stopEventWorkers()

	if mutationOutboxWorkersEnabled || galaWorkersEnabled {
		jobClient, ok := dbClient.Job.(*riverqueue.Client)
		if !ok {
			log.Warn().Bool("mutation_outbox", mutationOutboxWorkersEnabled).Bool("gala_worker", galaWorkersEnabled).Msg("event workers enabled but job client is not riverqueue client")
		} else {
			workerClient := jobClient.GetRiverClient()
			if err := workerClient.Start(ctx); err != nil {
				return err
			}

			stopEventWorker = func() {
				stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if stopErr := workerClient.Stop(stopCtx); stopErr != nil &&
					!errors.Is(stopErr, context.Canceled) &&
					!errors.Is(stopErr, context.DeadlineExceeded) {
					log.Error().Err(stopErr).Msg("error stopping event worker client")
				}
			}

			log.Info().
				Bool("mutation_outbox_enabled", mutationOutboxWorkersEnabled).
				Bool("gala_worker_enabled", galaWorkersEnabled).
				Int("mutation_outbox_worker_count", mutationOutboxWorkerCount).
				Int("gala_worker_count", galaWorkerCount).
				Msg("event worker client started")
		}
	}

	if so.Config.Settings.Workflows.Enabled {
		if wfEngine, ok := dbClient.WorkflowEngine.(*engine.WorkflowEngine); ok {
			so.AddServerOptions(serveropts.WithWorkflows(wfEngine))
			log.Info().Msg("workflow engine initialized")
		}
	}

	if so.Config.Settings.CampaignWebhook.Enabled {
		so.AddServerOptions(serveropts.WithCampaignWebhookConfig())
	}

	so.AddServerOptions(serveropts.WithCloudflareConfig())

	go func() {
		<-ctx.Done()

		log.Ctx(ctx).Info().Msg("waiting for in-flight uploads to complete")
		pkgobjects.WaitForUploads()
		log.Ctx(ctx).Info().Msg("all uploads completed")

		stopEventWorkers()

		if err := entdb.GracefulClose(context.Background(), dbClient, time.Second); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error closing database")
		}

		if err := soiree.ShutdownAll(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error shutting down event pools")
		}
	}()

	defer entdb.GracefulClose(context.Background(), dbClient, time.Second)

	// add auth session manager
	so.Config.Handler.AuthManager = authmanager.New(dbClient)

	// Add Driver to the Handlers Config
	so.Config.Handler.DBClient = dbClient

	// Add redis client to Handlers Config
	so.Config.Handler.RedisClient = redisClient

	// set dev flag
	so.Config.Handler.IsDev = so.Config.Settings.Server.Dev

	// set the logging config options based on flags
	so.Config.Settings.Server.Debug = k.Bool("debug")
	so.Config.Settings.Server.Pretty = k.Bool("pretty")

	// add default trust center domain
	so.AddServerOptions(
		serveropts.WithDefaultTrustCenterDomain(),
	)

	// add ready checks
	so.AddServerOptions(
		serveropts.WithReadyChecks(dbClient.Config, fgaClient, redisClient, dbClient.Job),
	)

	// add auth and integration options
	so.AddServerOptions(
		serveropts.WithAuth(),
		serveropts.WithIntegrationStore(dbClient),
		serveropts.WithIntegrationBroker(),
		serveropts.WithIntegrationClients(),
		serveropts.WithIntegrationOperations(),
		serveropts.WithIntegrationIngestEvents(dbClient),
		serveropts.WithIntegrationActivation(),
	)

	// add session manager
	so.AddServerOptions(
		serveropts.WithSessionMiddleware(),
	)

	// add csrf protection after auth and session middleware
	so.AddServerOptions(
		serveropts.WithCSRF(),
	)

	srv, err := server.NewServer(so.Config)
	if err != nil {
		return err
	}

	// Setup Graph API Handlers
	so.AddServerOptions(
		serveropts.WithGraphRoute(srv, dbClient),
		serveropts.WithHistoryGraphRoute(srv, historyClient),
	)

	if err := srv.StartEchoServer(ctx); err != nil {
		log.Error().Err(err).Msg("failed to run server")
	}

	return nil
}
