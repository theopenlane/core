package interceptors

import (
	"context"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
)

// historyObjectOwnerTables maps a history object owner to the parent table whose owner_id scopes it to an organization
var historyObjectOwnerTables = map[string]string{
	entityops.SchemaGroup.Snake:       group.Table,
	entityops.SchemaProgram.Snake:     program.Table,
	entityops.SchemaTrustCenter.Snake: trustcenter.Table,
}

// HistoryAccess is a traversal interceptor that checks if the user has the required role for the organization
func HistoryAccess(relation string, orgOwned, userOwed bool, objectOwner string) ent.Interceptor {
	return TraverseFunc(func(ctx context.Context, q Query) error {
		caller, ok := auth.CallerFromContext(ctx)
		if !ok {
			return auth.ErrNoAuthUser
		}

		// check if the user has the audit log role for the organization
		req := fgax.AccessCheck{
			Relation:    relation,
			SubjectID:   caller.SubjectID,
			SubjectType: caller.SubjectType(),
			ObjectType:  generated.TypeOrganization,
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

		return addFilter(ctx, q, userOwed, objectOwner, allowedOrgs)
	})
}

// addFilter adds a filter to the query based on the authenticated user's organization
func addFilter(ctx context.Context, q Query, userOwed bool, objectOwner string, allowedOrgs []string) error {
	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return err
	}

	switch {
	// special cases of top level owned objects of organization or user
	case q.Type() == "OrganizationHistory":
		q.WhereP(
			sql.FieldIn("ref", allowedOrgs...),
		)
	case q.Type() == "OrganizationSettingHistory" || q.Type() == "OrgMembershipHistory":
		q.WhereP(
			sql.FieldIn("organization_id", allowedOrgs...),
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
	case objectOwner != "":
		// fga can grant audit log access beyond the caller's orgs (e.g. public trust centers), so also limit to parents the caller's orgs own
		parentTable, ok := historyObjectOwnerTables[objectOwner]
		if !ok {
			return ErrUnscopedHistory
		}

		q.WhereP(parentOwnedByOrgsP(parentTable, objectOwner+"_id", allowedOrgs))
	default:
		// history of schemas with no organization owner, such as system managed mappable domains, is only for system admins
		// fallback here, its better to not return data than expose
		// cross org data
		schema, ok := entityops.LookupSchema(strings.TrimSuffix(q.Type(), "History"))
		if !ok || schema.OwnerField == "" {
			if !auth.IsSystemAdminFromContext(ctx) {
				return ErrUnscopedHistory
			}

			return nil
		}

		q.WhereP(
			sql.FieldIn(schema.OwnerField, allowedOrgs...),
		)
	}

	return nil
}
