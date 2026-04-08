package controls_test

import (
	"testing"

	"github.com/theopenlane/core/internal/controls"
	"github.com/theopenlane/core/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGetControlIDFromRefCode(t *testing.T) {
	t.Run("match by control refCode", func(t *testing.T) {
		matchControls := []*generated.Control{
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

		id, isSubcontrol := controls.GetControlIDFromRefCode("REF-1", matchControls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(false, isSubcontrol))
	})

	t.Run("match by control alias", func(t *testing.T) {
		matchControls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
				Aliases: []string{"ALIAS-1", "ALIAS-2"},
			},
		}

		id, isSubcontrol := controls.GetControlIDFromRefCode("ALIAS-2", matchControls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(false, isSubcontrol))
	})

	t.Run("match by subcontrol refCode", func(t *testing.T) {
		matchControls := []*generated.Control{
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

		id, isSubcontrol := controls.GetControlIDFromRefCode("SUBREF-1", matchControls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(true, isSubcontrol))
	})

	t.Run("match by subcontrol alias", func(t *testing.T) {
		matchControls := []*generated.Control{
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

		id, isSubcontrol := controls.GetControlIDFromRefCode("SUBALIAS-1", matchControls)

		assert.Assert(t, id != nil)
		assert.Check(t, is.Equal("control-1", *id))
		assert.Check(t, is.Equal(true, isSubcontrol))
	})

	t.Run("no match found", func(t *testing.T) {
		matchControls := []*generated.Control{
			{
				ID:      "control-1",
				RefCode: "REF-1",
			},
		}

		id, isSubcontrol := controls.GetControlIDFromRefCode("NON-EXIST	ENT", matchControls)

		assert.Assert(t, id == nil)
		assert.Check(t, is.Equal(false, isSubcontrol))
	})
}
