package schema

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/mixin"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/entx/accessmap"
)

// fgaModelPath is the path to the FGA model file on disk
var fgaModelPath = "../../../fga/model/model.fga"

// loadCreateObjectTypes loads the list of object types that support group-based create permissions.
// It reads the FGA model from disk at runtime and returns only the parsed types.
// If reading or parsing fails, it returns an empty list (no fallback to defaults).
func loadCreateObjectTypes() []string {
	model, err := os.ReadFile(fgaModelPath)
	if err != nil {
		log.Debug().Err(err).Msgf("failed to read FGA model from %s", fgaModelPath)
		return nil
	}

	parsed, err := creatorTypesFromModel(model)
	if err != nil {
		return nil
	}

	return parsed
}

// creatorTypesFromModel parses the FGA model file and returns a list of object types
// that have a creator relation granting access via group membership.
// Only relations defined as "define [object]_creator: [group#member]" are included.
// Returns a slice of object type names, or an error if parsing fails.
func creatorTypesFromModel(model []byte) ([]string, error) {
	var (
		types     []string
		inOrgType bool
		creatorRE = regexp.MustCompile(`^define\s+([a-z0-9_]+)_creator:\s*(.+)$`)
	)

	scanner := bufio.NewScanner(bytes.NewReader(model))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "type ") {
			// detect if we are entering or leaving the organization relations section
			if strings.HasPrefix(line, "type organization") {
				inOrgType = true
				continue
			}

			if inOrgType {
				break
			}

			continue
		}

		if !inOrgType {
			continue
		}

		matches := creatorRE.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		// only include creator relations that grant access via group membership
		if !strings.Contains(matches[2], "group#member") {
			continue
		}

		types = append(types, matches[1])
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return types, nil
}

// GroupBasedCreateAccessMixin is a mixin for group permissions for creation of an entity
// that should be added to both the to schema (Group) and the from schema (Organization)
// the object type must be included in the fga model for this to work:
//
//	#     define [object]_creator: [group#member]
//	#     define can_create_[object]: can_edit or [object]_creator
//
// once these exist in the model, the object type will be picked up automatically by the createObjectTypes list
// above and the mixin will add the edges and hooks to the schema that will create
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

	for _, t := range loadCreateObjectTypes() {
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

	for _, objectType := range loadCreateObjectTypes() {
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
