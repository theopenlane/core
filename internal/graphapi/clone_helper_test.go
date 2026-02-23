package graphapi

import (
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGetControlIDFromRefCode(t *testing.T) {
	t.Parallel()

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

		id, isSubcontrol := getControlIDFromRefCode("REF-1", controls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(false, isSubcontrol))
	})

	t.Run("match by control alias", func(t *testing.T) {
		controls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
				Aliases: []string{"ALIAS-1", "ALIAS-2"},
			},
		}

		id, isSubcontrol := getControlIDFromRefCode("ALIAS-2", controls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(false, isSubcontrol))
	})

	t.Run("match by subcontrol refCode", func(t *testing.T) {
		controls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
				Edges: generated.ControlEdges{
					Subcontrols: []*generated.Subcontrol{
						{
							ID:      "subcontrol-1",
							RefCode: "SUBREF-1",
							Aliases: []string{"SUBALIAS-1"},
						},
					},
				},
			},
		}

		id, isSubcontrol := getControlIDFromRefCode("SUBREF-1", controls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(true, isSubcontrol))
	})

	t.Run("match by subcontrol alias", func(t *testing.T) {
		controls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
				Edges: generated.ControlEdges{
					Subcontrols: []*generated.Subcontrol{
						{
							ID:      "subcontrol-1",
							RefCode: "SUBREF-1",
							Aliases: []string{"SUBALIAS-1"},
						},
					},
				},
			},
		}

		id, isSubcontrol := getControlIDFromRefCode("SUBALIAS-1", controls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(true, isSubcontrol))
	})

	t.Run("no match found", func(t *testing.T) {
		controls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
			},
		}

		id, isSubcontrol := getControlIDFromRefCode("NON-EXIST	ENT", controls)

		assert.Assert(t, id == nil)
		assert.Check(t, is.Equal(false, isSubcontrol))
	})
}
