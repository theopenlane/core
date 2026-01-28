package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"

	"github.com/theopenlane/core/internal/workflows"
)

// CELEvaluator handles CEL expression compilation and evaluation with caching
type CELEvaluator struct {
	// env is the CEL environment for compiling expressions
	env *cel.Env
	// config contains CEL configuration options
	config *workflows.Config
	// programCache stores compiled CEL programs for reuse
	programCache sync.Map
}

// NewCELEvaluator creates a new CEL evaluator with the provided environment and configuration
func NewCELEvaluator(env *cel.Env, config *workflows.Config) *CELEvaluator {
	return &CELEvaluator{
		env:    env,
		config: config,
	}
}

// Validate validates that a CEL expression compiles successfully
func (e *CELEvaluator) Validate(expression string) error {
	ast, issues := e.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("%w: %w", ErrCELCompilationFailed, issues.Err())
	}

	if _, err := e.env.Program(ast); err != nil {
		return fmt.Errorf("%w: %w", ErrCELProgramCreationFailed, err)
	}

	return nil
}

// Evaluate evaluates a CEL expression with the given variables, using caching and timeout
func (e *CELEvaluator) Evaluate(ctx context.Context, expression string, vars map[string]any) (bool, error) {
	prg, err := e.getOrCompileProgram(expression)
	if err != nil {
		return false, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	evalCtx, cancel := context.WithTimeout(ctx, e.config.CEL.Timeout)
	defer cancel()

	out, _, evalErr := prg.ContextEval(evalCtx, vars)
	if evalErr != nil {
		if errors.Is(evalErr, context.DeadlineExceeded) || errors.Is(evalErr, context.Canceled) || evalCtx.Err() != nil {
			return false, ErrEvaluationTimeout
		}

		return false, fmt.Errorf("%w: %v", ErrConditionFailed, evalErr)
	}

	if out == nil {
		return false, ErrCELNilOutput
	}

	if out.Type() != types.BoolType {
		return false, ErrCELTypeMismatch
	}

	result, ok := out.Value().(bool)
	if !ok {
		result = out.Equal(types.True) == types.True
	}

	return result, nil
}

// getOrCompileProgram retrieves a compiled CEL program from cache or compiles it
func (e *CELEvaluator) getOrCompileProgram(expression string) (cel.Program, error) {
	if cached, ok := e.programCache.Load(expression); ok {
		return cached.(cel.Program), nil
	}

	ast, issues := e.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("%w: %v", ErrCELCompilationFailed, issues.Err())
	}

	opts := []cel.ProgramOption{
		cel.InterruptCheckFrequency(e.config.CEL.InterruptCheckFrequency),
	}

	if e.config.CEL.CostLimit > 0 {
		opts = append(opts, cel.CostLimit(e.config.CEL.CostLimit))
	}

	opts = append(opts, e.buildEvalOptions()...)

	prg, err := e.env.Program(ast, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCELProgramCreationFailed, err)
	}

	e.programCache.Store(expression, prg)

	return prg, nil
}

// buildEvalOptions constructs CEL evaluation options from configuration
func (e *CELEvaluator) buildEvalOptions() []cel.ProgramOption {
	var evalOpts []cel.EvalOption

	if e.config.CEL.EvalOptimize {
		evalOpts = append(evalOpts, cel.OptOptimize)
	}

	if e.config.CEL.TrackState {
		evalOpts = append(evalOpts, cel.OptTrackState)
	}

	if len(evalOpts) == 0 {
		return nil
	}

	return []cel.ProgramOption{cel.EvalOptions(evalOpts...)}
}
