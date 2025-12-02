package graphapi

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/control"
	"github.com/theopenlane/ent/generated/subcontrol"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/ent/generated/predicate"
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
		return nil, parseRequestError(ctx, err, action{action: ActionGet, object: "categories"})
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
		return nil, parseRequestError(ctx, err, action{action: ActionGet, object: "categories"})
	}

	tmp := map[string]map[string]bool{}
	for _, r := range res {
		refFramework := "Custom"
		if r.ReferenceFramework != nil {
			refFramework = *r.ReferenceFramework
		}

		if _, ok := tmp[refFramework]; !ok {
			if fieldName == control.FieldCategory && r.Category != "" {
				tmp[refFramework] = map[string]bool{
					r.Category: true,
				}
			} else if fieldName == control.FieldSubcategory && r.Subcategory != "" {
				tmp[refFramework] = map[string]bool{
					r.Subcategory: true,
				}
			}
		} else {
			if fieldName == control.FieldCategory && r.Category != "" {
				if _, ok := tmp[refFramework][r.Category]; !ok {
					tmp[refFramework][r.Category] = true
				}
			} else if fieldName == control.FieldSubcategory && r.Subcategory != "" {
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

// getStandardRefCodes processes a slice of strings formatted as "standard_short_name::control_ref_code"
// and returns a map where the keys are standard short names and the values are slices of control reference codes.
// If the input slice is empty, it returns nil.
func getStandardRefCodes(data []string) (map[string][]string, error) {
	if len(data) == 0 {
		return nil, nil
	}

	result := make(map[string][]string)

	for _, controlData := range data {
		parts := strings.Split(controlData, "::")
		if len(parts) != 2 { //nolint:mnd
			return nil, fmt.Errorf("%w: expected format 'standard_short_name::control_ref_code'", ErrInvalidInput)
		}

		standardShortName := parts[0]
		controlRefCode := parts[1]

		// add the mapping to the result
		if _, ok := result[standardShortName]; !ok {
			result[standardShortName] = []string{}
		}

		result[standardShortName] = append(result[standardShortName], controlRefCode)
	}

	return result, nil
}

// constructWherePredicatesFromStandardRefCodes constructs a slice of predicates based on the provided
// standard reference codes map. The map keys are standard short names and the values are slices of control reference codes.
// It returns a slice of predicates that can be used in queries to filter controls or subcontrols.
func constructWherePredicatesFromStandardRefCodes[T predicate.Control | predicate.Subcontrol](ctx context.Context, standardRefCodes map[string][]string) []T {
	predicates := []T{}

	// use to determine if we should filter by system owned controls
	systemOwned := auth.IsSystemAdminFromContext(ctx)

	for standardShortName, refCodes := range standardRefCodes {
		switch any(*new(T)).(type) {
		case predicate.Control:
			if standardShortName == "CUSTOM" {
				predicates = append(predicates, any(control.And(
					control.ReferenceFrameworkIsNil(),
					control.RefCodeIn(refCodes...),
				)).(T))
			} else {
				predicates = append(predicates, any(control.And(
					control.ReferenceFrameworkEqualFold(standardShortName),
					control.RefCodeIn(refCodes...),
				)).(T))
			}

		case predicate.Subcontrol:
			if standardShortName == "CUSTOM" {
				predicates = append(predicates, any(subcontrol.And(
					subcontrol.ReferenceFrameworkIsNil(),
					subcontrol.RefCodeIn(refCodes...),
				)).(T))
			} else {
				predicates = append(predicates, any(subcontrol.And(
					subcontrol.ReferenceFrameworkEqualFold(standardShortName),
					subcontrol.RefCodeIn(refCodes...),
				)).(T))
			}

		}
	}

	// wrap all predicates in an OR clause
	if len(predicates) > 1 {
		switch any(*new(T)).(type) {
		case predicate.Control:
			orPredicates := make([]predicate.Control, len(predicates))
			for i, p := range predicates {
				orPredicates[i] = any(p).(predicate.Control)
			}
			predicates = []T{any(control.Or(orPredicates...)).(T)}
		case predicate.Subcontrol:
			orPredicates := make([]predicate.Subcontrol, len(predicates))
			for i, p := range predicates {
				orPredicates[i] = any(p).(predicate.Subcontrol)
			}
			predicates = []T{any(subcontrol.Or(orPredicates...)).(T)}
		}
	}

	switch any(*new(T)).(type) {
	case predicate.Control:
		predicates = append(predicates, any(
			control.SystemOwned(systemOwned),
		).(T))
	case predicate.Subcontrol:
		predicates = append(predicates, any(
			subcontrol.SystemOwned(systemOwned),
		).(T))
	}

	return predicates
}

// getControlIDsFromRefCodes retrieves control or subcontrol IDs based on the provided reference codes which include standard short names
// for example "ISO27001::A.5.1.1"
func getControlIDsFromRefCodes[T predicate.Control | predicate.Subcontrol](ctx context.Context, data []string) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}

	mappedData, err := getStandardRefCodes(data)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get standard ref codes")

		return nil, err
	}

	where := constructWherePredicatesFromStandardRefCodes[T](ctx, mappedData)

	switch any(*new(T)).(type) {
	case predicate.Control:
		whereControl := make([]predicate.Control, len(where))
		for i, w := range where {
			whereControl[i] = any(w).(predicate.Control)
		}
		return withTransactionalMutation(ctx).Control.Query().Where(whereControl...).IDs(ctx)
	case predicate.Subcontrol:
		whereSubcontrol := make([]predicate.Subcontrol, len(where))
		for i, w := range where {
			whereSubcontrol[i] = any(w).(predicate.Subcontrol)
		}
		return withTransactionalMutation(ctx).Subcontrol.Query().Where(whereSubcontrol...).IDs(ctx)
	}

	return nil, nil
}
