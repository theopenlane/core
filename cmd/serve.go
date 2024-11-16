package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/theopenlane/beacon/otelx"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/cache"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the core api server",
	Run: func(cmd *cobra.Command, args []string) {
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
		serveropts.WithConfigProvider(&config.ConfigProviderWithRefresh{}),
		serveropts.WithHTTPS(),
		serveropts.WithEmailConfig(),
		serveropts.WithMiddleware(),
		serveropts.WithRateLimiter(),
		serveropts.WithSecureMW(),
		serveropts.WithCacheHeaders(),
		serveropts.WithCORS(),
		serveropts.WithObjectStorage(),
		serveropts.WithEntitlements(),
	)

	so := serveropts.NewServerOptions(serverOpts, k.String("config"))

	// Create keys for development
	if so.Config.Settings.Auth.Token.GenerateKeys {
		so.AddServerOptions(serveropts.WithGeneratedKeys())
	}

	// add auth session manager
	so.Config.Handler.AuthManager = authmanager.New()

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
	defer redisClient.Close()

	// add session manager
	so.AddServerOptions(
		serveropts.WithSessionManager(redisClient),
	)

	// add otp manager, after redis is setup
	so.AddServerOptions(
		serveropts.WithOTP(),
	)

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
	)

	// Setup DB connection
	log.Info().Interface("db", so.Config.Settings.DB).Msg("connecting to database")

	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(so.Config.Settings.JobQueue.ConnectionURI),
	}

	dbClient, err := entdb.New(ctx, so.Config.Settings.DB, jobOpts, entOpts...)
	if err != nil {
		return err
	}

	defer dbClient.CloseAll() // nolint: errcheck

	// Add Driver to the Handlers Config
	so.Config.Handler.DBClient = dbClient

	// Add redis client to Handlers Config
	so.Config.Handler.RedisClient = redisClient

	// add ready checks
	so.AddServerOptions(
		serveropts.WithReadyChecks(dbClient.Config, fgaClient, redisClient, dbClient.Job),
	)

	// add auth options
	so.AddServerOptions(
		serveropts.WithAuth(),
	)

	// add session manager
	so.AddServerOptions(
		serveropts.WithSessionMiddleware(),
	)

	srv := server.NewServer(so.Config)

	// Setup Graph API Handlers
	so.AddServerOptions(serveropts.WithGraphRoute(srv, dbClient))

	if err := srv.StartEchoServer(ctx); err != nil {
		log.Error().Err(err).Msg("failed to run server")
	}

	return nil
}
