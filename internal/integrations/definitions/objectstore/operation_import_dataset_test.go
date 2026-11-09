package objectstore

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/placeholder"
	"github.com/theopenlane/utils/jsonx"
)

const (
	demoTestBucket  = "../../../graphapi/testharness/testdata/objectstore"
	demoTestDataset = "minimal"
)

func loadTestDemoManifest(t *testing.T) (storage.Provider, string, datasetManifest) {
	t.Helper()

	ctx := context.Background()

	provider, err := newProvider(ctx, storage.DiskProvider, storage.Providers{Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: demoTestBucket}}})
	assert.NilError(t, err)

	prefix, err := DemoDataset(demoTestDataset).Prefix()
	assert.NilError(t, err)

	manifest, err := loadDatasetManifest(ctx, provider, prefix)
	assert.NilError(t, err)

	return provider, prefix, manifest
}

func TestDatasetPrefix(t *testing.T) {
	testCases := []struct {
		prefix      string
		expected    string
		expectedErr error
	}{
		{prefix: "demo/minimal", expected: "demo/minimal"},
		{prefix: "templates/controls", expected: "templates/controls"},
		{prefix: "risk", expected: "risk"},
		{prefix: "demo/soc2/", expected: "demo/soc2"},
		{prefix: "", expectedErr: ErrDatasetInvalid},
		{prefix: "../vendors", expectedErr: ErrDatasetInvalid},
		{prefix: "demo/../..", expectedErr: ErrDatasetInvalid},
		{prefix: "/demo/soc2", expectedErr: ErrDatasetInvalid},
		{prefix: "Demo/soc2", expectedErr: ErrDatasetInvalid},
		{prefix: "demo/soc 2", expectedErr: ErrDatasetInvalid},
	}

	for _, tc := range testCases {
		t.Run(tc.prefix, func(t *testing.T) {
			prefix, err := datasetPrefix(tc.prefix)
			if tc.expectedErr != nil {
				assert.Check(t, errors.Is(err, tc.expectedErr))
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tc.expected, prefix)
		})
	}
}

func TestLoadDatasetPayloadSet(t *testing.T) {
	ctx := context.Background()
	provider, prefix, manifest := loadTestDemoManifest(t)

	assert.Check(t, len(manifest.Order) > 0)

	contracted := map[string]bool{}
	for _, contract := range importContracts {
		contracted[contract.Schema] = true
	}

	fill := placeholder.Variables{"ORG_NAME": "Acme Corp", "ORG_DOMAIN": "acme.test", "ORG_OWNER_USER_ID": "user_1", "TODAY": time.Now().UTC().Format(time.DateOnly)}

	for _, name := range manifest.Order {
		payloadSet, err := loadDatasetPayloadSet(ctx, provider, prefix, name, fill)
		assert.NilError(t, err, name)
		assert.Check(t, contracted[payloadSet.Schema], "%s is not an import contract", payloadSet.Schema)
		assert.Check(t, payloadSet.SourceKey == prefix)
		assert.Check(t, len(payloadSet.Envelopes) > 0, name)

		for _, envelope := range payloadSet.Envelopes {
			assert.Check(t, externalID(t, envelope) != "", envelope.Resource)
			assert.Check(t, !strings.Contains(string(envelope.Payload), "{{"), envelope.Resource)
		}
	}

	_, err := loadDatasetPayloadSet(ctx, provider, prefix, "not_a_schema", fill)
	assert.Check(t, errors.Is(err, ErrDatasetSchemaUnknown))
}

func TestDatasetProfiles(t *testing.T) {
	ctx := context.Background()
	provider, prefix, manifest := loadTestDemoManifest(t)

	testCases := []struct {
		name             string
		profile          string
		expectedStatus   string
		expectedApprover string
		expectedErr      error
	}{
		{name: "no profile keeps the record values", expectedStatus: "PUBLISHED", expectedApprover: "Admins"},
		{name: "demo profile keeps the record values", profile: "DEMO", expectedStatus: "PUBLISHED", expectedApprover: "Admins"},
		{name: "draft profile", profile: "DRAFT", expectedStatus: "DRAFT", expectedApprover: ""},
		{name: "unknown profile", profile: "MISSING", expectedErr: ErrDatasetProfileUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			variables, err := manifest.profileVariables(tc.profile)
			if tc.expectedErr != nil {
				assert.Check(t, errors.Is(err, tc.expectedErr))
				return
			}

			assert.NilError(t, err)

			fill := placeholder.Variables{"TODAY": time.Now().UTC().Format(time.DateOnly)}.With(variables)

			payloadSet, err := loadDatasetPayloadSet(ctx, provider, prefix, "internal_policy", fill)
			assert.NilError(t, err)

			var record struct {
				Status   string `json:"status"`
				Approver string `json:"approver_name"`
			}

			assert.NilError(t, json.Unmarshal(payloadSet.Envelopes[0].Payload, &record))
			assert.Equal(t, tc.expectedStatus, record.Status)
			assert.Equal(t, tc.expectedApprover, record.Approver)
		})
	}
}

func TestLinkStandardControls(t *testing.T) {
	refCodes := map[string][]string{"program::soc2": {"CC1.1", "CC1.2"}}

	testCases := []struct {
		name     string
		schema   string
		payload  string
		expected []string
	}{
		{name: "program keeps its own ref codes and gains the cloned ones", schema: entityops.SchemaProgram.Name, payload: `{"external_id":"program::soc2","control_ref_codes":["OL-01.01"]}`, expected: []string{"OL-01.01", "CC1.1", "CC1.2"}},
		{name: "program with only cloned controls", schema: entityops.SchemaProgram.Name, payload: `{"external_id":"program::soc2"}`, expected: []string{"CC1.1", "CC1.2"}},
		{name: "program without cloned controls is untouched", schema: entityops.SchemaProgram.Name, payload: `{"external_id":"program::other","control_ref_codes":["OL-01.01"]}`, expected: []string{"OL-01.01"}},
		{name: "other schemas are untouched", schema: entityops.SchemaTask.Name, payload: `{"external_id":"program::soc2"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payloadSets := []types.IngestPayloadSet{{Schema: tc.schema, Envelopes: []types.MappingEnvelope{{Payload: json.RawMessage(tc.payload)}}}}

			assert.NilError(t, linkStandardControls(payloadSets, refCodes))

			linked, _ := jsonx.DecodeObjectKey[[]string](payloadSets[0].Envelopes[0].Payload, recordControlRefCodesKey)
			assert.DeepEqual(t, tc.expected, linked)
		})
	}
}

func TestDatasetBuiltinVariables(t *testing.T) {
	variables, err := DatasetBuiltinVariables{OrgName: "Acme", OrgSlug: "acme", OrgDomain: "acme.test", OwnerUserID: "user_1", OwnerEmail: "owner@acme.test", OwnerName: "Owner", Today: "2026-11-04"}.variables()
	assert.NilError(t, err)

	assert.DeepEqual(t, placeholder.Variables{
		"ORG_NAME":          "Acme",
		"ORG_SLUG":          "acme",
		"ORG_DOMAIN":        "acme.test",
		"ORG_OWNER_USER_ID": "user_1",
		"ORG_OWNER_EMAIL":   "owner@acme.test",
		"ORG_OWNER_NAME":    "Owner",
		"TODAY":             "2026-11-04",
	}, variables)

	schema := jsonx.SchemaFrom[DatasetBuiltinVariables]()
	for name := range variables {
		assert.Check(t, strings.Contains(string(schema), `"`+name+`"`), name)
	}
}

func externalID(t *testing.T, envelope types.MappingEnvelope) string {
	t.Helper()

	var record struct {
		ExternalID string `json:"external_id"`
	}

	assert.NilError(t, json.Unmarshal(envelope.Payload, &record))

	return record.ExternalID
}
