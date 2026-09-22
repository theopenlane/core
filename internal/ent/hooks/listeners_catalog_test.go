package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func TestRegisterGalaCatalogListeners(t *testing.T) {
	t.Parallel()

	runtime, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 3})
	assert.NilError(t, err)

	ids, err := gala.Register(runtime, CatalogListeners()...)
	assert.NilError(t, err)
	assert.Equal(t, len(ids), 1)

	topic := entityops.MutationTopicName(entityops.MutationConcernDirect, generated.TypeEntity)
	assert.Check(t, runtime.InterestedIn(topic, ent.OpUpdate.String()))
	assert.Check(t, runtime.InterestedIn(topic, ent.OpUpdateOne.String()))
	assert.Check(t, !runtime.InterestedIn(topic, ent.OpCreate.String()))
	assert.Check(t, !runtime.InterestedIn(topic, ent.OpDelete.String()))
}

func TestCatalogListenerGate(t *testing.T) {
	t.Parallel()

	listener, ok := CatalogListeners()[0].(entityops.MutationListener)
	assert.Check(t, ok)

	gate := listener.Definition().Gate

	changed := func(field string, value any) entityops.MutationPayload {
		return entityops.MutationPayload{
			MutationType: generated.TypeEntity,
			Operation:    entityops.OpUpdateOne,
			EntityID:     "entity-1",
			ChangeSet: entityops.ChangeSet{
				ChangedFields:   []string{field},
				ProposedChanges: map[string]any{field: value},
			},
		}
	}

	assert.Check(t, gate(context.Background(), changed(entity.FieldDisplayName, "Stripe")))
	assert.Check(t, gate(context.Background(), changed(entity.FieldDomains, []string{"stripe.com"})))
	assert.Check(t, !gate(context.Background(), changed(entity.FieldTags, []string{"payments"})))
}
