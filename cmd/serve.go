package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theopenlane/beacon/otelx"
	"go.uber.org/zap"

	dbx "github.com/theopenlane/dbx/pkg/dbxclient"
	"github.com/theopenlane/iam/fgax"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
	"github.com/theopenlane/utils/cache"
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
	entOpts := []ent.Option{ent.Logger(*logger)}

	serverOpts := []serveropts.ServerOption{}
	serverOpts = append(serverOpts,
		serveropts.WithConfigProvider(&config.ConfigProviderWithRefresh{}),
		serveropts.WithLogger(logger),
		serveropts.WithHTTPS(),
		serveropts.WithEmailManager(),
		serveropts.WithTaskManager(),
		serveropts.WithMiddleware(),
		serveropts.WithRateLimiter(),
		serveropts.WithSecureMW(),
		serveropts.WithCacheHeaders(),
		serveropts.WithCORS(),
		serveropts.WithAnalytics(),
		serveropts.WithEventPublisher(),
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
		logger.Fatalw("failed to initialize tracer", "error", err)
	}

	// setup Authz connection
	// this must come before the database setup because the FGA Client
	// is used as an ent dependency
	fgaClient, err = fgax.CreateFGAClientWithStore(ctx, so.Config.Settings.Authz, so.Config.Logger)
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

	// Setup Geodetic client
	if so.Config.Settings.DBx.Enabled {
		gc := so.Config.Settings.DBx.NewDefaultClient()

		entOpts = append(entOpts, ent.DBx(gc.(*dbx.Client)))
	}

	// add otp manager, after redis is setup
	so.AddServerOptions(
		serveropts.WithOTP(),
	)

	// add additional ent dependencies
	entOpts = append(
		entOpts,
		ent.Authz(*fgaClient),
		ent.Emails(so.Config.Handler.EmailManager),
		ent.Marionette(so.Config.Handler.TaskMan),
		ent.Analytics(so.Config.Handler.AnalyticsClient),
		ent.TOTP(so.Config.Handler.OTPManager),
		ent.TokenManager(so.Config.Handler.TokenManager),
		ent.SessionConfig(so.Config.Handler.SessionConfig),
		ent.EntConfig(&so.Config.Settings.EntConfig),
	)

	// Setup DB connection
	entdbClient, dbConfig, err := entdb.NewMultiDriverDBClient(ctx, so.Config.Settings.DB, logger, entOpts)
	if err != nil {
		return err
	}

	defer entdbClient.Close()

	// Add Driver to the Handlers Config
	so.Config.Handler.DBClient = entdbClient

	// Add redis client to Handlers Config
	so.Config.Handler.RedisClient = redisClient

	// add ready checks
	so.AddServerOptions(
		serveropts.WithReadyChecks(dbConfig, fgaClient, redisClient),
	)

	// add auth options
	so.AddServerOptions(
		serveropts.WithAuth(),
	)

	// add session manager
	so.AddServerOptions(
		serveropts.WithSessionMiddleware(),
	)

	srv := server.NewServer(so.Config, so.Config.Logger)

	// Setup Graph API Handlers
	so.AddServerOptions(serveropts.WithGraphRoute(srv, entdbClient))

	if err := srv.StartEchoServer(ctx); err != nil {
		logger.Error("failed to run server", zap.Error(err))
	}

	return nil
}
