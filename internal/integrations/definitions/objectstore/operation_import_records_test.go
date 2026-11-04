package objectstore

import (
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestObjectSchema(t *testing.T) {
	testCases := []struct {
		name        string
		prefix      string
		key         string
		expected    *entityops.Schema
		expectedErr error
	}{
		{name: "directory names a schema", prefix: "records/", key: "records/controls/soc2.json", expected: entityops.SchemaControl},
		{name: "prefix without a trailing slash", prefix: "records", key: "records/programs/programs.json", expected: entityops.SchemaProgram},
		{name: "snake case directory", prefix: "records/", key: "records/internal_policies/policies.json", expected: entityops.SchemaInternalPolicy},
		{name: "entities directory", prefix: "records/", key: "records/entities/customers.json", expected: entityops.SchemaEntity},
		{name: "nested directory uses its parent", prefix: "records/", key: "records/2026/controls/soc2.json", expected: entityops.SchemaControl},
		{name: "vendors directory is entity", prefix: "records/", key: "records/vendors/catalog.json", expected: entityops.SchemaEntity},
		{name: "object under a vendors prefix is entity", prefix: "vendors/", key: "vendors/catalog.json", expected: entityops.SchemaEntity},
		{name: "object under a prefix naming no schema is entity", prefix: "vendors-updated/", key: "vendors-updated/catalog.json", expected: entityops.SchemaEntity},
		{name: "object under a prefix naming a schema", prefix: "controls/", key: "controls/soc2.json", expected: entityops.SchemaControl},
		{name: "singular directory", prefix: "records/", key: "records/control/soc2.json", expectedErr: ErrRecordSchemaUnknown},
		{name: "unknown directory", prefix: "records/", key: "records/controlz/soc2.json", expectedErr: ErrRecordSchemaUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := objectSchema(tc.prefix, tc.key)
			if tc.expectedErr != nil {
				assert.Check(t, errors.Is(err, tc.expectedErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.expected.Name, schema.Name)
		})
	}
}

func TestRecordObjectsImportTargetsFirst(t *testing.T) {
	objects, err := recordObjects("records/", []string{"records/controls/soc2.json", "records/custom_type_enums/kinds.json", "records/vendors/catalog.json"})
	assert.NilError(t, err)

	keys := lo.Map(objects, func(object recordObject, _ int) string { return object.key })

	assert.Check(t, lo.IndexOf(keys, "records/custom_type_enums/kinds.json") < lo.IndexOf(keys, "records/controls/soc2.json"))
}

func TestRecordObjectsUnknownDirectory(t *testing.T) {
	_, err := recordObjects("records/", []string{"records/controls/soc2.json", "records/bananas/fruit.json"})
	assert.Check(t, errors.Is(err, ErrRecordSchemaUnknown))
}
