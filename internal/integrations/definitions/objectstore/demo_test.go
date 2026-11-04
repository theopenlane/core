package objectstore

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestDemoTestDatasetIsDeletable(t *testing.T) {
	_, _, manifest := loadTestDemoManifest(t)

	for _, name := range manifest.Order {
		schema, ok := entityops.LookupSchema(name)
		assert.Assert(t, ok, name)
		assert.Check(t, schema.DeleteBySource != nil, "%s has no delete by source", name)
	}
}

func TestDemoDatasetPrefix(t *testing.T) {
	testCases := []struct {
		dataset     DemoDataset
		prefix      string
		expectedErr error
	}{
		{dataset: demoTestDataset, prefix: "demo/minimal"},
		{dataset: "soc2", prefix: "demo/soc2"},
		{dataset: "", expectedErr: ErrDemoDatasetInvalid},
		{dataset: "../vendors", expectedErr: ErrDemoDatasetInvalid},
		{dataset: "soc2/../..", expectedErr: ErrDemoDatasetInvalid},
		{dataset: "SOC2", expectedErr: ErrDemoDatasetInvalid},
		{dataset: "soc 2", expectedErr: ErrDemoDatasetInvalid},
		{dataset: "demo/soc2", expectedErr: ErrDemoDatasetInvalid},
	}

	for _, tc := range testCases {
		t.Run(string(tc.dataset), func(t *testing.T) {
			prefix, err := tc.dataset.Prefix()
			if tc.expectedErr != nil {
				assert.Check(t, errors.Is(err, tc.expectedErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.prefix, prefix)
		})
	}
}

func TestDemoSeedUniqueKey(t *testing.T) {
	testCases := []struct {
		name     string
		profile  string
		expected string
	}{
		{name: "default profile", expected: "demo-seed:org_1:soc2:"},
		{name: "draft profile", profile: "DRAFT", expected: "demo-seed:org_1:soc2:DRAFT"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, DemoSeedUniqueKey("org_1", "soc2", tc.profile))
		})
	}
}

func TestDemoProfileName(t *testing.T) {
	testCases := []struct {
		name        string
		profile     enums.DemoDataProfile
		expected    string
		expectedErr error
	}{
		{name: "omitted profile uses the record defaults", expected: ""},
		{name: "demo profile uses the record defaults", profile: enums.DemoDataProfileDemo, expected: ""},
		{name: "draft profile", profile: enums.DemoDataProfileDraft, expected: "DRAFT"},
		{name: "unknown profile", profile: enums.DemoDataProfileInvalid, expectedErr: ErrDemoProfileUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile, err := demoProfileName(tc.profile)
			if tc.expectedErr != nil {
				assert.Check(t, errors.Is(err, tc.expectedErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.expected, profile)
		})
	}
}
