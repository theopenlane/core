package engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"

	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/celx"
)

// CELEvaluator handles CEL expression compilation and evaluation with caching
type CELEvaluator struct {
	// evaluator handles CEL compilation and evaluation with caching
	evaluator *celx.Evaluator
}

// NewCELEvaluator creates a new CEL evaluator with the provided environment and configuration
func NewCELEvaluator(env *cel.Env, config *workflows.Config) *CELEvaluator {
	evalCfg := celx.EvalConfig{
		Timeout:                 config.CEL.Timeout,
		CostLimit:               config.CEL.CostLimit,
		InterruptCheckFrequency: config.CEL.InterruptCheckFrequency,
		EvalOptimize:            config.CEL.EvalOptimize,
		TrackState:              config.CEL.TrackState,
	}

	return &CELEvaluator{
		evaluator: celx.NewEvaluator(env, evalCfg),
	}
}

// Validate validates that a CEL expression compiles successfully
func (e *CELEvaluator) Validate(expression string) error {
	ast, issues := e.evaluator.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("%w: %w", ErrCELCompilationFailed, issues.Err())
	}

	if _, err := e.evaluator.Program(ast); err != nil {
		return fmt.Errorf("%w: %w", ErrCELProgramCreationFailed, err)
	}

	return nil
}

// Evaluate evaluates a CEL expression with the given variables, using caching and timeout
func (e *CELEvaluator) Evaluate(ctx context.Context, expression string, vars map[string]any) (bool, error) {
	out, _, evalErr := e.evaluator.Evaluate(ctx, expression, vars)
	if evalErr != nil {
		if errors.Is(evalErr, context.DeadlineExceeded) || errors.Is(evalErr, context.Canceled) {
			return false, ErrEvaluationTimeout
		}

		return false, fmt.Errorf("%w: %v", ErrConditionFailed, evalErr)
	}

	result, boolErr := celx.BoolResult(out)
	if boolErr != nil {
		switch {
		case errors.Is(boolErr, celx.ErrNilOutput):
			return false, ErrCELNilOutput
		default:
			return false, ErrCELTypeMismatch
		}
	}

	return result, nil
}

// EvaluateJSONMap evaluates a CEL expression and converts the result to a JSON object map.
func (e *CELEvaluator) EvaluateJSONMap(ctx context.Context, expression string, vars map[string]any) (map[string]any, error) {
	out, evalErr := e.evaluator.EvaluateJSONMap(ctx, expression, vars)
	if evalErr != nil {
		if errors.Is(evalErr, context.DeadlineExceeded) || errors.Is(evalErr, context.Canceled) {
			return nil, ErrEvaluationTimeout
		}

		return nil, evalErr
	}

	return out, nil
}

// EvaluateValue evaluates a CEL expression and returns the JSON-compatible value.
func (e *CELEvaluator) EvaluateValue(ctx context.Context, expression string, vars map[string]any) (any, error) {
	out, _, evalErr := e.evaluator.Evaluate(ctx, expression, vars)
	if evalErr != nil {
		if errors.Is(evalErr, context.DeadlineExceeded) || errors.Is(evalErr, context.Canceled) {
			return nil, ErrEvaluationTimeout
		}

		return nil, evalErr
	}

	return celx.ToJSON(out)
}
