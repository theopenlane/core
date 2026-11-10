package workflows

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestConfigOptions(t *testing.T) {
	cfg := NewDefaultConfig(
		WithEnabled(true),
		WithCELTimeout(250*time.Millisecond),
		WithCELCostLimit(99),
		WithCELInterruptCheckFrequency(7),
		WithCELParserRecursionLimit(5),
		WithCELParserExpressionSizeLimit(123),
		WithCELComprehensionNestingLimit(3),
		WithCELExtendedValidations(false),
		WithCELOptionalTypes(true),
		WithCELIdentifierEscapeSyntax(true),
		WithCELCrossTypeNumericComparisons(true),
		WithCELMacroCallTracking(true),
		WithCELEvalOptimize(false),
		WithCELTrackState(true),
	)

	assert.Check(t, cfg.Enabled)
	assert.Check(t, is.Equal(250*time.Millisecond, cfg.CEL.Timeout))
	assert.Check(t, is.Equal(uint64(99), cfg.CEL.CostLimit))
	assert.Check(t, is.Equal(uint(7), cfg.CEL.InterruptCheckFrequency))
	assert.Check(t, is.Equal(5, cfg.CEL.ParserRecursionLimit))
	assert.Check(t, is.Equal(123, cfg.CEL.ParserExpressionSizeLimit))
	assert.Check(t, is.Equal(3, cfg.CEL.ComprehensionNestingLimit))
	assert.Check(t, !cfg.CEL.ExtendedValidations)
	assert.Check(t, cfg.CEL.OptionalTypes)
	assert.Check(t, cfg.CEL.IdentifierEscapeSyntax)
	assert.Check(t, cfg.CEL.CrossTypeNumericComparisons)
	assert.Check(t, cfg.CEL.MacroCallTracking)
	assert.Check(t, !cfg.CEL.EvalOptimize)
	assert.Check(t, cfg.CEL.TrackState)
	assert.Check(t, is.Equal("events", cfg.Gala.QueueName))

	override := Config{
		Enabled: false,
		CEL: CELConfig{
			Timeout:   time.Second,
			CostLimit: 12,
		},
		Gala: GalaConfig{
			Enabled:            true,
			WorkerCount:        11,
			MaxRetries:         13,
			FailOnEnqueueError: true,
			QueueName:          "events",
		},
		WebhookAllowPrivateAddresses: true,
	}
	cfg = NewDefaultConfig(WithConfig(override))
	assert.Check(t, is.Equal(override.Enabled, cfg.Enabled))
	assert.Check(t, is.Equal(override.CEL.Timeout, cfg.CEL.Timeout))
	assert.Check(t, is.Equal(override.CEL.CostLimit, cfg.CEL.CostLimit))
	assert.Check(t, is.Equal(override.Gala.Enabled, cfg.Gala.Enabled))
	assert.Check(t, is.Equal(override.Gala.WorkerCount, cfg.Gala.WorkerCount))
	assert.Check(t, is.Equal(override.Gala.MaxRetries, cfg.Gala.MaxRetries))
	assert.Check(t, is.Equal(override.Gala.FailOnEnqueueError, cfg.Gala.FailOnEnqueueError))
	assert.Check(t, is.Equal(override.Gala.QueueName, cfg.Gala.QueueName))
	assert.Check(t, is.Equal(override.WebhookAllowPrivateAddresses, cfg.WebhookAllowPrivateAddresses))
}

func TestConfigIsEnabledNil(t *testing.T) {
	var cfg *Config
	assert.Check(t, !cfg.IsEnabled())
}
