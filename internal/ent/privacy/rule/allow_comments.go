package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slateparser"
)

// CheckIfCommentOnly is a rule that returns allow decision if the mutation is a comment-only operation
func CheckIfCommentOnly() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		if m.Op().Is(ent.OpCreate) {
			logx.FromContext(ctx).Warn().Msg("mutation is a create operation, allowing for comments")
			return privacy.Skipf("mutation is a create operation, skipping bypass")
		}

		// get the list of added and removed edges and fields in the mutation
		addedEdges := m.AddedEdges()
		removedEdges := m.RemovedEdges()
		fields := m.Fields()
		addedFields := m.AddedFields() // get numeric fields

		ignoreFields := []string{"updated_at", "updated_by", "owner_id"}
		allowedEdges := []string{"comments"}

		// remove ignored fields from the list of fields being set in the mutation
		fields = lo.Without(fields, ignoreFields...)

		// remove allowed edges from the list of added and removed edges
		addedEdges = lo.Without(addedEdges, allowedEdges...)
		removedEdges = lo.Without(removedEdges, allowedEdges...)

		if len(addedEdges) == 0 && len(removedEdges) == 0 && len(fields) == 0 && len(addedFields) == 0 {
			logx.FromContext(ctx).Warn().Strs("fields", fields).Strs("added_edges", addedEdges).Strs("removed_edges", removedEdges).Strs("added_fields", addedFields).Msg("mutation has no changes beyond allowed edges and fields, allowing for comments")
			return privacy.Allowf("mutation has no changes beyond allowed edges, allowing")
		}

		detailsJSONFieldName := "details_json"
		if !lo.Contains(fields, detailsJSONFieldName) {
			// try description_json instead
			detailsJSONFieldName = "description_json"
		}

		// if just one fields changed, check the details_json, this is done for plate comments
		if len(fields) == 1 && lo.Contains(fields, detailsJSONFieldName) {
			// get the old details json value from the mutation
			oldDetailsJSON, _ := m.OldField(ctx, detailsJSONFieldName)
			newDetailsJSON, _ := m.Field(detailsJSONFieldName)

			oldDetailsTyped, _ := oldDetailsJSON.([]any)
			newDetailsTyped, _ := newDetailsJSON.([]any)

			if slateparser.OnlyCommentsAdded(oldDetailsTyped, newDetailsTyped) {
				logx.FromContext(ctx).Warn().Strs("fields", fields).Strs("added_edges", addedEdges).Strs("removed_edges", removedEdges).Strs("added_fields", addedFields).Msg("mutation has only comments added to details_json, allowing")
				return privacy.Allowf("mutation has only comments added to details_json, allowing")
			}
		}

		logx.FromContext(ctx).Warn().Strs("fields", fields).Strs("added_edges", addedEdges).Strs("removed_edges", removedEdges).Strs("added_fields", addedFields).Msg("mutation has changes beyond allowed edges and fields, skipping")
		// if we reach here, changes are beyond scope of comments and we should fall to next rule
		return privacy.Skipf("mutation has changes, skipping")
	})
}
