package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/iam/auth"
)

// InterceptorTrustCenter is middleware to change the TrustCenter query
func InterceptorTrustCenter() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			q.WhereP(trustcenter.IDEQ(anon.TrustCenterID))
		}
		return nil
	})
}
