package schema

import (
	"fmt"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/mixin"
	"github.com/theopenlane/iam/fgax"

	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/entx/accessmap"
)

// createObjectTypes is derived from the model scopes for service subjects.
var createObjectTypes = func() []string {
	opts, err := fgamodel.CreateOptions()
	if err != nil {
		return nil
	}

	return opts
}()

// GroupBasedCreateAccessMixin is a mixin for group permissions for creation of an entity
// that should be added to both the to schema (Group) and the from schema (Organization)
// the object type must be included in the fga model for this to work:
//
//	#     define [object]_creator: [group#member]
//	#     define can_create_[object]: can_edit or [object]_creator
//
// once these exist in the model, the object type can be added to the createObjectTypes list
// above and the mixin will automatically add the edges and hooks to the schema that will create
// the tuples upon association with the organization
type GroupBasedCreateAccessMixin struct {
	mixin.Schema
}

// NewGroupBasedCreateAccessMixin creates a new GroupBasedCreateAccessMixin with the specified edges
func NewGroupBasedCreateAccessMixin() GroupBasedCreateAccessMixin {
	return GroupBasedCreateAccessMixin{}
}

// Edges of the GroupBasedCreateAccessMixin
func (c GroupBasedCreateAccessMixin) Edges() []ent.Edge {
	edges := []ent.Edge{}

	for _, t := range createObjectTypes {
		toName := strings.ToLower(fmt.Sprintf("%s_creators", t))

		edge := edge.To(toName, Group.Type).
			Comment(fmt.Sprintf("groups that are allowed to create %ss", t)).
			Annotations(
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
				accessmap.EdgeViewCheck(Group{}.Name()),
			)

		edges = append(edges, edge)
	}

	return edges
}

// Hooks of the GroupBasedCreateAccessMixin
func (c GroupBasedCreateAccessMixin) Hooks() []ent.Hook {
	var h []ent.Hook

	for _, objectType := range createObjectTypes {
		idField := fmt.Sprintf("%s_creator_id", objectType)

		relation := fgax.Relation(objectType + "_creator")

		hook := hook.On(
			hooks.HookRelationTuples(map[string]string{
				idField: "group",
			}, relation),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		)

		h = append(h, hook)
	}

	return h
}
