package serveropts

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	echoprometheus "github.com/theopenlane/echo-prometheus"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/echox/middleware"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/utils/cache"
	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/marionette"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/utils/echocontext"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/pkg/analytics"
	"github.com/theopenlane/core/pkg/events/kafka/publisher"
	authmw "github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/core/pkg/middleware/cachecontrol"
	"github.com/theopenlane/core/pkg/middleware/cors"
	"github.com/theopenlane/core/pkg/middleware/mime"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/pkg/middleware/redirect"
	"github.com/theopenlane/core/pkg/middleware/secure"
)

type ServerOption interface {
	apply(*ServerOptions)
}

type applyFunc struct {
	applyInternal func(*ServerOptions)
}

func (fso *applyFunc) apply(s *ServerOptions) {
	fso.applyInternal(s)
}

func newApplyFunc(apply func(option *ServerOptions)) *applyFunc {
	return &applyFunc{
		applyInternal: apply,
	}
}

// WithConfigProvider supplies the config for the server
func WithConfigProvider(cfgProvider config.ConfigProvider) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		s.ConfigProvider = cfgProvider
	})
}

// WithHTTPS sets up TLS config settings for the server
func WithHTTPS() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if !s.Config.Settings.Server.TLS.Enabled {
			// this is set to enabled by WithServer
			// if TLS is not enabled, move on
			return
		}

		s.Config.WithTLSDefaults()

		if !s.Config.Settings.Server.TLS.AutoCert {
			s.Config.WithTLSCerts(s.Config.Settings.Server.TLS.CertFile, s.Config.Settings.Server.TLS.CertKey)
		}
	})
}

// WithGeneratedKeys will generate a public/private key pair
// that can be used for jwt signing.
// This should only be used in a development environment
func WithGeneratedKeys() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		privFileName := "private_key.pem"

		// generate a new private key if one doesn't exist
		if _, err := os.Stat(privFileName); err != nil {
			// Generate a new RSA private key with 2048 bits
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048) //nolint:mnd
			if err != nil {
				log.Panic().Err(err).Msg("Error generating RSA private key")
			}

			// Encode the private key to the PEM format
			privateKeyPEM := &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}

			privateKeyFile, err := os.Create(privFileName)
			if err != nil {
				log.Panic().Err(err).Msg("Error creating private key file")
			}

			if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
				log.Panic().Err(err).Msg("unable to encode pem on startup")
			}

			privateKeyFile.Close()
		}

		keys := map[string]string{}

		// check if kid was passed in
		kidPriv := s.Config.Settings.Auth.Token.KID

		// if we didn't get a kid in the settings, assign one
		if kidPriv == "" {
			kidPriv = ulids.New().String()
		}

		keys[kidPriv] = fmt.Sprintf("%v", privFileName)

		s.Config.Settings.Auth.Token.Keys = keys
	})
}

// WithTokenManager sets up the token manager for the server
func WithTokenManager() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Setup token manager
		tm, err := tokens.New(s.Config.Settings.Auth.Token)
		if err != nil {
			panic(err)
		}

		keys, err := tm.Keys()
		if err != nil {
			panic(err)
		}

		// pass to the REST handlers
		s.Config.Handler.JWTKeys = keys
		s.Config.Handler.TokenManager = tm
		s.Config.Handler.AuthManager.SetTokenManager(tm)
	})
}

// WithAuth supplies the authn and jwt config for the server
func WithAuth() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// add oauth providers
		s.Config.Handler.OauthProvider = s.Config.Settings.Auth.Providers

		// add auth middleware
		conf := authmw.NewAuthOptions(
			authmw.WithAudience(s.Config.Settings.Auth.Token.Audience),
			authmw.WithIssuer(s.Config.Settings.Auth.Token.Issuer),
			authmw.WithJWKSEndpoint(s.Config.Settings.Auth.Token.JWKSEndpoint),
			authmw.WithDBClient(s.Config.Handler.DBClient),
			authmw.WithCookieConfig(s.Config.SessionConfig.CookieConfig),
		)

		s.Config.Handler.WebAuthn = webauthn.NewWithConfig(s.Config.Settings.Auth.Providers.Webauthn)

		s.Config.GraphMiddleware = append(s.Config.GraphMiddleware, authmw.Authenticate(&conf))
		s.Config.Handler.AuthMiddleware = append(s.Config.Handler.AuthMiddleware, authmw.Authenticate(&conf))
	})
}

// WithReadyChecks adds readiness checks to the server
func WithReadyChecks(c *entx.EntClientConfig, f *fgax.Client, r *redis.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Always add a check to the primary db connection
		s.Config.Handler.AddReadinessCheck("db_primary", entx.Healthcheck(c.GetPrimaryDB()))

		// Check the secondary db, if enabled
		if s.Config.Settings.DB.MultiWrite {
			s.Config.Handler.AddReadinessCheck("db_secondary", entx.Healthcheck(c.GetSecondaryDB()))
		}

		// Check the connection to openFGA, if enabled
		if s.Config.Settings.Authz.Enabled {
			s.Config.Handler.AddReadinessCheck("fga", fgax.Healthcheck(*f))
		}

		// Check the connection to redis, if enabled
		if s.Config.Settings.Redis.Enabled {
			s.Config.Handler.AddReadinessCheck("redis", cache.Healthcheck(r))
		}
	})
}

// WithGraphRoute adds the graph handler to the server
func WithGraphRoute(srv *server.Server, c *generated.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Setup Graph API Handlers
		r := graphapi.NewResolver(c).
			WithExtensions(s.Config.Settings.Server.EnableGraphExtensions)

		// add pool to the resolver to manage the number of goroutines
		r.WithPool(
			s.Config.Settings.Server.GraphPool.MaxWorkers,
			s.Config.Settings.Server.GraphPool.MaxCapacity,
		)

		handler := r.Handler(s.Config.Settings.Server.Dev)

		// Add Graph Handler
		srv.AddHandler(handler)
	})
}

// WithMiddleware adds the middleware to the server
func WithMiddleware() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Initialize middleware if null
		if s.Config.DefaultMiddleware == nil {
			s.Config.DefaultMiddleware = []echo.MiddlewareFunc{}
		}

		// default middleware
		s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware,
			middleware.RequestID(), // add request id
			middleware.Recover(),   // recover server from any panic/fatal error gracefully
			middleware.LoggerWithConfig(middleware.LoggerConfig{
				Format: "remote_ip=${remote_ip}, method=${method}, uri=${uri}, status=${status}, session=${header:Set-Cookie}, host=${host}, referer=${referer}, user_agent=${user_agent}, route=${route}, path=${path}, auth=${header:Authorization}\n",
			}),
			echoprometheus.MetricsMiddleware(), // add prometheus metrics
			// echozap.ZapLogger(log.Desugar()),                                     // add logger middleware
			echocontext.EchoContextToContextMiddleware(),                                       // adds echo context to parent
			mime.NewWithConfig(mime.Config{DefaultContentType: httpsling.ContentTypeJSONUTF8}), // add mime middleware
		)
	})
}

// WithEventPublisher sets up the default Kafka event publisher
func WithEventPublisher() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		ep := publisher.KafkaPublisher{
			Config: s.Config.Settings.Events,
		}

		publisher := publisher.NewKafkaPublisher(ep.Config.Addresses)

		s.Config.Handler.EventManager = publisher
	})
}

// WithEmailManager sets up the default SendGrid email manager to be used to send emails to users
// on registration, password reset, etc
func WithEmailManager() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		em, err := emails.New(s.Config.Settings.Email)
		if err != nil {
			panic(err)
		}

		if err := s.Config.Settings.Email.ConsoleURLConfig.Validate(); err != nil {
			panic(err)
		}

		em.ConsoleURLConfig = s.Config.Settings.Email.ConsoleURLConfig

		if err := s.Config.Settings.Email.MarketingURLConfig.Validate(); err != nil {
			panic(err)
		}

		em.MarketingURLConfig = s.Config.Settings.Email.MarketingURLConfig

		s.Config.Handler.EmailManager = em
	})
}

// WithTaskManager sets up the default Marionette task manager to be used for delegating background tasks
func WithTaskManager() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// Start task manager
		tmConfig := marionette.Config{}

		tm := marionette.New(tmConfig)

		tm.Start()

		s.Config.Handler.TaskMan = tm
	})
}

// WithSessionManager sets up the default session manager with a 10 minute ttl
// with persistence to redis
func WithSessionManager(rc *redis.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		cc := sessions.DefaultCookieConfig

		// In order for things to work in dev mode with localhost
		// we need to se the debug cookie config
		if s.Config.Settings.Server.Dev {
			cc = &sessions.DebugOnlyCookieConfig
		} else {
			cc.Name = sessions.DefaultCookieName
		}

		if s.Config.Settings.Sessions.Domain != "" {
			cc.Domain = s.Config.Settings.Sessions.Domain
		}

		sm := sessions.NewCookieStore[map[string]any](cc,
			[]byte(s.Config.Settings.Sessions.SigningKey),
			[]byte(s.Config.Settings.Sessions.EncryptionKey),
		)

		// add session middleware, this has to be added after the authMiddleware so we have the user id
		// when we get to the session. this is also added here so its only added to the graph routes
		// REST routes are expected to add the session middleware, as required
		sessionConfig := sessions.NewSessionConfig(
			sm,
			sessions.WithPersistence(rc),
			sessions.WithSkipperFunc(authmw.SessionSkipperFunc),
		)

		// set cookie config to be used
		sessionConfig.CookieConfig = cc

		// Make the cookie session store available
		// to graph and REST endpoints
		s.Config.Handler.SessionConfig = &sessionConfig
		s.Config.SessionConfig = &sessionConfig
		s.Config.Handler.AuthManager.SetSessionConfig(&sessionConfig)
	})
}

// WithSessionMiddleware sets up the session middleware for the server
func WithSessionMiddleware() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		// add session middleware, this has to be added after the authMiddleware
		s.Config.GraphMiddleware = append(s.Config.GraphMiddleware,
			sessions.LoadAndSaveWithConfig(*s.Config.SessionConfig),
		)
	})
}

// WithAnalytics sets up the PostHog analytics manager
func WithAnalytics() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		ph := s.Config.Settings.PostHog.Init()
		if ph == nil {
			s.Config.Handler.AnalyticsClient = &analytics.EventManager{
				Enabled: false,
				Handler: nil,
			}

			return
		}

		s.Config.Handler.AnalyticsClient = &analytics.EventManager{
			Enabled: true,
			Handler: ph,
		}
	})
}

// WithOTP sets up the OTP provider
func WithOTP() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.TOTP.Enabled {
			if s.Config.Settings.TOTP.Secret == "" {
				s.Config.Settings.TOTP.Secret = ulids.New().String()
			}

			opts := []totp.ConfigOption{
				totp.WithCodeLength(s.Config.Settings.TOTP.CodeLength),
				totp.WithIssuer(s.Config.Settings.TOTP.Issuer),
				totp.WithSecret(totp.Secret{
					Version: 0,
					Key:     s.Config.Settings.TOTP.Secret,
				}),
			}

			// append redis client if enabed
			if s.Config.Settings.TOTP.WithRedis {
				opts = append(opts, totp.WithRedis(s.Config.Handler.RedisClient))
			}

			// setup new opt manager
			otpMan := totp.NewOTP(
				opts...,
			)

			s.Config.Handler.OTPManager = &totp.Manager{
				TOTPManager: otpMan,
			}
		}
	})
}

// WithRateLimiter sets up the rate limiter for the server
func WithRateLimiter() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Ratelimit.Enabled {
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, ratelimit.RateLimiterWithConfig(&s.Config.Settings.Ratelimit))
		}
	})
}

// WithSecureMW sets up the secure middleware for the server
func WithSecureMW() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.Secure.Enabled {
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, secure.Secure(&s.Config.Settings.Server.Secure))
		}
	})
}

// WithRedirects sets up the redirects for the server
func WithRedirects() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.Redirects.Enabled {
			redirects := s.Config.Settings.Server.Redirects
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, redirect.NewWithConfig(redirects))
		}
	})
}

// WithCacheHeaders sets up the cache control headers for the server
func WithCacheHeaders() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.CacheControl.Enabled {
			cacheConfig := s.Config.Settings.Server.CacheControl
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, cachecontrol.NewWithConfig(cacheConfig))
		}
	})
}

// WithCORS sets up the CORS middleware for the server
func WithCORS() ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if s.Config.Settings.Server.CORS.Enabled {
			s.Config.DefaultMiddleware = append(s.Config.DefaultMiddleware, cors.New(s.Config.Settings.Server.CORS.AllowOrigins))
		}
	})
}
