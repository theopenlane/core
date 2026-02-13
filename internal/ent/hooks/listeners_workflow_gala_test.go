package hooks_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterGalaWorkflowListeners(t *testing.T) {
	registry := gala.NewRegistry()

	ids, err := hooks.RegisterGalaWorkflowListeners(registry)
	require.NoError(t, err)
	require.Len(t, ids, len(enums.WorkflowObjectTypes)+1)
}
