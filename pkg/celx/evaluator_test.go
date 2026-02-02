package celx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/stretchr/testify/assert"
)

func TestEvaluatorCompileProgram(t *testing.T) {
	env, err := NewEnv(EnvConfig{})
	assert.NoError(t, err)

	eval := NewEvaluator(env, EvalConfig{})

	ast, issues := eval.Compile("1 + 1")
	assert.NoError(t, issues.Err())

	_, err = eval.Program(ast)
	assert.NoError(t, err)

	_, issues = eval.Compile("1 +")
	assert.Error(t, issues.Err())
}

func TestEvaluatorEvaluateValue(t *testing.T) {
	env, err := NewEnv(EnvConfig{})
	assert.NoError(t, err)

	eval := NewEvaluator(env, EvalConfig{Timeout: 100 * time.Millisecond})

	out, _, err := eval.Evaluate(context.Background(), "1 + 1", nil)
	assert.NoError(t, err)
	assert.Equal(t, types.Int(2), out)
}

func TestEvaluatorEvaluateTimeout(t *testing.T) {
	env, err := NewEnv(EnvConfig{ComprehensionNestingLimit: 2})
	assert.NoError(t, err)

	eval := NewEvaluator(env, EvalConfig{
		Timeout:                 1 * time.Nanosecond,
		InterruptCheckFrequency: 1,
	})

	// Force evaluation work in a comprehension to trigger the timeout check

	_, _, err = eval.Evaluate(context.Background(), "[1].map(x, [2].map(y, x + y))", nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
}

func TestEvaluatorEvaluateJSONMap(t *testing.T) {
	env, err := NewEnv(EnvConfig{})
	assert.NoError(t, err)

	eval := NewEvaluator(env, EvalConfig{Timeout: 100 * time.Millisecond})

	out, err := eval.EvaluateJSONMap(context.Background(), `{"status": "ok", "count": 2}`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "ok", out["status"])
	assert.Equal(t, float64(2), out["count"])

	_, err = eval.EvaluateJSONMap(context.Background(), `"nope"`, nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrJSONMapExpected))
}
