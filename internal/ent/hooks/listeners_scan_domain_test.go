package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterGalaDomainScanListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewInMemory()
	assert.NilError(t, err)

	ids, err := gala.Register(runtime, DomainScanListeners()...)
	assert.NilError(t, err)
	assert.Equal(t, len(ids), 2)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, generated.TypeScan)
	assert.Check(t, runtime.InterestedIn(topic, ent.OpCreate.String()))
	assert.Check(t, !runtime.InterestedIn(topic, ent.OpUpdate.String()))
	assert.Check(t, !runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestDomainScanSubmitGate(t *testing.T) {
	t.Parallel()

	listener, ok := DomainScanListeners()[0].(entityops.MutationListener)
	assert.Check(t, ok)

	gate := listener.Definition().Gate

	proposed := func(scanType, status, performedBy string) entityops.MutationPayload {
		return entityops.MutationPayload{
			MutationType: generated.TypeScan,
			Operation:    entityops.OpCreate,
			EntityID:     "scan-1",
			ChangeSet: entityops.ChangeSet{
				ProposedChanges: map[string]any{
					scan.FieldScanType:    scanType,
					scan.FieldStatus:      status,
					scan.FieldPerformedBy: performedBy,
				},
			},
		}
	}

	assert.Check(t, gate(context.Background(), proposed(enums.ScanTypeDomain.String(), enums.ScanStatusPending.String(), cloudflare.DomainScanPerformedBy)))
	assert.Check(t, !gate(context.Background(), proposed(enums.ScanTypeVulnerability.String(), enums.ScanStatusPending.String(), cloudflare.DomainScanPerformedBy)))
	assert.Check(t, !gate(context.Background(), proposed(enums.ScanTypeDomain.String(), enums.ScanStatusCompleted.String(), cloudflare.DomainScanPerformedBy)))
	assert.Check(t, !gate(context.Background(), proposed(enums.ScanTypeDomain.String(), enums.ScanStatusPending.String(), "third-party-pentest")))
}
