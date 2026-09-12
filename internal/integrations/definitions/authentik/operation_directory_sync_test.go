package authentik

import (
	"testing"
	"time"

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

	since := time.Now()

	tests := []struct {
		name      string
		userSince *time.Time
		want      map[string]bool
	}{
		{
			name:      "full user fetch marks all sets complete",
			userSince: nil,
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name:    true,
				entityops.SchemaDirectoryGroup.Name:      true,
				entityops.SchemaDirectoryMembership.Name: true,
			},
		},
		{
			name:      "incremental user fetch marks all sets incomplete",
			userSince: &since,
			want: map[string]bool{
				entityops.SchemaDirectoryAccount.Name:    false,
				entityops.SchemaDirectoryGroup.Name:      false,
				entityops.SchemaDirectoryMembership.Name: false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sets := append([]types.IngestPayloadSet{accountPayloadSet(nil, tc.userSince)}, groupPayloadSets(nil, nil, tc.userSince)...)

			assert.DeepEqual(t, snapshotFlagsBySchema(sets), tc.want)
		})
	}
}
