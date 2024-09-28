package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/entx"
)

// isDeleteOp checks if the mutation is a deletion operation.
// which includes soft delete, delete, and delete one.
func isDeleteOp(ctx context.Context, m ent.Mutation) bool {
	return entx.CheckIsSoftDelete(ctx) || m.Op().Is(ent.OpDelete) || m.Op().Is(ent.OpDeleteOne)
}
