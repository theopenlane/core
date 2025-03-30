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

		opts.zcontext = opts.zcontext.Logger().Level(zlvl).With()
		opts.level = elvl
	}
}

// WithField sets a field on a Logger
func WithField(name string, value any) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Interface(name, value)
	}
}

// WithFields sets fields on a Logger
func WithFields(fields map[string]any) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Fields(fields)
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

// WithCallerWithSkipFrameCount sets the caller with a skip frame count on a Logger
func WithCallerWithSkipFrameCount(skipFrameCount int) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.CallerWithSkipFrameCount(skipFrameCount)
	}
}

// WithPrefix sets the prefix on a Logger
func WithPrefix(prefix string) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Str("prefix", prefix)
	}
}

// WithHook sets the hook on a Logger
func WithHook(hook zerolog.Hook) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Logger().Hook(hook).With()
	}
}

// WithHookFunc sets the hook function on a Logger
func WithHookFunc(hook zerolog.HookFunc) ConfigSetter {
	return func(opts *Options) {
		opts.zcontext = opts.zcontext.Logger().Hook(hook).With()
	}
}
