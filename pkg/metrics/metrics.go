package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Metrics struct holds a distinct echo instance to report system metrics
type Metrics struct {
	e    *echo.Echo
	port string
	reg  *sync.Once
}

// New creates a new Metrics instance
func New(port string) *Metrics {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	customCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "custom_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method",
		},
	)

	m := &Metrics{
		e:    e,
		port: port,
		reg:  &sync.Once{},
	}

	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		DoNotUseRequestPathFor404: true,
		AfterNext: func(_ echo.Context, err error) {
			customCounter.Inc() // use our custom metric in middleware after every request increment the counter
		},
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/metrics" || c.Path() == "/health"
		},
		Subsystem: "openlane",
		HistogramOptsFunc: func(opts prometheus.HistogramOpts) prometheus.HistogramOpts {
			if strings.HasSuffix(opts.Name, "_duration_seconds") {
				opts.Buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}
			} else if strings.HasSuffix(opts.Name, "_size_bytes") {
				opts.Buckets = []float64{
					512.0,       // 512 B
					1024.0,      // 1,024 B
					2048.0,      // 2.0 KiB
					5120.0,      // 5.0 KiB
					10240.0,     // 10.0 KiB
					25600.0,     // 25.0 KiB
					51200.0,     // 50.0 KiB
					102400.0,    // 100.0 KiB
					256000.0,    // 250.0 KiB
					512000.0,    // 500.0 KiB
					1048576.0,   // 1024.0 KiB
					2097152.0,   // 2.0 MiB
					5242880.0,   // 5.0 MiB
					10485760.0,  // 10.0 MiB
					26214400.0,  // 25.0 MiB
					52428800.0,  // 50.0 MiB
					104857600.0, // 100.0 MiB
				}
			}

			return opts
		},
		LabelFuncs: map[string]echoprometheus.LabelValueFunc{
			"url": func(c echo.Context, _ error) string {
				return c.Path()
			},
		},
	}))

	m.e.GET("/metrics", echoprometheus.NewHandler())

	return m
}

// Start starts the metrics server
func (m *Metrics) Start(ctx context.Context) error {
	zerolog.Ctx(ctx).Info().Msg("starting metrics server")

	if err := m.e.Start(fmt.Sprintf(":%d", m.port)); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

// Stop stops the metrics server
func (m *Metrics) Stop(ctx context.Context) error {
	zerolog.Ctx(ctx).Info().Msg("stopping metrics server")

	return m.e.Shutdown(ctx)
}

// Register registers metrics to the default registry
func (m *Metrics) Register(metrics []prometheus.Collector) error {
	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			return err
		}
	}

	return nil
}
