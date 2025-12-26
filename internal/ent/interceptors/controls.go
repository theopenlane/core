package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
)

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
