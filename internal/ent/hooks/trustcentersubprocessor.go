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

// HookTrustCenterSubprocessor adds parent relationship tuples on create of trust center subprocessors for the subprocessor, allowing trust center access
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
				logx.FromContext(ctx).Error().Msg("unexpected type for trust center subprocessor mutation return value")

				return retVal, nil
			}

			tcSpID := tcSubprocessor.ID

			subprocessorID, ok := m.SubprocessorID()
			if !ok {
				logx.FromContext(ctx).Error().Msg("missing subprocessor id in trust center subprocessor mutation")
				return retVal, nil
			}

			// add parent trust center subprocessor relation
			req := fgax.TupleRequest{
				SubjectID:   tcSpID,
				SubjectType: strcase.SnakeCase(m.Type()),
				ObjectID:    subprocessorID,
				ObjectType:  strcase.SnakeCase(generated.TypeSubprocessor),
				Relation:    fgax.ParentRelation,
			}

			if _, err := m.Authz.WriteTupleKeys(ctx, []fgax.TupleKey{fgax.GetTupleKey(req)}, nil); err != nil {
				log.Error().Err(err).Interface("writes", req).Msg("failed to create relationship tuples for trust center subprocessor parent")

				return nil, err
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}
