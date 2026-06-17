package graphapi

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
)

const (
	noCategoryLabel = "No Category"
	customFramework = "CUSTOM"
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
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "categories"})
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
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "categories"})
	}

	tmp := map[string]map[string]bool{}
	for _, r := range res {
		refFramework := getFrameworkName(r)

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
			return nil, fmt.Errorf("%w: expected format 'standard_short_name::control_ref_code'", common.ErrInvalidInput)
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
	caller, _ := auth.CallerFromContext(ctx)
	systemOwned := caller != nil && caller.Has(auth.CapSystemAdmin)

	for standardShortName, refCodes := range standardRefCodes {
		switch any(*new(T)).(type) {
		case predicate.Control:
			if strings.EqualFold(standardShortName, customFramework) {
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
			if strings.EqualFold(standardShortName, customFramework) {
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

// getFrameworkName returns the string form of the reference framework
func getFrameworkName(c *generated.Control) string {
	return normalizeFramework(c.ReferenceFramework)
}

// normalizeFramework returns the string form of the reference framework
func normalizeFramework(framework *string) string {
	if framework != nil {
		return *framework
	}

	return customFramework
}

// getMappedControlsBySubcontrolID returns mapped control records that include the given subcontrol ID on either side
func getMappedControlsBySubcontrolID(ctx context.Context, subcontrolID string) ([]*generated.MappedControl, error) {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ownership := subcontrol.Or(
		subcontrol.SystemOwned(true),
		subcontrol.OwnerIDIn(orgIDs...),
	)

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return withTransactionalMutation(ctx).MappedControl.Query().
		Where(
			mappedcontrol.Or(
				mappedcontrol.HasFromSubcontrolsWith(subcontrol.ID(subcontrolID), ownership),
				mappedcontrol.HasToSubcontrolsWith(subcontrol.ID(subcontrolID), ownership),
			),
		).
		WithFromControls().WithToControls().WithFromSubcontrols().WithToSubcontrols().
		All(allowCtx)
}

// getControlMappings returns the controls and subcontrols mapped to a control based on the ref code and framework
func getControlMappings(ctx context.Context, refCode string, framework *string, parentControlID *string) ([]*generated.MappedControl, error) {
	fullWhere, err := prepMappedControlQuery(ctx, refCode, framework, parentControlID)
	if err != nil {
		return nil, err
	}

	// skip filters, this is already filtered on organization
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	res, err := withTransactionalMutation(ctx).MappedControl.Query().
		Where(
			fullWhere...,
		).WithFromControls().WithToControls().WithFromSubcontrols().WithToSubcontrols().All(allowCtx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// prepMappedControlQuery gets the predicate for the mapped control query
func prepMappedControlQuery(ctx context.Context, refCode string, framework *string, parentControlID *string) ([]predicate.MappedControl, error) {
	// get orgs to filter, this will allow us to skip expensive authz checks
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// controls have no parent control, whereas subcontrols do
	if parentControlID == nil {
		controlWhere := []predicate.Control{
			control.RefCode(refCode),
			control.Or(
				control.SystemOwned(true),
				control.OwnerIDIn(orgIDs...),
			),
		}

		if framework == nil {
			controlWhere = append(controlWhere, control.ReferenceFrameworkIsNil())
		} else {
			controlWhere = append(controlWhere, control.ReferenceFramework(*framework))
		}

		return []predicate.MappedControl{
			mappedcontrol.Or(
				mappedcontrol.HasToControlsWith(
					controlWhere...,
				),
				mappedcontrol.HasFromControlsWith(
					controlWhere...,
				),
			),
		}, nil
	}

	subControlWhere := []predicate.Subcontrol{
		subcontrol.RefCode(refCode),
		subcontrol.ControlID(*parentControlID),
		subcontrol.Or(
			subcontrol.SystemOwned(true),
			subcontrol.OwnerIDIn(orgIDs...),
		),
	}

	return []predicate.MappedControl{
		mappedcontrol.Or(
			mappedcontrol.HasToSubcontrolsWith(
				subControlWhere...,
			),
			mappedcontrol.HasFromSubcontrolsWith(
				subControlWhere...,
			),
		),
	}, nil
}

func getOrgMappedControlsInfo(ctx context.Context, controls map[string]*model.ControlInfo, refCode string, framework *string) []*model.ControlInfo {
	if len(controls) == 0 {
		return nil
	}

	filteredControls := make(map[string]*model.ControlInfo, len(controls))
	for k, c := range controls {
		if isSameControlInfo(refCode, framework, c) {
			continue
		}
		filteredControls[k] = c
	}

	if len(filteredControls) == 0 {
		return nil
	}

	orgControls, ok := findOrganizationControlInfoForMappings(ctx, filteredControls)
	if !ok {
		return nil
	}

	return orgControls

}

// findOrganizationControlInfoForMappings
func findOrganizationControlInfoForMappings(ctx context.Context, controls map[string]*model.ControlInfo) ([]*model.ControlInfo, bool) {
	// get orgs to filter, this will allow us to skip expensive authz checks
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, false
	}

	controlRefCodes := map[string][]string{}
	subcontrolRefCodes := map[string][]string{}

	results := []*model.ControlInfo{}

	for _, c := range controls {
		if c.ReferenceFramework == nil {
			logx.FromContext(ctx).Warn().Str("id", c.ID).Str("ref_code", c.RefCode).Msg("found system control without a reference framework")

			continue
		}

		fw := normalizeFramework(c.ReferenceFramework)
		if c.IsSubcontrol {
			subcontrolRefCodes[fw] = append(subcontrolRefCodes[fw], c.RefCode)
		} else {
			controlRefCodes[fw] = append(controlRefCodes[fw], c.RefCode)
		}

	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	if len(subcontrolRefCodes) > 0 {
		orClauses := make([]predicate.Subcontrol, 0, len(subcontrolRefCodes))
		for fw, refCodes := range subcontrolRefCodes {
			orClauses = append(orClauses, subcontrol.And(
				subcontrol.ReferenceFrameworkEQ(fw),
				subcontrol.RefCodeIn(refCodes...),
			))
		}

		mappedFromSystem, err := withTransactionalMutation(ctx).Subcontrol.Query().Where(
			subcontrol.SystemOwned(false),
			subcontrol.OwnerIDIn(orgIDs...),
			subcontrol.Or(orClauses...),
		).All(allowCtx)
		if err != nil {
			// only log errors that are not because it does not exist
			if !generated.IsNotFound(err) {
				logx.FromContext(ctx).Error().Err(err).Msg("error getting mapped subcontrol")
			}
			return nil, false
		}

		for _, sc := range mappedFromSystem {
			results = append(results, subcontrolEdgeToControlInfo(sc))
		}

	}

	if len(controlRefCodes) > 0 {
		orClauses := make([]predicate.Control, 0, len(subcontrolRefCodes))
		for fw, refCodes := range controlRefCodes {
			orClauses = append(orClauses, control.And(
				control.ReferenceFrameworkEQ(fw),
				control.RefCodeIn(refCodes...),
			))
		}

		mappedFromSystem, err := withTransactionalMutation(ctx).Control.Query().
			Where(
				control.SystemOwned(false),
				control.OwnerIDIn(orgIDs...),
				control.Or(orClauses...),
			).All(allowCtx)
		if err != nil {
			// only log errors that are not because it does not exist
			if !generated.IsNotFound(err) {
				logx.FromContext(ctx).Error().Err(err).Msg("error getting mapped control")
			}
			return nil, false
		}
		for _, sc := range mappedFromSystem {
			results = append(results, controlEdgeToControlInfo(sc))
		}
	}

	return results, len(results) > 0
}

func isSameControlInfo(refCode string, framework *string, mappedControl *model.ControlInfo) bool {
	if refCode != mappedControl.RefCode {
		return false
	}

	// ref codes match, and both are custom framework, same control
	if framework == nil && mappedControl.ReferenceFramework == nil {
		return true
	}

	mappedFramework := normalizeFramework(mappedControl.ReferenceFramework)
	currentFramework := normalizeFramework(framework)

	return currentFramework == mappedFramework
}

func generateMapControlKey(refCode string, framework *string) string {
	f := normalizeFramework(framework)

	return fmt.Sprintf("%s::%s", refCode, f)
}

// getStandardsInOrg gets unique frameworks in the organization
func getStandardsInOrg(ctx context.Context) ([]string, error) {
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return withTransactionalMutation(ctx).Control.Query().
		Where(
			control.SystemOwned(false),
			control.OwnerIDIn(orgIDs...),
			control.ReferenceFrameworkNotNil(),
		).
		Unique(true).Select(control.FieldReferenceFramework).Strings(allowCtx)
}
