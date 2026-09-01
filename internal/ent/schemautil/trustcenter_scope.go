package schemautil

import (
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/logx"
)

// TrustCenterScopePredicate returns a predicate that scopes trust center joins
// and filters out correctly using the trustcenter_id
func TrustCenterScopePredicate() func(*sql.Selector) {
	return func(s *sql.Selector) {
		ctx := s.Context()

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			logx.FromContext(ctx).Debug().Msg("could not fetch caller when scoping trustcenter")
			return
		}

		if caller.Has(auth.CapSystemAdmin) {
			return
		}

		if tcID, _, ok := auth.TrustCenterScopeFromContext(ctx); ok {
			s.Where(sql.EQ(s.C("trust_center_id"), tcID))
			return
		}

		orgIDs := caller.OrgIDs()

		t := sql.Table(trustcenter.Table)

		anys := make([]any, len(orgIDs))
		for i, id := range orgIDs {
			anys[i] = id
		}

		s.Where(
			sql.In(
				s.C("trust_center_id"),
				sql.Select(t.C(trustcenter.FieldID)).From(t).Where(
					sql.In(
						t.C(trustcenter.FieldOwnerID), anys...,
					),
				),
			),
		)
	}
}
