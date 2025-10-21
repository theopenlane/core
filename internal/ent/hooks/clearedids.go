package hooks

import (
	"context"
	"sort"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/theopenlane/utils/contextx"
)

// RelationType represents the kind of relation that was modified
type RelationType string

const (
	RelationTypeEdge  RelationType = "edge"
	RelationTypeField RelationType = "field"
)

// RelationAction describes how a relation changed during a mutation
type RelationAction string

const (
	ActionEdgeRemoved   RelationAction = "edge_removed"
	ActionEdgeCleared   RelationAction = "edge_cleared"
	ActionFieldCleared  RelationAction = "field_cleared"
	ActionFieldReplaced RelationAction = "field_replaced"
)

// ClearedRelationEntry captures metadata about IDs that were removed from a relation
type ClearedRelationEntry struct {
	Name   string
	Type   RelationType
	Action RelationAction
	IDs    []string
}

// ClearedMutation stores the relation entries captured for a single mutation execution
type ClearedMutation struct {
	MutationType string
	Op           ent.Op
	Entries      []ClearedRelationEntry
}

type clearedIDsState struct {
	Mutations []ClearedMutation
}

// HookCaptureClearedIDs captures IDs removed or cleared from relations and stores them on the context
func HookCaptureClearedIDs() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if !(m.Op().Is(ent.OpUpdate) || m.Op().Is(ent.OpUpdateOne) || m.Op().Is(ent.OpDelete) || m.Op().Is(ent.OpDeleteOne)) {
				return next.Mutate(ctx, m)
			}

			state, ctx := ensureClearedState(ctx)
			entries := collectClearedRelationEntries(ctx, m)
			if len(entries) > 0 {
				state.Mutations = append(state.Mutations, ClearedMutation{
					MutationType: m.Type(),
					Op:           m.Op(),
					Entries:      entries,
				})
			}

			return next.Mutate(ctx, m)
		})
	}
}

// GetClearedMutations returns any cleared relation metadata stored on the context
func GetClearedMutations(ctx context.Context) []ClearedMutation {
	state, ok := contextx.From[*clearedIDsState](ctx)
	if !ok || state == nil {
		return nil
	}
	return append([]ClearedMutation(nil), state.Mutations...)
}

// ConsumeClearedMutations returns all collected mutations and clears the state on the returned context.
func ConsumeClearedMutations(ctx context.Context) (context.Context, []ClearedMutation) {
	state, ok := contextx.From[*clearedIDsState](ctx)
	if !ok || state == nil || len(state.Mutations) == 0 {
		return ctx, nil
	}

	mutations := state.Mutations
	state.Mutations = nil
	return ctx, mutations
}

func collectClearedRelationEntries(ctx context.Context, m ent.Mutation) []ClearedRelationEntry {
	var entries []ClearedRelationEntry

	// RemovedEdges captures explicit RemoveXIDs() calls that target specific IDs.
	for _, edge := range m.RemovedEdges() {
		ids := valuesToStrings(m.RemovedIDs(edge))
		if len(ids) == 0 {
			continue
		}

		sort.Strings(ids)
		entries = append(entries, ClearedRelationEntry{
			Name:   edge,
			Type:   RelationTypeEdge,
			Action: ActionEdgeRemoved,
			IDs:    ids,
		})
	}

	// ClearedEdges covers helper methods like ClearFiles() that remove all edge rows.
	for _, edge := range m.ClearedEdges() {
		entries = append(entries, ClearedRelationEntry{
			Name:   edge,
			Type:   RelationTypeEdge,
			Action: ActionEdgeCleared,
		})
	}

	if cleared, ok := m.(interface{ ClearedFields() []string }); ok {
		for _, field := range cleared.ClearedFields() {
			// OldField() gives us the pre-mutation ID for single-valued fields (avatar, logo, etc.).
			ids, err := oldFieldIDs(ctx, m, field)
			if err != nil {
				zerolog.Ctx(ctx).Debug().Err(err).Str("field", field).Msg("unable to load cleared field value")
				continue
			}

			sort.Strings(ids)
			if len(ids) == 0 {
				continue
			}

			entries = append(entries, ClearedRelationEntry{
				Name:   field,
				Type:   RelationTypeField,
				Action: ActionFieldCleared,
				IDs:    ids,
			})
		}
	}

	// For ordinary Set operations, compute the delta between previous and new IDs.
	for _, field := range m.Fields() {
		if !isIDLikeField(field) {
			continue
		}

		oldIDs, err := oldFieldIDs(ctx, m, field)
		if err != nil {
			zerolog.Ctx(ctx).Debug().Err(err).Str("field", field).Msg("unable to load prior field value")
			continue
		}
		if len(oldIDs) == 0 {
			continue
		}

		newIDs := newFieldIDs(m, field)
		removed := lo.Without(oldIDs, newIDs...)
		if len(removed) == 0 {
			continue
		}

		sort.Strings(removed)
		entries = append(entries, ClearedRelationEntry{
			Name:   field,
			Type:   RelationTypeField,
			Action: ActionFieldReplaced,
			IDs:    removed,
		})
	}

	if len(entries) == 0 {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return entries[i].Type < entries[j].Type
		}
		if entries[i].Name != entries[j].Name {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Action < entries[j].Action
	})

	return entries
}

func oldFieldIDs(ctx context.Context, m ent.Mutation, field string) ([]string, error) {
	mutation, ok := m.(interface {
		OldField(context.Context, string) (ent.Value, error)
	})
	if !ok {
		return nil, nil
	}

	value, err := mutation.OldField(ctx, field)
	if err != nil {
		// When Ent cannot read the prior value (e.g. during creates) we ignore this field.
		return nil, err
	}

	return singleValueToStrings(value), nil
}

func newFieldIDs(m ent.Mutation, field string) []string {
	value, ok := m.Field(field)
	if !ok {
		return nil
	}

	return singleValueToStrings(value)
}

func singleValueToStrings(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []string:
		return lo.Filter(v, func(id string, _ int) bool { return id != "" })
	case *string:
		if v == nil || *v == "" {
			return nil
		}
		return []string{*v}
	default:
		return nil
	}
}

func isIDLikeField(field string) bool {
	return strings.HasSuffix(field, "_id") || strings.HasSuffix(field, "_ids")
}

func valuesToStrings(values []ent.Value) []string {
	var ids []string
	for _, v := range values {
		if str, ok := v.(string); ok && str != "" {
			ids = append(ids, str)
		}
	}
	return ids
}

func ensureClearedState(ctx context.Context) (*clearedIDsState, context.Context) {
	state, ok := contextx.From[*clearedIDsState](ctx)
	if ok && state != nil {
		return state, ctx
	}

	state = &clearedIDsState{}
	ctx = contextx.With(ctx, state)
	return state, ctx
}

// HasClearedFileRelations reports whether the current context contains cleared relations that appear to reference files.
func HasClearedFileRelations(ctx context.Context) bool {
	return hasFileClears(GetClearedMutations(ctx))
}

func hasFileClears(mutations []ClearedMutation) bool {
	for _, mutation := range mutations {
		for _, entry := range mutation.Entries {
			if isFileRelationName(entry.Name) {
				return true
			}
		}
	}

	return false
}

func fileIDsFromMutations(mutations []ClearedMutation) []string {
	if len(mutations) == 0 {
		return nil
	}

	// Use a set to deduplicate the ids gathered from different relations.
	ids := make(map[string]struct{})
	for _, mutation := range mutations {
		for _, entry := range mutation.Entries {
			if !isFileRelationName(entry.Name) {
				continue
			}

			for _, id := range entry.IDs {
				if id == "" {
					continue
				}
				ids[id] = struct{}{}
			}
		}
	}

	if len(ids) == 0 {
		return nil
	}

	// Produce a stable list for event payloads / deterministic logging.
	collected := make([]string, 0, len(ids))
	for id := range ids {
		collected = append(collected, id)
	}

	sort.Strings(collected)
	return collected
}

func isFileRelationName(name string) bool {
	lower := strings.ToLower(name)
	if lower == "file" || lower == "files" {
		return true
	}

	if strings.Contains(lower, "_file_") || strings.Contains(lower, "file_id") {
		return true
	}

	if strings.HasSuffix(lower, "_file") || strings.HasSuffix(lower, "_files") || strings.HasSuffix(lower, "_file_id") || strings.HasSuffix(lower, "_file_ids") {
		return true
	}

	if strings.HasPrefix(lower, "file_") || strings.HasPrefix(lower, "files_") {
		return true
	}

	return false
}
