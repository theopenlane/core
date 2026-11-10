package utils

import (
	"context"
	"sync"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"
)

// bulkAuthzKey holds the request-scoped FGA batcher
var bulkAuthzKey = contextx.NewKey[*BulkAuthz]()

// bulkCheckKey identifies an access check that has no request-specific context
type bulkCheckKey struct {
	objectType  fgax.Kind
	objectID    string
	subjectID   string
	subjectType string
	relation    string
}

// BulkAuthz reduces FGA round trips during a bulk create of a single object type
type BulkAuthz struct {
	objectType string
	mu         sync.Mutex
	checks     map[bulkCheckKey]bool
	writes     map[fgax.TupleKey]struct{}
	deletes    map[fgax.TupleKey]struct{}
}

// WithBulkAuthz adds a batcher for creates of objectType to the context
func WithBulkAuthz(ctx context.Context, objectType string) (context.Context, *BulkAuthz) {
	batch := &BulkAuthz{
		objectType: objectType,
		checks:     map[bulkCheckKey]bool{},
		writes:     map[fgax.TupleKey]struct{}{},
		deletes:    map[fgax.TupleKey]struct{}{},
	}

	return bulkAuthzKey.Set(ctx, batch), batch
}

// DeferTuples queues tuples for the batch flush and reports whether they were queued
func DeferTuples(ctx context.Context, m ent.Mutation, writes, deletes []fgax.TupleKey) bool {
	batch, ok := bulkAuthzKey.Get(ctx)
	if !ok || m.Type() != batch.objectType || !m.Op().Is(ent.OpCreate) {
		return false
	}

	batch.mu.Lock()
	defer batch.mu.Unlock()

	for _, tuple := range writes {
		batch.writes[tuple] = struct{}{}
	}

	for _, tuple := range deletes {
		batch.deletes[tuple] = struct{}{}
	}

	return true
}

// Flush writes the queued tuples, the keys are unique because FGA rejects duplicates in one request
func (b *BulkAuthz) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.writes) == 0 && len(b.deletes) == 0 {
		return nil
	}

	writes := make([]fgax.TupleKey, 0, len(b.writes))
	for tuple := range b.writes {
		writes = append(writes, tuple)
	}

	deletes := make([]fgax.TupleKey, 0, len(b.deletes))
	for tuple := range b.deletes {
		deletes = append(deletes, tuple)
	}

	if _, err := AuthzClientFromContext(ctx).WriteTupleKeys(ctx, writes, deletes); err != nil {
		return err
	}

	clear(b.writes)
	clear(b.deletes)

	return nil
}

// CheckAccess runs an access check once per distinct check when a batcher is in the context
func CheckAccess(ctx context.Context, client *fgax.Client, ac fgax.AccessCheck) (bool, error) {
	batch, ok := bulkAuthzKey.Get(ctx)
	if !ok || ac.Context != nil || len(ac.ContextualTuples) > 0 {
		return client.CheckAccess(ctx, ac)
	}

	key := bulkCheckKey{
		objectType:  ac.ObjectType,
		objectID:    ac.ObjectID,
		subjectID:   ac.SubjectID,
		subjectType: ac.SubjectType,
		relation:    ac.Relation,
	}

	batch.mu.Lock()
	defer batch.mu.Unlock()

	if access, found := batch.checks[key]; found {
		return access, nil
	}

	access, err := client.CheckAccess(ctx, ac)
	if err != nil {
		return false, err
	}

	batch.checks[key] = access

	return access, nil
}
