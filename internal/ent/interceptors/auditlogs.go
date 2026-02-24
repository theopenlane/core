package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// HistoryAccess is a traversal interceptor that checks if the user has the required role for the organization
func HistoryAccess(relation string, orgOwned, userOwed bool, objectOwner string) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		// check if the user has the audit log role for the organization
		req := fgax.AccessCheck{
			Relation:    relation,
			SubjectID:   caller.SubjectID,
			SubjectType: caller.SubjectType(),
			ObjectType:  generated.TypeOrganization,
			Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
		}

		var allowedOrgs []string

		for _, orgID := range caller.OrgIDs() {
			req.ObjectID = orgID

			allowed, err := utils.AuthzClientFromContext(ctx).CheckAccess(ctx, req)
			if err != nil {
				return err
			}

			if allowed {
				allowedOrgs = append(allowedOrgs, orgID)
			}
		}

		if len(allowedOrgs) == 0 {
			return rout.ErrPermissionDenied
		}

		if objectOwner != "" {
			filter, err := GetAuthorizedObjectIDs(ctx, objectOwner, fgax.CanViewAuditLog)
			if err != nil {
				return err
			}

			idField := strings.ToLower(objectOwner + "_id")

			q.WhereP(
				sql.FieldIn(idField, filter...),
			)
		}

		return addFilter(ctx, q, orgOwned, userOwed, allowedOrgs)
	})
}

// addFilter adds a filter to the query based on the authenticated user's organization
func addFilter(ctx context.Context, q Query, orgOwned, userOwed bool, allowedOrgs []string) error {
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return err
	}

	switch {
	case q.Type() == "OrganizationHistory":
		q.WhereP(
			sql.FieldIn("ref", allowedOrgs...),
		)
	case q.Type() == "OrganizationSettingHistory" || q.Type() == "OrgMembershipHistory":
		q.WhereP(
			sql.FieldIn("organization_id", allowedOrgs...),
		)
	case orgOwned:
		q.WhereP(
			sql.FieldIn("owner_id", allowedOrgs...),
		)
	case q.Type() == "UserHistory":
		q.WhereP(
			sql.FieldIn("ref", userID),
		)
	case q.Type() == "UserSettingHistory":
		q.WhereP(
			sql.FieldIn("user_id", userID),
		)
	case userOwed:
		q.WhereP(
			sql.FieldIn("owner_id", userID),
		)
	}

	return nil
}
