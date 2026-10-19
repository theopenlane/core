package azuresecuritycenter

import (
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	mappingtest.AssertExpressionsValid(t, def.Mappings)
}

func TestAssessmentExample(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	findingSpec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaFinding.Name)

	t.Run("unhealthy_maps_to_finding", func(t *testing.T) {
		a := loadAssessment(t, "assessment.json")
		a.Properties.Status.Code = "Unhealthy"
		a.ComplianceControls = []string{"CC7.2", "IR.2"}
		a.ComplianceStandards = []string{"Azure-SOC-2023", "Microsoft-cloud-security-benchmark"}

		payload, err := buildPayload(a)
		assert.NilError(t, err)

		envelope := marshalExampleEnvelope(t, payload)
		assert.Check(t, mappingtest.AssertFiltered(t, findingSpec, envelope))

		mapped := mappingtest.EvalMap(t, findingSpec, envelope)

		resourceID := "/subscriptions/20ff7fc3-e762-44dd-bd96-b71116dcdc23/resourceGroups/myRg/providers/Microsoft.Compute/virtualMachineScaleSets/vmss2"

		assert.Check(t, mapped["external_id"] == resourceID+"/providers/Microsoft.Security/assessments/21300918-b2e3-0346-785f-c77ff57d243b")
		assert.Check(t, mapped["external_owner_id"] == "")
		assert.Check(t, mapped["internal_owner"] == nil)
		assert.Check(t, mapped["display_name"] == "Install endpoint protection solution on virtual machine scale sets")
		assert.Check(t, mapped["finding_status_name"] == "Unhealthy")
		assert.Check(t, mapped["open"] == true)
		assert.Check(t, mapped["external_uri"] == "https://www.portal.azure.com/?fea#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/21300918-b2e3-0346-785f-c77ff57d243b")
		assert.DeepEqual(t, []any{resourceID}, mapped["targets"])
		assert.DeepEqual(t, []any{"CC7.2", "IR.2"}, mapped["categories"])
		assert.DeepEqual(t, []any{"Azure-SOC-2023", "Microsoft-cloud-security-benchmark"}, mapped["tags"])
		assert.Check(t, mapped["reported_at"] == nil)
		assert.Check(t, mapped["resource_name"] == resourceID)
		assert.DeepEqual(t, map[string]any{"resources": []any{map[string]any{
			"id":              resourceID,
			"source":          "Azure",
			"name":            "vmss2",
			"type":            "Microsoft.Compute/virtualMachineScaleSets",
			"subscription_id": "20ff7fc3-e762-44dd-bd96-b71116dcdc23",
			"resource_group":  "myRg",
		}}}, mapped["target_details"])
		assert.Check(t, mapped["metadata"] == nil)
		assert.Check(t, mapped["source"] == mappingtest.Installation.Name)
	})

	t.Run("capitalized_resource_details_decode", func(t *testing.T) {
		payload, err := buildPayload(loadAssessment(t, "assessment_capitalized_resource_details.json"))
		assert.NilError(t, err)
		assert.Check(t, payload.RecommendationCategory == "")

		envelope := marshalExampleEnvelope(t, payload)
		mapped := mappingtest.EvalMap(t, findingSpec, envelope)

		assert.Check(t, mapped["category"] == "")
		assert.DeepEqual(t, []any{}, mapped["categories"])

		subscriptionID := "359fcb95-1c6d-453b-9140-5a8254ec83fd"
		resourceID := "/subscriptions/" + subscriptionID

		assert.Check(t, mapped["external_owner_id"] == "")
		assert.Check(t, mapped["resource_name"] == resourceID)
		assert.DeepEqual(t, []any{resourceID}, mapped["targets"])
		assert.DeepEqual(t, map[string]any{"resources": []any{map[string]any{
			"id":              resourceID,
			"source":          "azure",
			"name":            subscriptionID,
			"type":            "Microsoft.Resources/subscriptions",
			"subscription_id": subscriptionID,
		}}}, mapped["target_details"])
		assert.Check(t, mapped["display_name"] == "Microsoft Defender for Storage plan should be enabled with Malware Scanning and Sensitive Data Threat Detection")
		assert.Check(t, mapped["external_uri"] == "https://portal.azure.com/#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/00000000-0000-0000-0000-000000000000")

		metadata, ok := mapped["metadata"].(map[string]any)
		assert.Check(t, ok)
		assert.Check(t, metadata["policyDefinitionId"] == "/providers/microsoft.authorization/policyDefinitions/640d2586-54d2-465f-877f-9ffc1d2109f4")
	})
}

func TestVulnerabilityExample(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	vulnSpec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaVulnerability.Name)

	payload, err := buildPayload(loadAssessment(t, "vulnerability_recommendation.json"))
	assert.NilError(t, err)
	assert.Check(t, len(payload.CVEDetails) == 2)
	assert.Check(t, payload.FixedVersion == "1.4.10")
	assert.Check(t, payload.RecommendationCategory == "SoftwareUpdate")
	assert.Check(t, payload.PackageName == "apt")
	assert.Check(t, payload.PackageEcosystem == "debian")
	assert.DeepEqual(t, []string{"1.4.9"}, payload.DetectedVersions)

	payloads := cvePayloads(payload)
	assert.Check(t, len(payloads) == 2)
	assert.Check(t, payloads[1].CveID == "CVE-2019-3462")
	assert.Check(t, payloads[1].Severity == "High")
	assert.Check(t, payloads[1].FixedVersion == "1.4.9")

	envelope := marshalExampleEnvelope(t, payloads[0])
	assert.Check(t, mappingtest.AssertFiltered(t, vulnSpec, envelope))

	mapped := mappingtest.EvalMap(t, vulnSpec, envelope)

	resourceID := "/subscriptions/359fcb95-1c6d-453b-9140-5a8254ec83fd/resourceGroups/prod-rg/providers/Microsoft.Compute/virtualMachines/web-01"
	recommendationID := resourceID + "/providers/Microsoft.Security/assessments/5f1c3a2e-9b7d-4e61-8a4f-2c0d6b9e1a73"

	assert.Check(t, mapped["external_id"] == recommendationID+"/cves/CVE-2020-3810")
	assert.Check(t, mapped["external_owner_id"] == "secops@example.com")
	assert.Check(t, mapped["internal_owner"] == "secops@example.com")
	assert.Check(t, mapped["cve_id"] == "CVE-2020-3810")
	assert.Check(t, mapped["display_name"] == "CVE-2020-3810")
	assert.Check(t, mapped["summary"] == "Update apt")
	assert.Check(t, mapped["description"] == "Vulnerable versions of apt were detected on this machine.")
	assert.Check(t, mapped["severity"] == "Medium")
	assert.Check(t, mapped["vulnerability_status_name"] == "Unhealthy")
	assert.Check(t, mapped["open"] == true)
	assert.Check(t, mapped["package_name"] == "apt")
	assert.Check(t, mapped["package_ecosystem"] == "debian")
	assert.Check(t, mapped["score"] == float64(5.5))
	assert.Check(t, mapped["first_patched_version"] == "1.4.10")
	assert.Check(t, mapped["fix_available"] == true)
	assert.Check(t, mapped["discovered_at"] == "2026-08-02T10:15:00Z")
	assert.Check(t, mapped["source_updated_at"] == "2026-08-03T09:30:00Z")
	assert.Check(t, mapped["external_uri"] == "https://portal.azure.com/#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/5f1c3a2e-9b7d-4e61-8a4f-2c0d6b9e1a73")
	assert.Check(t, mapped["source"] == mappingtest.Installation.Name)

	metadata, ok := mapped["metadata"].(map[string]any)
	assert.Check(t, ok)

	resourceDetails, ok := metadata["resource_details"].(map[string]any)
	assert.Check(t, ok)
	assert.Check(t, resourceDetails["id"] == resourceID)
	assert.Check(t, resourceDetails["name"] == "web-01")
	assert.DeepEqual(t, []any{"1.4.9"}, metadata["detected_versions"])
}

func TestComplianceExample(t *testing.T) {
	def, err := Builder(Config{})()
	assert.NilError(t, err)

	observedAt := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	var row complianceAssessment
	assert.NilError(t, json.Unmarshal(mappingtest.LoadExample(t, "examples", "compliance_assessment.json"), &row))

	payload := buildCompliancePayload(row, observedAt)
	assert.Check(t, payload.Standard == "Azure-SOC-2023")
	assert.Check(t, payload.Control == "CC7.4")
	assert.Check(t, payload.Status == "FAIL")

	t.Run("maps_to_check_result", func(t *testing.T) {
		checkSpec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaCheckResult.Name)

		envelope := marshalExampleEnvelope(t, payload)
		assert.Check(t, mappingtest.AssertFiltered(t, checkSpec, envelope))

		mapped := mappingtest.EvalMap(t, checkSpec, envelope)

		assert.Check(t, mapped["parent_external_id"] == row.ID)
		assert.Check(t, mapped["status"] == "FAIL")
		assert.Check(t, mapped["details"] == "Microsoft Defender for Resource Manager should be enabled")
		assert.Check(t, mapped["external_uri"] == "https://portal.azure.com/#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/77758c9d-8a56-5f54-6ff7-69a762ca6004")
		assert.Check(t, mapped["last_observed_at"] == "2026-10-16T12:00:00Z")
		assert.DeepEqual(t, []any{"Azure-SOC-2023", "CC7.4"}, mapped["tags"])
		assert.Check(t, mapped["source"] == mappingtest.Installation.Name)
	})

	t.Run("failed_policy_maps_to_finding", func(t *testing.T) {
		findingSpec := mappingtest.MappingSpec(t, def.Mappings, entityops.SchemaFinding.Name)

		findings := complianceFindings([]CompliancePayload{payload})
		assert.Check(t, len(findings) == 1)

		envelope := marshalExampleEnvelope(t, findings[0])
		assert.Check(t, mappingtest.AssertFiltered(t, findingSpec, envelope))

		mapped := mappingtest.EvalMap(t, findingSpec, envelope)

		assert.Check(t, mapped["external_id"] == "/subscriptions/359fcb95-1c6d-453b-9140-5a8254ec83fd/providers/Microsoft.Security/assessments/77758c9d-8a56-5f54-6ff7-69a762ca6004")
		assert.Check(t, mapped["display_name"] == "Microsoft Defender for Resource Manager should be enabled")
		assert.Check(t, mapped["category"] == "")
		assert.DeepEqual(t, []any{"CC7.4"}, mapped["categories"])
		assert.DeepEqual(t, []any{"Azure-SOC-2023"}, mapped["tags"])
		assert.Check(t, mapped["finding_status_name"] == "Unhealthy")
		assert.Check(t, mapped["open"] == true)
		assert.Check(t, mapped["external_owner_id"] == "secops@example.com")
		assert.Check(t, mapped["target_details"] == nil)
		assert.Check(t, mapped["source_updated_at"] == "2026-10-16T19:23:35.8120901Z")
		assert.Check(t, mapped["severity"] == "High")
		assert.Check(t, mapped["description"] == "Microsoft Defender for Resource Manager automatically monitors the resource management operations in your organization.")
		assert.Check(t, mapped["recommended_actions"] == "Enable Microsoft Defender for Resource Manager in the Defender plans of the subscription's environment settings.")
		assert.DeepEqual(t, []any{"/subscriptions/359fcb95-1c6d-453b-9140-5a8254ec83fd"}, mapped["targets"])
		assert.Check(t, mapped["internal_owner"] == "secops@example.com")

		metadata, ok := mapped["metadata"].(map[string]any)
		assert.Check(t, ok)
		assert.Check(t, metadata["failed_resources"] == float64(1))
		assert.Check(t, metadata["state"] == "Failed")
	})

	t.Run("controls_sharing_an_assessment_merge_into_one_finding", func(t *testing.T) {
		other := payload
		other.Control = "CC7.2"
		other.ID = "/subscriptions/359fcb95-1c6d-453b-9140-5a8254ec83fd/providers/Microsoft.Security/regulatoryComplianceStandards/Azure-SOC-2023/regulatoryComplianceControls/CC7.2/regulatoryComplianceAssessments/77758c9d-8a56-5f54-6ff7-69a762ca6004"

		findings := complianceFindings([]CompliancePayload{payload, other, payload})
		assert.Check(t, len(findings) == 1)
		assert.DeepEqual(t, []string{"CC7.4", "CC7.2"}, findings[0].Categories)
	})
}

func loadAssessment(t *testing.T, name string) assessment {
	t.Helper()

	var a assessment
	assert.NilError(t, json.Unmarshal(mappingtest.LoadExample(t, "examples", name), &a))

	return a
}

func marshalExampleEnvelope(t *testing.T, payload any) types.MappingEnvelope {
	t.Helper()

	envelope, err := providerkit.MarshalEnvelope("", payload, ErrIngestPayloadEncode)
	assert.NilError(t, err)

	return envelope
}
