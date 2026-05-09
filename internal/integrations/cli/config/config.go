//go:build examples

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DefaultConfigDir is the repo-relative directory that holds the CLI config
// file when no explicit --config flag is provided
const DefaultConfigDir = "internal/integrations/cli/config"

// DefaultConfigFile is the default config file name resolved inside
// DefaultConfigDir
const DefaultConfigFile = ".integrations.yaml"

// Defaults applied to persistent flags and struct tags. These match the
// values produced by `task cli:user:all` in the top-level cli Taskfile.
const (
	DefaultHost     = "http://localhost:17608"
	DefaultEmail    = "mitb@theopenlane.io"
	DefaultPassword = "mattisthebest1234"
	DefaultAuthMode = "auto"
)

// Options configures the CLI config loader
type Options struct {
	// AppName is used for default config file resolution (~/.<appname>.yaml)
	AppName string
	// Config is the koanf instance to populate
	Config *koanf.Koanf
	// ConfigFile is a pointer to the --config flag value
	ConfigFile *string
	// EnvPrefix is the environment variable prefix whose stripped, lowercased
	// form becomes the top-level koanf namespace (OPENLANE_ → openlane)
	EnvPrefix string
}

// Loader handles env/file/flag loading and logging setup
type Loader struct {
	appName    string
	config     *koanf.Koanf
	configFile *string
	envPrefix  string
}

// New returns a Loader configured with the provided options
func New(opts Options) *Loader {
	return &Loader{
		appName:    opts.AppName,
		config:     opts.Config,
		configFile: opts.ConfigFile,
		envPrefix:  opts.EnvPrefix,
	}
}

// InitSources loads env and config file sources. Call from cobra.OnInitialize.
func (l *Loader) InitSources() error {
	if err := l.loadEnv(); err != nil {
		return err
	}

	return l.loadConfigFile()
}

// Apply loads flag values for the invoked command and configures logging
func (l *Loader) Apply(cmd *cobra.Command) error {
	if err := l.loadFlags(cmd); err != nil {
		return err
	}

	l.setupLogging()

	return nil
}

// loadConfigFile loads the first available config file path
func (l *Loader) loadConfigFile() error {
	if l.config == nil || l.configFile == nil {
		return nil
	}

	path, err := l.resolveConfigPath()
	if err != nil || path == "" {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return l.config.Load(file.Provider(path), yaml.Parser())
}

// resolveConfigPath returns the config file path: the explicit --config flag
// value when set, else DefaultConfigDir/DefaultConfigFile resolved relative to
// the current working directory (repo root)
func (l *Loader) resolveConfigPath() (string, error) {
	if explicit := strings.TrimSpace(*l.configFile); explicit != "" {
		return explicit, nil
	}

	path := filepath.Join(DefaultConfigDir, DefaultConfigFile)
	*l.configFile = path

	return path, nil
}

// loadEnv loads environment variables under the configured prefix into a
// koanf namespace derived from the prefix (OPENLANE_AUTH__EMAIL →
// openlane.auth.email). Double underscores separate path segments; single
// underscores become dashes.
func (l *Loader) loadEnv() error {
	prefix := strings.TrimSpace(l.envPrefix)
	if prefix == "" {
		return nil
	}

	namespace := strings.ToLower(strings.TrimRight(prefix, "_"))

	return l.config.Load(env.ProviderWithValue(prefix, ".", func(s, v string) (string, any) {
		key := normalizeEnvKey(prefix, namespace, s)

		if strings.Contains(v, ",") {
			return key, strings.Split(v, ",")
		}

		return key, v
	}), nil)
}

// normalizeEnvKey converts OPENLANE_AUTH__EMAIL into openlane.auth.email
func normalizeEnvKey(prefix, namespace, key string) string {
	trimmed := strings.ToLower(strings.TrimPrefix(key, prefix))
	trimmed = strings.ReplaceAll(trimmed, "__", ".")
	trimmed = strings.ReplaceAll(trimmed, "_", "-")

	if namespace == "" {
		return trimmed
	}

	return namespace + "." + trimmed
}

// loadFlags loads command flags into koanf, routing annotated flags to their
// declared config path so flat CLI flags land at their nested location
func (l *Loader) loadFlags(cmd *cobra.Command) error {
	if l.config == nil || cmd == nil {
		return nil
	}

	fs := cmd.Flags()

	return l.config.Load(posflag.ProviderWithFlag(fs, l.config.Delim(), l.config, func(f *pflag.Flag) (string, any) {
		key := f.Name
		if annotated := f.Annotations[ConfigKeyAnnotation]; len(annotated) > 0 {
			key = annotated[0]
		}

		return key, posflag.FlagVal(fs, f)
	}), nil)
}

// setupLogging configures the zerolog logger using logging.debug / logging.pretty
func (l *Loader) setupLogging() {
	if l.config == nil {
		return
	}

	log.Logger = zerolog.New(os.Stderr).
		With().Timestamp().
		Logger().
		With().Str("app", l.appName).
		Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if l.config.Bool("logging.debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if l.config.Bool("logging.pretty") {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			FormatCaller: func(i any) string {
				return filepath.Base(fmt.Sprintf("%s", i))
			},
		})
	}
}

// Config defines the structured CLI configuration
type Config struct {
	// Openlane holds API host and authentication defaults used by every
	// subcommand that talks to the Openlane API
	Openlane OpenlaneConfig `json:"openlane" koanf:"openlane"`
}

// OpenlaneConfig holds connection and auth defaults
type OpenlaneConfig struct {
	// Host is the base URL for the Openlane API
	Host string `json:"host" koanf:"host" default:"http://localhost:17608"`
	// Auth controls authentication for API requests
	Auth OpenlaneAuthConfig `json:"auth" koanf:"auth"`
}

// OpenlaneAuthConfig defines credentials and auth strategy
type OpenlaneAuthConfig struct {
	// Mode selects the auth strategy: auto prefers token/pat when present and
	// falls back to credential login
	Mode string `json:"mode" koanf:"mode" default:"credentials"`
	// Token is the API bearer token (PAT or API token)
	Token string `json:"token" koanf:"token" default:"" sensitive:"true"`
	// PAT is a personal access token; used as a fallback when Token is empty
	PAT string `json:"pat" koanf:"pat" default:"" sensitive:"true"`
	// Email is the login email used for auth-mode=credentials
	Email string `json:"email" koanf:"email" default:"mitb@theopenlane.io"`
	// Password is the login password used for auth-mode=credentials
	Password string `json:"password" koanf:"password" default:"mattisthebest1234" sensitive:"true"`
}
