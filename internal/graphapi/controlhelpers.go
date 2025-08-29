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

	var categories []struct {
		Category           string  `json:"category,omitempty"`
		Subcategory        *string `json:"subcategory,omitempty"`
		ReferenceFramework *string `json:"reference_framework,omitempty"`
	}

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

	if err := withTransactionalMutation(ctx).Control.Query().
		Select(fieldName,
			control.FieldReferenceFramework).
		Where(whereFilter).
		GroupBy(fieldName, control.FieldReferenceFramework).Scan(ctx, &categories); err != nil {
		return nil, parseRequestError(err, action{action: ActionGet, object: "categories"})
	}

	if len(categories) == 0 {
		return resp, nil // No categories found
	}

	for _, category := range categories {
		referenceFramework := "Custom"
		if category.ReferenceFramework != nil {
			referenceFramework = *category.ReferenceFramework
		}

		cat := category.Category
		if fieldName == control.FieldSubcategory {
			cat = *category.Subcategory
		}
		resp = append(resp, &model.ControlCategoryEdge{
			Node: &model.ControlCategory{
				Name:               cat,
				ReferenceFramework: &referenceFramework,
			},
		})
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
