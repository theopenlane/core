package providerkit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	celtypes "github.com/google/cel-go/common/types"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/celx"
)

const (
	// celParserRecursionLimit caps the CEL parser recursion depth to prevent stack overflows
	celParserRecursionLimit = 250
	// celExpressionSizeLimit caps the maximum CEL expression byte length
	celExpressionSizeLimit = 100_000
	// celInterruptCheckFrequency controls how often the CEL evaluator checks for context cancellation
	celInterruptCheckFrequency = 100
	// celTimeout is the maximum wall-clock time allowed for a single CEL evaluation
	celTimeout = 100 * time.Millisecond

	// celVarEnvelope is the CEL variable name bound to the alert envelope map
	celVarEnvelope = "envelope"
	// celVarVariant is the CEL variable name bound to the envelope variant field
	celVarVariant = "variant"
	// celVarResource is the CEL variable name bound to the envelope resource field
	celVarResource = "resource"
	// celVarAction is the CEL variable name bound to the envelope action field
	celVarAction = "action"
	// celVarPayload is the CEL variable name bound to the envelope payload field
	celVarPayload = "payload"
)

var (
	// celEvalOnce guards single initialization of the shared CEL evaluator
	celEvalOnce sync.Once
	// celEval is the lazily-initialized shared CEL evaluator instance
	celEval *celx.Evaluator
	// celEvalErr holds the initialization error, if any, from the first celEvalOnce.Do call
	celEvalErr error
)

// getEvaluator returns the shared CEL evaluator, initializing it once on first call
func getEvaluator() (*celx.Evaluator, error) {
	celEvalOnce.Do(func() {
		env, err := buildEnvelopeEnv()
		if err != nil {
			celEvalErr = err
			return
		}

		celEval = celx.NewEvaluator(env, celx.EvalConfig{
			Timeout:                 celTimeout,
			InterruptCheckFrequency: celInterruptCheckFrequency,
			EvalOptimize:            true,
		})
	})

	return celEval, celEvalErr
}

// buildEnvelopeEnv constructs the CEL environment declaring the envelope variable
func buildEnvelopeEnv() (*cel.Env, error) {
	cfg := celx.EnvConfig{
		ParserRecursionLimit:        celParserRecursionLimit,
		ParserExpressionSizeLimit:   celExpressionSizeLimit,
		ExtendedValidations:         true,
		CrossTypeNumericComparisons: true,
	}

	return celx.NewEnv(cfg, cel.VariableDecls(
		decls.NewVariable(celVarEnvelope, celtypes.DynType),
		decls.NewVariable(celVarVariant, celtypes.DynType),
		decls.NewVariable(celVarResource, celtypes.DynType),
		decls.NewVariable(celVarAction, celtypes.DynType),
		decls.NewVariable(celVarPayload, celtypes.DynType),
	))
}

// envelopeToVars converts a MappingEnvelope into the CEL variable map
func envelopeToVars(envelope types.MappingEnvelope) map[string]any {
	var payload any

	if len(envelope.Payload) > 0 {
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			payload = string(envelope.Payload)
		}
	}

	return map[string]any{
		celVarEnvelope: map[string]any{
			"variant":  envelope.Variant,
			"resource": envelope.Resource,
			"action":   envelope.Action,
			"payload":  payload,
		},
		celVarVariant:  envelope.Variant,
		celVarResource: envelope.Resource,
		celVarAction:   envelope.Action,
		celVarPayload:  payload,
	}
}

// EvalFilter evaluates a CEL filter expression against a MappingEnvelope
// An empty expr returns true (pass-through). Returns false when the expression excludes the envelope,
// or a wrapped ErrFilterExprEval on evaluation failure
func EvalFilter(ctx context.Context, expr string, envelope types.MappingEnvelope) (bool, error) {
	if expr == "" {
		return true, nil
	}

	ev, err := getEvaluator()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrFilterExprEval, err)
	}

	out, _, err := ev.Evaluate(ctx, expr, envelopeToVars(envelope))
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrFilterExprEval, err)
	}

	if out.Type() != celtypes.BoolType {
		return false, ErrFilterExprEval
	}

	value, ok := out.Value().(bool)
	if !ok {
		value = out.Equal(celtypes.True) == celtypes.True
	}

	return value, nil
}

// EvalMap evaluates a CEL map expression against a MappingEnvelope and returns a JSON payload
// An empty expr returns the original envelope.Payload (pass-through)
// Returns a wrapped ErrMapExprEval on failure
func EvalMap(ctx context.Context, expr string, envelope types.MappingEnvelope) (json.RawMessage, error) {
	if expr == "" {
		return envelope.Payload, nil
	}

	ev, err := getEvaluator()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMapExprEval, err)
	}

	result, err := ev.EvaluateJSONMap(ctx, expr, envelopeToVars(envelope))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMapExprEval, err)
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMapExprEval, err)
	}

	return raw, nil
}

// ValidateExpr parses and type-checks a CEL expression without evaluating it
func ValidateExpr(expr string) error {
	ev, err := getEvaluator()
	if err != nil {
		return err
	}

	_, issues := ev.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return issues.Err()
	}

	return nil
}
