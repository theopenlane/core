package graphapi

import (
	"cmp"
	"context"
	"slices"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/graphapi/model"
)

const (
	noCategoryLabel = "No Category"
)

// getAllCategories retrieves all control categories based on the provided field name and where filter
// It returns a slice of ControlCategoryEdge containing the categories and their reference frameworks.
// If the field name is not provided, it defaults to control.FieldCategory.
func (r *queryResolver) getAllCategories(ctx context.Context, fieldName string, where *generated.ControlWhereInput) ([]*model.ControlCategoryEdge, error) {
	// fallback to default field name if not provided
	if fieldName == "" {
		fieldName = control.FieldCategory
	}

	resp := []*model.ControlCategoryEdge{}

	whereP, err := getControlWherePredicate(where)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "categories"})
	}

	whereFilter := control.CategoryNEQ("")
	if fieldName == control.FieldSubcategory {
		whereFilter = control.SubcategoryNEQ("")
	}

	if whereP != nil {
		whereFilter = control.And(
			append([]predicate.Control{whereFilter}, whereP)...,
		)
	}

	res, err := withTransactionalMutation(ctx).Control.Query().
		Select(fieldName,
			control.FieldReferenceFramework).
		Where(whereFilter).All(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "categories"})
	}

	tmp := map[string]map[string]bool{}
	for _, r := range res {
		refFramework := "Custom"
		if r.ReferenceFramework != nil {
			refFramework = *r.ReferenceFramework
		}

		if _, ok := tmp[refFramework]; !ok {
			tmp[refFramework] = map[string]bool{
				r.Category: true,
			}
		} else {
			if fieldName == control.FieldCategory && r.Category == "" {
				if _, ok := tmp[refFramework][r.Category]; !ok {
					tmp[refFramework][r.Category] = true
				}
			} else if fieldName == control.FieldSubcategory && r.Subcategory == "" {
				if _, ok := tmp[refFramework][r.Subcategory]; !ok {
					tmp[refFramework][r.Subcategory] = true
				}
			}
		}
	}

	if len(tmp) == 0 {
		return resp, nil // No categories found
	}

	for refFramework, categories := range tmp {
		for cat := range categories {
			resp = append(resp, &model.ControlCategoryEdge{
				Node: &model.ControlCategory{
					Name:               cat,
					ReferenceFramework: &refFramework,
				},
			})
		}
	}

	// sort the categories to ensure consistent order
	slices.SortFunc(resp, func(a, b *model.ControlCategoryEdge) int {
		return cmp.Compare(a.Node.Name, b.Node.Name)
	})

	return resp, nil
}

// getControlWherePredicate extracts the predicate from the ControlWhereInput.
// It returns nil if the where input is nil or if there are no predicates.
func getControlWherePredicate(where *generated.ControlWhereInput) (predicate.Control, error) {
	if where == nil {
		return nil, nil
	}

	whereP, err := where.P()
	if err != nil {
		return nil, err
	}

	if whereP == nil {
		return nil, nil
	}

	return whereP, nil
}
