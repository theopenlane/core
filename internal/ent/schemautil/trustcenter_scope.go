package schemautil

import (
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/logx"
)

// TrustCenterScopePredicate returns a predicate that scopes trust center joins
// and filters out correctly using the trustcenter_id
func TrustCenterScopePredicate() func(*sql.Selector) {
	return func(s *sql.Selector) {
		ctx := s.Context()

		if auth.IsSystemAdminFromContext(ctx) {
			return
		}

		if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			if anon.TrustCenterID != "" {
				s.Where(sql.EQ(s.C("trust_center_id"), anon.TrustCenterID))
				return
			}
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			logx.FromContext(ctx).Err(err).Msg("could not fetch org ids when scoping trustcenter")
			return
		}

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
