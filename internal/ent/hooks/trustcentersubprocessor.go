package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/fgax"
)

// HookTrustCenterSubprocessor process files for trust center subprocessors
func HookTrustCenterSubprocessor() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSubprocessorFunc(func(ctx context.Context, m *generated.TrustCenterSubprocessorMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center subprocessor hook")

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// get the ids of the subprocessor and tc subprocessor
			tcSubprocessor, ok := retVal.(*generated.TrustCenterSubprocessor)
			if !ok {
				return retVal, nil
			}

			tcSpID := tcSubprocessor.ID

			subprocessorID, ok := m.SubprocessorID()
			if !ok {
				return retVal, nil
			}

			/// add parent trust center subprocessor relation
			req := fgax.TupleRequest{
				SubjectID:   tcSpID,
				SubjectType: strcase.SnakeCase(m.Type()),
				ObjectID:    subprocessorID,
				ObjectType:  strcase.SnakeCase(generated.TypeSubprocessor),
				Relation:    fgax.ParentRelation,
			}

			w := fgax.GetTupleKey(req)
			if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{w}, nil); err != nil {
				log.Error().Err(err).Interface("writes", w).Msg("failed to create relationship tuples for trust center subprocessor parent")

				return nil, err
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}
