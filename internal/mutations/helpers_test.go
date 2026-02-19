package mutations

import (
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMutation struct {
	fields  []string
	cleared []string
	values  map[string]any
}

func (m testMutation) Fields() []string {
	return m.fields
}

func (m testMutation) ClearedFields() []string {
	return m.cleared
}

func (m testMutation) Field(name string) (ent.Value, bool) {
	value, ok := m.values[name]
	return value, ok
}

func TestNormalizeStrings(t *testing.T) {
	require.Nil(t, NormalizeStrings(nil))
	require.Nil(t, NormalizeStrings([]string{}))
	assert.Equal(t, []string{"b", "a"}, NormalizeStrings([]string{" b ", "", "a", "b", "  "}))
}

func TestChangedAndClearedFields(t *testing.T) {
	changed, cleared := ChangedAndClearedFields(testMutation{
		fields:  []string{"name", "", "status"},
		cleared: []string{" status ", "owner", ""},
	})

	assert.Equal(t, []string{"status", "owner"}, cleared)
	assert.Equal(t, []string{"name", "status", "owner"}, changed)
}

func TestBuildProposedChanges(t *testing.T) {
	changes := BuildProposedChanges(testMutation{
		fields:  []string{"fieldA"},
		cleared: []string{"fieldB"},
		values: map[string]any{
			"fieldA": "value",
		},
	}, []string{"fieldA", "fieldB", "fieldC"})

	assert.Equal(t, map[string]any{
		"fieldA": "value",
		"fieldB": nil,
	}, changes)
}

func TestChangeSetClone(t *testing.T) {
	original := ChangeSet{
		ChangedFields: []string{"a"},
		ChangedEdges:  []string{"edge"},
		AddedIDs: map[string][]string{
			"edge": {"id1"},
		},
		RemovedIDs: map[string][]string{
			"edge": {"id2"},
		},
		ProposedChanges: map[string]any{
			"name": "value",
		},
	}

	cloned := original.Clone()
	require.Equal(t, original, cloned)

	cloned.ChangedFields[0] = "mutated"
	cloned.AddedIDs["edge"][0] = "mutated"
	cloned.ProposedChanges["name"] = "mutated"

	assert.Equal(t, "a", original.ChangedFields[0])
	assert.Equal(t, "id1", original.AddedIDs["edge"][0])
	assert.Equal(t, "value", original.ProposedChanges["name"])
}
