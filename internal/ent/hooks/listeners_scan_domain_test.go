package hooks

import (
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterGalaDomainScanSubmitListeners(t *testing.T) {
	t.Parallel()

	registry := gala.NewRegistry()

	ids, err := RegisterGalaDomainScanSubmitListeners(registry)
	assert.NilError(t, err)
	assert.Equal(t, len(ids), 1)

	topic := eventqueue.MutationTopicName(eventqueue.MutationConcernDirect, generated.TypeScan)
	assert.Check(t, registry.InterestedIn(topic, ent.OpCreate.String()))
	assert.Check(t, !registry.InterestedIn(topic, ent.OpUpdate.String()))
	assert.Check(t, !registry.InterestedIn(topic, ent.OpDelete.String()))
}

func TestIsPendingDomainScan(t *testing.T) {
	tests := []struct {
		name string
		scan *generated.Scan
		want bool
	}{
		{
			name: "domain scan awaiting submission",
			scan: &generated.Scan{ScanType: enums.ScanTypeDomain, Status: enums.ScanStatusProcessing, PerformedBy: operations.DomainScanPerformedBy},
			want: true,
		},
		{
			name: "domain scan queued via the REST endpoint",
			scan: &generated.Scan{ScanType: enums.ScanTypeDomain, Status: enums.ScanStatusPending, PerformedBy: operations.DomainScanPerformedBy},
			want: true,
		},
		{
			name: "non-domain scan is ignored",
			scan: &generated.Scan{ScanType: enums.ScanTypeVulnerability, Status: enums.ScanStatusProcessing, PerformedBy: operations.DomainScanPerformedBy},
			want: false,
		},
		{
			name: "already-completed historical record is not resubmitted",
			scan: &generated.Scan{ScanType: enums.ScanTypeDomain, Status: enums.ScanStatusCompleted, PerformedBy: operations.DomainScanPerformedBy},
			want: false,
		},
		{
			name: "already-failed historical record is not resubmitted",
			scan: &generated.Scan{ScanType: enums.ScanTypeDomain, Status: enums.ScanStatusFailed, PerformedBy: operations.DomainScanPerformedBy},
			want: false,
		},
		{
			name: "domain scan not marked as openlane-managed is ignored",
			scan: &generated.Scan{ScanType: enums.ScanTypeDomain, Status: enums.ScanStatusProcessing, PerformedBy: "third-party-pentest"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPendingDomainScan(tt.scan)

			assert.Equal(t, tt.want, got)
		})
	}
}
