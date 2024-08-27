package otelx

import (
	"context"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

const (
	StdOutProvider   = "stdout"
	OTLPHTTPProvider = "otlphttp"
	OTLPGRPCProvider = "otlpgrpc"
)

// Config defines the configuration settings for opentelemetry tracing
type Config struct {
	// Enabled to enable tracing
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// Provider to use for tracing
	Provider string `json:"provider" koanf:"provider" default:"stdout"`
	// Environment to set for the service
	Environment string `json:"environment" koanf:"environment" default:"development"`
	// StdOut settings for the stdout provider
	StdOut StdOut `json:"stdout" koanf:"stdout"`
	// OTLP settings for the otlp provider
	OTLP OTLP `json:"otlp" koanf:"otlp"`
}

// StdOut settings for the stdout provider
type StdOut struct {
	// Pretty enables pretty printing of the output
	Pretty bool `json:"pretty" koanf:"pretty" default:"true"`
	// DisableTimestamp disables the timestamp in the output
	DisableTimestamp bool `json:"disableTimestamp" koanf:"disableTimestamp" default:"false"`
}

// OTLP settings for the otlp provider
type OTLP struct {
	// Endpoint to send the traces to
	Endpoint string `json:"endpoint" koanf:"endpoint" default:"localhost:4317"`
	// Insecure to disable TLS
	Insecure bool `json:"insecure" koanf:"insecure" default:"true"`
	// Certificate to use for TLS
	Certificate string `json:"certificate" koanf:"certificate"`
	// Headers to send with the request
	Headers []string `json:"headers" koanf:"headers"`
	// Compression to use for the request
	Compression string `json:"compression" koanf:"compression"`
	// Timeout for the request
	Timeout time.Duration `json:"timeout" koanf:"timeout" default:"10s"`
}

func NewTracer(c Config, name string, logger *zap.SugaredLogger) error {
	if !c.Enabled {
		logger.Debug("Tracing disabled")
		return nil
	}

	exp, err := newTraceExporter(c)
	if err != nil {
		logger.Debugw("Failed to create trace exporter", "error", err)
		return err
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		// Record information about this application in a resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			attribute.String("environment", c.Environment),
		)),
	}

	if exp != nil {
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

func newTraceExporter(c Config) (sdktrace.SpanExporter, error) {
	switch c.Provider {
	case StdOutProvider:
		return newStdoutProvider(c)
	case OTLPHTTPProvider:
		return newOTLPHTTPProvider(c)
	case OTLPGRPCProvider:
		return newOTLPGRPCProvider(c)
	default:
		return nil, newUnknownProviderError(c.Provider)
	}
}

func newStdoutProvider(c Config) (sdktrace.SpanExporter, error) {
	opts := []stdouttrace.Option{}

	if c.StdOut.Pretty {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	if c.StdOut.DisableTimestamp {
		opts = append(opts, stdouttrace.WithoutTimestamps())
	}

	return stdouttrace.New(opts...)
}

func newOTLPHTTPProvider(c Config) (sdktrace.SpanExporter, error) {
	_, err := url.Parse(c.OTLP.Endpoint)
	if err != nil {
		return nil, newTraceConfigError(err)
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(c.OTLP.Endpoint),
		otlptracehttp.WithTimeout(c.OTLP.Timeout),
	}

	if c.OTLP.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}

func newOTLPGRPCProvider(c Config) (sdktrace.SpanExporter, error) {
	_, err := url.Parse(c.OTLP.Endpoint)
	if err != nil {
		return nil, newTraceConfigError(err)
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(c.OTLP.Endpoint),
		otlptracegrpc.WithTimeout(c.OTLP.Timeout),
	}

	if c.OTLP.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))
}
