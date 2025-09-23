package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/labstack/gommon/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
)

// InterceptorControlFieldSort is middleware to change the control field sort query
func InterceptorControlFieldSort() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, q ent.Query) (ent.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			c := getControlType(v)
			if c != nil {
				// sort the the fields to ensure consistent order
				for i, ctrl := range c {
					c[i].ImplementationGuidance = models.Sort(ctrl.ImplementationGuidance)
					c[i].AssessmentMethods = models.Sort(ctrl.AssessmentMethods)
					c[i].ExampleEvidence = models.Sort(ctrl.ExampleEvidence)
					c[i].AssessmentObjectives = models.Sort(ctrl.AssessmentObjectives)
				}
			} else {
				sc := getSubControlType(v)
				if sc != nil {
					// sort the the fields to ensure consistent order
					for i, subctrl := range sc {
						sc[i].ImplementationGuidance = models.Sort(subctrl.ImplementationGuidance)
						sc[i].AssessmentMethods = models.Sort(subctrl.AssessmentMethods)
						sc[i].ExampleEvidence = models.Sort(subctrl.ExampleEvidence)
						sc[i].AssessmentObjectives = models.Sort(subctrl.AssessmentObjectives)
					}
				} else {
					log.Warn("InterceptorControlFieldSort: could not determine type for sorting")
				}
			}

			return c, nil
		})
	})
}

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
