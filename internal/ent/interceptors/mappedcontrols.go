package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// InterceptorMappedControls handles returning internal only fields for mapped controls
func InterceptorMappedControls() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.MappedControlFunc(func(ctx context.Context, q *generated.MappedControlQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			mappedControls, ok := v.([]*generated.MappedControl)
			// Skip all query types besides node queries (e.g., Count, Scan, GroupBy).
			if !ok {
				return v, nil
			}

			admin, err := rule.CheckIsSystemAdminWithContext(ctx)
			if err != nil {
				return nil, err
			}

			if admin {
				return mappedControls, nil
			}

			// if not a system admin, remove internal only fields
			for _, mc := range mappedControls {
				mc.SystemInternalID = nil
				mc.InternalNotes = nil
			}

			return mappedControls, nil
		})
	})
}
