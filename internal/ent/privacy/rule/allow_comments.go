package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/pkg/slateparser"
)

// CheckIfCommentOnly is a rule that returns allow decision if the mutation is a comment-only operation
func CheckIfCommentOnly() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		if m.Op().Is(ent.OpCreate) {
			return privacy.Skipf("mutation is a create operation, skipping bypass")
		}

		// get the list of added and removed edges and fields in the mutation
		addedEdges := m.AddedEdges()
		removedEdges := m.RemovedEdges()
		fields := m.Fields()
		addedFields := m.AddedFields() // get numeric fields

		ignoreFields := []string{"updated_at", "updated_by", "owner_id"}
		allowedEdges := []string{"comments", "notes"}

		// remove ignored fields from the list of fields being set in the mutation
		fields = lo.Without(fields, ignoreFields...)

		// remove allowed edges from the list of added and removed edges
		addedEdges = lo.Without(addedEdges, allowedEdges...)
		removedEdges = lo.Without(removedEdges, allowedEdges...)

		if len(addedEdges) == 0 && len(removedEdges) == 0 && len(fields) == 0 && len(addedFields) == 0 {
			return privacy.Allowf("mutation has no changes beyond allowed edges, allowing")
		}

		// pick up the matching text/json fields the mutation uses
		// either details or description
		jsonFieldName := "details_json"
		textFieldName := "details"
		if !lo.Contains(fields, jsonFieldName) {
			jsonFieldName = "description_json"
			textFieldName = "description"
		}

		if len(fields) == 1 && lo.Contains(fields, jsonFieldName) {
			oldDetailsJSON, _ := m.OldField(ctx, jsonFieldName)
			newDetailsJSON, _ := m.Field(jsonFieldName)

			oldDetailsTyped, _ := oldDetailsJSON.([]any)
			newDetailsTyped, _ := newDetailsJSON.([]any)

			if slateparser.NoDetailsChanged(oldDetailsTyped, newDetailsTyped) {
				mergedComments, ok := slateparser.MergeComments(oldDetailsTyped, newDetailsTyped)
				if !ok {
					return privacy.Allowf("mutation has only comments added to %s, allowing", jsonFieldName)
				}

				if err := m.SetField(jsonFieldName, mergedComments); err != nil {
					return privacy.Denyf("unable to merge comment markers")
				}

				return privacy.Allowf("mutation has only comments added to %s, allowing", jsonFieldName)
			}

			if len(oldDetailsTyped) == 0 {
				oldText, _ := m.OldField(ctx, textFieldName)
				oldTextTyped, _ := oldText.(string)
				if slateparser.DoesMarkdownMatchSlate(oldTextTyped, newDetailsTyped) {
					return privacy.Allowf("mutation initializes %s with comments without changing %s, allowing", jsonFieldName, textFieldName)
				}
			}
		}

		// if we reach here, changes are beyond scope of comments and we should fall to next rule
		return privacy.Skipf("mutation has changes, skipping")
	})
}
