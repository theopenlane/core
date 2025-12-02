package logx

import (
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

// Options is a struct that holds the options for a Logger
type Options struct {
	zcontext zerolog.Context
	level    log.Lvl
	prefix   string
}

// ConfigSetter is a function that sets an option on a Logger
type ConfigSetter func(opts *Options)

// newOptions returns a new Options instance
func newOptions(log zerolog.Logger, setters []ConfigSetter) *Options {
	elvl, _ := MatchZeroLevel(log.GetLevel())

	opts := &Options{
		zcontext: log.With(),
		level:    elvl,
	}

	for _, set := range setters {
		set(opts)
	}

	return opts
}

// WithLevel sets the level on a Logger
func WithLevel(level log.Lvl) ConfigSetter {
	return func(opts *Options) {
		zlvl, elvl := MatchEchoLevel(level)

		applyZeroLevel(opts, zlvl, elvl)
	}
}

// WithTimestamp sets the timestamp on a Logger
func WithTimestamp() ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Timestamp()
	}
}

// WithCaller sets the caller on a Logger
func WithCaller() ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Caller()
	}
}

// WithZeroLevel sets the zerolog level on a Logger while keeping the echo level in sync
func WithZeroLevel(level zerolog.Level) ConfigSetter {
	return func(opts *Options) {
		elvl, _ := MatchZeroLevel(level)

		applyZeroLevel(opts, level, elvl)
	}
}

func applyZeroLevel(opts *Options, zlvl zerolog.Level, elvl log.Lvl) {
	opts.zcontext = opts.zcontext.Logger().Level(zlvl).With()
	opts.level = elvl
}

func withPrefix(prefix string) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Str("prefix", prefix)
		opts.prefix = prefix
	}
}
