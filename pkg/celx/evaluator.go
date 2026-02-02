package celx

import (
	"context"
	"errors"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

// Evaluator compiles and evaluates CEL expressions with caching

type Evaluator struct {
	env          *cel.Env
	config       EvalConfig
	programCache sync.Map
}

// NewEvaluator creates a new CEL evaluator with the provided environment and configuration

func NewEvaluator(env *cel.Env, config EvalConfig) *Evaluator {
	return &Evaluator{
		env:    env,
		config: config,
	}
}

// Compile parses and type-checks a CEL expression using the evaluator environment

func (e *Evaluator) Compile(expression string) (*cel.Ast, *cel.Issues) {
	return e.env.Compile(expression)
}

// Program builds a CEL program from an AST without evaluation options

func (e *Evaluator) Program(ast *cel.Ast) (cel.Program, error) {
	return e.env.Program(ast)
}

// Evaluate runs a CEL expression against the provided variables

func (e *Evaluator) Evaluate(ctx context.Context, expression string, vars map[string]any) (ref.Val, *cel.EvalDetails, error) {
	prg, err := e.getOrCompileProgram(expression)
	if err != nil {
		return nil, nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	evalCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	out, details, evalErr := prg.ContextEval(evalCtx, vars)
	if evalErr != nil {
		if errors.Is(evalErr, context.DeadlineExceeded) || errors.Is(evalErr, context.Canceled) || evalCtx.Err() != nil {
			return nil, details, evalCtx.Err()
		}

		return nil, details, evalErr
	}

	return out, details, nil
}

// EvaluateJSONMap evaluates a CEL expression and converts the result to a JSON object map

func (e *Evaluator) EvaluateJSONMap(ctx context.Context, expression string, vars map[string]any) (map[string]any, error) {
	out, _, err := e.Evaluate(ctx, expression, vars)
	if err != nil {
		return nil, err
	}

	return ToJSONMap(out)
}

// getOrCompileProgram returns a cached program or compiles a new one with evaluator options

func (e *Evaluator) getOrCompileProgram(expression string) (cel.Program, error) {
	if cached, ok := e.programCache.Load(expression); ok {
		return cached.(cel.Program), nil
	}

	ast, issues := e.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	opts := []cel.ProgramOption{
		cel.InterruptCheckFrequency(e.config.InterruptCheckFrequency),
	}

	if e.config.CostLimit > 0 {
		opts = append(opts, cel.CostLimit(e.config.CostLimit))
	}

	if programOpts := e.buildEvalOptions(); len(programOpts) > 0 {
		opts = append(opts, programOpts...)
	}

	prg, err := e.env.Program(ast, opts...)
	if err != nil {
		return nil, err
	}

	e.programCache.Store(expression, prg)

	return prg, nil
}

// buildEvalOptions assembles program options from evaluation configuration

func (e *Evaluator) buildEvalOptions() []cel.ProgramOption {
	var evalOpts []cel.EvalOption

	if e.config.EvalOptimize {
		evalOpts = append(evalOpts, cel.OptOptimize)
	}

	if e.config.TrackState {
		evalOpts = append(evalOpts, cel.OptTrackState)
	}

	if len(evalOpts) == 0 {
		return nil
	}

	return []cel.ProgramOption{cel.EvalOptions(evalOpts...)}
}
