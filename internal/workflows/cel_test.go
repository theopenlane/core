package workflows

import (
	"testing"

	"github.com/google/cel-go/cel"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestNewCELEnv verifies CEL env creation with defaults
func TestNewCELEnv(t *testing.T) {
	env, err := NewCELEnv(nil, CELScopeAction)
	assert.Check(t, err)

	ast, issues := env.Compile(`user_id != "" && size(changed_fields) >= 0 && object != null`)
	assert.Check(t, issues.Err())

	_, err = env.Program(ast)
	assert.Check(t, err)
}

// TestCelEnv_ParserRecursionLimit verifies parser recursion limit config
func TestCelEnv_ParserRecursionLimit(t *testing.T) {
	env, err := newTestCELEnv(WithCELParserRecursionLimit(10))
	assert.Check(t, err)

	_, iss := env.Compile("0 + 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 + 11")
	assert.Check(t, iss.Err() != nil)
	assert.Check(t, is.Contains(iss.Err().Error(), "max recursion depth exceeded"))
}

// TestCelEnv_ParserExpressionSizeLimit verifies parser expression size limit config
func TestCelEnv_ParserExpressionSizeLimit(t *testing.T) {
	env, err := newTestCELEnv(WithCELParserExpressionSizeLimit(10))
	assert.Check(t, err)

	_, iss := env.Parse("'greetings'")
	assert.Check(t, iss.Err() != nil)
	assert.Check(t, is.Contains(iss.Err().Error(), "size exceeds limit"))
}

// TestCelEnv_ComprehensionNestingLimit verifies comprehension nesting limit config
func TestCelEnv_ComprehensionNestingLimit(t *testing.T) {
	env, err := newTestCELEnv(WithCELComprehensionNestingLimit(1))
	assert.Check(t, err)

	_, iss := env.Compile("[1].map(i, [2].map(j, i + j))")
	assert.Check(t, iss.Err() != nil)
	assert.Check(t, is.Contains(iss.Err().Error(), "comprehension exceeds nesting limit"))
}

// TestCelEnv_ExtendedValidations verifies extended validations config
func TestCelEnv_ExtendedValidations(t *testing.T) {
	expr := `"test".matches("x++")`

	env, err := newTestCELEnv(WithCELExtendedValidations(true))
	assert.Check(t, err)

	_, iss := env.Compile(expr)
	assert.Check(t, iss.Err() != nil)
	assert.Check(t, is.Contains(iss.Err().Error(), "invalid matches argument"))

	env, err = newTestCELEnv(WithCELExtendedValidations(false))
	assert.Check(t, err)

	_, iss = env.Compile(expr)
	assert.Check(t, iss.Err())
}

// TestCelEnv_OptionalTypes verifies optional types config
func TestCelEnv_OptionalTypes(t *testing.T) {
	env, err := newTestCELEnv(WithCELOptionalTypes(true))
	assert.Check(t, err)
	assert.Check(t, env.HasLibrary("cel.lib.optional"))
	assert.Check(t, env.HasFunction("optional.of"))

	env, err = newTestCELEnv(WithCELOptionalTypes(false))
	assert.Check(t, err)
	assert.Check(t, !env.HasLibrary("cel.lib.optional"))
	assert.Check(t, !env.HasFunction("optional.of"))
}

// TestCelEnv_IdentifierEscapeSyntax verifies identifier escape syntax config
func TestCelEnv_IdentifierEscapeSyntax(t *testing.T) {
	expr := "{'key-1': 64}.`key-1`"

	env, err := newTestCELEnv(WithCELIdentifierEscapeSyntax(true))
	assert.Check(t, err)

	_, iss := env.Compile(expr)
	assert.Check(t, iss.Err())

	env, err = newTestCELEnv(WithCELIdentifierEscapeSyntax(false))
	assert.Check(t, err)

	_, iss = env.Compile(expr)
	// this seems to always be true now for the Parser https://github.com/google/cel-go/pull/1295, so we no longer
	// expect an error
	assert.Check(t, err)
}

// TestCelEnv_CrossTypeNumericComparisons verifies cross-type numeric comparisons config
func TestCelEnv_CrossTypeNumericComparisons(t *testing.T) {
	expr := "1.0 < 2"

	env, err := newTestCELEnv(WithCELCrossTypeNumericComparisons(true))
	assert.Check(t, err)

	_, iss := env.Compile(expr)
	assert.Check(t, iss.Err())

	env, err = newTestCELEnv(WithCELCrossTypeNumericComparisons(false))
	assert.Check(t, err)

	_, iss = env.Compile(expr)
	assert.Check(t, iss.Err() != nil)
}

// TestCelEnv_MacroCallTracking verifies macro call tracking config
func TestCelEnv_MacroCallTracking(t *testing.T) {
	expr := "[1].exists(i, i > 0)"

	env, err := newTestCELEnv(WithCELMacroCallTracking(true))
	assert.Check(t, err)

	ast, iss := env.Parse(expr)
	assert.Check(t, iss.Err())

	parsed, err := cel.AstToParsedExpr(ast)
	assert.Check(t, err)
	assert.Check(t, len(parsed.GetSourceInfo().GetMacroCalls()) != 0)

	env, err = newTestCELEnv(WithCELMacroCallTracking(false))
	assert.Check(t, err)

	ast, iss = env.Parse(expr)
	assert.Check(t, iss.Err())

	parsed, err = cel.AstToParsedExpr(ast)
	assert.Check(t, err)
	assert.Check(t, is.Len(parsed.GetSourceInfo().GetMacroCalls(), 0))
}

func newTestCELEnv(opts ...ConfigOpts) (*cel.Env, error) {
	return NewCELEnv(NewDefaultConfig(opts...), CELScopeAction)
}
