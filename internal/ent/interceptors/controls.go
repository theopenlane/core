package interceptors

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// InterceptorTrustCenterControl is middleware that filters control queries based on user context:
// - anonymous trust center users only see publicly visible trust center controls
// - authenticated users with only the trust center module (not compliance) only see trust center controls
func InterceptorTrustCenterControl() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// skip if modules are not enabled
		client := generated.FromContext(ctx)
		if !utils.ModulesEnabled(client) {
			return nil
		}

		// anonymous trust center users can only see controls that are:
		// 1. marked as trust center controls (cloned from the trust center standard)
		// 2. have public visibility
		if _, ok := auth.ActiveTrustCenterIDKey.Get(ctx); ok {
			q.WhereP(
				sql.FieldEQ(control.FieldIsTrustCenterControl, true),
				sql.FieldEQ(control.FieldTrustCenterVisibility, enums.TrustCenterControlVisibilityPubliclyVisible),
			)

			return nil
		}

		// For internal/service/system-admin contexts, skip module-based filtering.
		// These paths are used by hooks/workflow internals and must see all controls.
		if rule.ShouldSkipFeatureCheck(ctx) {
			return nil
		}

		// for authenticated users with only the trust center module (not compliance),
		// limit to trust center controls only
		hasCompliance, _, err := rule.HasAnyFeature(ctx, models.CatalogComplianceModule)
		if err != nil {
			return nil
		}

		if !hasCompliance {
			q.WhereP(sql.FieldEQ(control.FieldIsTrustCenterControl, true))
		}

		return nil
	})
}

// InterceptorControlFieldSort sorts the custom model slice fields
// on controls and subcontrols to ensure consistent order of results
func InterceptorControlFieldSort() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if controlSortSkipType(v) {
				return v, nil
			}

			// determine if the type is Control or Subcontrol
			c := getControlType(v)
			if c != nil {
				// sort the the fields to ensure consistent order
				for i, ctrl := range c {
					c[i].ImplementationGuidance = models.Sort(ctrl.ImplementationGuidance)
					c[i].AssessmentMethods = models.Sort(ctrl.AssessmentMethods)
					c[i].ExampleEvidence = models.Sort(ctrl.ExampleEvidence)
					c[i].AssessmentObjectives = models.Sort(ctrl.AssessmentObjectives)
					c[i].References = models.Sort(ctrl.References)
				}

				return c, nil
			}

			sc := getSubControlType(v)
			if sc != nil {
				// sort the the fields to ensure consistent order
				for i, subctrl := range sc {
					sc[i].ImplementationGuidance = models.Sort(subctrl.ImplementationGuidance)
					sc[i].AssessmentMethods = models.Sort(subctrl.AssessmentMethods)
					sc[i].ExampleEvidence = models.Sort(subctrl.ExampleEvidence)
					sc[i].AssessmentObjectives = models.Sort(subctrl.AssessmentObjectives)
					sc[i].References = models.Sort(subctrl.References)
				}

				return sc, nil
			}

			return v, nil
		})
	})
}

// controlSortSkipType checks if the type is one we should skip sorting for
// this includes primitive types and slices of primitive types
func controlSortSkipType(c any) bool {
	switch c.(type) {
	case []string, string, *string, int, *int:
		return true
	default:
		return false
	}
}

// getControlType attempts to cast the input to a Control or slice of Controls
func getControlType(c any) []*generated.Control {
	switch v := c.(type) {
	case []*generated.Control:
		return v
	case *generated.Control:
		return []*generated.Control{v}
	default:
		return nil
	}
}

// getSubControlType attempts to cast the input to a Subcontrol or slice of Subcontrols
func getSubControlType(c any) []*generated.Subcontrol {
	switch v := c.(type) {
	case []*generated.Subcontrol:
		return v
	case *generated.Subcontrol:
		return []*generated.Subcontrol{v}
	default:
		return nil
	}
}
