package graphapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/internal/ent/generated"
)

func TestGetControlIDFromRefCode(t *testing.T) {
	ctx := context.Background()

	t.Run("match by control refCode", func(t *testing.T) {
		controls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
				Aliases: []string{"ALIAS-1", "ALIAS-2"},
			},
			{
				ID:      "control-2",
				RefCode: "REF-2",
				Aliases: []string{"ALIAS-3"},
			},
		}

		id, isSubcontrol := getControlIDFromRefCode(ctx, "REF-1", controls)

		assert.NotNil(t, id)
		assert.Equal(t, "control-1", *id)
		assert.False(t, isSubcontrol)
	})

	t.Run("match by control alias", func(t *testing.T) {
		controls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
				Aliases: []string{"ALIAS-1", "ALIAS-2"},
			},
		}

		id, isSubcontrol := getControlIDFromRefCode(ctx, "ALIAS-2", controls)

		assert.NotNil(t, id)
		assert.Equal(t, "control-1", *id)
		assert.False(t, isSubcontrol)
	})
}
