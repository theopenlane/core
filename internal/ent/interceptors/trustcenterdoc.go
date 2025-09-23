package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
)

// InterceptorTrustCenterDoc is middleware to change the TrustCenterDoc query
func InterceptorTrustCenterDoc() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		zerolog.Ctx(ctx).Debug().Msg("InterceptorTrustCenterDoc")
		if _, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			// If this is a trust center user, don't show the "not visibible" documents
			q.WhereP(trustcenterdoc.VisibilityNEQ(enums.TrustCenterDocumentVisibilityNotVisible))
		}
		return nil
	})
}
