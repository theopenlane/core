package entityops

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/theopenlane/core/v2/pkg/celx"
)

// NewEvaluator binds the provided cel expressions to the target and source.
func NewEvaluator(targetType, sourceType reflect.Type) (*celx.NativeEntityEvaluator, error) {
	if targetType == nil {
		return nil, ErrEvaluatorBuildFailed
	}

	cfg := celx.StrictEnvConfig()
	cfg.CrossTypeNumericComparisons = true

	eval, err := celx.NewNativeEntityEvaluator(cfg, celx.FastEvalConfig(), targetType, sourceType)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEvaluatorBuildFailed, err)
	}

	return eval, nil
}

// MatchSelector evaluates the provided expression and tries to match them
func MatchSelector(ctx context.Context, eval *celx.NativeEntityEvaluator, expression string, entity any) (bool, error) {
	if strings.TrimSpace(expression) == "" {
		return true, nil
	}

	data, err := json.Marshal(entity)
	if err != nil {
		return false, fmt.Errorf("marshal selector entity: %w", err)
	}

	match, err := eval.EvaluateBool(ctx, expression, data)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrEvaluationFailed, err)
	}

	return match, nil
}
