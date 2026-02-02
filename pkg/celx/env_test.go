package celx

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvOptionsApplied(t *testing.T) {
	cfg := EnvConfig{
		OptionalTypes:          true,
		IdentifierEscapeSyntax: true,
	}

	env, err := NewEnv(cfg,
		cel.Variable("obj", cel.DynType),
	)
	assert.NoError(t, err)

	assert.True(t, env.HasLibrary("cel.lib.optional"))
	assert.True(t, env.HasFunction("optional.of"))

	// Identifier escape syntax allows backtick identifiers

	_, iss := env.Compile("{'key-1': 1}.`key-1`")
	assert.NoError(t, iss.Err())
}
