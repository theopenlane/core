package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOldestRevision(t *testing.T) {
	v1 := "v1.0.0"
	v2 := "v2.0.0"
	v1Minor := "v1.1.0"

	controls := []standardControl{
		{
			ID:                         "control-1",
			ReferenceFrameworkRevision: &v2,
		},
		{
			ID:                         "control-2",
			ReferenceFrameworkRevision: &v1,
		},
		{
			ID:                         "control-3",
			ReferenceFrameworkRevision: &v1Minor,
		},
	}

	assert.Equal(t, v1, pickOldestRevision(controls))
}

func TestSubcontrolsForControls(t *testing.T) {
	controls := []standardControl{
		{ID: "control-1"},
		{ID: "control-2"},
	}

	grouped := map[string][]standardSubcontrol{
		"control-1": []standardSubcontrol{
			{standardControl: standardControl{ID: "subcontrol-1"}, ControlID: "control-1"},
		},
		"control-2": []standardSubcontrol{
			{standardControl: standardControl{ID: "subcontrol-2"}, ControlID: "control-2"},
			{standardControl: standardControl{ID: "subcontrol-3"}, ControlID: "control-2"},
		},
	}

	assert.Equal(t, []standardSubcontrol{
		{standardControl: standardControl{ID: "subcontrol-1"}, ControlID: "control-1"},
		{standardControl: standardControl{ID: "subcontrol-2"}, ControlID: "control-2"},
		{standardControl: standardControl{ID: "subcontrol-3"}, ControlID: "control-2"},
	}, fetchSubcontrolsOwnedByControl(controls, grouped))
}
