package logx

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/shared/logx/consolelog"
)

// LoggerConfig controls how loggers produced by Configure behave.
type LoggerConfig struct {
	// Level sets the zerolog level. Use zerolog.NoLevel to keep the existing global default.
	Level zerolog.Level
	// Pretty toggles human-readable console output instead of structured JSON.
	Pretty bool
	// Writer controls where logs are written. Defaults to os.Stdout.
	Writer io.Writer
	// IncludeCaller attaches caller information to each log entry.
	IncludeCaller bool
	// Hooks are applied to every log event.
	Hooks []zerolog.HookFunc
	// WithEcho instructs Configure to build an echo-compatible logger.
	WithEcho bool
	// SetGlobal updates zerolog's global logger (`log.Logger`) using the configured settings.
	SetGlobal bool
}

// Result contains the loggers produced by Configure.
type LoggerSet struct {
	Logger zerolog.Logger
	Echo   *Logger
}

// Configure builds loggers according to the supplied configuration.
func Configure(cfg LoggerConfig) LoggerSet {
	root, setters, output := buildRootLogger(cfg)

	if cfg.SetGlobal {
		if cfg.Level != zerolog.NoLevel {
			zerolog.SetGlobalLevel(cfg.Level)
		}

		log.Logger = root
	}

	var echoLogger *Logger
	if cfg.WithEcho {
		echoLogger = newLoggerFromExisting(root, output, setters)
	}

	return LoggerSet{
		Logger: root,
		Echo:   echoLogger,
	}
}

func buildRootLogger(cfg LoggerConfig) (zerolog.Logger, []ConfigSetter, io.Writer) {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	output := writer
	if cfg.Pretty {
		cw := consolelog.NewConsoleWriter()
		output = &cw
	}

	setters := []ConfigSetter{
		WithTimestamp(),
		WithSeverityMapping(),
	}

	if cfg.IncludeCaller {
		setters = append(setters, WithCaller())
	}

	if cfg.Level != zerolog.NoLevel {
		setters = append(setters, WithZeroLevel(cfg.Level))
	}

	for _, hook := range cfg.Hooks {
		if hook != nil {
			setters = append(setters, func(opts *Options) {
				opts.zcontext = opts.zcontext.Logger().Hook(hook).With()
			})
		}
	}

	opts := newOptions(zerolog.New(output), setters)

	return opts.zcontext.Logger(), append([]ConfigSetter(nil), setters...), output
}
