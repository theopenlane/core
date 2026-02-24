package policy

import (
	"context"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/authzgenerated"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// CheckCreateAccess checks if the user has access to create an object in the org
// for create operations
func CheckCreateAccess() privacy.MutationRule {
	return privacy.OnMutationOperation(
		rule.CheckGroupBasedObjectCreationAccess(),
		ent.OpCreate,
	)
}

// AllowCreate is mutation that allows any user to create a that mutation type
// this is only for the actual mutation type; edges are checked by edge access
// hooks
func AllowCreate() privacy.MutationRule {
	return privacy.OnMutationOperation(
		privacy.MutationRuleFunc(func(_ context.Context, m generated.Mutation) error {
			if m.Op() == ent.OpCreate {
				return privacy.Allow
			}

			return privacy.Skip
		}),

		ent.OpCreate,
	)
}

// CanCreateObjectsUnderParents will check edit access to the specified
// edges that allow the user to create objects under.
// If the mutation has no permission edges a privacy.Skip will be returned
// If there are edges, if the user does not have edit
// access to one of the edges, a privacy.Deny will be returned
// If there are parent permission edges, and the user has edit access to all of them
// they will be allowed to create the object
func CanCreateObjectsUnderParents(edges []string) privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		if m.Op() != generated.OpCreate {
			return privacy.Skip
		}

		addedEdges := m.AddedEdges()

		edgesToCheck := mapx.MapIntersectionUnique(edges, addedEdges)

		if len(edgesToCheck) == 0 {
			return privacy.Skipf("no parent permission edges, cannot authorize creation")
		}

		if err := CheckEdgesForAddedAccess(ctx, m, edgesToCheck); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("access not allowed to parent, cannot authorize creation")

			return privacy.Deny
		}

		return privacy.Allow
	})
}

// CheckOrgReadAccess checks if the requestor has access to read the organization
func CheckOrgReadAccess() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		if _, hasAnon := auth.ContextValue(ctx, auth.AnonymousTrustCenterUserKey); hasAnon {
			return privacy.Deny
		}
		// check if the user has access to view the organization
		// check the query first for the IDS
		query, ok := q.(*generated.OrganizationQuery)
		if ok {
			// if we get an error (privacy.Deny or privacy.Allow are both "errors")
			// return as the result
			// if its nil, we didn't check anything and should continue to the next check
			if err := rule.CheckOrgAccessBasedOnRequest(ctx, fgax.CanView, query); err != nil {
				return err
			}
		}

		// otherwise check against the current context
		return rule.CheckCurrentOrgAccess(ctx, nil, fgax.CanView)
	})
}

// CheckOrgReadAccess checks if the requestor has access to edit access the organization for
// some query operations
func CheckOrgEditAccess() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, _ ent.Query) error {
		// otherwise check against the current context
		return rule.CheckCurrentOrgAccess(ctx, nil, fgax.CanEdit)
	})
}

// CheckOrgWriteAccess checks if the requestor has access to edit the organization
func CheckOrgWriteAccess() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking org write access")
		return rule.CheckCurrentOrgAccess(ctx, m, fgax.CanEdit)
	})
}

// CheckOrgAccess checks if the requestor has access to read the organization
func CheckOrgAccess() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking org read access")
		return rule.CheckCurrentOrgAccess(ctx, m, fgax.CanView)
	})
}

// DenyQueryIfNotAuthenticated denies a query if the user is not authenticated
func DenyQueryIfNotAuthenticated() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, _ ent.Query) error {
		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			logx.FromContext(ctx).Info().Msg("unable to get authenticated user context")

			return auth.ErrNoAuthUser
		}

		return nil
	})
}

// DenyMutationIfNotAuthenticated denies a mutation if the user is not authenticated
func DenyMutationIfNotAuthenticated() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			logx.FromContext(ctx).Info().Msg("unable to get authenticated user context")

			return auth.ErrNoAuthUser
		}

		return nil
	})
}

// CheckFeatureAccess ensures the requesting organization has the given feature enabled.
func CheckFeatureAccess(feature string) privacy.QueryMutationRule {
	return rule.AllowIfHasFeature(feature)
}

// CheckAnyFeatureAccess ensures the requesting organization has at least one of the provided features enabled.
func CheckAnyFeatureAccess(features ...models.OrgModule) privacy.QueryMutationRule {
	return rule.AllowIfHasAnyFeature(features...)
}

// CheckEdgesForAddedAccess checks if the user has access to the object they are trying to add permissions to
// it will look at the AddedEdges and check if the user has access to the object
func CheckEdgesForAddedAccess(ctx context.Context, m ent.Mutation, edges []string) error {
	return checkEdgesEditAccess(ctx, m, edges, true)
}

// CheckEdgesForRemovedAccess checks if the user has access to the object they are trying to remove permissions from
func CheckEdgesForRemovedAccess(ctx context.Context, m ent.Mutation, edges []string) error {
	return checkEdgesEditAccess(ctx, m, edges, false)
}

// checkEdgesEditAccess takes a list of edges and looks for the permissions edges to confirm the user has edit access
func checkEdgesEditAccess(ctx context.Context, m ent.Mutation, edges []string, added bool) error {
	actor, ok := auth.CallerFromContext(ctx)
	if !ok || actor == nil || actor.IsAnonymous() {
		logx.FromContext(ctx).Error().Msg("unable to get caller from context")

		return auth.ErrNoAuthUser
	}

	for _, edge := range edges {
		relationCheck := fgax.CanEdit

		edgeMap := mapEdgeToObjectType(ctx, m.Type(), edge)
		if edgeMap.SkipEditCheck {
			if edgeMap.CheckViewAccess {
				relationCheck = fgax.CanView
			} else {
				// not required to check the edge, so skip
				continue
			}
		}

		var ids []ent.Value
		if added {
			ids = m.AddedIDs(edge)
		} else {
			ids = m.RemovedIDs(edge)
		}

		for _, id := range ids {
			idStr, idOk := id.(string)
			if !idOk {
				logx.FromContext(ctx).Warn().Interface("id", id).Msg("id is not a string, unable to check access")

				continue
			}

			if idStr == "" {
				logx.FromContext(ctx).Debug().Msg("id is empty, nothing to check, validation will catch this later")

				continue
			}

			// If the object type of the edge to check is an organization, we need to ensure both that
			// the object is in the organization and that the user has edit access
			// to the organization
			if edgeMap.ObjectType == organization.Label {
				orgID := actor.OrganizationID
				if orgID == "" {
					logx.FromContext(ctx).Error().Msg("unable to get organization id from context")

					return auth.ErrNoAuthUser
				}

				if err := ensureObjectInOrganization(ctx, m, edge, idStr, orgID); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("object is not part of the organization")

					return err
				}

				idStr = orgID
			}

			ac := fgax.AccessCheck{
				Relation:    relationCheck,
				ObjectID:    idStr,
				ObjectType:  fgax.Kind(edgeMap.ObjectType),
				SubjectID:   actor.SubjectID,
				SubjectType: actor.SubjectType(),
				Context:     utils.NewOrganizationContextKey(actor.SubjectEmail),
			}

			if allow, err := utils.AuthzClient(ctx, m).CheckAccess(ctx, ac); err != nil || !allow {
				logx.FromContext(ctx).Error().Err(err).Str("edge", edge).Str("relation", ac.Relation).Str("object_id", ac.ObjectID).Str("object_type", edgeMap.ObjectType).Msg("user does not have access to the object for edge permissions")

				return generated.ErrPermissionDenied
			}
		}
	}

	return nil
}

// mapEdgeToObjectType maps the edge to the object type and returns the EdgeAccess
// based on the generated access map
func mapEdgeToObjectType(ctx context.Context, schema string, edge string) authzgenerated.EdgeAccess {
	logx.FromContext(ctx).Debug().Str("schema", schema).Str("edge", edge).Msg("mapping edge to object type")
	schemaType := strcase.SnakeCase(schema)

	schemaMap, ok := authzgenerated.EdgeAccessMap[schemaType]
	if !ok {
		logx.FromContext(ctx).Error().Str("schema", schema).Msg("schema not found in edge access map")
		return authzgenerated.EdgeAccess{}
	}

	edgeAccess, ok := schemaMap[edge]
	if !ok {
		logx.FromContext(ctx).Error().Str("edge", edge).Msg("edge not found in edge access map for schema")

		return authzgenerated.EdgeAccess{}
	}

	return edgeAccess
}

// ensureObjectInOrganization checks if the object is in the organization based on the requested edge
func ensureObjectInOrganization(ctx context.Context, m ent.Mutation, edge string, objectID, orgID string) error {
	// also ensure the id is part of the organization
	mut, ok := m.(utils.GenericMutation)
	if !ok {
		logx.FromContext(ctx).Error().Msg("unable to determine access")
		return privacy.Denyf("unable to determine access")
	}

	// check view access to the organization instead if the edge is an organization
	if edge == organization.Label {
		if err := rule.CheckCurrentOrgAccess(ctx, m, fgax.CanView); errors.Is(err, privacy.Allow) {
			return nil
		}

		logx.FromContext(ctx).Error().Msg("user does not have access to the organization")

		return privacy.Denyf("user does not have access to the requested organization")
	}

	// check if the object is in the organization
	table := pluralize.NewClient().Plural(edge)
	query := "SELECT EXISTS (SELECT 1 FROM " + table + " WHERE id = $1 and (owner_id = $2 or owner_id IS NULL))"

	var rows sql.Rows
	if err := mut.Client().Driver().Query(ctx, query, []any{objectID, orgID}, &rows); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to check for object in organization")

		return privacy.Denyf("failed to check for object in organization: %v", err)
	}

	defer rows.Close()

	if rows.Next() {
		var exists bool
		if err := rows.Scan(&exists); err == nil && exists {
			return nil
		}
	}

	// fall back to deny if the object is not in the organization
	return privacy.Denyf("requested object not in organization")
}
