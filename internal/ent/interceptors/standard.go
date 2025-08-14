package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentercompliance"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// TraverseStandard only returns public standards and standards owned by the organization
func TraverseStandard() ent.Interceptor {
	return intercept.TraverseStandard(func(ctx context.Context, q *generated.StandardQuery) error {
		anon, isAnon := auth.AnonymousTrustCenterUserFromContext(ctx)
		if isAnon {
			q.Where(
				standard.HasTrustCenterCompliancesWith(
					trustcentercompliance.HasTrustCenterWith(
						trustcenter.OwnerID(anon.OrganizationID),
					),
				),
			)
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		systemStandardPredicates := []predicate.Standard{
			standard.OwnerIDIsNil(),
			standard.SystemOwned(true),
		}

		admin, err := rule.CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if !admin {
			// if the user is a not-system admin, restrict to only public standards
			systemStandardPredicates = append(systemStandardPredicates, standard.IsPublic(true))
		}

		// filter to return system owned standards and standards owned by the organization
		q.Where(
			standard.Or(
				standard.And(
					systemStandardPredicates...,
				),
				standard.OwnerIDIn(orgIDs...),
			),
		)

		return nil
	})
}

// func getAllowedTrustCenterStandards(ctx context.Context, tcID string) ([]string, error) {
// 	req := fgax.ListRequest{
// 		SubjectID:   tcID,
// 		SubjectType: "trust_center",
// 		ObjectType:  "standard",
// 		Relation:    "associated_with",
// 	}

// 	zerolog.Ctx(ctx).Debug().Interface("req", req).Msg("getting authorized object ids")

// 	resp, err := utils.AuthzClientFromContext(ctx).ListObjectsRequest(ctx, req)
// 	if err != nil {
// 		return []string{}, err
// 	}
// 	standardIDs := []string{}
// 	for _, obj := range resp.Objects {
// 		entity, err := fgax.ParseEntity(obj)
// 		if err != nil {
// 			return []string{}, nil
// 		}
// 		standardIDs = append(standardIDs, entity.Identifier)
// 	}

// 	return standardIDs, nil
// }
