package schema

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/mixin"

	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx/accessmap"
)

// GroupPermissionsMixin is a mixin for group permissions on an entity
// This allows for editor + blocked_groups, and optionally viewer groups
// to be added as edges to the entity. The hooks are added to create the tuples in
// FGA for the groups
// After adding this mixin to a schema, you must also update the EdgeInfo
// to include the reverse edges on the group schema
type GroupPermissionsMixin struct {
	mixin.Schema
	// ViewPermissions adds view permission for a group
	ViewPermissions bool
	// IncludeInterceptorFilter is used to skip the interceptor filter
	// this is used for more complex view permissions that are not solely based
	// on the group membership
	IncludeInterceptorFilter bool
	// WorkflowEdgeEligible marks group permission edges as workflow-eligible.
	WorkflowEdgeEligible bool
}

// GroupPermissionsEdgesMixin is a mixin for the reverse edges on the group schema
// this should be used in conjunction with the GroupPermissionsMixin on the entity schema
type GroupPermissionsEdgesMixin struct {
	mixin.Schema

	EdgeInfo []EdgeInfo
}

// EdgeInfo is used to define the edge information for the reverse edges (group schema)
type EdgeInfo struct {
	Schema          any
	ViewPermissions bool
}

// NewGroupPermissionsEdgesMixin creates a new GroupPermissionsEdgesMixin with options applied
func newGroupPermissionsEdgesMixin(opts ...groupEdgesOptions) GroupPermissionsEdgesMixin {
	g := GroupPermissionsEdgesMixin{
		EdgeInfo: []EdgeInfo{},
	}

	for _, opt := range opts {
		opt(&g)
	}

	return g
}

// groupEdgesOptions is a function that can be used to modify the GroupPermissionsEdgesMixin
type groupEdgesOptions func(*GroupPermissionsEdgesMixin)

// withEdges sets the group edges to the schemas provided with view permissions enabled
func withEdges(schemas ...any) groupEdgesOptions {
	return func(g *GroupPermissionsEdgesMixin) {
		for _, schema := range schemas {
			g.EdgeInfo = append(g.EdgeInfo, EdgeInfo{
				Schema:          schema,
				ViewPermissions: true,
			})
		}
	}
}

// withEdgesNoView sets the group edges to the schemas provided with view permissions not included
func withEdgesNoView(schemas ...any) groupEdgesOptions {
	return func(g *GroupPermissionsEdgesMixin) {
		for _, schema := range schemas {
			g.EdgeInfo = append(g.EdgeInfo, EdgeInfo{
				Schema:          schema,
				ViewPermissions: false,
			})
		}
	}
}

// newGroupPermissionsMixin creates a new GroupPermissionsMixin with options applied
// by default view permissions are enabled
func newGroupPermissionsMixin(opts ...groupPermissionsOption) GroupPermissionsMixin {
	g := GroupPermissionsMixin{
		ViewPermissions: true,
	}

	for _, opt := range opts {
		opt(&g)
	}

	return g
}

// groupPermissionsOption is a function that can be used to modify the GroupPermissionsMixin
type groupPermissionsOption func(*GroupPermissionsMixin)

// withSkipViewPermissions disables view permissions for the group setup
func withSkipViewPermissions() groupPermissionsOption {
	return func(g *GroupPermissionsMixin) {
		g.ViewPermissions = false
	}
}

// withGroupPermissionsInterceptor skips the interceptor filter for the group permissions
func withGroupPermissionsInterceptor() groupPermissionsOption {
	return func(g *GroupPermissionsMixin) {
		g.IncludeInterceptorFilter = true
	}
}

// withWorkflowGroupEdges marks group permission edges as workflow-eligible.
func withWorkflowGroupEdges() groupPermissionsOption {
	return func(g *GroupPermissionsMixin) {
		g.WorkflowEdgeEligible = true
	}
}

// Edges of the GroupPermissionsMixin
func (g GroupPermissionsMixin) Edges() []ent.Edge {
	blockEdge := edge.To("blocked_groups", Group.Type).
		Comment("groups that are blocked from viewing or editing the risk").
		Annotations(
			entgql.RelayConnection(),
			entgql.QueryField(),
			entgql.MultiOrder(),
			accessmap.EdgeViewCheck(Group{}.Name()),
		)

	editEdge := edge.To("editors", Group.Type).
		Comment("provides edit access to the risk to members of the group").
		Annotations(
			entgql.RelayConnection(),
			entgql.QueryField(),
			entgql.MultiOrder(),
			accessmap.EdgeViewCheck(Group{}.Name()),
		)

	viewEdge := edge.To("viewers", Group.Type).
		Comment("provides view access to the risk to members of the group").
		Annotations(
			entgql.RelayConnection(),
			entgql.QueryField(),
			entgql.MultiOrder(),
			accessmap.EdgeViewCheck(Group{}.Name()),
		)

	edges := []ent.Edge{blockEdge, editEdge}

	// add the view edge if the view permissions are enabled
	if g.ViewPermissions {
		edges = append(edges, viewEdge)
	}

	return edges
}

// Interceptors of the GroupPermissionsMixin
func (g GroupPermissionsMixin) Interceptors() []ent.Interceptor {
	if !g.IncludeInterceptorFilter {
		return []ent.Interceptor{}
	}

	// this interceptor is used to limit the results returned by the query to not include
	// results that have a blocked group that the user is a member of
	// this can be used to prevent extra queries to fga for objects that are view by default
	// except for blocked groups (e.g. controls)
	return []ent.Interceptor{intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// add a filter to exclude results that have a blocked group that the user is a member of
		au, err := auth.GetAuthenticatedUserFromContext(ctx)
		if err != nil {
			return err
		}

		if skip := groupPermissionInterceptorSkipper(ctx, au); skip {
			return nil
		}

		groupIDs, err := generated.FromContext(ctx).Group.Query().Where(
			group.HasMembersWith(
				groupmembership.UserID(au.SubjectID),
			),
		).IDs(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get group IDs for user")

			return err
		}

		addBlockedGroupPredicate(q, groupIDs)

		if g.ViewPermissions {
			addViewGroupPredicate(q, groupIDs)
		}

		return nil
	})}
}

func groupPermissionInterceptorSkipper(ctx context.Context, au *auth.AuthenticatedUser) bool {
	// bypass if request is set to allowed
	if _, allow := privacy.DecisionFromContext(ctx); allow {
		return true
	}

	// if its a service account, we don't need to filter by groups
	if au.AuthenticationType == auth.APITokenAuthentication {
		return true
	}

	// skip for org owners, they might not have explicit access to the object, but they can view all objects in the org
	if err := rule.CheckCurrentOrgAccess(ctx, nil, fgax.OwnerRelation); errors.Is(err, privacy.Allow) {
		return true
	}

	return false
}

// addBlockedGroupPredicate adds a predicate to the query to filter out results
// that have a blocked group that the user is a member of
func addBlockedGroupPredicate(q intercept.Query, groupIDs []string) {
	objectSnakeCase := strcase.SnakeCase(q.Type())
	tableName := fmt.Sprintf("%s_blocked_groups", objectSnakeCase)

	q.WhereP(func(s *sql.Selector) {
		t := sql.Table(tableName)
		s.LeftJoin(t).On(
			s.C("id"), t.C(fmt.Sprintf("%s_id", objectSnakeCase)),
		)
		s.Where(
			sql.Or(
				sql.IsNull(t.C("group_id")),
				sql.NotIn(
					t.C("group_id"), lo.ToAnySlice(groupIDs)...,
				),
			),
		)
	})
}

// addViewGroupPredicate adds a predicate to the query to include
// results that have a viewer group that the user is a member of
// or results with no viewer group at all
func addViewGroupPredicate(q intercept.Query, groupIDs []string) {
	objectSnakeCase := strcase.SnakeCase(q.Type())
	tableName := fmt.Sprintf("%s_viewers", objectSnakeCase)

	q.WhereP(func(s *sql.Selector) {
		t := sql.Table(tableName)
		s.Join(t).On(
			s.C("id"), t.C(fmt.Sprintf("%s_id", objectSnakeCase)),
		)
		s.Where(
			sql.In(
				t.C("group_id"), lo.ToAnySlice(groupIDs)...,
			),
		)
	})
}

// Hooks of the GroupPermissionsMixin
func (g GroupPermissionsMixin) Hooks() (hooks []ent.Hook) {
	hooks = append(hooks, groupWriteOnlyHooks...)

	if g.ViewPermissions {
		hooks = append(hooks, groupReadOnlyHooks...)
	}

	return
}

// groupReadOnlyHooks are the hooks that are used to add the viewer tuples
// based on a group
var groupReadOnlyHooks = []ent.Hook{
	hook.On(
		hooks.HookRelationTuples(map[string]string{
			"viewer_id": "group",
		}, fgax.ViewerRelation), // add viewer tuples for associated groups
		ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
	),
}

// groupWriteOnlyHooks are the hooks that are used to add the editor and blocked tuples
// based on a group
var groupWriteOnlyHooks = []ent.Hook{
	hook.On(
		hooks.HookRelationTuples(map[string]string{
			"editor_id": "group",
		}, fgax.EditorRelation), // add editor tuples for associated groups
		ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
	),
	hook.On(
		hooks.HookRelationTuples(map[string]string{
			"blocked_group_id": "group",
		}, fgax.BlockedRelation), // add block tuples for associated groups
		ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
	),
}

// Edges of the GroupPermissionsEdgesMixin
func (g GroupPermissionsEdgesMixin) Edges() []ent.Edge {
	var edges []ent.Edge

	for _, schema := range g.EdgeInfo {
		sch := toSchemaFuncs(schema.Schema)

		var defaultReverseEdges = []ent.Edge{
			edge.From(fmt.Sprintf("%s_editors", sch.Name()), sch.GetType()).
				Ref("editors").
				Annotations(
					entgql.RelayConnection(),
					entgql.QueryField(),
					entgql.MultiOrder(),
					accessmap.EdgeAuthCheck(sch.Name()),
				),
			edge.From(fmt.Sprintf("%s_blocked_groups", sch.Name()), sch.GetType()).
				Ref("blocked_groups").
				Annotations(
					entgql.RelayConnection(),
					entgql.QueryField(),
					entgql.MultiOrder(),
					accessmap.EdgeAuthCheck(sch.Name()),
				),
		}

		edges = append(edges, defaultReverseEdges...)

		// add the view edge if the view permissions are enabled
		if schema.ViewPermissions {
			viewerEdge := edge.From(fmt.Sprintf("%s_viewers", sch.Name()), sch.GetType()).
				Ref("viewers").
				Annotations(
					entgql.RelayConnection(),
					entgql.QueryField(),
					entgql.MultiOrder(),
					accessmap.EdgeAuthCheck(sch.Name()),
				)

			edges = append(edges, viewerEdge)
		}
	}

	return edges
}

// Hooks of the GroupPermissionsEdgesMixin
func (g GroupPermissionsEdgesMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookGroupPermissionsTuples(),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}
