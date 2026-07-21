package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
)

// TestSubprocessorEntries verifies changed join rows are coalesced per subprocessor and classified by
// their net change relative to the last notification floor: a vendor removed after the floor is a
// removal, newly created is an addition, updated in place is an update, a deleted-and-recreated vendor
// renders once as an update rather than a removal plus an addition, and a vendor both added and removed
// within the window is dropped
func TestSubprocessorEntries(t *testing.T) {
	t.Parallel()

	floor := time.Now().Add(-24 * time.Hour)

	vendor := func(id, name string) *ent.Subprocessor {
		return &ent.Subprocessor{ID: id, Name: name}
	}

	row := func(sp *ent.Subprocessor, createdAt, deletedAt time.Time) *ent.TrustCenterSubprocessor {
		return &ent.TrustCenterSubprocessor{
			SubprocessorID: sp.ID,
			CreatedAt:      createdAt,
			DeletedAt:      deletedAt,
			Edges:          ent.TrustCenterSubprocessorEdges{Subprocessor: sp},
		}
	}

	var (
		none      = time.Time{}
		before    = floor.Add(-time.Hour)
		after     = time.Now()
		aws       = vendor("sp-aws", "Amazon Web Services")
		gcp       = vendor("sp-gcp", "Google Cloud")
		datadog   = vendor("sp-datadog", "Datadog")
		snowflake = vendor("sp-snowflake", "Snowflake")
		ephemeral = vendor("sp-ephemeral", "Ephemeral")
	)

	entries := subprocessorEntries([]*ent.TrustCenterSubprocessor{
		// removed: existed at the floor, soft-deleted in the window
		row(aws, before, after),
		// added: created in the window and still live
		row(gcp, after, none),
		// updated in place: existed at the floor and still live
		row(datadog, before, none),
		// deleted and recreated in the window: nets to a single update
		row(snowflake, before, after),
		row(snowflake, after, none),
		// added and removed within the window: nets to no change
		row(ephemeral, after, after),
	}, floor)

	assert.Equal(t, []SubprocessorEntry{
		{Name: "Amazon Web Services", Change: "Removed"},
		{Name: "Google Cloud", Change: "Added"},
		{Name: "Datadog", Change: "Updated"},
		{Name: "Snowflake", Change: "Updated"},
	}, entries)
}
