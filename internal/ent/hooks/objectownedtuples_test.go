package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/internal/ent/generated/group"
)

func TestMapEdgeToObjectType(t *testing.T) {
	tests := []struct {
		name     string
		edge     string
		wantType string
		wantOk   bool
	}{
		{
			name:     "owner edge",
			edge:     "owner",
			wantType: "",
			wantOk:   false,
		},
		{
			name:     "setting edge",
			edge:     "setting",
			wantType: "",
			wantOk:   false,
		},
		{
			name:     "approver edge",
			edge:     "approver",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "delegate edge",
			edge:     "delegate",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "stakeholder edge",
			edge:     "stakeholder",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "editors edge",
			edge:     "editors",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "viewers edge",
			edge:     "viewers",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "blocked_groups edge",
			edge:     "blocked_groups",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "membership edge",
			edge:     "program_membership",
			wantType: "",
			wantOk:   false,
		},
		{
			name:     "program_editors edge",
			edge:     "program_editors",
			wantType: "program",
			wantOk:   true,
		},
		{
			name:     "program_viewers edge",
			edge:     "program_viewers",
			wantType: "program",
			wantOk:   true,
		},
		{
			name:     "program_blocked_groups edge",
			edge:     "program_blocked_groups",
			wantType: "program",
			wantOk:   true,
		},
		{
			name:     "program_creators edge",
			edge:     "program_creators",
			wantType: group.Label,
			wantOk:   true,
		},
		{
			name:     "plural edge",
			edge:     "programs",
			wantType: "program",
			wantOk:   true,
		},
		{
			name:     "singular edge",
			edge:     "project",
			wantType: "project",
			wantOk:   true,
		},
		{
			name:     "unknown edge",
			edge:     "foobar",
			wantType: "foobar",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOk := mapEdgeToObjectType(tt.edge)
			assert.Equal(t, tt.wantType, gotType)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}
