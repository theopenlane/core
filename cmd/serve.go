package cmd

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/theopenlane/beacon/otelx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/cache"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
	"github.com/theopenlane/core/pkg/events/soiree"
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

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("config", "./config/.config.yaml", "config file location")
}

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
		serveropts.WithWindmill(),
		serveropts.WithKeyDirOption(),
		serveropts.WithSecretManagerKeysOption(),
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

	// Setup Pond Pool if max workers is greater than 0
	var pool *soiree.PondPool
	if so.Config.Settings.EntConfig.MaxPoolSize > 0 {
		pool = soiree.NewPondPool(
			soiree.WithMaxWorkers(so.Config.Settings.EntConfig.MaxPoolSize),
			soiree.WithName("ent_client_pool"),
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
		PreviewZoneID: so.Config.Settings.Server.TrustCenterPreviewZoneID,
		CnameTarget:   so.Config.Settings.Server.TrustCenterCnameTarget,
	})

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
		ent.Windmill(so.Config.Handler.Windmill),
		ent.PondPool(pool),
		ent.EmailVerifier(verifier),
	)

	// Setup DB connection
	log.Info().Interface("db", so.Config.Settings.DB.DatabaseName).Msg("connecting to database")

	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(so.Config.Settings.JobQueue.ConnectionURI),
	}

	dbClient, err := entdb.New(ctx, so.Config.Settings.DB, jobOpts, entOpts...)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		log.Ctx(ctx).Info().Msg("waiting for in-flight uploads to complete")
		pkgobjects.WaitForUploads()
		log.Ctx(ctx).Info().Msg("all uploads completed")

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

	// add auth options
	so.AddServerOptions(serveropts.WithAuth())
	so.AddServerOptions(serveropts.WithIntegrationStore(dbClient))
	so.AddServerOptions(serveropts.WithIntegrationBroker())
	so.AddServerOptions(serveropts.WithIntegrationClients())
	so.AddServerOptions(serveropts.WithIntegrationOperations())
	so.AddServerOptions(serveropts.WithKeymaker())

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
	so.AddServerOptions(serveropts.WithGraphRoute(srv, dbClient))

	if err := srv.StartEchoServer(ctx); err != nil {
		log.Error().Err(err).Msg("failed to run server")
	}

	return nil
}
