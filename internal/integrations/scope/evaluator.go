package scope

import (
	"context"
	"errors"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/theopenlane/core/pkg/celx"
)

// ConditionEvaluator defines the behavior required for integration scope condition evaluation
type ConditionEvaluator interface {
	// Validate validates expression compilation and program creation
	Validate(expression string) error
	// EvaluateCondition evaluates a bool condition expression against variables
	EvaluateCondition(ctx context.Context, expression string, vars map[string]any) (bool, error)
	// EvaluateConditionWithVars evaluates a bool condition expression using typed scope vars
	EvaluateConditionWithVars(ctx context.Context, expression string, vars ScopeVars) (bool, error)
}

// Evaluator evaluates integration scope conditions with a dedicated CEL environment
type Evaluator struct {
	evaluator             *celx.Evaluator
	emptyExpressionResult bool
}

// NewEvaluator creates an integration scope evaluator with provided config
func NewEvaluator(config EvaluatorConfig) (*Evaluator, error) {
	normalized := normalizeEvaluatorConfig(config)

	env, err := newScopeEnv(normalized)
	if err != nil {
		return nil, err
	}

	evaluator := celx.NewEvaluator(env, celx.EvalConfig{
		Timeout:                 normalized.Timeout,
		InterruptCheckFrequency: normalized.InterruptCheckFrequency,
		EvalOptimize:            true,
	})

	return &Evaluator{
		evaluator:             evaluator,
		emptyExpressionResult: normalized.EmptyExpressionResult,
	}, nil
}

// Validate validates expression compilation and program creation
func (e *Evaluator) Validate(expression string) error {
	if expression == "" {
		return ErrScopeExpressionRequired
	}

	ast, issues := e.evaluator.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return ErrScopeCompilationFailed
	}

	if _, err := e.evaluator.Program(ast); err != nil {
		return ErrScopeProgramCreationFailed
	}

	return nil
}

// EvaluateCondition evaluates a bool condition expression against variables.
// The underlying celx.Evaluator caches compiled programs, so repeated calls
// for the same expression do not recompile. A compilation check is performed
// first to return ErrScopeCompilationFailed for syntax or type errors distinct
// from runtime evaluation failures.
func (e *Evaluator) EvaluateCondition(ctx context.Context, expression string, vars map[string]any) (bool, error) {
	if expression == "" {
		return e.emptyExpressionResult, nil
	}

	if _, issues := e.evaluator.Compile(expression); issues != nil && issues.Err() != nil {
		return false, ErrScopeCompilationFailed
	}

	value, err := e.evaluateRaw(ctx, expression, vars)
	if err != nil {
		return false, err
	}

	return boolResultFromValue(value)
}

// EvaluateConditionWithVars evaluates a bool condition expression using typed scope vars
func (e *Evaluator) EvaluateConditionWithVars(ctx context.Context, expression string, vars ScopeVars) (bool, error) {
	return e.EvaluateCondition(ctx, expression, vars.Map())
}

// normalizeEvaluatorConfig applies defaults for zero-valued configuration fields
func normalizeEvaluatorConfig(config EvaluatorConfig) EvaluatorConfig {
	defaults := DefaultEvaluatorConfig()

	if config.Timeout <= 0 {
		config.Timeout = defaults.Timeout
	}
	if config.InterruptCheckFrequency == 0 {
		config.InterruptCheckFrequency = defaults.InterruptCheckFrequency
	}
	if config.ParserRecursionLimit == 0 {
		config.ParserRecursionLimit = defaults.ParserRecursionLimit
	}
	if config.ParserExpressionSizeLimit == 0 {
		config.ParserExpressionSizeLimit = defaults.ParserExpressionSizeLimit
	}

	return config
}

// newScopeEnv creates the CEL environment for integration scope evaluation
func newScopeEnv(config EvaluatorConfig) (*cel.Env, error) {
	return celx.NewEnv(
		celx.EnvConfig{
			ParserRecursionLimit:      config.ParserRecursionLimit,
			ParserExpressionSizeLimit: config.ParserExpressionSizeLimit,
			ExtendedValidations:       true,
		},
		cel.VariableDecls(
			decls.NewVariable(VariablePayload, celtypes.DynType),
			decls.NewVariable(VariableResource, celtypes.StringType),
			decls.NewVariable(VariableProvider, celtypes.StringType),
			decls.NewVariable(VariableOperation, celtypes.StringType),
			decls.NewVariable(VariableConfig, celtypes.DynType),
			decls.NewVariable(VariableIntegrationConfig, celtypes.DynType),
			decls.NewVariable(VariableProviderState, celtypes.DynType),
			decls.NewVariable(VariableOrgID, celtypes.StringType),
			decls.NewVariable(VariableIntegrationID, celtypes.StringType),
		),
	)
}

// evaluateRaw evaluates an expression and returns the raw CEL value
func (e *Evaluator) evaluateRaw(ctx context.Context, expression string, vars map[string]any) (ref.Val, error) {
	value, _, err := e.evaluator.Evaluate(ctx, expression, vars)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrScopeEvaluationTimeout
		}

		return nil, ErrScopeEvaluationFailed
	}
	if value == nil {
		return nil, ErrScopeEvaluationOutputNil
	}

	return value, nil
}

// boolResultFromValue extracts a bool result from a CEL value
func boolResultFromValue(value ref.Val) (bool, error) {
	if value.Type() != celtypes.BoolType {
		return false, ErrScopeConditionType
	}

	boolValue, ok := value.Value().(bool)
	if ok {
		return boolValue, nil
	}

	return value.Equal(celtypes.True) == celtypes.True, nil
}
