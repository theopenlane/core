package operations

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func TestSourceDeleteOrder(t *testing.T) {
	t.Parallel()

	deletable := deletableSchemas()
	assert.Check(t, len(deletable) > 0)

	testCases := []struct {
		name    string
		keep    []string
		want    int
		wantErr error
	}{
		{name: "keep none", want: len(deletable)},
		{name: "keep one", keep: []string{"internal_policy"}, want: len(deletable) - 1},
		{name: "keep unknown schema", keep: []string{"custom_type_enums"}, wantErr: ErrSourceKeepUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := sourceDeleteOrder(tc.keep)
			if tc.wantErr != nil {
				assert.Check(t, errors.Is(err, tc.wantErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.want, len(got))

			for _, name := range tc.keep {
				assert.Check(t, !lo.Contains(got, name), name)
			}

			for _, name := range got {
				schema, ok := entityops.LookupSchema(name)
				assert.Assert(t, ok, name)
				assert.Check(t, schema.DeleteBySource != nil, name)
			}
		})
	}
}

func TestSourceDeleteOrderReferencingRowsFirst(t *testing.T) {
	t.Parallel()

	got, err := sourceDeleteOrder(nil)
	assert.NilError(t, err)

	position := lo.SliceToMap(got, func(name string) (string, int) { return name, lo.IndexOf(got, name) })

	for _, name := range got {
		schema, _ := entityops.LookupSchema(name)

		for _, edge := range schema.Edges {
			target, deletable := position[lo.FromPtr(edge.Target).Snake]
			if edge.Field == "" || edge.Target == schema || !deletable {
				continue
			}

			if _, cyclic := lo.Find(edge.Target.Edges, func(back entityops.EdgeDescriptor) bool { return back.Field != "" && back.Target == schema }); cyclic {
				continue
			}

			assert.Check(t, position[name] < target, "%s must be deleted before %s", name, edge.Target.Snake)
		}
	}

	assert.Check(t, position[entityops.SchemaControl.Snake] < position[entityops.SchemaCustomTypeEnum.Snake])
}

func TestSourceDefinitionID(t *testing.T) {
	t.Parallel()

	runtimeCtx := gala.OperationContextKey.Set(context.Background(), types.NewOperationContext("", DeleteSourceOp.Name(), types.IntegrationSource{DefinitionID: "def_runtime", Runtime: true}))

	testCases := []struct {
		name        string
		ctx         context.Context
		request     types.OperationRequest
		expected    string
		expectedErr error
	}{
		{name: "installation definition", ctx: context.Background(), request: types.OperationRequest{Integration: &ent.Integration{DefinitionID: "def_installed"}}, expected: "def_installed"},
		{name: "runtime definition", ctx: runtimeCtx, expected: "def_runtime"},
		{name: "runtime without an operation context", ctx: context.Background(), expectedErr: ErrSourceOperationContextRequired},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := sourceDefinitionID(tc.ctx, tc.request)
			if tc.expectedErr != nil {
				assert.Check(t, errors.Is(err, tc.expectedErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
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

			deleted, err := deleteSourceRecords(context.Background(), nil, tc.filter, nil)
			assert.Check(t, errors.Is(err, ErrSourceFilterRequired))
			assert.Check(t, len(deleted) == 0)
		})
	}
}
