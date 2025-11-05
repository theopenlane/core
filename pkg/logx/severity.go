package logx

import "github.com/rs/zerolog"

// Mapping of zerolog levels to GCP severity strings
var levelToSeverity = map[zerolog.Level]string{
	zerolog.TraceLevel: "DEBUG",
	zerolog.DebugLevel: "DEBUG",
	zerolog.InfoLevel:  "INFO",
	zerolog.WarnLevel:  "WARNING",
	zerolog.ErrorLevel: "ERROR",
	zerolog.FatalLevel: "CRITICAL",
	zerolog.PanicLevel: "ALERT",
	zerolog.NoLevel:    "DEFAULT",
	zerolog.Disabled:   "DEFAULT",
}

// severityForLevel returns the GCP severity string for a given zerolog level
func severityForLevel(level zerolog.Level) string {
	if severity, ok := levelToSeverity[level]; ok {
		return severity
	}

	return "DEFAULT"
}

// WithSeverityMapping ensures each log event has a severity field aligned with GCP expectations
func WithSeverityMapping() ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Logger().Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, _ string) {
			if e == nil {
				return
			}

			e.Str("severity", severityForLevel(level))
		})).With()
	}
}
