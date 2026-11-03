package graphapi

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func (r *mutationResolver) validateTrustcenterAccess(ctx context.Context, trustCenter *generated.TrustCenter) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller.IsAnonymous() {
		return auth.ErrNoAuthUser
	}

	if caller.Has(auth.CapSystemAdmin) {
		return nil
	}

	if caller.Has(auth.CapOrgSupport) && caller.OrganizationID == trustCenter.OwnerID {
		return nil
	}

	ok, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		Relation:    fgax.CanEdit,
		ObjectType:  "trust_center",
		ObjectID:    trustCenter.ID,
		SubjectType: caller.SubjectType(),
		SubjectID:   caller.SubjectID,
	})
	if err != nil {
		return err
	}

	if !ok {
		logx.FromContext(ctx).Warn().Msg("access denied to create a trust center preview domain")
		return newPermissionDeniedError()
	}

	return nil
}
