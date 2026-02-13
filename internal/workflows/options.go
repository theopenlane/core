package workflows

import (
	"time"

	"github.com/mcuadros/go-defaults"
	"github.com/samber/lo"
)

// GalaTopicMode controls migration behavior for one mutation topic.
type GalaTopicMode string

const (
	// GalaTopicModeSoireeOnly keeps topic processing on legacy Soiree only.
	GalaTopicModeSoireeOnly GalaTopicMode = "soiree_only"
	// GalaTopicModeDualEmit emits to both legacy Soiree and Gala.
	GalaTopicModeDualEmit GalaTopicMode = "dual_emit"
	// GalaTopicModeV2Only prefers Gala emission only (with runtime-level fallback behavior).
	GalaTopicModeV2Only GalaTopicMode = "v2_only"
)

// IsValid reports whether a topic mode is supported.
func (m GalaTopicMode) IsValid() bool {
	return m == GalaTopicModeSoireeOnly || m == GalaTopicModeDualEmit || m == GalaTopicModeV2Only
}

// Config contains the configuration for the workflows engine
type Config struct {
	// Enabled determines if the workflows engine is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// CEL contains configuration for CEL evaluation and validation
	CEL CELConfig `json:"cel" koanf:"cel"`
	// MutationOutbox enables optional River-backed mutation dispatch for Soiree listeners
	MutationOutbox MutationOutboxConfig `json:"mutationoutbox" koanf:"mutationoutbox"`
	// Gala enables optional River-backed durable gala runtime and dual-emit behavior
	Gala GalaConfig `json:"gala" koanf:"gala"`
}

// CELConfig contains CEL evaluation and validation settings for workflows
type CELConfig struct {
	// Timeout is the maximum duration allowed for evaluating a CEL expression
	Timeout time.Duration `json:"timeout" koanf:"timeout" default:"100ms"`
	// CostLimit caps the runtime cost of CEL evaluation, 0 disables the limit
	CostLimit uint64 `json:"costlimit" koanf:"costlimit" default:"0"`
	// InterruptCheckFrequency controls how often CEL checks for interrupts during comprehensions
	InterruptCheckFrequency uint `json:"interruptcheckfrequency" koanf:"interruptcheckfrequency" default:"100"`
	// ParserRecursionLimit caps the parser recursion depth, 0 uses CEL defaults
	ParserRecursionLimit int `json:"parserrecursionlimit" koanf:"parserrecursionlimit" default:"250"`
	// ParserExpressionSizeLimit caps expression size (code points), 0 uses CEL defaults
	ParserExpressionSizeLimit int `json:"parserexpressionsizelimit" koanf:"parserexpressionsizelimit" default:"100000"`
	// ComprehensionNestingLimit caps nested comprehensions, 0 disables the check
	ComprehensionNestingLimit int `json:"comprehensionnestinglimit" koanf:"comprehensionnestinglimit" default:"0"`
	// ExtendedValidations enables extra AST validations (regex, duration, timestamps, homogeneous aggregates)
	ExtendedValidations bool `json:"extendedvalidations" koanf:"extendedvalidations" default:"true"`
	// OptionalTypes enables CEL optional types and optional field syntax
	OptionalTypes bool `json:"optionaltypes" koanf:"optionaltypes" default:"false"`
	// IdentifierEscapeSyntax enables backtick escaped identifiers
	IdentifierEscapeSyntax bool `json:"identifierescapesyntax" koanf:"identifierescapesyntax" default:"false"`
	// CrossTypeNumericComparisons enables comparisons across numeric types
	CrossTypeNumericComparisons bool `json:"crosstypenumericcomparisons" koanf:"crosstypenumericcomparisons" default:"false"`
	// MacroCallTracking records macro calls in AST source info for debugging
	MacroCallTracking bool `json:"macrocalltracking" koanf:"macrocalltracking" default:"false"`
	// EvalOptimize enables evaluation-time optimizations for repeated program runs
	EvalOptimize bool `json:"evaloptimize" koanf:"evaloptimize" default:"true"`
	// TrackState enables evaluation state tracking for debugging
	TrackState bool `json:"trackstate" koanf:"trackstate" default:"false"`
}

// MutationOutboxConfig controls optional River-backed delivery for mutation listeners.
type MutationOutboxConfig struct {
	// Enabled toggles River-backed mutation dispatch
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// WorkerCount configures the default queue worker concurrency when enabled
	WorkerCount int `json:"workercount" koanf:"workercount" default:"10"`
	// MaxRetries sets River job max attempts for mutation dispatch jobs
	MaxRetries int `json:"maxretries" koanf:"maxretries" default:"5"`
	// FailOnEnqueueError enables strict-mode logging when outbox enqueue fails
	FailOnEnqueueError bool `json:"failonenqueueerror" koanf:"failonenqueueerror" default:"false"`
	// Topics optionally scopes outbox dispatch to specific mutation topics; empty means all topics
	Topics []string `json:"topics" koanf:"topics"`
}

// GalaConfig controls optional gala runtime wiring and mutation dual-emit behavior.
type GalaConfig struct {
	// Enabled toggles gala worker and runtime initialization
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// DualEmit toggles mutation dual-emit into gala alongside legacy Soiree inline emission
	DualEmit bool `json:"dualemit" koanf:"dualemit" default:"false"`
	// WorkerCount configures default queue worker concurrency when gala workers are enabled
	WorkerCount int `json:"workercount" koanf:"workercount" default:"10"`
	// MaxRetries sets River job max attempts for gala dispatch jobs
	MaxRetries int `json:"maxretries" koanf:"maxretries" default:"5"`
	// FailOnEnqueueError enables strict-mode logging when gala enqueue fails during dual emit
	FailOnEnqueueError bool `json:"failonenqueueerror" koanf:"failonenqueueerror" default:"false"`
	// Topics optionally scopes gala dual emit to specific mutation topics; empty means all topics
	Topics []string `json:"topics" koanf:"topics"`
	// TopicModes overrides global gala migration behavior by topic (soiree_only, dual_emit, v2_only)
	TopicModes map[string]GalaTopicMode `json:"topicmodes" koanf:"topicmodes"`
	// QueueName optionally overrides queue selection for durable gala dispatch jobs
	QueueName string `json:"queuename" koanf:"queuename" default:"default"`
}

// NewDefaultConfig creates a new workflows config with default values applied.
func NewDefaultConfig(opts ...ConfigOpts) *Config {
	c := &Config{}
	defaults.SetDefaults(c)

	lo.ForEach(opts, func(opt ConfigOpts, _ int) { opt(c) })

	return c
}

// ConfigOpts configures the Config
type ConfigOpts func(*Config)

// WithEnabled sets the enabled field
func WithEnabled(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithCELTimeout sets the CEL evaluation timeout
func WithCELTimeout(timeout time.Duration) ConfigOpts {
	return func(c *Config) {
		c.CEL.Timeout = timeout
	}
}

// WithCELCostLimit sets the CEL cost limit
func WithCELCostLimit(limit uint64) ConfigOpts {
	return func(c *Config) {
		c.CEL.CostLimit = limit
	}
}

// WithCELInterruptCheckFrequency sets the CEL interrupt check frequency
func WithCELInterruptCheckFrequency(freq uint) ConfigOpts {
	return func(c *Config) {
		c.CEL.InterruptCheckFrequency = freq
	}
}

// WithCELParserRecursionLimit sets the CEL parser recursion limit
func WithCELParserRecursionLimit(limit int) ConfigOpts {
	return func(c *Config) {
		c.CEL.ParserRecursionLimit = limit
	}
}

// WithCELParserExpressionSizeLimit sets the CEL parser expression size limit
func WithCELParserExpressionSizeLimit(limit int) ConfigOpts {
	return func(c *Config) {
		c.CEL.ParserExpressionSizeLimit = limit
	}
}

// WithCELComprehensionNestingLimit sets the CEL comprehension nesting limit
func WithCELComprehensionNestingLimit(limit int) ConfigOpts {
	return func(c *Config) {
		c.CEL.ComprehensionNestingLimit = limit
	}
}

// WithCELExtendedValidations toggles CEL extended validations
func WithCELExtendedValidations(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.ExtendedValidations = enabled
	}
}

// WithCELOptionalTypes toggles CEL optional types
func WithCELOptionalTypes(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.OptionalTypes = enabled
	}
}

// WithCELIdentifierEscapeSyntax toggles CEL identifier escape syntax
func WithCELIdentifierEscapeSyntax(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.IdentifierEscapeSyntax = enabled
	}
}

// WithCELCrossTypeNumericComparisons toggles CEL cross-type numeric comparisons
func WithCELCrossTypeNumericComparisons(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.CrossTypeNumericComparisons = enabled
	}
}

// WithCELMacroCallTracking toggles CEL macro call tracking
func WithCELMacroCallTracking(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.MacroCallTracking = enabled
	}
}

// WithCELEvalOptimize toggles CEL evaluation optimizations
func WithCELEvalOptimize(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.EvalOptimize = enabled
	}
}

// WithCELTrackState toggles CEL evaluation state tracking
func WithCELTrackState(enabled bool) ConfigOpts {
	return func(c *Config) {
		c.CEL.TrackState = enabled
	}
}

// WithConfig applies all settings from a Config struct
func WithConfig(cfg Config) ConfigOpts {
	return func(c *Config) {
		c.Enabled = cfg.Enabled
		c.CEL = cfg.CEL
		c.MutationOutbox = cfg.MutationOutbox
		c.Gala = cfg.Gala
	}
}

// IsEnabled checks if the workflows feature is enabled
func (c *Config) IsEnabled() bool {
	if c == nil {
		return false
	}

	return c.Enabled
}
