package workflows

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewCelEnv verifies CEL env creation with defaults
func TestNewCelEnv(t *testing.T) {
	env, err := NewCelEnv()
	require.NoError(t, err)

	ast, issues := env.Compile(`user_id != "" && size(changed_fields) >= 0 && object != null`)
	require.NoError(t, issues.Err())

	_, err = env.Program(ast)
	require.NoError(t, err)
}
