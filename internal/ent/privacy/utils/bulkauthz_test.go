package utils

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func testClient() *generated.Client {
	return generated.NewClient(generated.EntConfig(&entconfig.Config{}))
}

func controlTuple(id string) fgax.TupleKey {
	return fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   "org-1",
		SubjectType: "organization",
		ObjectID:    id,
		ObjectType:  "control",
		Relation:    fgax.ParentRelation,
	})
}

func TestDeferTuplesWithoutBatcher(t *testing.T) {
	m := testClient().Control.Create().Mutation()

	assert.Check(t, !DeferTuples(context.Background(), m, []fgax.TupleKey{controlTuple("ctl-1")}, nil))
}

func TestDeferTuplesOnlyMatchesConfiguredCreates(t *testing.T) {
	tests := map[string]struct {
		mutation ent.Mutation
		deferred bool
	}{
		"control create": {testClient().Control.Create().Mutation(), true},
		"control update": {testClient().Control.UpdateOneID("ctl-1").Mutation(), false},
		"nested create":  {testClient().CustomTypeEnum.Create().Mutation(), false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, _ := WithBulkAuthz(context.Background(), "Control")

			assert.Check(t, is.Equal(tc.deferred, DeferTuples(ctx, tc.mutation, []fgax.TupleKey{controlTuple("ctl-1")}, nil)))
		})
	}
}

func TestDeferTuplesRemovesDuplicates(t *testing.T) {
	ctx, batch := WithBulkAuthz(context.Background(), "Control")
	m := testClient().Control.Create().Mutation()

	assert.Check(t, DeferTuples(ctx, m, []fgax.TupleKey{controlTuple("ctl-1"), controlTuple("ctl-2")}, nil))
	assert.Check(t, DeferTuples(ctx, m, []fgax.TupleKey{controlTuple("ctl-1")}, nil))

	assert.Check(t, is.Len(batch.writes, 2))
}

func TestFlushWithNothingQueuedSkipsAuthz(t *testing.T) {
	ctx, batch := WithBulkAuthz(context.Background(), "Control")

	assert.NilError(t, batch.Flush(ctx))
}
