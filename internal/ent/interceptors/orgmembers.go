package interceptors

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/logx"

	fgamodel "github.com/theopenlane/core/v2/fga/model"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/ent/historygenerated/orgmembershiphistory"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
)

const (
	isDeletedProperty = "is_deleted"
)

// TraverseOrgMembers is middleware to change the Org Members query
func TraverseOrgMembers() ent.Interceptor {
	return intercept.TraverseOrgMembership(func(ctx context.Context, q *generated.OrgMembershipQuery) error {
		ctxQuery := ent.QueryFromContext(ctx)

		if isRequestIncludingHistory(ctx) {
			// needed because we cannot always include this else it would fail if
			// client requests totalCount or others because the query would not contain this.
			q.Modify(withOrgMembershipHistoryModifier(ctxQuery.Op == "All"))
		}

		if auth.HasCrossOrgCapabilities(ctx) {
			return nil
		}

		// Organization list queries should not be deduped, they
		// will show up under each org they are a member of
		if ctxQuery.Type == generated.TypeOrganization {
			return nil
		}

		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Msg("unable to get authenticated user context while traversing org members")
			return auth.ErrNoAuthUser
		}

		// get all parent orgs to ensure we get all OrgMembers in the org tree
		allOrgsIDs, err := getAllParentOrgIDs(ctx, orgIDs)
		if err != nil {
			return err
		}

		// sets the organization id on the query for the current organization
		q.Where(orgmembership.OrganizationIDIn(allOrgsIDs...))

		return nil
	})
}

// InterceptorOrgMember is middleware to change the OrgMember query result
func InterceptorOrgMember() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.OrgMembershipFunc(func(ctx context.Context, q *generated.OrgMembershipQuery) (generated.Value, error) {
			// run the query
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if orgMembersSkipInterceptor(ctx) {
				return v, nil
			}

			// deduplicate the org members if the query is for org members
			members, ok := v.([]*generated.OrgMembership)
			if !ok || len(members) == 0 {
				return v, nil
			}

			if isRequestIncludingHistory(ctx) {
				if err := includeRemovedUserEdges(ctx, members); err != nil {
					return nil, err
				}

				for _, om := range members {
					if v, err := om.Value(isDeletedProperty); err == nil {
						om.IsDeleted, _ = v.(bool)
					}
				}
			}

			res, err := dedupeOrgMembers(ctx, members)
			if err != nil {
				return nil, err
			}

			if !graphutils.CheckForRequestedField(ctx, "additionalRoles") {
				return res, nil
			}

			// get additional roles per result
			for i, r := range res {
				res[i].AdditionalRoles, err = getFunctionalRoles(ctx, r.UserID)
				if err != nil {
					return nil, err
				}
			}

			return res, nil
		})
	})
}

func getFunctionalRoles(ctx context.Context, userID string) ([]string, error) {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return []string{}, nil
	}

	return GetFunctionalRolesForSubject(ctx, auth.UserSubjectType, userID, orgID)
}

// GetFunctionalRolesForSubject returns the display names of the functional/organization roles
// assigned to the given subject on the organization. The subject is a user or a group tupleset
// (e.g. subjectType "group" with subjectID "<groupID>#member"), matching how the roles were assigned.
func GetFunctionalRolesForSubject(ctx context.Context, subjectType, subjectID, orgID string) ([]string, error) {
	_, ok := auth.CallerFromContext(ctx)
	if !ok {
		return []string{}, nil
	}

	roles, err := fgamodel.OrganizationRoles()
	if err != nil {
		return []string{}, err
	}

	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.ID)
	}

	req := fgax.ListAccess{
		SubjectType: subjectType,
		SubjectID:   subjectID,
		ObjectID:    orgID,
		ObjectType:  fgax.Kind(generated.TypeOrganization),
		Relations:   ids,
	}

	client := utils.AuthzClientFromContext(ctx)
	if client == nil {
		return []string{}, nil
	}

	assignedRoles, err := client.ListRelations(ctx, req, fgax.WithHighConsistency())
	if err != nil {
		return []string{}, err
	}

	return fgamodel.GetOrganizationRoleStrings(roles, assignedRoles), nil
}

// dedupeOrgMembers removes duplicate org members from the list
func dedupeOrgMembers(ctx context.Context, members []*generated.OrgMembership) ([]*generated.OrgMembership, error) {

	missingOrgs := lo.Uniq(lo.FilterMap(members, func(om *generated.OrgMembership, _ int) (string, bool) {
		return om.OrganizationID, om.Edges.Organization == nil && om.OrganizationID != ""
	}))

	var orgsMap map[string]*generated.Organization

	if len(missingOrgs) > 0 {
		orgs, err := generated.FromContext(ctx).Organization.Query().
			Where(organization.IDIn(missingOrgs...)).
			Select(organization.FieldID, organization.FieldParentOrganizationID).
			All(entx.SkipSoftDelete(auth.WithInternalReadContext(ctx)))
		if err != nil {
			return nil, err
		}

		orgsMap = lo.KeyBy(orgs, func(o *generated.Organization) string {
			return o.ID
		})
	}

	seen := map[string]*generated.OrgMembership{}
	deduped := []*generated.OrgMembership{}

	for _, om := range members {
		org := om.Edges.Organization
		if org == nil {
			org = orgsMap[om.OrganizationID]
		}

		if org == nil {
			deduped = append(deduped, om)
			seen[om.UserID] = om

			continue
		}

		// we dedupe for hierarchical orgs, we should skip this if the org has no parent or children
		children, err := org.QueryChildren().All(ctx)
		if err != nil {
			return nil, err
		}

		if org.ParentOrganizationID == "" && len(children) == 0 {
			deduped = append(deduped, om)
			seen[om.UserID] = om

			continue
		}

		if _, ok := seen[om.UserID]; !ok {
			deduped = append(deduped, om)
			seen[om.UserID] = om
		}
	}

	return deduped, nil
}

// orgMembersSkipInterceptor includes conditions to skip the org members interceptor
func orgMembersSkipInterceptor(ctx context.Context) bool {
	// internal requests don't need to dedupe the org members
	if auth.IsInternalRequest(ctx) {
		return true
	}

	// Organization list queries should not be deduped, they
	// will show up under each org they are a member of
	ctxQuery := ent.QueryFromContext(ctx)

	return ctxQuery.Type == generated.TypeOrganization
}

func isRequestIncludingHistory(ctx context.Context) bool {
	fCtx := graphql.GetFieldContext(ctx)
	if fCtx == nil {
		return false
	}

	where, ok := fCtx.Args["where"].(*generated.OrgMembershipWhereInput)
	if !ok || where == nil {
		return false
	}

	return lo.FromPtr(where.IncludeDeleted)
}

func includeRemovedUserEdges(ctx context.Context, members []*generated.OrgMembership) error {
	missingMembers := lo.Filter(members, func(om *generated.OrgMembership, _ int) bool {
		_, err := om.Edges.UserOrErr()
		return om.UserID != "" && generated.IsNotFound(err)
	})
	if len(missingMembers) == 0 {
		return nil
	}

	userField, ok := getOrgMembershipUserField(ctx)
	if !ok {
		return nil
	}

	userIDs := lo.Map(missingMembers, func(om *generated.OrgMembership, _ int) string {
		return om.UserID
	})

	readCtx := entx.SkipSoftDelete(auth.WithInternalReadContext(ctx))

	userCtx := graphql.WithFieldContext(readCtx, &graphql.FieldContext{
		Field: userField,
	})

	uq, err := generated.FromContext(ctx).User.Query().
		Where(user.IDIn(lo.Uniq(userIDs)...)).
		CollectFields(userCtx)
	if err != nil {
		return err
	}

	users, err := uq.All(readCtx)
	if err != nil {
		return err
	}

	groupByID := lo.KeyBy(users, func(u *generated.User) string {
		return u.ID
	})

	for _, om := range missingMembers {
		om.Edges.User = groupByID[om.UserID]
	}

	return nil
}

func getOrgMembershipUserField(ctx context.Context) (graphql.CollectedField, bool) {
	opCtx := graphql.GetOperationContext(ctx)
	field := graphql.GetFieldContext(ctx).Field

	for _, name := range []string{"edges", "node", "user"} {
		var match graphql.CollectedField

		for _, child := range graphql.CollectFields(opCtx, field.Selections, nil) {
			if child.Name != name {
				continue
			}

			if match.Field == nil {
				match.Field = child.Field
			}

			match.Selections = append(match.Selections, child.Selections...)
		}

		if match.Field == nil {
			return graphql.CollectedField{}, false
		}

		field = match
	}

	return field, true
}

// withOrgMembershipHistoryModifier modifies the query to include the org_membership_history while preserving the
// pagination among others else we'd need to change how pagination loading works for history
// but since this is one off, we don't need to do that and others will ideally just go through the history
// schemas directly
func withOrgMembershipHistoryModifier(includeDeleted bool) func(*sql.Selector) {
	return func(s *sql.Selector) {
		cols := lo.Replace(orgmembership.Columns, orgmembership.FieldID,
			sql.As(orgmembershiphistory.FieldRef, orgmembership.FieldID), 1)

		sel := sql.Select(orgmembership.Columns...).
			AppendSelectExprAs(sql.Raw("FALSE"), isDeletedProperty).
			From(sql.Table(orgmembership.Table))

		histSel := sql.Select(cols...).
			AppendSelectExprAs(sql.Raw("TRUE"), isDeletedProperty).
			From(sql.Table(orgmembershiphistory.Table)).
			Where(sql.EQ(orgmembershiphistory.FieldOperation, history.OpTypeDelete))

		s.From(sel.UnionAll(histSel).
			As(orgmembership.Table))

		if includeDeleted {
			s.AppendSelect(s.C(isDeletedProperty))
		}
	}
}
