package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 250*time.Millisecond, cfg.CEL.Timeout)
	assert.Equal(t, uint64(99), cfg.CEL.CostLimit)
	assert.Equal(t, uint(7), cfg.CEL.InterruptCheckFrequency)
	assert.Equal(t, 5, cfg.CEL.ParserRecursionLimit)
	assert.Equal(t, 123, cfg.CEL.ParserExpressionSizeLimit)
	assert.Equal(t, 3, cfg.CEL.ComprehensionNestingLimit)
	assert.False(t, cfg.CEL.ExtendedValidations)
	assert.True(t, cfg.CEL.OptionalTypes)
	assert.True(t, cfg.CEL.IdentifierEscapeSyntax)
	assert.True(t, cfg.CEL.CrossTypeNumericComparisons)
	assert.True(t, cfg.CEL.MacroCallTracking)
	assert.False(t, cfg.CEL.EvalOptimize)
	assert.True(t, cfg.CEL.TrackState)
	assert.Equal(t, "events", cfg.EventingQueueName)

	override := Config{
		Enabled: false,
		CEL: CELConfig{
			Timeout:   time.Second,
			CostLimit: 12,
		},
		EventingEnabled:     true,
		EventingWorkerCount: 11,
		EventingMaxRetries:  13,
		EventingQueueName:   "events",
	}
	cfg = NewDefaultConfig(WithConfig(override))
	assert.Equal(t, override.Enabled, cfg.Enabled)
	assert.Equal(t, override.CEL.Timeout, cfg.CEL.Timeout)
	assert.Equal(t, override.CEL.CostLimit, cfg.CEL.CostLimit)
	assert.Equal(t, override.EventingEnabled, cfg.EventingEnabled)
	assert.Equal(t, override.EventingWorkerCount, cfg.EventingWorkerCount)
	assert.Equal(t, override.EventingMaxRetries, cfg.EventingMaxRetries)
	assert.Equal(t, override.EventingQueueName, cfg.EventingQueueName)
}

func TestConfigIsEnabledNil(t *testing.T) {
	var cfg *Config
	assert.False(t, cfg.IsEnabled())
}
