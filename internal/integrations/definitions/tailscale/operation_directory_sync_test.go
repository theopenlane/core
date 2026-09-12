package tailscale

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// snapshotFlagsBySchema indexes payload set completeness by schema name
func snapshotFlagsBySchema(sets []types.IngestPayloadSet) map[string]bool {
	flags := make(map[string]bool, len(sets))

	for _, set := range sets {
		flags[set.Schema] = set.SnapshotComplete
	}

	return flags
}

// TestDirectorySyncSnapshotComplete verifies account, group, and membership payload sets carry the expected completeness flags
func TestDirectorySyncSnapshotComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		membershipsComplete bool
		want                map[string]bool
	}{
		{
			name:                "policy file fetched marks all sets complete",
			membershipsComplete: true,
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name:    true,
				entityops.SchemaDirectoryGroup.Name:      true,
				entityops.SchemaDirectoryMembership.Name: true,
			},
		},
		{
			name:                "policy file fetch failed marks group and membership sets incomplete",
			membershipsComplete: false,
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name:    true,
				entityops.SchemaDirectoryGroup.Name:      false,
				entityops.SchemaDirectoryMembership.Name: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sets := append([]types.IngestPayloadSet{accountPayloadSet(nil)}, groupPayloadSets(nil, nil, tc.membershipsComplete)...)

			assert.DeepEqual(t, snapshotFlagsBySchema(sets), tc.want)
		})
	}
}
