package celx

import (
	"context"
	"errors"
	"testing"
	"time"

	"cel.dev/cel-go/common/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestEvaluatorCompileProgram(t *testing.T) {
	env, err := NewEnv(EnvConfig{})
	assert.Check(t, err)

	eval := NewEvaluator(env, EvalConfig{})

	ast, issues := eval.Compile("1 + 1")
	assert.Check(t, issues.Err())

	_, err = eval.Program(ast)
	assert.Check(t, err)

	_, issues = eval.Compile("1 +")
	assert.Check(t, issues.Err() != nil)
}

func TestEvaluatorEvaluateValue(t *testing.T) {
	env, err := NewEnv(EnvConfig{})
	assert.Check(t, err)

	eval := NewEvaluator(env, EvalConfig{Timeout: 100 * time.Millisecond})

	out, _, err := eval.Evaluate(context.Background(), "1 + 1", nil)
	assert.Check(t, err)
	assert.Check(t, is.Equal(types.Int(2), out))
}

func TestEvaluatorEvaluateTimeout(t *testing.T) {
	env, err := NewEnv(EnvConfig{ComprehensionNestingLimit: 2})
	assert.Check(t, err)

	eval := NewEvaluator(env, EvalConfig{
		Timeout:                 1 * time.Nanosecond,
		InterruptCheckFrequency: 1,
	})

	// Force evaluation work in a comprehension to trigger the timeout check

	_, _, err = eval.Evaluate(context.Background(), "[1].map(x, [2].map(y, x + y))", nil)
	assert.Check(t, err != nil)
	assert.Check(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
}

func TestEvaluatorEvaluateJSONMap(t *testing.T) {
	env, err := NewEnv(EnvConfig{})
	assert.Check(t, err)

	eval := NewEvaluator(env, EvalConfig{Timeout: 100 * time.Millisecond})

	out, err := eval.EvaluateJSONMap(context.Background(), `{"status": "ok", "count": 2}`, nil)
	assert.Check(t, err)
	assert.Check(t, is.Equal("ok", out["status"]))
	assert.Check(t, is.Equal(float64(2), out["count"]))

	_, err = eval.EvaluateJSONMap(context.Background(), `"nope"`, nil)
	assert.Check(t, err != nil)
	assert.Check(t, errors.Is(err, ErrJSONMapExpected))
}
