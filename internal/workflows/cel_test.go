package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewCelEnv verifies CEL env creation with defaults
func TestNewCelEnv(t *testing.T) {
	env, err := NewCelEnv()
	assert.NoError(t, err)

	ast, issues := env.Compile(`user_id != "" && size(changed_fields) >= 0 && object != null`)
	assert.NoError(t, issues.Err())

	_, err = env.Program(ast)
	assert.NoError(t, err)
}
