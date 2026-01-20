package workflows

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/require"
)

func TestCelEnv_ParserRecursionLimit(t *testing.T) {
	env, err := NewCelEnv(WithCELParserRecursionLimit(10))
	require.NoError(t, err)

	_, iss := env.Compile("0 + 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 + 11")
	require.Error(t, iss.Err())
	require.Contains(t, iss.Err().Error(), "max recursion depth exceeded")
}

func TestCelEnv_ParserExpressionSizeLimit(t *testing.T) {
	env, err := NewCelEnv(WithCELParserExpressionSizeLimit(10))
	require.NoError(t, err)

	_, iss := env.Parse("'greetings'")
	require.Error(t, iss.Err())
	require.Contains(t, iss.Err().Error(), "size exceeds limit")
}

func TestCelEnv_ComprehensionNestingLimit(t *testing.T) {
	env, err := NewCelEnv(WithCELComprehensionNestingLimit(1))
	require.NoError(t, err)

	_, iss := env.Compile("[1].map(i, [2].map(j, i + j))")
	require.Error(t, iss.Err())
	require.Contains(t, iss.Err().Error(), "comprehension exceeds nesting limit")
}

func TestCelEnv_ExtendedValidations(t *testing.T) {
	expr := `"test".matches("x++")`

	env, err := NewCelEnv(WithCELExtendedValidations(true))
	require.NoError(t, err)

	_, iss := env.Compile(expr)
	require.Error(t, iss.Err())
	require.Contains(t, iss.Err().Error(), "invalid matches argument")

	env, err = NewCelEnv(WithCELExtendedValidations(false))
	require.NoError(t, err)

	_, iss = env.Compile(expr)
	require.NoError(t, iss.Err())
}

func TestCelEnv_OptionalTypes(t *testing.T) {
	env, err := NewCelEnv(WithCELOptionalTypes(true))
	require.NoError(t, err)
	require.True(t, env.HasLibrary("cel.lib.optional"))
	require.True(t, env.HasFunction("optional.of"))

	env, err = NewCelEnv(WithCELOptionalTypes(false))
	require.NoError(t, err)
	require.False(t, env.HasLibrary("cel.lib.optional"))
	require.False(t, env.HasFunction("optional.of"))
}

func TestCelEnv_IdentifierEscapeSyntax(t *testing.T) {
	expr := "{'key-1': 64}.`key-1`"

	env, err := NewCelEnv(WithCELIdentifierEscapeSyntax(true))
	require.NoError(t, err)

	_, iss := env.Compile(expr)
	require.NoError(t, iss.Err())

	env, err = NewCelEnv(WithCELIdentifierEscapeSyntax(false))
	require.NoError(t, err)

	_, iss = env.Compile(expr)
	require.Error(t, iss.Err())
}

func TestCelEnv_CrossTypeNumericComparisons(t *testing.T) {
	expr := "1.0 < 2"

	env, err := NewCelEnv(WithCELCrossTypeNumericComparisons(true))
	require.NoError(t, err)

	_, iss := env.Compile(expr)
	require.NoError(t, iss.Err())

	env, err = NewCelEnv(WithCELCrossTypeNumericComparisons(false))
	require.NoError(t, err)

	_, iss = env.Compile(expr)
	require.Error(t, iss.Err())
}

func TestCelEnv_MacroCallTracking(t *testing.T) {
	expr := "[1].exists(i, i > 0)"

	env, err := NewCelEnv(WithCELMacroCallTracking(true))
	require.NoError(t, err)

	ast, iss := env.Parse(expr)
	require.NoError(t, iss.Err())

	parsed, err := cel.AstToParsedExpr(ast)
	require.NoError(t, err)
	require.NotEmpty(t, parsed.GetSourceInfo().GetMacroCalls())

	env, err = NewCelEnv(WithCELMacroCallTracking(false))
	require.NoError(t, err)

	ast, iss = env.Parse(expr)
	require.NoError(t, iss.Err())

	parsed, err = cel.AstToParsedExpr(ast)
	require.NoError(t, err)
	require.Empty(t, parsed.GetSourceInfo().GetMacroCalls())
}
