package logx

import (
	"fmt"
	"io"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/pkg/logx/consolelog"
)

// Logger is a wrapper around zerolog.Logger that provides an implementation of the echo.Logger interface
type Logger struct {
	log     zerolog.Logger
	out     io.Writer
	level   log.Lvl
	prefix  string
	setters []ConfigSetter
}

// New returns a new Logger instance
func New(out io.Writer, setters ...ConfigSetter) *Logger {
	switch l := out.(type) {
	case zerolog.Logger:
		return newLogger(l, setters)
	default:
		return newLogger(zerolog.New(out), setters)
	}
}

// CreateLogger creates a new logger for the echo server based on the provided configuration
func CreateLogger(level log.Lvl, pretty bool) *Logger {
	var logger *Logger

	setters := []ConfigSetter{
		WithLevel(level),
		WithTimestamp(),
		WithCaller(),
	}

	// if PrettyLog is enabled, use the console writer for pretty logging
	// otherwise, use the default stdout writer (json format)
	if pretty {
		cw := consolelog.NewConsoleWriter()
		logger = New(
			&cw,
			setters...,
		)
	} else {
		logger = New(
			os.Stdout,
			setters...,
		)
	}

	return logger
}

// From returns a new Logger instance using existing zerolog log
func From(log zerolog.Logger, setters ...ConfigSetter) *Logger {
	return newLogger(log, setters)
}

// newLogger returns a new Logger instance
func newLogger(log zerolog.Logger, setters []ConfigSetter) *Logger {
	opts := newOptions(log, setters)

	return &Logger{
		log:     opts.zcontext.Logger(),
		out:     nil,
		level:   opts.level,
		prefix:  opts.prefix,
		setters: setters,
	}
}

// Write implements the io.Writer interface for Logger
func (l *Logger) Write(p []byte) (n int, err error) {
	return l.log.Write(p)
}

// Debug implements echo.Logger interface
func (l *Logger) Debug(i ...any) {
	l.log.Debug().Msg(fmt.Sprint(i...))
}

// Debugf implements echo.Logger interface
func (l *Logger) Debugf(format string, i ...any) {
	l.log.Debug().Msgf(format, i...)
}

// Debugj implements echo.Logger interface
func (l *Logger) Debugj(j log.JSON) {
	l.logJSON(l.log.Debug(), j)
}

// Info implements echo.Logger interface
func (l *Logger) Info(i ...any) {
	l.log.Info().Msg(fmt.Sprint(i...))
}

// Infof implements echo.Logger interface
func (l *Logger) Infof(format string, i ...any) {
	l.log.Info().Msgf(format, i...)
}

// Infoj implements echo.Logger interface
func (l *Logger) Infoj(j log.JSON) {
	l.logJSON(l.log.Info(), j)
}

// Warn implements echo.Logger interface
func (l *Logger) Warn(i ...any) {
	l.log.Warn().Msg(fmt.Sprint(i...))
}

// Warnf implements echo.Logger interface
func (l *Logger) Warnf(format string, i ...any) {
	l.log.Warn().Msgf(format, i...)
}

// Warnj implements echo.Logger interface
func (l *Logger) Warnj(j log.JSON) {
	l.logJSON(l.log.Warn(), j)
}

// Error implements echo.Logger interface
func (l *Logger) Error(err error) {
	l.log.Error().Err(err).Send()
}

// Errorf implements echo.Logger interface
func (l *Logger) Errorf(format string, i ...any) {
	l.log.Error().Msgf(format, i...)
}

// Errorj implements echo.Logger interface
func (l *Logger) Errorj(j log.JSON) {
	l.logJSON(l.log.Error(), j)
}

// Fatal implements echo.Logger interface
func (l *Logger) Fatal(i ...any) {
	l.log.Fatal().Msg(fmt.Sprint(i...))
}

// Fatalf implements echo.Logger interface
func (l *Logger) Fatalf(format string, i ...any) {
	l.log.Fatal().Msgf(format, i...)
}

// Fatalj implements echo.Logger interface
func (l *Logger) Fatalj(j log.JSON) {
	l.logJSON(l.log.Fatal(), j)
}

// Panic implements echo.Logger interface
func (l *Logger) Panic(i ...any) {
	l.log.Panic().Msg(fmt.Sprint(i...))
}

// Panicf implements echo.Logger interface
func (l *Logger) Panicf(format string, i ...any) {
	l.log.Panic().Msgf(format, i...)
}

// Panicj implements echo.Logger interface
func (l *Logger) Panicj(j log.JSON) {
	l.logJSON(l.log.Panic(), j)
}

// Print implements echo.Logger interface
func (l *Logger) Print(i ...any) {
	l.log.WithLevel(zerolog.NoLevel).Str("level", "-").Msg(fmt.Sprint(i...))
}

// Printf implements echo.Logger interface
func (l *Logger) Printf(format string, i ...any) {
	l.log.WithLevel(zerolog.NoLevel).Str("level", "-").Msgf(format, i...)
}

// Printj implements echo.Logger interface
func (l *Logger) Printj(j log.JSON) {
	l.logJSON(l.log.WithLevel(zerolog.NoLevel).Str("level", "-"), j)
}

// Output implements echo.Logger interface
func (l *Logger) Output() io.Writer {
	return l.out
}

// SetOutput implements echo.Logger interface
func (l *Logger) SetOutput(newOut io.Writer) {
	l.out = newOut
	l.log = l.log.Output(newOut)
}

// Level implements echo.Logger interface
func (l *Logger) Level() log.Lvl {
	return l.level
}

// SetLevel implements echo.Logger interface
func (l *Logger) SetLevel(level log.Lvl) {
	zlvl, elvl := MatchEchoLevel(level)

	l.setters = append(l.setters, WithLevel(elvl))
	l.level = elvl
	l.log = l.log.Level(zlvl)
}

// Prefix implements echo.Logger interface
func (l *Logger) Prefix() string {
	return l.prefix
}

// SetHeader implements echo.Logger interface
func (l *Logger) SetHeader(_ string) {
	// not implemented
}

// SetPrefix implements echo.Logger interface
func (l *Logger) SetPrefix(newPrefix string) {
	l.setters = append(l.setters, WithPrefix(newPrefix))

	opts := newOptions(l.log, l.setters)

	l.prefix = newPrefix
	l.log = opts.zcontext.Logger()
}

// Unwrap returns the underlying zerolog.Logger
func (l *Logger) Unwrap() zerolog.Logger {
	return l.log
}

// logJSON logs a JSON object
func (l *Logger) logJSON(event *zerolog.Event, j log.JSON) {
	for k, v := range j {
		event = event.Interface(k, v)
	}

	event.Msg("")
}
