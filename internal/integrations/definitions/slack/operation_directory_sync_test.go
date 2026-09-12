package slack

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// TestDirectorySyncSnapshotComplete verifies the account payload set is complete only for a full user fetch
func TestDirectorySyncSnapshotComplete(t *testing.T) {
	t.Parallel()

	lastRun := time.Now()

	tests := []struct {
		name      string
		lastRunAt *time.Time
		want      bool
	}{
		{
			name:      "full user fetch marks account set complete",
			lastRunAt: nil,
			want:      true,
		},
		{
			name:      "incremental user fetch marks account set incomplete",
			lastRunAt: &lastRun,
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			set := accountPayloadSet(nil, tc.lastRunAt)

			assert.Equal(t, set.Schema, entityops.SchemaDirectoryAccount.Name)
			assert.Equal(t, set.SnapshotComplete, tc.want)
		})
	}
}
