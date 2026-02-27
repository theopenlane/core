package rule

import (
	"context"
	"maps"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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

		ignore_fields := []string{"updated_at", "updated_by", "owner_id"}
		allowed_edges := []string{"comments"}

		// remove ignored fields from the list of fields being set in the mutation
		fields = lo.Without(fields, ignore_fields...)

		// remove allowed edges from the list of added and removed edges
		addedEdges = lo.Without(addedEdges, allowed_edges...)
		removedEdges = lo.Without(removedEdges, allowed_edges...)

		if len(addedEdges) == 0 && len(removedEdges) == 0 && len(fields) == 0 && len(addedFields) == 0 {
			return privacy.Allowf("mutation has no changes beyond allowed edges, allowing")
		}

		// if just one fields changed, check the details_json, this is done for plate comments
		if len(fields) == 1 && lo.Contains(fields, "details_json") {
			// get the old details json value from the mutation
			oldDetailsJSON, _ := m.OldField(ctx, "details_json")
			newDetailsJSON, _ := m.Field("details_json")

			oldDetailsMap, ok := oldDetailsJSON.(map[string]any)
			if !ok {
				return privacy.Skipf("old details json is not a map, skipping")
			}

			newDetailsMap, ok := newDetailsJSON.(map[string]any)
			if !ok {
				return privacy.Skipf("new details json is not a map, skipping")
			}

			if maps.Equal(oldDetailsMap, newDetailsMap) {
				return privacy.Allowf("mutation has no changes, skipping")
			}
		}

		// if we reach here, changes are beyond scope of comments and we should fall to next rule
		return privacy.Skipf("mutation has changes, skipping")
	})
}
