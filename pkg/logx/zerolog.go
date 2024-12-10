package logx

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var log zerolog.Logger

func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.InfoLevel) // default to INFO
		}

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			FieldsExclude: []string{
				"user_agent",
				"git_revision",
				"go_version",
			},
		}

		// If the app is not running in development mode, write logs to a file and use Stderr as the output
		// using console writer out in production dramatically slows log output
		if os.Getenv("APP_ENV") != "development" {
			fileLogger := &lumberjack.Logger{
				Filename:   "openlane.log",
				MaxSize:    5,  // nolint:mnd
				MaxBackups: 10, // nolint:mnd
				MaxAge:     14, // nolint:mnd
				Compress:   true,
			}

			output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		}

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Str("git_revision", gitRevision).
			Str("go_version", buildInfo.GoVersion).
			Logger()

		zerolog.DefaultContextLogger = &log
	})

	return log
}

type correlation struct {
	correlationID string
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l := Get()

		corr := correlation{ulids.New().String()}

		r = r.WithContext(contextx.With(r.Context(), corr))

		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("correlation_id", corr.correlationID)
		})

		w.Header().Add("X-Correlation-ID", corr.correlationID)

		lrw := newLoggingResponseWriter(w)

		// add to the request context so that it is accessible in all handler functions
		r = r.WithContext(l.WithContext(r.Context()))

		defer func() {
			panicVal := recover()
			if panicVal != nil {
				lrw.statusCode = http.StatusInternalServerError // ensure that the status code is updated

				panic(panicVal)
			}

			l.
				Info().
				Str("method", r.Method).
				Str("url", r.URL.RequestURI()).
				Str("user_agent", r.UserAgent()).
				Int("status_code", lrw.statusCode).
				Dur("elapsed_ms", time.Since(start)).
				Msg("incoming request")
		}()

		next.ServeHTTP(lrw, r)
	})
}

// The loggingResponseWriter struct embeds the http.ResponseWriter type and adds a statusCode property which defaults to http.StatusOK (200) when newLoggingResponseWriter() is called. A WriteHeader() method is defined on the loggingResponseWriter which updates the value of the statusCode property and calls the WriteHeader() method of the embedded http.ResponseWriter instance. This allows us to capture the status code of the response written by the handler function
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func consoleWriter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}
}
