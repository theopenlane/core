package auth

import (
	"context"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/echox/middleware"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"

	api "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
)

// Option allows users to optionally supply configuration to the Authorization middleware.
type Option func(opts *Options)

// Options is constructed from variadic AuthOption arguments with reasonable defaults.
type Options struct {
	// KeysURL endpoint to the JWKS public keys on the server
	KeysURL string `default:"http://localhost:17608/.well-known/jwks.json"`
	// Audience to verify on tokens
	Audience string `default:"http://localhost:17608"`
	// Issuer to verify on tokens
	Issuer string `default:"http://localhost:17608"`
	// MinRefreshInterval to cache the JWKS public keys
	MinRefreshInterval time.Duration `default:"5m"`
	// Context to control the lifecycle of the background fetch routine
	Context context.Context

	// CookieConfig to set the cookie configuration for the auth middleware
	CookieConfig *sessions.CookieConfig

	//  validator constructed by the auth options (can be directly supplied by the user).
	validator tokens.Validator
	// reauth constructed by the auth options (can be directly supplied by the user).
	reauth Reauthenticator

	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper
	// BeforeFunc  defines a function which is executed just before the middleware
	BeforeFunc     middleware.BeforeFunc
	AllowAnonymous bool

	// Used to check other auth types like personal access tokens
	DBClient *ent.Client
	// RedisClient is used to set the permission cache in the context
	RedisClient *redis.Client
}

// Reauthenticator generates new access and refresh pair given a valid refresh token.
type Reauthenticator interface {
	Refresh(context.Context, *api.RefreshRequest) (*api.LoginReply, error)
}

// DefaultAuthOptions is the default auth options used by the middleware.
var DefaultAuthOptions = Options{
	KeysURL:            "http://localhost:17608/.well-known/jwks.json",
	Audience:           "http://localhost:17608",
	Issuer:             "http://localhost:17608",
	MinRefreshInterval: 5 * time.Minute, //nolint:mnd
	Skipper:            middleware.DefaultSkipper,
	CookieConfig:       sessions.DefaultCookieConfig,
}

// NewAuthOptions creates an AuthOptions object with reasonable defaults and any user
// supplied input from the AuthOption variadic arguments.
func NewAuthOptions(opts ...Option) (conf Options) {
	conf = DefaultAuthOptions

	for _, opt := range opts {
		opt(&conf)
	}

	// Create a context if one has not been supplied by the user.
	if conf.Context == nil && conf.validator == nil {
		conf.Context = context.Background()
	}

	return conf
}

// Validator returns the user supplied validator or constructs a new JWKS Cache
// Validator from the supplied options. If the options are invalid or the validator
// cannot be created an error is returned
func (conf *Options) Validator() (tokens.Validator, error) {
	if conf.validator != nil {
		return conf.validator, nil
	}

	httprcclient := httprc.NewClient()

	cache, err := jwk.NewCache(conf.Context, httprcclient)
	if err != nil {
		return nil, err
	}

	if err := cache.Register(context.Background(), conf.KeysURL, jwk.WithMinInterval(conf.MinRefreshInterval)); err != nil {
		return nil, ErrUnableToConstructValidator
	}

	conf.validator, err = tokens.NewCachedJWKSValidator(cache, conf.KeysURL, conf.Audience, conf.Issuer)
	if err != nil {
		return nil, err
	}

	return conf.validator, nil
}

// WithLocalValidator returns a new JWKS Validator constructed from the supplied options using
// the local keys instead of fetching them from the server
func (conf *Options) WithLocalValidator() error {
	if conf.validator != nil {
		return nil
	}

	httprcclient := httprc.NewClient()

	cache, err := jwk.NewCache(conf.Context, httprcclient)
	if err != nil {
		return err
	}

	if err := cache.Register(conf.Context, conf.KeysURL, jwk.WithMinInterval(conf.MinRefreshInterval)); err != nil {
		return ErrUnableToConstructValidator
	}

	keys, err := conf.DBClient.TokenManager.Keys()
	if err != nil {
		return err
	}

	conf.validator = tokens.NewJWKSValidator(keys, conf.Audience, conf.Issuer)

	return nil
}

// WithAuthOptions allows the user to update the default auth options with an auth
// options struct to set many options values at once. Zero values are ignored, so if
// using this option, the defaults will still be preserved if not set on the input.
func WithAuthOptions(opts Options) Option {
	return func(conf *Options) {
		if opts.KeysURL != "" {
			conf.KeysURL = opts.KeysURL
		}

		if opts.Audience != "" {
			conf.Audience = opts.Audience
		}

		if opts.Issuer != "" {
			conf.Issuer = opts.Issuer
		}

		if opts.MinRefreshInterval != 0 {
			conf.MinRefreshInterval = opts.MinRefreshInterval
		}

		if opts.Context != nil {
			conf.Context = opts.Context
		}
	}
}

// WithAllowAnonymous allows anonymous access to the API.
func WithAllowAnonymous(allow bool) Option {
	return func(opts *Options) {
		opts.AllowAnonymous = allow
	}
}

// WithJWKSEndpoint allows the user to specify an alternative endpoint to fetch the JWKS
// public keys from. This is useful for testing or for different environments.
func WithJWKSEndpoint(url string) Option {
	return func(opts *Options) {
		opts.KeysURL = url
	}
}

// WithAudience allows the user to specify an alternative audience.
func WithAudience(audience string) Option {
	return func(opts *Options) {
		opts.Audience = audience
	}
}

// WithIssuer allows the user to specify an alternative issuer.
func WithIssuer(issuer string) Option {
	return func(opts *Options) {
		opts.Issuer = issuer
	}
}

// WithMinRefreshInterval allows the user to specify an alternative minimum duration
// between cache refreshes to control refresh behavior for the JWKS public keys.
func WithMinRefreshInterval(interval time.Duration) Option {
	return func(opts *Options) {
		opts.MinRefreshInterval = interval
	}
}

// WithContext allows the user to specify an external, cancelable context to control
// the background refresh behavior of the JWKS cache.
func WithContext(ctx context.Context) Option {
	return func(opts *Options) {
		opts.Context = ctx
	}
}

// WithValidator allows the user to specify an alternative validator to the auth
// middleware. This is particularly useful for testing authentication.
func WithValidator(validator tokens.Validator) Option {
	return func(opts *Options) {
		opts.validator = validator
	}
}

// WithReauthenticator allows the user to specify a reauthenticator to the auth
// middleware.
func WithReauthenticator(reauth Reauthenticator) Option {
	return func(opts *Options) {
		opts.reauth = reauth
	}
}

// WithSkipperFunc allows the user to specify a skipper function for the middleware
func WithSkipperFunc(skipper middleware.Skipper) Option {
	return func(opts *Options) {
		opts.Skipper = skipper
	}
}

// WithBeforeFunc allows the user to specify a function to happen before the auth middleware
func WithBeforeFunc(before middleware.BeforeFunc) Option {
	return func(opts *Options) {
		opts.BeforeFunc = before
	}
}

// WithDBClient is a function that returns an AuthOption function which sets the DBClient field of AuthOptions.
// The DBClient field is used to specify the database client to be to check authentication with personal access tokens.
func WithDBClient(client *ent.Client) Option {
	return func(opts *Options) {
		opts.DBClient = client
	}
}

// WithCookieConfig allows the user to specify a cookie configuration for the auth middleware
// in order to override the default cookie configuration.
func WithCookieConfig(cookieConfig *sessions.CookieConfig) Option {
	return func(opts *Options) {
		opts.CookieConfig = cookieConfig
	}
}

// WithRedisClient allows the user to specify a Redis client for the auth middleware
// in order to set the permission cache in the context.
func WithRedisClient(redisClient *redis.Client) Option {
	return func(opts *Options) {
		opts.RedisClient = redisClient
	}
}
