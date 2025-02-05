package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/mixin"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/iam/fgax"
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
}

// GroupPermissionsEdgesMixin is a mixin for the reverse edges on the group schema
// this should be used in conjunction with the GroupPermissionsMixin on the entity schema
type GroupPermissionsEdgesMixin struct {
	mixin.Schema

	EdgeInfo []EdgeInfo
}

// EdgeInfo is used to define the edge information for the reverse edges (group schema)
type EdgeInfo struct {
	Name            string
	ViewPermissions bool
	Type            any
}

// NewGroupPermissionsMixin creates a new GroupPermissionsMixin with optional viewer permissions
func NewGroupPermissionsMixin(viewPermissions bool) GroupPermissionsMixin {
	return GroupPermissionsMixin{
		ViewPermissions: viewPermissions,
	}
}

// Edges of the GroupPermissionsMixin
func (g GroupPermissionsMixin) Edges() []ent.Edge {
	blockEdge := edge.To("blocked_groups", Group.Type).
		Comment("groups that are blocked from viewing or editing the risk")

	editEdge := edge.To("editors", Group.Type).
		Comment("provides edit access to the risk to members of the group")

	viewEdge := edge.To("viewers", Group.Type).
		Comment("provides view access to the risk to members of the group")

	edges := []ent.Edge{blockEdge, editEdge}

	// add the view edge if the view permissions are enabled
	if g.ViewPermissions {
		edges = append(edges, viewEdge)
	}

	return edges
}

// Hooks of the GroupPermissionsMixin
func (g GroupPermissionsMixin) Hooks() (hooks []ent.Hook) {
	hooks = append(hooks, groupWriteOnlyHooks...)

	if g.ViewPermissions {
		hooks = append(hooks, groupReadOnlyHooks...)
	}

	return
}

// groupReadWriteHooks are the hooks that are used to add the editor, blocked, and viewer tuples
// based on a group
var groupReadWriteHooks = append(groupWriteOnlyHooks, groupReadOnlyHooks...)

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
		var defaultReverseEdges = []ent.Edge{
			edge.From(fmt.Sprintf("%s_editors", schema.Name), schema.Type).
				Ref("editors"),
			edge.From(fmt.Sprintf("%s_blocked_groups", schema.Name), schema.Type).
				Ref("blocked_groups"),
		}

		edges = append(edges, defaultReverseEdges...)

		// add the view edge if the view permissions are enabled
		if schema.ViewPermissions {
			viewerEdge := edge.From(fmt.Sprintf("%s_viewers", schema.Name), schema.Type).
				Ref("viewers")

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
