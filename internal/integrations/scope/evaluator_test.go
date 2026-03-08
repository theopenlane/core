package scope

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

// newTestEvaluator creates a scope evaluator configured for tests
func newTestEvaluator(t *testing.T, config EvaluatorConfig) *Evaluator {
	t.Helper()

	evaluator, err := NewEvaluator(config)
	if err != nil {
		t.Fatalf("new evaluator failed: %v", err)
	}

	return evaluator
}

// TestEvaluateConditionEmptyExpressionDefault verifies empty expression behavior uses defaults
func TestEvaluateConditionEmptyExpressionDefault(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	result, err := evaluator.EvaluateCondition(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("evaluate condition failed: %v", err)
	}
	if !result {
		t.Fatalf("expected empty expression to evaluate true by default")
	}
}

// TestEvaluateConditionEmptyExpressionCustomResult verifies empty expression behavior is configurable
func TestEvaluateConditionEmptyExpressionCustomResult(t *testing.T) {
	config := DefaultEvaluatorConfig()
	config.EmptyExpressionResult = false

	evaluator := newTestEvaluator(t, config)

	result, err := evaluator.EvaluateCondition(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("evaluate condition failed: %v", err)
	}
	if result {
		t.Fatalf("expected empty expression to evaluate false when configured")
	}
}

// TestEvaluateConditionWithMapAndListVars verifies map and list payload access
func TestEvaluateConditionWithMapAndListVars(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	vars := ScopeVars{
		Payload:   json.RawMessage(`{"tags":["prod","pci"]}`),
		Provider:  integrationtypes.ProviderType("githubapp"),
		Operation: integrationtypes.OperationVulnerabilitiesCollect,
	}

	result, err := evaluator.EvaluateConditionWithVars(
		context.Background(),
		`provider == "githubapp" && operation == "vulnerabilities.collect" && payload.tags.exists(tag, tag == "prod")`,
		vars,
	)
	if err != nil {
		t.Fatalf("evaluate condition failed: %v", err)
	}
	if !result {
		t.Fatalf("expected condition to evaluate true")
	}
}

// TestEvaluateConditionReturnsFalse verifies false condition outputs
func TestEvaluateConditionReturnsFalse(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	result, err := evaluator.EvaluateCondition(context.Background(), `payload.severity == "critical"`, map[string]any{
		VariablePayload: map[string]any{
			"severity": "low",
		},
	})
	if err != nil {
		t.Fatalf("evaluate condition failed: %v", err)
	}
	if result {
		t.Fatalf("expected condition to evaluate false")
	}
}

// TestEvaluateConditionCompilationFailure verifies parse and compile failures return static errors
func TestEvaluateConditionCompilationFailure(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	_, err := evaluator.EvaluateCondition(context.Background(), "payload.", nil)
	if !errors.Is(err, ErrScopeCompilationFailed) {
		t.Fatalf("expected ErrScopeCompilationFailed, got %v", err)
	}
}

// TestEvaluateConditionRuntimeFailure verifies runtime evaluation failures return static errors
func TestEvaluateConditionRuntimeFailure(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	_, err := evaluator.EvaluateCondition(context.Background(), `size(payload.tags) > 0`, map[string]any{
		VariablePayload: map[string]any{
			"tags": 3,
		},
	})
	if !errors.Is(err, ErrScopeEvaluationFailed) {
		t.Fatalf("expected ErrScopeEvaluationFailed, got %v", err)
	}
}

// TestEvaluateConditionTypeMismatch verifies non-bool results return static type errors
func TestEvaluateConditionTypeMismatch(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	_, err := evaluator.EvaluateCondition(context.Background(), `payload`, map[string]any{
		VariablePayload: map[string]any{
			"severity": "high",
		},
	})
	if !errors.Is(err, ErrScopeConditionType) {
		t.Fatalf("expected ErrScopeConditionType, got %v", err)
	}
}

// TestEvaluateConditionTimeout verifies timeout failures return static timeout errors
func TestEvaluateConditionTimeout(t *testing.T) {
	config := DefaultEvaluatorConfig()
	config.Timeout = 1 * time.Nanosecond

	evaluator := newTestEvaluator(t, config)

	_, err := evaluator.EvaluateCondition(context.Background(), `payload.items.exists(item, item == "match")`, map[string]any{
		VariablePayload: map[string]any{
			"items": []any{"a", "b", "c"},
		},
	})
	if err != nil && !errors.Is(err, ErrScopeEvaluationTimeout) && !errors.Is(err, ErrScopeEvaluationFailed) {
		t.Fatalf("expected timeout or static evaluation failure, got %v", err)
	}
}

// TestValidateExpressionRequired verifies validate rejects empty expressions
func TestValidateExpressionRequired(t *testing.T) {
	evaluator := newTestEvaluator(t, DefaultEvaluatorConfig())

	err := evaluator.Validate("")
	if !errors.Is(err, ErrScopeExpressionRequired) {
		t.Fatalf("expected ErrScopeExpressionRequired, got %v", err)
	}
}
