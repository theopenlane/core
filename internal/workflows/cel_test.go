package workflows

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
)

// TestNewCELEnv verifies CEL env creation with defaults
func TestNewCELEnv(t *testing.T) {
	env, err := NewCELEnv(nil, CELScopeAction)
	assert.NoError(t, err)

	ast, issues := env.Compile(`user_id != "" && size(changed_fields) >= 0 && object != null`)
	assert.NoError(t, issues.Err())

	_, err = env.Program(ast)
	assert.NoError(t, err)
}

// TestCelEnv_ParserRecursionLimit verifies parser recursion limit config
func TestCelEnv_ParserRecursionLimit(t *testing.T) {
	env, err := newTestCELEnv(WithCELParserRecursionLimit(10))
	assert.NoError(t, err)

	_, iss := env.Compile("0 + 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 + 11")
	assert.Error(t, iss.Err())
	assert.Contains(t, iss.Err().Error(), "max recursion depth exceeded")
}

// TestCelEnv_ParserExpressionSizeLimit verifies parser expression size limit config
func TestCelEnv_ParserExpressionSizeLimit(t *testing.T) {
	env, err := newTestCELEnv(WithCELParserExpressionSizeLimit(10))
	assert.NoError(t, err)

	_, iss := env.Parse("'greetings'")
	assert.Error(t, iss.Err())
	assert.Contains(t, iss.Err().Error(), "size exceeds limit")
}

// TestCelEnv_ComprehensionNestingLimit verifies comprehension nesting limit config
func TestCelEnv_ComprehensionNestingLimit(t *testing.T) {
	env, err := newTestCELEnv(WithCELComprehensionNestingLimit(1))
	assert.NoError(t, err)

	_, iss := env.Compile("[1].map(i, [2].map(j, i + j))")
	assert.Error(t, iss.Err())
	assert.Contains(t, iss.Err().Error(), "comprehension exceeds nesting limit")
}

// TestCelEnv_ExtendedValidations verifies extended validations config
func TestCelEnv_ExtendedValidations(t *testing.T) {
	expr := `"test".matches("x++")`

	env, err := newTestCELEnv(WithCELExtendedValidations(true))
	assert.NoError(t, err)

	_, iss := env.Compile(expr)
	assert.Error(t, iss.Err())
	assert.Contains(t, iss.Err().Error(), "invalid matches argument")

	env, err = newTestCELEnv(WithCELExtendedValidations(false))
	assert.NoError(t, err)

	_, iss = env.Compile(expr)
	assert.NoError(t, iss.Err())
}

// TestCelEnv_OptionalTypes verifies optional types config
func TestCelEnv_OptionalTypes(t *testing.T) {
	env, err := newTestCELEnv(WithCELOptionalTypes(true))
	assert.NoError(t, err)
	assert.True(t, env.HasLibrary("cel.lib.optional"))
	assert.True(t, env.HasFunction("optional.of"))

	env, err = newTestCELEnv(WithCELOptionalTypes(false))
	assert.NoError(t, err)
	assert.False(t, env.HasLibrary("cel.lib.optional"))
	assert.False(t, env.HasFunction("optional.of"))
}

// TestCelEnv_IdentifierEscapeSyntax verifies identifier escape syntax config
func TestCelEnv_IdentifierEscapeSyntax(t *testing.T) {
	expr := "{'key-1': 64}.`key-1`"

	env, err := newTestCELEnv(WithCELIdentifierEscapeSyntax(true))
	assert.NoError(t, err)

	_, iss := env.Compile(expr)
	assert.NoError(t, iss.Err())

	env, err = newTestCELEnv(WithCELIdentifierEscapeSyntax(false))
	assert.NoError(t, err)

	_, iss = env.Compile(expr)
	assert.Error(t, iss.Err())
}

// TestCelEnv_CrossTypeNumericComparisons verifies cross-type numeric comparisons config
func TestCelEnv_CrossTypeNumericComparisons(t *testing.T) {
	expr := "1.0 < 2"

	env, err := newTestCELEnv(WithCELCrossTypeNumericComparisons(true))
	assert.NoError(t, err)

	_, iss := env.Compile(expr)
	assert.NoError(t, iss.Err())

	env, err = newTestCELEnv(WithCELCrossTypeNumericComparisons(false))
	assert.NoError(t, err)

	_, iss = env.Compile(expr)
	assert.Error(t, iss.Err())
}

// TestCelEnv_MacroCallTracking verifies macro call tracking config
func TestCelEnv_MacroCallTracking(t *testing.T) {
	expr := "[1].exists(i, i > 0)"

	env, err := newTestCELEnv(WithCELMacroCallTracking(true))
	assert.NoError(t, err)

	ast, iss := env.Parse(expr)
	assert.NoError(t, iss.Err())

	parsed, err := cel.AstToParsedExpr(ast)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed.GetSourceInfo().GetMacroCalls())

	env, err = newTestCELEnv(WithCELMacroCallTracking(false))
	assert.NoError(t, err)

	ast, iss = env.Parse(expr)
	assert.NoError(t, iss.Err())

	parsed, err = cel.AstToParsedExpr(ast)
	assert.NoError(t, err)
	assert.Empty(t, parsed.GetSourceInfo().GetMacroCalls())
}

func newTestCELEnv(opts ...ConfigOpts) (*cel.Env, error) {
	return NewCELEnv(NewDefaultConfig(opts...), CELScopeAction)
}
