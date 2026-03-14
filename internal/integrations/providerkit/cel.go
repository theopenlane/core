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

	"github.com/theopenlane/core/pkg/celx"
)

const (
	celParserRecursionLimit    = 250
	celExpressionSizeLimit     = 100_000
	celInterruptCheckFrequency = 100
	celTimeout                 = 100 * time.Millisecond

	// celVarEnvelope is the CEL variable name bound to the alert envelope map
	celVarEnvelope = "envelope"
)

var (
	celEvalOnce sync.Once
	celEval     *celx.Evaluator
	celEvalErr  error
)

// AlertEnvelope wraps a webhook alert payload for CEL filter and map evaluation
type AlertEnvelope struct {
	// AlertType identifies the alert category (dependabot, code_scanning, etc.)
	AlertType string `json:"alertType"`
	// Resource identifies the alert resource (repo, project, etc.)
	Resource string `json:"resource,omitempty"`
	// Action indicates the webhook action (created, resolved, etc.)
	Action string `json:"action,omitempty"`
	// Payload is the raw alert payload as received from the provider
	Payload json.RawMessage `json:"payload,omitempty"`
}

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

	return celx.NewEnv(cfg,
		cel.VariableDecls(
			decls.NewVariable(celVarEnvelope, celtypes.DynType),
		),
	)
}

// envelopeToVars converts an AlertEnvelope into the CEL variable map.
// The payload field is decoded as a map when valid JSON; otherwise it is included as a string.
func envelopeToVars(envelope AlertEnvelope) map[string]any {
	var payload any

	if len(envelope.Payload) > 0 {
		var m map[string]any
		if err := json.Unmarshal(envelope.Payload, &m); err == nil {
			payload = m
		} else {
			payload = string(envelope.Payload)
		}
	}

	return map[string]any{
		celVarEnvelope: map[string]any{
			"alertType": envelope.AlertType,
			"resource":  envelope.Resource,
			"action":    envelope.Action,
			"payload":   payload,
		},
	}
}

// EvalFilter evaluates a CEL filter expression against an AlertEnvelope.
// An empty expr returns true (pass-through). Returns false when the expression excludes the envelope,
// or a wrapped ErrFilterExprEval on evaluation failure.
func EvalFilter(ctx context.Context, expr string, envelope AlertEnvelope) (bool, error) {
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

// EvalMap evaluates a CEL map expression against an AlertEnvelope and returns a JSON payload.
// An empty expr returns the original envelope.Payload (pass-through).
// Returns a wrapped ErrMapExprEval on failure.
func EvalMap(ctx context.Context, expr string, envelope AlertEnvelope) (json.RawMessage, error) {
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
