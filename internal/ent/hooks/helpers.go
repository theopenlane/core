package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/entx"
)

// isDeleteOp checks if the mutation is a deletion operation.
// which includes soft delete, delete, and delete one.
func isDeleteOp(ctx context.Context, mutation ent.Mutation) bool {
	return entx.CheckIsSoftDelete(ctx) || mutation.Op().Is(ent.OpDelete) || mutation.Op().Is(ent.OpDeleteOne)
}
