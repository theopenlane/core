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

	override := Config{
		Enabled: false,
		CEL: CELConfig{
			Timeout:   time.Second,
			CostLimit: 12,
		},
		MutationOutbox: MutationOutboxConfig{
			Enabled:            true,
			WorkerCount:        7,
			MaxRetries:         9,
			FailOnEnqueueError: true,
			Topics:             []string{"Organization", "WorkflowAssignment"},
		},
		Gala: GalaConfig{
			Enabled:            true,
			DualEmit:           true,
			WorkerCount:        11,
			MaxRetries:         13,
			FailOnEnqueueError: true,
			Topics:             []string{"Organization", "OrganizationSetting"},
			TopicModes: map[string]GalaTopicMode{
				"Organization":        GalaTopicModeDualEmit,
				"OrganizationSetting": GalaTopicModeV2Only,
			},
			QueueName: "default",
		},
	}
	cfg = NewDefaultConfig(WithConfig(override))
	assert.Equal(t, override.Enabled, cfg.Enabled)
	assert.Equal(t, override.CEL.Timeout, cfg.CEL.Timeout)
	assert.Equal(t, override.CEL.CostLimit, cfg.CEL.CostLimit)
	assert.Equal(t, override.MutationOutbox.Enabled, cfg.MutationOutbox.Enabled)
	assert.Equal(t, override.MutationOutbox.WorkerCount, cfg.MutationOutbox.WorkerCount)
	assert.Equal(t, override.MutationOutbox.MaxRetries, cfg.MutationOutbox.MaxRetries)
	assert.Equal(t, override.MutationOutbox.FailOnEnqueueError, cfg.MutationOutbox.FailOnEnqueueError)
	assert.Equal(t, override.MutationOutbox.Topics, cfg.MutationOutbox.Topics)
	assert.Equal(t, override.Gala.Enabled, cfg.Gala.Enabled)
	assert.Equal(t, override.Gala.DualEmit, cfg.Gala.DualEmit)
	assert.Equal(t, override.Gala.WorkerCount, cfg.Gala.WorkerCount)
	assert.Equal(t, override.Gala.MaxRetries, cfg.Gala.MaxRetries)
	assert.Equal(t, override.Gala.FailOnEnqueueError, cfg.Gala.FailOnEnqueueError)
	assert.Equal(t, override.Gala.Topics, cfg.Gala.Topics)
	assert.Equal(t, override.Gala.TopicModes, cfg.Gala.TopicModes)
	assert.Equal(t, override.Gala.QueueName, cfg.Gala.QueueName)
}

func TestConfigIsEnabledNil(t *testing.T) {
	var cfg *Config
	assert.False(t, cfg.IsEnabled())
}
