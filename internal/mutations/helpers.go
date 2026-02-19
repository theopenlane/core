package mutations

import (
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/helpers"
)

// ChangeSet captures mutation deltas shared across eventing and workflow trigger contexts
type ChangeSet struct {
	// ChangedFields captures updated/cleared fields for the mutation
	ChangedFields []string
	// ChangedEdges captures changed edge names
	ChangedEdges []string
	// AddedIDs captures edge IDs added by edge name
	AddedIDs map[string][]string
	// RemovedIDs captures edge IDs removed by edge name
	RemovedIDs map[string][]string
	// ProposedChanges captures field-level proposed values
	ProposedChanges map[string]any
}

// NewChangeSet builds a cloned change set from raw change payload values
func NewChangeSet(changedFields, changedEdges []string, addedIDs, removedIDs map[string][]string, proposedChanges map[string]any) ChangeSet {
	return ChangeSet{
		ChangedFields:   append([]string(nil), changedFields...),
		ChangedEdges:    append([]string(nil), changedEdges...),
		AddedIDs:        CloneStringSliceMap(addedIDs),
		RemovedIDs:      CloneStringSliceMap(removedIDs),
		ProposedChanges: CloneAnyMap(proposedChanges),
	}
}

// Clone returns a copy of the change set and its map-backed values
func (set ChangeSet) Clone() ChangeSet {
	return NewChangeSet(set.ChangedFields, set.ChangedEdges, set.AddedIDs, set.RemovedIDs, set.ProposedChanges)
}

// FieldChangeSource captures the mutation accessors needed to derive changed and cleared field lists
type FieldChangeSource interface {
	Fields() []string
	ClearedFields() []string
}

// ProposedChangeSource captures the mutation accessors needed to build proposed changes
type ProposedChangeSource interface {
	ClearedFields() []string
	Field(string) (ent.Value, bool)
}

// ChangedAndClearedFields returns normalized changed and cleared field lists from a mutation source
func ChangedAndClearedFields(source FieldChangeSource) (changed []string, cleared []string) {
	if source == nil {
		return nil, nil
	}

	cleared = NormalizeStrings(source.ClearedFields())
	changed = NormalizeStrings(append(append([]string(nil), source.Fields()...), cleared...))
	return changed, cleared
}

// BuildProposedChanges materializes mutation values including explicit clears
func BuildProposedChanges(source ProposedChangeSource, changedFields []string) map[string]any {
	if source == nil || len(changedFields) == 0 {
		return nil
	}

	clearedSet := lo.SliceToMap(NormalizeStrings(source.ClearedFields()), func(field string) (string, struct{}) {
		return field, struct{}{}
	})

	proposed := make(map[string]any, len(changedFields))
	for _, field := range changedFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		if value, ok := source.Field(field); ok {
			proposed[field] = value
			continue
		}

		if _, ok := clearedSet[field]; ok {
			proposed[field] = nil
		}
	}

	if len(proposed) == 0 {
		return nil
	}

	return proposed
}

// NormalizeStrings trims, deduplicates, and drops empty string values
func NormalizeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalized := lo.Uniq(lo.FilterMap(values, func(value string, _ int) (string, bool) {
		value = strings.TrimSpace(value)
		return value, value != ""
	}))
	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

// CloneStringSliceMap deep-copies map values while dropping blank keys
func CloneStringSliceMap(values map[string][]string) map[string][]string {
	if len(values) == 0 {
		return nil
	}

	filtered := lo.PickBy(values, func(key string, _ []string) bool { return strings.TrimSpace(key) != "" })
	cloned := lo.MapValues(filtered, func(list []string, _ string) []string { return append([]string(nil), list...) })
	if len(cloned) == 0 {
		return nil
	}

	return cloned
}

// CloneAnyMap deep-copies map values while dropping blank keys
func CloneAnyMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	filtered := lo.PickBy(values, func(key string, _ any) bool { return strings.TrimSpace(key) != "" })
	if len(filtered) == 0 {
		return nil
	}

	return helpers.DeepCloneMap(filtered)
}
