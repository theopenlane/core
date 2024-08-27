package otelx_test

import (
	"testing"

	"github.com/theopenlane/core/pkg/otelx"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewTracer(t *testing.T) {
	testCases := []struct {
		name          string
		config        otelx.Config
		expectedError error
	}{
		{
			name: "enabled",
			config: otelx.Config{
				Enabled:     true,
				Provider:    otelx.OTLPHTTPProvider,
				Environment: "development",
				OTLP: otelx.OTLP{
					Endpoint: "localhost:4317",
					Insecure: true,
				},
			},
		},
		{
			name: "disabled",
			config: otelx.Config{
				Enabled:     false,
				Provider:    otelx.OTLPHTTPProvider,
				Environment: "development",
				OTLP: otelx.OTLP{
					Endpoint: "localhost:4317",
					Insecure: true,
				},
			},
		},
		{
			name: "otlphttp - invalid endpoint",
			config: otelx.Config{
				Enabled:     true,
				Provider:    otelx.OTLPHTTPProvider,
				Environment: "development",
				OTLP: otelx.OTLP{
					Endpoint: "    a-b-://1234",
					Insecure: true,
				},
			},
			expectedError: otelx.ErrInvalidConfig,
		},
		{
			name: "otlpgrpc",
			config: otelx.Config{
				Enabled:     true,
				Provider:    otelx.OTLPGRPCProvider,
				Environment: "development",
				OTLP: otelx.OTLP{
					Endpoint: "localhost:4317",
					Insecure: true,
				},
			},
		},
		{
			name: "otlpgrpc - invalid endpoint",
			config: otelx.Config{
				Enabled:     true,
				Provider:    otelx.OTLPGRPCProvider,
				Environment: "development",
				OTLP: otelx.OTLP{
					Endpoint: "    a-b-://1234",
					Insecure: true,
				},
			},
			expectedError: otelx.ErrInvalidConfig,
		},
		{
			name: "stdout",
			config: otelx.Config{
				Enabled:     true,
				Provider:    otelx.StdOutProvider,
				Environment: "development",
				StdOut: otelx.StdOut{
					Pretty:           true,
					DisableTimestamp: true,
				},
			},
		},
		{
			name: "invalid provider",
			config: otelx.Config{
				Enabled:     true,
				Provider:    "invalid",
				Environment: "development",
			},
			expectedError: otelx.ErrUnknownProvider,
		},
	}

	for _, tc := range testCases {
		t.Run("Trace "+tc.name, func(t *testing.T) {
			err := otelx.NewTracer(tc.config, "test", zap.NewNop().Sugar())

			if tc.expectedError != nil {
				assert.NotNil(t, err)
				assert.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
