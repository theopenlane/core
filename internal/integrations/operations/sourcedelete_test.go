package operations

import (
	"context"
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestSourceDeleteOrder(t *testing.T) {
	t.Parallel()

	order := []string{"custom_type_enum", "program", "control", "internal_policy", "procedure"}

	testCases := []struct {
		name    string
		keep    []string
		want    []string
		wantErr error
	}{
		{
			name: "keep none",
			want: []string{"procedure", "internal_policy", "control", "program", "custom_type_enum"},
		},
		{
			name: "keep one",
			keep: []string{"internal_policy"},
			want: []string{"procedure", "control", "program", "custom_type_enum"},
		},
		{
			name: "keep all but one",
			keep: []string{"custom_type_enum", "program", "control", "internal_policy"},
			want: []string{"procedure"},
		},
		{
			name:    "keep unknown schema",
			keep:    []string{"custom_type_enums"},
			wantErr: ErrSourceKeepUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := sourceDeleteOrder(order, tc.keep)
			if tc.wantErr != nil {
				assert.Check(t, errors.Is(err, tc.wantErr))
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, tc.want, got)
			assert.DeepEqual(t, []string{"custom_type_enum", "program", "control", "internal_policy", "procedure"}, order)
		})
	}
}

func TestSourceDeleteOrderDefaultsToDeletableSchemas(t *testing.T) {
	t.Parallel()

	deletable := deletableSchemas()
	assert.Check(t, len(deletable) > 0)

	got, err := sourceDeleteOrder(nil, nil)
	assert.NilError(t, err)
	assert.Check(t, len(got) == len(deletable))

	for _, name := range got {
		schema, ok := entityops.LookupSchema(name)
		assert.Assert(t, ok, name)
		assert.Check(t, schema.DeleteBySource != nil, name)
	}

	kept, err := sourceDeleteOrder(nil, []string{deletable[0]})
	assert.NilError(t, err)
	assert.Check(t, len(kept) == len(deletable)-1)

	_, err = sourceDeleteOrder(nil, []string{"not_a_schema"})
	assert.Check(t, errors.Is(err, ErrSourceKeepUnknown))
}

func TestDeleteSourceRecordsFilterRequired(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		filter entityops.SourceFilter
	}{
		{name: "empty filter"},
		{name: "missing instance and key", filter: entityops.SourceFilter{DefinitionID: "def_objectstore"}},
		{name: "missing definition with instance", filter: entityops.SourceFilter{InstanceID: "tenant-1"}},
		{name: "missing definition with key", filter: entityops.SourceFilter{Key: "openlane_demodata_minimal_test"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deleted, err := deleteSourceRecords(context.Background(), nil, tc.filter, []string{"program"}, nil)
			assert.Check(t, errors.Is(err, ErrSourceFilterRequired))
			assert.Check(t, len(deleted) == 0)
		})
	}
}
