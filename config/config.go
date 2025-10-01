package config

import (
	"crypto/tls"
	"os"
	"reflect"

	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/mcuadros/go-defaults"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/beacon/otelx"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/theopenlane/utils/cache"

	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/internal/ent/hush/crypto"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/middleware/cachecontrol"
	"github.com/theopenlane/core/pkg/middleware/cors"
	"github.com/theopenlane/core/pkg/middleware/csrf"
	"github.com/theopenlane/core/pkg/middleware/mime"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/pkg/middleware/redirect"
	"github.com/theopenlane/core/pkg/middleware/secure"
	"github.com/theopenlane/core/pkg/objects"
)

// Config contains the configuration for the core server
type Config struct {
	// Domain provides a global domain value for other modules to inherit
	Domain string `json:"domain" koanf:"domain" default:""`
	// RefreshInterval determines how often to reload the config
	RefreshInterval time.Duration `json:"refreshInterval" koanf:"refreshInterval" default:"10m"`
	// Server contains the echo server settings
	Server Server `json:"server" koanf:"server"`
	// EntConfig contains the ent configuration used by the ent middleware
	EntConfig entconfig.Config `json:"entConfig" koanf:"entConfig"`
	// Auth contains the authentication token settings and provider(s)
	Auth Auth `json:"auth" koanf:"auth"`
	// Authz contains the authorization settings for fine grained access control
	Authz fgax.Config `json:"authz" koanf:"authz"`
	// DB contains the database configuration for the ent client
	DB entx.Config `json:"db" koanf:"db"`
	// JobQueue contains the configuration for the job queue (river) client
	JobQueue riverqueue.Config `json:"jobQueue" koanf:"jobQueue"`
	// Redis contains the redis configuration for the key-value store
	Redis cache.Config `json:"redis" koanf:"redis"`
	// Tracer contains the tracing config for opentelemetry
	Tracer otelx.Config `json:"tracer" koanf:"tracer"`
	// Email contains email sending configuration for the server
	Email emailtemplates.Config `json:"email" koanf:"email"`
	// Sessions config for user sessions and cookies
	Sessions sessions.Config `json:"sessions" koanf:"sessions"`
	// TOTP contains the configuration for the TOTP provider
	TOTP totp.Config `json:"totp" koanf:"totp"`
	// Ratelimit contains the configuration for the rate limiter
	Ratelimit ratelimit.Config `json:"ratelimit" koanf:"ratelimit"`
	// ObjectStorage contains the configuration for the object storage backend
	ObjectStorage objects.Config `json:"objectStorage" koanf:"objectStorage"`
	// Entitlements contains the configuration for the entitlements service
	Entitlements entitlements.Config `json:"subscription" koanf:"subscription"`
	// Keywatcher contains the configuration for the key watcher that manages JWT signing keys
	Keywatcher KeyWatcher `json:"keywatcher" koanf:"keywatcher"`
	// Slack contains settings for Slack notifications
	Slack Slack `json:"slack" koanf:"slack"`
	// IntegrationOauthProvider contains the OAuth provider configuration for integrations (separate from auth.providers)
	IntegrationOauthProvider handlers.IntegrationOauthProviderConfig `json:"integrationOauthProvider" koanf:"integrationOauthProvider"`
}

// Server settings for the echo server
type Server struct {
	// Debug enables debug mode for the server, set via the command line flag
	Debug bool `json:"-" koanf:"-" default:"false"`
	// Pretty enables pretty logging output, defaults to json format, set via the command line flag
	Pretty bool `json:"-" koanf:"-" default:"false"`
	// Dev enables echo's dev mode options
	Dev bool `json:"dev" koanf:"dev" default:"false"`
	// Listen sets the listen address to serve the echo server on
	Listen string `json:"listen" koanf:"listen" jsonschema:"required" default:":17608"`
	// MetricsPort sets the port for the metrics endpoint
	MetricsPort string `json:"metricsPort" koanf:"metricsPort" default:":17609"`
	// ShutdownGracePeriod sets the grace period for in flight requests before shutting down
	ShutdownGracePeriod time.Duration `json:"shutdownGracePeriod" koanf:"shutdownGracePeriod" default:"10s"`
	// ReadTimeout sets the maximum duration for reading the entire request including the body
	ReadTimeout time.Duration `json:"readTimeout" koanf:"readTimeout" default:"15s"`
	// WriteTimeout sets the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `json:"writeTimeout" koanf:"writeTimeout" default:"15s"`
	// IdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled
	IdleTimeout time.Duration `json:"idleTimeout" koanf:"idleTimeout" default:"30s"`
	// ReadHeaderTimeout sets the amount of time allowed to read request headers
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" koanf:"readHeaderTimeout" default:"2s"`
	// TLS contains the tls configuration settings
	TLS TLS `json:"tls" koanf:"tls"`
	// CORS contains settings to allow cross origin settings and insecure cookies
	CORS cors.Config `json:"cors" koanf:"cors"`
	// Secure contains settings for the secure middleware
	Secure secure.Config `json:"secure" koanf:"secure"`
	// Redirect contains settings for the redirect middleware
	Redirects redirect.Config `json:"redirects" koanf:"redirects"`
	// CacheControl contains settings for the cache control middleware
	CacheControl cachecontrol.Config `json:"cacheControl" koanf:"cacheControl"`
	// Mime contains settings for the mime middleware
	Mime mime.Config `json:"mime" koanf:"mime"`
	// GraphPool contains settings for the goroutine pool used by the graph resolvers
	GraphPool PondPool `json:"graphPool" koanf:"graphPool"`
	// EnableGraphExtensions enables the graph extensions for the graph resolvers
	EnableGraphExtensions bool `json:"enableGraphExtensions" koanf:"enableGraphExtensions" default:"true"`
	// ComplexityLimit sets the maximum complexity allowed for a query
	ComplexityLimit int `json:"complexityLimit" koanf:"complexityLimit" default:"100"`
	// MaxResultLimit sets the maximum number of results allowed for a query
	MaxResultLimit int `json:"maxResultLimit" koanf:"maxResultLimit" default:"100"`
	// CSRFProtection enables CSRF protection for the server
	CSRFProtection csrf.Config `json:"csrfProtection" koanf:"csrfProtection"`
	// SecretManagerSecret is the name of the GCP Secret Manager secret containing the JWT signing key
	SecretManagerSecret string `json:"secretManager" koanf:"secretManager" default:"" sensitive:"true"`
	// DefaultTrustCenterDomain is the default domain to use for the trust center if no custom domain is set
	DefaultTrustCenterDomain string `json:"defaultTrustCenterDomain" koanf:"defaultTrustCenterDomain" default:""`
	// FieldLevelEncryption contains the configuration for field level encryption
	FieldLevelEncryption crypto.Config `json:"fieldLevelEncryption" koanf:"fieldLevelEncryption"`
	// TrustCenterCnameTarget is the cname target for the trust center
	// Used for mapping the vanity domains to the trust centers
	TrustCenterCnameTarget string `json:"trustCenterCnameTarget" koanf:"trustCenterCnameTarget" default:""`
}

// KeyWatcher contains settings for the key watcher that manages JWT signing keys
type KeyWatcher struct {
	// Enabled indicates whether the key watcher is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// KeyDir is the path to the directory containing PEM keys for JWT signing
	KeyDir string `json:"keyDir" koanf:"keyDir" default:"./keys"`
	// ExternalSecretsIntegration enables integration with external secret management systems (specifically GCP secret manager today)
	ExternalSecretsIntegration bool `json:"externalSecretsIntegration" koanf:"externalSecretsIntegration" default:"false"`
	// SecretManagerSecret is the name of the GCP Secret Manager secret containing the JWT signing key
	SecretManagerSecret string `json:"secretManager" koanf:"secretManager" default:"" sensitive:"true"`
}

// Auth settings including oauth2 providers and token configuration
type Auth struct {
	// Enabled authentication on the server, not recommended to disable
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Token contains the token config settings for the issued tokens
	Token tokens.Config `json:"token" koanf:"token" jsonschema:"required" alias:"tokenconfig"`
	// SupportedProviders are the supported oauth providers that have been configured
	SupportedProviders []string `json:"supportedProviders" koanf:"supportedProviders"`
	// Providers contains supported oauth2 providers configuration
	Providers handlers.OauthProviderConfig `json:"providers" koanf:"providers"`
}

// TLS settings for the server for secure connections
type TLS struct {
	// Config contains the tls.Config settings
	Config *tls.Config `json:"config" koanf:"config" jsonschema:"-"`
	// Enabled turns on TLS settings for the server
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// CertFile location for the TLS server
	CertFile string `json:"certFile" koanf:"certFile" default:"server.crt"`
	// CertKey file location for the TLS server
	CertKey string `json:"certKey" koanf:"certKey" default:"server.key"`
	// AutoCert generates the cert with letsencrypt, this does not work on localhost
	AutoCert bool `json:"autoCert" koanf:"autoCert" default:"false"`
}

// PondPool contains the settings for the goroutine pool
type PondPool struct {
	// MaxWorkers is the maximum number of workers in the pool
	MaxWorkers int `json:"maxWorkers" koanf:"maxWorkers" default:"100"`
}

// Slack contains settings for Slack notifications
type Slack struct {
	// WebhookURL is the Slack webhook to post messages to
	WebhookURL string `json:"webhookURL" koanf:"webhookURL" sensitive:"true"`
	// NewSubscriberMessageFile is the path to the template used for new subscriber notifications
	NewSubscriberMessageFile string `json:"newSubscriberMessageFile" koanf:"newSubscriberMessageFile"`
	// NewUserMessageFile is the path to the template used for new user notifications
	NewUserMessageFile string `json:"newUserMessageFile" koanf:"newUserMessageFile"`
}

var (
	defaultConfigFilePath = "./config/.config.yaml"
)

// Option configures the Config
type Option func(*Config)

// New creates a Config with the supplied options applied
func New(opts ...Option) *Config {
	cfg := &Config{}
	defaults.SetDefaults(cfg)

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithDomain sets a global domain value
func WithDomain(domain string) Option {
	return func(c *Config) {
		c.Domain = domain
	}
}

// applyDomainOverrides sets related domain values when unset
func (c *Config) applyDomainOverrides() {
	if c.Domain == "" {
		return
	}

	applyDomain(reflect.ValueOf(c).Elem(), c.Domain)
}

// applyDomain recursively applies the domain to all fields that are tagged with `domain:"inherit"`
// It also handles domain prefixes and suffixes if specified in the struct tags
func applyDomain(v reflect.Value, domain string) {
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Ptr:
		// If it's a pointer, dereference and recurse
		if !v.IsNil() {
			applyDomain(v.Elem(), domain)
		}
	case reflect.Struct:
		// Iterate over all struct fields
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			sf := v.Type().Field(i)

			// Skip unexported fields
			if sf.PkgPath != "" {
				continue
			}

			// Handle domain inheritance with prefix and suffix support
			if sf.Tag.Get("domain") == "inherit" && f.CanSet() && f.String() == "" {
				domainValue := domain

				// Apply domain prefix if specified
				if prefix := sf.Tag.Get("domainPrefix"); prefix != "" {
					// Handle multiple prefixes for slices of strings
					if f.Kind() == reflect.Slice && f.Type().Elem().Kind() == reflect.String {
						prefixes := strings.Split(prefix, ",")

						var values []reflect.Value

						// Build slice values with prefix + domain
						for _, p := range prefixes {
							values = append(values, reflect.ValueOf(strings.TrimSpace(p)+"."+domain))
						}

						slice := reflect.MakeSlice(f.Type(), len(values), len(values))

						for i, val := range values {
							slice.Index(i).Set(val)
						}

						f.Set(slice)

						continue
					}
					// For string fields, just prepend prefix
					domainValue = prefix + "." + domain
				}

				// Apply domain suffix if specified
				if suffix := sf.Tag.Get("domainSuffix"); suffix != "" {
					domainValue += suffix
				}

				// Set the string field value
				if f.Kind() == reflect.String {
					f.SetString(domainValue)
				}

				continue
			}
			// Recurse into nested structs and pointers
			applyDomain(f, domain)
		}
	}
}

// Load is responsible for loading the configuration from a YAML file and environment variables.
// If the `cfgFile` is empty or nil, it sets the default configuration file path.
// Config settings are taken from default values, then from the config file, and finally from environment
// the later overwriting the former.
func Load(cfgFile *string) (*Config, error) {
	k := koanf.New(".")

	if cfgFile == nil || *cfgFile == "" {
		*cfgFile = defaultConfigFilePath
	}

	if _, err := os.Stat(*cfgFile); err != nil {
		if os.IsNotExist(err) {
			log.Warn().Err(err).Msg("config file not found, proceeding with default configuration")
		}
	}

	conf := New()

	// parse yaml config
	if err := k.Load(file.Provider(*cfgFile), yaml.Parser()); err != nil {
		log.Warn().Err(err).Msg("failed to load config file - ensure the .config.yaml is present and valid or use environment variables to set the configuration")
	}

	// unmarshal the config
	if err := k.Unmarshal("", &conf); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config file")
	}

	// load env vars
	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "CORE_",
		TransformFunc: func(key, v string) (string, any) {
			key = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(key, "CORE_")), "_", ".")

			if strings.Contains(v, ",") {
				return key, strings.Split(v, ",")
			}

			return key, v
		},
	}), nil); err != nil {
		log.Warn().Err(err).Msg("failed to load env vars, some settings may not be applied")
	}

	// unmarshal the env vars
	if err := k.Unmarshal("", &conf); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal env vars")
	}

	conf.applyDomainOverrides()

	return conf, nil
}
