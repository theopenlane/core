package mutations

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/mapx"
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
	// OldValues captures pre-mutation values for changed fields on update operations
	OldValues map[string]any
}

// NewChangeSet builds a cloned change set from raw change payload values
func NewChangeSet(changedFields, changedEdges []string, addedIDs, removedIDs map[string][]string, proposedChanges, oldValues map[string]any) ChangeSet {
	return ChangeSet{
		ChangedFields:   append([]string(nil), changedFields...),
		ChangedEdges:    append([]string(nil), changedEdges...),
		AddedIDs:        mapx.CloneMapStringSlice(addedIDs),
		RemovedIDs:      mapx.CloneMapStringSlice(removedIDs),
		ProposedChanges: mapx.DeepCloneMapAny(proposedChanges),
		OldValues:       mapx.DeepCloneMapAny(oldValues),
	}
}

// Clone returns a copy of the change set and its map-backed values
func (set ChangeSet) Clone() ChangeSet {
	return NewChangeSet(set.ChangedFields, set.ChangedEdges, set.AddedIDs, set.RemovedIDs, set.ProposedChanges, set.OldValues)
}

// OldValueSource captures the mutation accessors needed to snapshot pre-update field values
type OldValueSource interface {
	Op() ent.Op
	OldField(ctx context.Context, name string) (ent.Value, error)
}

// BuildOldValues snapshots pre-mutation values for the given fields; it must be called
// before the mutation applies, while the database still holds the prior row. Only
// single-row updates carry old values — ent cannot resolve OldField for creates or
// bulk updates — and fields that fail to resolve are skipped
func BuildOldValues(ctx context.Context, source OldValueSource, fields []string) map[string]any {
	if source == nil || !source.Op().Is(ent.OpUpdateOne) || len(fields) == 0 {
		return nil
	}

	oldValues := make(map[string]any, len(fields))

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		if value, err := source.OldField(ctx, field); err == nil {
			oldValues[field] = value
		}
	}

	if len(oldValues) == 0 {
		return nil
	}

	return oldValues
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

	clearedSet := mapx.MapSetFromSlice(NormalizeStrings(source.ClearedFields()))

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
