package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/entx"
)

// isDeleteOp checks if the mutation is a deletion operation.
// which includes soft delete, delete, and delete one.
func isDeleteOp(ctx context.Context, m ent.Mutation) bool {
	return entx.CheckIsSoftDelete(ctx) || m.Op().Is(ent.OpDelete) || m.Op().Is(ent.OpDeleteOne)
}

var runtimeHooks []func(ent.Mutator) ent.Mutator

func AddPostMutationHook[T any](hook func(ctx context.Context, v T) error) {
	runtimeHooks = append(runtimeHooks, func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			entvalue, ok := v.(T)

			if ok {
				err2 := hook(ctx, entvalue)
				if err2 != nil {
					log.Debug().Ctx(ctx).Err(err2).Msg("post mutation hook error")
				}
			}

			return v, err
		})
	})
}

func AddPreMutationHook[T any](hook func(ctx context.Context, v T) error) {
	runtimeHooks = append(runtimeHooks, func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			entvalue, ok := m.(T)
			if ok {
				err := hook(ctx, entvalue)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	})
}

func AddMutationHook(hook ent.Hook) {
	runtimeHooks = append(runtimeHooks, func(next ent.Mutator) ent.Mutator {
		return hook(next)
	})
}

func AddMutationHooks(client *entgen.Client) {
	for _, hook := range runtimeHooks {
		client.Use(hook)
	}
}
