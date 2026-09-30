//go:build test

package eventstest_test

import (
	"encoding/json"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/findingcontrol"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// TestFindingRawPayloadOnlyReingestNoop verifies a Volatile-only raw_payload difference never writes on its own but rides along on a material change
func TestFindingRawPayloadOnlyReingestNoop(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	counts := catalogEventCounters(t, entityops.SchemaFinding)
	defer counts.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Raw Payload Test").
		SetKind("findrawpayloadtest").
		SetDefinitionID("def_findrawpayloadtest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findrawpayloadtest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	const externalID = "find-rawpayload-1"

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalID(externalID)).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := jsonx.EditObject(findingCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["raw_payload"] = json.RawMessage(`{"a":1}`)
		return true
	})

	result := ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, create)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Creates.Load()))
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()))

	created := findingByExternalID(ctx, t, externalID)
	assert.Check(t, is.Equal(float64(1), created.RawPayload["a"]))

	volatileOnly := jsonx.EditObject(findingCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["raw_payload"] = json.RawMessage(`{"a":2}`)
		doc["source_updated_at"] = json.RawMessage(`"2024-06-01T00:00:00Z"`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, volatileOnly)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Changed), "a volatile-only payload must not count as changed")
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()), "a volatile-only payload must not emit an update mutation event")

	afterVolatileOnly := findingByExternalID(ctx, t, externalID)
	assert.Check(t, afterVolatileOnly.UpdatedAt.Equal(created.UpdatedAt), "a volatile-only payload must not rewrite the row")
	assert.Check(t, is.Equal(float64(1), afterVolatileOnly.RawPayload["a"]), "an unwritten volatile field must keep its stored value")

	materialAndVolatile := jsonx.EditObject(findingCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["description"] = json.RawMessage(`"changed desc"`)
		doc["raw_payload"] = json.RawMessage(`{"a":2}`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, materialAndVolatile)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a material change must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Updates.Load()), "a material change must emit an update mutation event")

	after := findingByExternalID(ctx, t, externalID)
	assert.Check(t, is.Equal("changed desc", after.Description))
	assert.Check(t, is.Equal(float64(2), after.RawPayload["a"]), "the volatile field must ride along once a material field changed")
}

// TestVulnerabilityVolatileOnlyReingestNoop verifies a Volatile-only raw_payload difference never writes on its own but rides along on a material change
func TestVulnerabilityVolatileOnlyReingestNoop(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	counts := catalogEventCounters(t, entityops.SchemaVulnerability)
	defer counts.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Vulnerability Volatile Test").
		SetKind("vulnvolatiletest").
		SetDefinitionID("def_vulnvolatiletest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-vulnvolatiletest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	const externalID = "vuln-volonly-1"

	t.Cleanup(func() {
		vulns, err := suite.Client.DB.Vulnerability.Query().Where(vulnerability.ExternalID(externalID)).All(ctx)
		th.RequireNoError(t, err)

		if len(vulns) > 0 {
			(&th.Cleanup[*ent.VulnerabilityDeleteOne]{Client: suite.Client.DB.Vulnerability, IDs: lo.Map(vulns, func(v *ent.Vulnerability, _ int) string { return v.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := jsonx.EditObject(vulnerabilityCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["raw_payload"] = json.RawMessage(`{"a":1}`)
		return true
	})

	result := ingestCatalogPayload(ctx, t, installation, entityops.SchemaVulnerability.Name, create)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Creates.Load()))
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()))

	created := vulnerabilityByExternalID(ctx, t, externalID)
	assert.Check(t, is.Equal(float64(1), created.RawPayload["a"]))

	volatileOnly := jsonx.EditObject(vulnerabilityCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["raw_payload"] = json.RawMessage(`{"a":2}`)
		doc["source_updated_at"] = json.RawMessage(`"2024-06-01T00:00:00Z"`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaVulnerability.Name, volatileOnly)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Changed), "a volatile-only payload must not count as changed")
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()), "a volatile-only payload must not emit an update mutation event")

	afterVolatileOnly := vulnerabilityByExternalID(ctx, t, externalID)
	assert.Check(t, afterVolatileOnly.UpdatedAt.Equal(created.UpdatedAt), "a volatile-only payload must not rewrite the row")
	assert.Check(t, is.Equal(float64(1), afterVolatileOnly.RawPayload["a"]), "an unwritten volatile field must keep its stored value")

	materialAndVolatile := jsonx.EditObject(vulnerabilityCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["description"] = json.RawMessage(`"changed desc"`)
		doc["raw_payload"] = json.RawMessage(`{"a":2}`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaVulnerability.Name, materialAndVolatile)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a material change must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Updates.Load()), "a material change must emit an update mutation event")

	after := vulnerabilityByExternalID(ctx, t, externalID)
	assert.Check(t, is.Equal("changed desc", after.Description))
	assert.Check(t, is.Equal(float64(2), after.RawPayload["a"]), "the volatile field must ride along once a material field changed")
}

// TestAssetObservedAtOnlyReingestNoop verifies an observed_at-only difference never writes on its own but rides along on a material change
func TestAssetObservedAtOnlyReingestNoop(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	counts := catalogEventCounters(t, entityops.SchemaAsset)
	defer counts.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Asset Observed At Test").
		SetKind("assetobsattest").
		SetDefinitionID("def_assetobsattest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-assetobsattest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	const sourceIdentifier = "asset-obs-1"

	t.Cleanup(func() {
		assets, err := suite.Client.DB.Asset.Query().Where(asset.SourceIdentifier(sourceIdentifier)).All(ctx)
		th.RequireNoError(t, err)

		if len(assets) > 0 {
			(&th.Cleanup[*ent.AssetDeleteOne]{Client: suite.Client.DB.Asset, IDs: lo.Map(assets, func(a *ent.Asset, _ int) string { return a.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := jsonx.EditObject(assetCatalogPayload(sourceIdentifier), func(doc map[string]json.RawMessage) bool {
		doc["observed_at"] = json.RawMessage(`"2024-01-01T00:00:00Z"`)
		return true
	})

	result := ingestCatalogPayload(ctx, t, installation, entityops.SchemaAsset.Name, create)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Creates.Load()))
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()))

	created := catalogAssetBySourceIdentifier(ctx, t, sourceIdentifier)
	assert.Assert(t, created.ObservedAt != nil)

	volatileOnly := jsonx.EditObject(assetCatalogPayload(sourceIdentifier), func(doc map[string]json.RawMessage) bool {
		doc["observed_at"] = json.RawMessage(`"2024-06-01T00:00:00Z"`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaAsset.Name, volatileOnly)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Changed), "a volatile-only payload must not count as changed")
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()), "a volatile-only payload must not emit an update mutation event")

	afterVolatileOnly := catalogAssetBySourceIdentifier(ctx, t, sourceIdentifier)
	assert.Check(t, afterVolatileOnly.UpdatedAt.Equal(created.UpdatedAt), "a volatile-only payload must not rewrite the row")
	assert.Check(t, time.Time(*afterVolatileOnly.ObservedAt).Equal(time.Time(*created.ObservedAt)), "an unwritten volatile field must keep its stored value")

	materialAndVolatile := jsonx.EditObject(assetCatalogPayload(sourceIdentifier), func(doc map[string]json.RawMessage) bool {
		doc["name"] = json.RawMessage(`"Catalog Asset Two"`)
		doc["observed_at"] = json.RawMessage(`"2024-06-01T00:00:00Z"`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaAsset.Name, materialAndVolatile)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a material change must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Updates.Load()), "a material change must emit an update mutation event")

	after := catalogAssetBySourceIdentifier(ctx, t, sourceIdentifier)
	assert.Check(t, is.Equal("Catalog Asset Two", after.Name))
	assert.Check(t, time.Time(*after.ObservedAt).Equal(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)), "the volatile field must ride along once a material field changed")
}

// TestRiskObservedAtOnlyReingestNoop verifies an observed_at-only difference never writes on its own but rides along on a material change
func TestRiskObservedAtOnlyReingestNoop(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	counts := catalogEventCounters(t, entityops.SchemaRisk)
	defer counts.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Risk Observed At Test").
		SetKind("riskobsattest").
		SetDefinitionID("def_riskobsattest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-riskobsattest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	const externalID = "risk-obs-1"

	t.Cleanup(func() {
		risks, err := suite.Client.DB.Risk.Query().Where(risk.ExternalID(externalID)).All(ctx)
		th.RequireNoError(t, err)

		if len(risks) > 0 {
			(&th.Cleanup[*ent.RiskDeleteOne]{Client: suite.Client.DB.Risk, IDs: lo.Map(risks, func(r *ent.Risk, _ int) string { return r.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := jsonx.EditObject(riskCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["observed_at"] = json.RawMessage(`"2024-01-01T00:00:00Z"`)
		return true
	})

	result := ingestCatalogPayload(ctx, t, installation, entityops.SchemaRisk.Name, create)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Creates.Load()))
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()))

	created := riskByExternalID(ctx, t, externalID)
	assert.Assert(t, created.ObservedAt != nil)

	volatileOnly := jsonx.EditObject(riskCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["observed_at"] = json.RawMessage(`"2024-06-01T00:00:00Z"`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaRisk.Name, volatileOnly)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Changed), "a volatile-only payload must not count as changed")
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()), "a volatile-only payload must not emit an update mutation event")

	afterVolatileOnly := riskByExternalID(ctx, t, externalID)
	assert.Check(t, afterVolatileOnly.UpdatedAt.Equal(created.UpdatedAt), "a volatile-only payload must not rewrite the row")
	assert.Check(t, time.Time(*afterVolatileOnly.ObservedAt).Equal(time.Time(*created.ObservedAt)), "an unwritten volatile field must keep its stored value")

	materialAndVolatile := jsonx.EditObject(riskCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["name"] = json.RawMessage(`"Catalog Risk Two"`)
		doc["observed_at"] = json.RawMessage(`"2024-06-01T00:00:00Z"`)
		return true
	})

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaRisk.Name, materialAndVolatile)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a material change must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Updates.Load()), "a material change must emit an update mutation event")

	after := riskByExternalID(ctx, t, externalID)
	assert.Check(t, is.Equal("Catalog Risk Two", after.Name))
	assert.Check(t, time.Time(*after.ObservedAt).Equal(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)), "the volatile field must ride along once a material field changed")
}

// TestFindingFractionalTimestampReingestUnchanged pins the second-precision round-trip of a non-volatile models.DateTime field
func TestFindingFractionalTimestampReingestUnchanged(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	counts := catalogEventCounters(t, entityops.SchemaFinding)
	defer counts.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Fractional Timestamp Test").
		SetKind("findfractest").
		SetDefinitionID("def_findfractest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findfractest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	const externalID = "find-frac-1"

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalID(externalID)).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	payload := jsonx.EditObject(findingCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["reported_at"] = json.RawMessage(`"2025-06-22T03:11:22.561Z"`)
		return true
	})

	result := ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, payload)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(int64(1), counts.Creates.Load()))

	created := findingByExternalID(ctx, t, externalID)
	assert.Assert(t, created.ReportedAt != nil)

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, payload)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Changed), "an identical fractional-second timestamp must not count as changed")
	assert.Check(t, is.Equal(int64(0), counts.Updates.Load()), "an identical fractional-second timestamp must not emit an update mutation event")

	after := findingByExternalID(ctx, t, externalID)
	assert.Check(t, after.UpdatedAt.Equal(created.UpdatedAt), "an identical fractional-second timestamp must not rewrite the row")
}

// TestFindingControlLinkResyncDedupes verifies a Finding.controls through-edge link supplied on every ingest is applied once
func TestFindingControlLinkResyncDedupes(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Control Link Test").
		SetKind("findctrllinktest").
		SetDefinitionID("def_findctrllinktest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findctrllinktest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	control, err := suite.Client.DB.Control.Create().
		SetRefCode("CTRLLINK-" + ulids.New().String()).
		Save(ctx)
	th.RequireNoError(t, err)

	const externalID = "find-ctrl-1"

	t.Cleanup(func() {
		links, err := suite.Client.DB.FindingControl.Query().Where(findingcontrol.ControlID(control.ID)).All(ctx)
		th.RequireNoError(t, err)

		if len(links) > 0 {
			(&th.Cleanup[*ent.FindingControlDeleteOne]{Client: suite.Client.DB.FindingControl, IDs: lo.Map(links, func(fc *ent.FindingControl, _ int) string { return fc.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalID(externalID)).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	controlIDs, err := json.Marshal([]string{control.ID})
	th.RequireNoError(t, err)

	payload := jsonx.EditObject(findingCatalogPayload(externalID), func(doc map[string]json.RawMessage) bool {
		doc["control_ids"] = controlIDs
		return true
	})

	linkCount := func() int {
		count, err := suite.Client.DB.FindingControl.Query().Where(findingcontrol.ControlID(control.ID)).Count(ctx)
		th.RequireNoError(t, err)

		return count
	}

	result := ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, payload)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Failed), "linking an existing control must not fail")
	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(1, linkCount()), "the first ingest must create exactly one finding_controls join row")

	created := findingByExternalID(ctx, t, externalID)

	result = ingestCatalogPayload(ctx, t, installation, entityops.SchemaFinding.Name, payload)
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Failed), "a re-ingest carrying the same control link must not error")
	assert.Check(t, is.Equal(0, result.Changed), "a re-ingest with no material change must not count as changed")
	assert.Check(t, is.Equal(1, linkCount()), "a re-ingest of the same control link must not duplicate the join row")

	after := findingByExternalID(ctx, t, externalID)
	assert.Check(t, after.UpdatedAt.Equal(created.UpdatedAt), "a re-ingest with no material change must not rewrite the row")
}
