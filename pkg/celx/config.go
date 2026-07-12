package celx

import "time"

const (
	// DefaultParserRecursionLimit caps CEL parser recursion depth for hardened environments
	DefaultParserRecursionLimit = 250
	// DefaultParserExpressionSizeLimit caps CEL expression size (code points) for hardened environments
	DefaultParserExpressionSizeLimit = 100_000
	// DefaultInterruptCheckFrequency controls how often CEL checks for interrupts during comprehensions
	DefaultInterruptCheckFrequency uint = 100
	// DefaultEvalTimeout bounds the wall-clock time allowed for a single CEL evaluation
	DefaultEvalTimeout = 100 * time.Millisecond
)

// StrictEnvConfig returns a hardened environment config with parser limits and extended validations
// enabled. Callers tweak the returned value for environment-specific options such as cross-type
// numeric comparisons
func StrictEnvConfig() EnvConfig {
	return EnvConfig{
		ParserRecursionLimit:      DefaultParserRecursionLimit,
		ParserExpressionSizeLimit: DefaultParserExpressionSizeLimit,
		ExtendedValidations:       true,
	}
}

// FastEvalConfig returns an evaluation config tuned for short, cached, repeatedly-run expressions
func FastEvalConfig() EvalConfig {
	return EvalConfig{
		Timeout:                 DefaultEvalTimeout,
		InterruptCheckFrequency: DefaultInterruptCheckFrequency,
		EvalOptimize:            true,
	}
}

// EnvConfig configures CEL parsing, validation, and environment features

type EnvConfig struct {
	// ParserRecursionLimit caps the parser recursion depth, 0 uses CEL defaults
	ParserRecursionLimit int
	// ParserExpressionSizeLimit caps expression size (code points), 0 uses CEL defaults
	ParserExpressionSizeLimit int
	// ComprehensionNestingLimit caps nested comprehensions, 0 disables the check
	ComprehensionNestingLimit int
	// ExtendedValidations enables extra AST validations (regex, duration, timestamps, homogeneous aggregates)
	ExtendedValidations bool
	// OptionalTypes enables CEL optional types and optional field syntax
	OptionalTypes bool
	// IdentifierEscapeSyntax enables backtick escaped identifiers
	IdentifierEscapeSyntax bool
	// CrossTypeNumericComparisons enables comparisons across numeric types
	CrossTypeNumericComparisons bool
	// MacroCallTracking records macro calls in AST source info for debugging
	MacroCallTracking bool
}

// EvalConfig configures CEL evaluation behavior
type EvalConfig struct {
	// Timeout is the maximum duration allowed for evaluating a CEL expression
	Timeout time.Duration
	// CostLimit caps the runtime cost of CEL evaluation, 0 disables the limit
	CostLimit uint64
	// InterruptCheckFrequency controls how often CEL checks for interrupts during comprehensions
	InterruptCheckFrequency uint
	// EvalOptimize enables evaluation-time optimizations for repeated program runs
	EvalOptimize bool
	// TrackState enables evaluation state tracking for debugging
	TrackState bool
}
