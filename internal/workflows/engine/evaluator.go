package engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/iam/auth"
)

// EvaluateConditions checks if all conditions pass for a workflow
func (e *WorkflowEngine) EvaluateConditions(ctx context.Context, def *generated.WorkflowDefinition, obj *workflows.Object, eventType string, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string) (bool, error) {
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return false, err
	}

	vars := workflows.BuildCELVars(obj, changedFields, changedEdges, addedIDs, removedIDs, eventType, userID)

	for i, cond := range def.DefinitionJSON.Conditions {
		result, err := e.evaluateExpression(ctx, cond.Expression, vars)
		if err != nil {
			return false, fmt.Errorf("%w: condition %d: %v", ErrConditionFailed, i, err)
		}

		if !result {
			return false, nil
		}
	}

	return true, nil
}

// evaluateExpression evaluates a CEL expression with the given variables, using caching and timeout
func (e *WorkflowEngine) evaluateExpression(ctx context.Context, expression string, vars map[string]any) (bool, error) {
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

		return false, ErrConditionFailed
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
func (e *WorkflowEngine) getOrCompileProgram(expression string) (cel.Program, error) {
	if cached, ok := e.programCache.Load(expression); ok {
		return cached.(cel.Program), nil
	}

	ast, issues := e.env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, ErrCELCompilationFailed
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
		return nil, ErrCELProgramCreationFailed
	}

	e.programCache.Store(expression, prg)

	return prg, nil
}

// buildEvalOptions constructs CEL evaluation options from configuration
func (e *WorkflowEngine) buildEvalOptions() []cel.ProgramOption {
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
