package celx

import (
	"testing"

	"github.com/google/cel-go/cel"
	"gotest.tools/v3/assert"
)

func TestNewEnvOptionsApplied(t *testing.T) {
	cfg := EnvConfig{
		OptionalTypes:          true,
		IdentifierEscapeSyntax: true,
	}

	env, err := NewEnv(cfg,
		cel.Variable("obj", cel.DynType),
	)
	assert.Check(t, err)

	assert.Check(t, env.HasLibrary("cel.lib.optional"))
	assert.Check(t, env.HasFunction("optional.of"))

	// Identifier escape syntax allows backtick identifiers

	_, iss := env.Compile("{'key-1': 1}.`key-1`")
	assert.Check(t, iss.Err())
}
