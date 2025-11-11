package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

func HookCreateAssessmentResponse() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.AssessmentResponseFunc(func(ctx context.Context, m *generated.AssessmentResponseMutation) (generated.Value, error) {

			m.ClearDocumentDataID()

			retVal, err := next.Mutate(ctx, m)

			return retVal, err
		})
	}, ent.OpCreate)
}
