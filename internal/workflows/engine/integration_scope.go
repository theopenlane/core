package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	celtypes "github.com/google/cel-go/common/types"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/celx"
)

const (
	integrationScopeCELParserRecursionLimit = 250
	integrationScopeCELExpressionSizeLimit  = 100_000
	integrationScopeCELInterruptFrequency   = uint(100)
	integrationScopeCELTimeout              = 100 * time.Millisecond
	integrationScopeEmptyExpressionResult   = true
)

// IntegrationScopeEvaluator evaluates CEL scope conditions for integration actions
type IntegrationScopeEvaluator struct {
	evaluator             *celx.Evaluator
	emptyExpressionResult bool
}

// NewIntegrationScopeEvaluator builds a CEL evaluator scoped to integration scope variables
func NewIntegrationScopeEvaluator() (*IntegrationScopeEvaluator, error) {
	env, err := buildIntegrationScopeEnv()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCELCompilationFailed, err)
	}

	evaluator := celx.NewEvaluator(env, celx.EvalConfig{
		Timeout:                 integrationScopeCELTimeout,
		InterruptCheckFrequency: integrationScopeCELInterruptFrequency,
		EvalOptimize:            true,
	})

	return &IntegrationScopeEvaluator{
		evaluator:             evaluator,
		emptyExpressionResult: integrationScopeEmptyExpressionResult,
	}, nil
}

// Validate compiles an expression to confirm it is syntactically and type-correct.
// Returns ErrScopeExpressionRequired for empty expressions, ErrCELCompilationFailed
// for syntax errors, and ErrCELProgramCreationFailed for program construction failures.
func (e *IntegrationScopeEvaluator) Validate(expr string) error {
	if expr == "" {
		return ErrScopeExpressionRequired
	}

	ast, issues := e.evaluator.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("%w: %w", ErrCELCompilationFailed, issues.Err())
	}

	if _, err := e.evaluator.Program(ast); err != nil {
		return fmt.Errorf("%w: %w", ErrCELProgramCreationFailed, err)
	}

	return nil
}

// EvaluateConditionWithVars evaluates a CEL expression against the provided scope variables
func (e *IntegrationScopeEvaluator) EvaluateConditionWithVars(ctx context.Context, expr string, vars types.ScopeVars) (bool, error) {
	if expr == "" {
		return e.emptyExpressionResult, nil
	}

	// Validate compilation first to return ErrCELCompilationFailed distinctly from runtime errors
	if err := e.Validate(expr); err != nil {
		return false, err
	}

	out, _, err := e.evaluator.Evaluate(ctx, expr, vars.CELVars())
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrConditionFailed, err)
	}

	if out == nil {
		return false, ErrCELNilOutput
	}

	if out.Type() != celtypes.BoolType {
		return false, ErrCELTypeMismatch
	}

	result, ok := out.Value().(bool)
	if !ok {
		result = out.Equal(celtypes.True) == celtypes.True
	}

	return result, nil
}

// buildIntegrationScopeEnv creates a CEL environment that declares all integration scope variables
func buildIntegrationScopeEnv() (*cel.Env, error) {
	cfg := celx.EnvConfig{
		ParserRecursionLimit:      integrationScopeCELParserRecursionLimit,
		ParserExpressionSizeLimit: integrationScopeCELExpressionSizeLimit,
		ExtendedValidations:       true,
	}

	return celx.NewEnv(cfg,
		cel.VariableDecls(
			decls.NewVariable(types.ScopeVariablePayload, celtypes.DynType),
			decls.NewVariable(types.ScopeVariableResource, celtypes.StringType),
			decls.NewVariable(types.ScopeVariableDefinition, celtypes.StringType),
			decls.NewVariable(types.ScopeVariableOperation, celtypes.StringType),
			decls.NewVariable(types.ScopeVariableConfig, celtypes.DynType),
			decls.NewVariable(types.ScopeVariableInstallationConfig, celtypes.DynType),
			decls.NewVariable(types.ScopeVariableOrgID, celtypes.StringType),
			decls.NewVariable(types.ScopeVariableInstallationID, celtypes.StringType),
		),
	)
}
