package celx

import "time"

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
