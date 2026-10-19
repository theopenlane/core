package gemini

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/docextract/soc2"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
)

func TestAcceptedSectionsFiltersEditsAndAllowlists(t *testing.T) {
	report := map[string]any{
		soc2.ControlsPart: []any{
			map[string]any{"ref_code": "ACF-01", "title": "stored", "owner_id": "org_other"},
			map[string]any{"ref_code": "ACF-02", "title": "rejected"},
		},
		soc2.FindingsPart: []any{
			map[string]any{"external_id": "f1", "description": "exception", "review_external_id": "r1"},
		},
	}

	accepted := acceptedSections(report, ReportScanImport{
		Accepted: map[string][]string{soc2.ControlsPart: {"ACF-01"}, soc2.FindingsPart: {"f1"}},
		Edited:   map[string][]map[string]any{soc2.ControlsPart: {{"ref_code": "ACF-01", "title": "edited", "ingest_target_id": "ctl_other"}}},
	})

	assert.Check(t, cmp.Len(accepted[soc2.ControlsPart], 1))
	assert.Check(t, cmp.Equal(accepted[soc2.ControlsPart][0]["title"], "edited"))

	_, ownerKept := accepted[soc2.ControlsPart][0]["owner_id"]
	_, targetKept := accepted[soc2.ControlsPart][0]["ingest_target_id"]
	assert.Check(t, !ownerKept)
	assert.Check(t, !targetKept)

	assert.Check(t, cmp.Equal(accepted[soc2.FindingsPart][0]["external_id"], "f1"))
	assert.Check(t, cmp.Equal(accepted[soc2.FindingsPart][0]["review_external_id"], "r1"))
}

func TestVendorExternalIDPrefersExistingThenDomain(t *testing.T) {
	existing := map[string]string{"purr.example": "vendor-purr"}

	assert.Check(t, cmp.Equal(vendorExternalID(map[string]any{"name": "purr", "domains": []any{"purr.example"}}, existing), "vendor-purr"))
	assert.Check(t, cmp.Equal(vendorExternalID(map[string]any{"name": "meow", "domains": []any{"Cloud.Meow.example."}}, existing), "cloud.meow.example"))
	assert.Check(t, cmp.Equal(vendorExternalID(map[string]any{"name": "Purr"}, existing), "purr"))
}

func TestImportVendorKeysByNameAndDisplayName(t *testing.T) {
	keys := importVendorKeys([]map[string]any{{"name": "meow", "display_name": "Meow Cloud Services", "domains": []any{"cloud.meow.example"}}}, map[string]string{"cloud.meow.example": "vendor-meow"})

	assert.Check(t, cmp.Equal(keys["meow"], "vendor-meow"))
	assert.Check(t, cmp.Equal(keys["meow cloud services"], "vendor-meow"))
}

func TestVendorPayloadUsesExistingEntityExternalID(t *testing.T) {
	imp := reportImport{existingVendors: map[string]string{"cloud.meow.example": "vendor-meow"}}

	existing := imp.vendor(map[string]any{"name": "meow", "domains": []any{"cloud.meow.example"}})
	assert.Check(t, cmp.Equal(existing[entityops.EntityFields.ExternalID.Name], "vendor-meow"))

	_, domainsKept := existing["domains"]
	assert.Check(t, !domainsKept)

	created := imp.vendor(map[string]any{"name": "purr", "domains": []any{"purr.example"}})
	assert.Check(t, cmp.Equal(created[entityops.EntityFields.ExternalID.Name], "purr.example"))
	assert.Check(t, cmp.DeepEqual(created["domains"], []any{"purr.example"}))
}

func TestReportImportPayloads(t *testing.T) {
	imp := reportImport{programID: "prg_1", vendorTypeID: "type_vendor", vendorKeys: map[string]string{"meow cloud services": "cloud.meow.example"}}

	ctl := imp.control(map[string]any{"ref_code": "ACF-01", "soc2_mapping": []any{"CC6.1"}})
	_, mappingKept := ctl["soc2_mapping"]
	assert.Check(t, !mappingKept)
	assert.Check(t, cmp.Equal(ctl[entityops.ControlFields.Status.Name], enums.ControlStatusApproved))
	assert.Check(t, cmp.DeepEqual(ctl["program_ids"], []string{"prg_1"}))

	asset := imp.asset(map[string]any{"name": "prod-api", "entity_name": "Meow Cloud Services"})
	assert.Check(t, cmp.Equal(asset[entityops.AssetFields.SourceIdentifier.Name], "prod-api"))
	assert.Check(t, cmp.DeepEqual(asset[vendorKeysField], []string{"cloud.meow.example"}))

	finding := imp.finding(map[string]any{"external_id": "f1", "ref_codes": []any{"ACF-01", "CC6.1"}, "review_external_id": "r1", "test_details": "x"})
	assert.Check(t, cmp.DeepEqual(finding[controlRefCodesField], []any{"ACF-01", "CC6.1"}))
	assert.Check(t, cmp.DeepEqual(finding[reviewExternalIDsField], []string{"r1"}))

	_, detailsKept := finding["test_details"]
	assert.Check(t, !detailsKept)
}

func TestReportImportPayloadsWithoutProgram(t *testing.T) {
	imp := reportImport{}

	for _, payload := range []map[string]any{
		imp.control(map[string]any{"ref_code": "ACF-01"}),
		imp.systemDetail(map[string]any{"system_name": "Acme Platform"}),
		imp.review(map[string]any{"external_id": "r1"}),
		imp.finding(map[string]any{"external_id": "f1"}),
	} {
		_, linked := payload["program_ids"]
		assert.Check(t, !linked)
	}
}

func TestMappedControlsPairEachCriterionWithItsFrameworkControl(t *testing.T) {
	imp := reportImport{frameworkIDs: map[string]string{"CC6.1": "ctl_cc61", "CC6.2": "ctl_cc62"}}

	set, err := imp.mappedControls([]map[string]any{{"ref_code": "ACF-01", "soc2_mapping": []any{"CC6.1", "CC6.2", "CC9.9"}}}, "file1")
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(set.Schema, entityops.SchemaMappedControl.Name))
	assert.Check(t, cmp.Equal(set.SourceInstanceID, "file1"))
	assert.Check(t, cmp.Len(set.Envelopes, 2))
	assert.Check(t, cmp.Equal(set.Envelopes[0].Resource, "ACF-01>CC6.1"))
	assert.Check(t, cmp.Equal(entityops.FieldValue(set.Envelopes[0].Payload, entityops.MappedControlFields.MappingType.Name), "SUBSET"))
}

func TestReportScanImportMappingsLinksAreValid(t *testing.T) {
	for _, mapping := range reportScanImportMappings() {
		schema, ok := entityops.LookupSchema(mapping.Schema)
		assert.Assert(t, ok, mapping.Schema)
		assert.Check(t, registry.ValidateLinkRules(schema, mapping.Spec.Links), mapping.Schema)
	}
}
