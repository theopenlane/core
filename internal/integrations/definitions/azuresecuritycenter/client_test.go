package azuresecuritycenter

import (
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

const (
	testSubscriptionID   = "359fcb95-1c6d-453b-9140-5a8254ec83fd"
	testVMID             = "/subscriptions/" + testSubscriptionID + "/resourceGroups/prod-rg/providers/Microsoft.Compute/virtualMachines/web-01"
	testVMAssessmentID   = testVMID + "/providers/Microsoft.Security/assessments/abc"
	testRegistryID       = "/subscriptions/" + testSubscriptionID + "/resourceGroups/ol-defender-test/providers/Microsoft.ContainerRegistry/registries/openlane"
	testImageDigest      = "sha256:17a1998407746106c307c58c5089569bc1d0728567657b8c19ccffd0497c91ba"
	testImageID          = testRegistryID + "/securityentitydata/repositories-nginx-images-" + testImageDigest
	testRecommendationID = "/subscriptions/x/providers/Microsoft.Security/assessments/y"
)

func TestBuildPayloadCategories(t *testing.T) {
	tests := []struct {
		name         string
		metadata     string
		controls     []string
		wantCategory string
		wantList     []string
	}{
		{
			name:         "classic categories",
			metadata:     `{"categories":["Data","Compute"],"recommendationCategory":"Unknown"}`,
			wantCategory: "Data",
			wantList:     []string{"Data", "Compute"},
		},
		{
			name:         "unknown categories fall back to recommendation category",
			metadata:     `{"categories":["Unknown"],"recommendationCategory":"HostMisconfigurations"}`,
			wantCategory: "HostMisconfigurations",
			wantList:     []string{"HostMisconfigurations"},
		},
		{
			name:         "everything unknown stays empty",
			metadata:     `{"categories":["unknown"],"recommendationCategory":"Unknown"}`,
			wantCategory: "",
			wantList:     nil,
		},
		{
			name:         "empty categories and recommendation category",
			metadata:     `{"categories":[],"recommendationCategory":""}`,
			wantCategory: "",
			wantList:     nil,
		},
		{
			name:         "missing categories fields",
			metadata:     `{}`,
			wantCategory: "",
			wantList:     nil,
		},
		{
			name:         "compliance controls appended without duplicates",
			metadata:     `{"categories":["Data"]}`,
			controls:     []string{"CC7.2", "Data"},
			wantCategory: "Data",
			wantList:     []string{"Data", "CC7.2"},
		},
		{
			name:         "compliance controls without metadata",
			controls:     []string{"CC7.2"},
			wantCategory: "",
			wantList:     []string{"CC7.2"},
		},
		{
			name:         "empty compliance controls",
			controls:     []string{},
			wantCategory: "",
			wantList:     nil,
		},
		{
			name:         "no metadata or controls",
			wantCategory: "",
			wantList:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := assessment{ID: testRecommendationID, ComplianceControls: tc.controls}
			if tc.metadata != "" {
				a.Properties.Metadata = json.RawMessage(tc.metadata)
			}

			payload, err := buildPayload(a)
			assert.NilError(t, err)
			assert.Check(t, payload.Category == tc.wantCategory)
			assert.DeepEqual(t, tc.wantList, payload.Categories)
		})
	}
}

func TestBuildPayloadMetadata(t *testing.T) {
	tests := []struct {
		name         string
		metadata     string
		wantErr      bool
		wantSeverity string
	}{
		{
			name:         "valid metadata",
			metadata:     `{"severity":"High"}`,
			wantSeverity: "High",
		},
		{
			name:         "empty object",
			metadata:     `{}`,
			wantSeverity: "",
		},
		{
			name:         "json null",
			metadata:     `null`,
			wantSeverity: "",
		},
		{
			name:     "invalid json",
			metadata: `{not json`,
			wantErr:  true,
		},
		{
			name:     "wrong shape",
			metadata: `["High"]`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := buildPayload(assessment{ID: testRecommendationID, Properties: assessmentProperties{Metadata: json.RawMessage(tc.metadata)}})
			if tc.wantErr {
				assert.Assert(t, err != nil)
				assert.DeepEqual(t, AssessmentPayload{}, payload)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, payload.Severity == tc.wantSeverity)
		})
	}
}

func TestBuildPayloadEmptyAssessment(t *testing.T) {
	payload, err := buildPayload(assessment{})
	assert.NilError(t, err)
	assert.Check(t, payload.ID == "")
	assert.Check(t, payload.ExternalURI == "")
	assert.Check(t, payload.SubscriptionID == "")
	assert.Check(t, payload.ResourceID == "")
	assert.Check(t, payload.Resources == nil)
	assert.Check(t, payload.Categories == nil)
	assert.Check(t, payload.CVEDetails == nil)
	assert.Check(t, payload.CVSSScore == nil)
	assert.DeepEqual(t, &ResourceDetails{}, payload.ResourceDetails)
	assert.DeepEqual(t, []ResourceDetails{{}}, payload.AffectedResources)
}

func TestBuildPayloadAdditionalData(t *testing.T) {
	tests := []struct {
		name                 string
		data                 map[string]any
		wantFixedVersion     string
		wantPackageName      string
		wantPackageEcosystem string
		wantDetectedVersions []string
		wantCVEs             []cveDetail
	}{
		{
			name: "nil map",
		},
		{
			name: "empty map",
			data: map[string]any{},
		},
		{
			name: "populated",
			data: map[string]any{
				fixedVersionKey:     "1.4.10",
				softwareNameKey:     "apt",
				softwareVendorKey:   "debian",
				detectedVersionsKey: `["1.4.9"]`,
				cveDetailsKey:       `[{"CveId":"CVE-2020-3810"}]`,
			},
			wantFixedVersion:     "1.4.10",
			wantPackageName:      "apt",
			wantPackageEcosystem: "debian",
			wantDetectedVersions: []string{"1.4.9"},
			wantCVEs:             []cveDetail{{CveID: "CVE-2020-3810"}},
		},
		{
			name: "nil values",
			data: map[string]any{
				fixedVersionKey:     nil,
				softwareNameKey:     nil,
				softwareVendorKey:   nil,
				detectedVersionsKey: nil,
				cveDetailsKey:       nil,
			},
		},
		{
			name: "empty values",
			data: map[string]any{
				fixedVersionKey:     "",
				softwareNameKey:     "",
				softwareVendorKey:   "",
				detectedVersionsKey: "",
				cveDetailsKey:       "",
			},
		},
		{
			name: "wrong types ignored",
			data: map[string]any{
				fixedVersionKey:     10,
				softwareNameKey:     true,
				softwareVendorKey:   []string{"debian"},
				detectedVersionsKey: "not json",
				cveDetailsKey:       42,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := buildPayload(assessment{ID: testRecommendationID, Properties: assessmentProperties{AdditionalData: tc.data}})
			assert.NilError(t, err)
			assert.Check(t, payload.FixedVersion == tc.wantFixedVersion)
			assert.Check(t, payload.PackageName == tc.wantPackageName)
			assert.Check(t, payload.PackageEcosystem == tc.wantPackageEcosystem)
			assert.DeepEqual(t, tc.wantDetectedVersions, payload.DetectedVersions)
			assert.DeepEqual(t, tc.wantCVEs, payload.CVEDetails)
		})
	}
}

func TestBuildPayloadExternalURI(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		portalLink string
		want       string
	}{
		{
			name:       "portal link preferred",
			key:        "abc",
			portalLink: "https://portal.azure.com/custom",
			want:       "https://portal.azure.com/custom",
		},
		{
			name: "falls back to recommendation url",
			key:  "abc",
			want: "https://portal.azure.com/#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/abc",
		},
		{
			name:       "portal link without key",
			portalLink: "https://portal.azure.com/custom",
			want:       "https://portal.azure.com/custom",
		},
		{
			name: "no key or link",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := buildPayload(assessment{ID: testRecommendationID, Name: tc.key, Properties: assessmentProperties{Links: assessmentLinks{AzurePortalURI: tc.portalLink}}})
			assert.NilError(t, err)
			assert.Check(t, payload.ExternalURI == tc.want)
		})
	}
}

func TestBuildPayloadCVSSScore(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  *float64
	}{
		{
			name:  "json string",
			value: "7.5",
			want:  new(7.5),
		},
		{
			name:  "decoded number",
			value: float64(9.8),
			want:  new(9.8),
		},
		{
			name:  "nil value",
			value: nil,
			want:  nil,
		},
		{
			name:  "empty string",
			value: "",
			want:  nil,
		},
		{
			name:  "invalid string",
			value: "high",
			want:  nil,
		},
		{
			name:  "wrong type",
			value: map[string]any{"score": 1},
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := buildPayload(assessment{ID: testRecommendationID, Properties: assessmentProperties{AdditionalData: map[string]any{maxCvssScoreKey: tc.value}}})
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.want, payload.CVSSScore)
		})
	}
}

func TestBuildPayloadResource(t *testing.T) {
	tests := []struct {
		name               string
		assessmentID       string
		details            assessmentResourceDetails
		wantResourceID     string
		wantResources      []string
		wantSubscriptionID string
	}{
		{
			name:               "resource details id",
			assessmentID:       testVMAssessmentID,
			details:            assessmentResourceDetails{ID: testVMID},
			wantResourceID:     testVMID,
			wantResources:      []string{testVMID},
			wantSubscriptionID: testSubscriptionID,
		},
		{
			name:               "falls back to assessment scope",
			assessmentID:       testVMAssessmentID,
			wantResourceID:     testVMID,
			wantResources:      []string{testVMID},
			wantSubscriptionID: testSubscriptionID,
		},
		{
			name:               "unparsable assessment id",
			assessmentID:       "not-an-arm-id",
			wantResourceID:     "",
			wantResources:      nil,
			wantSubscriptionID: "",
		},
		{
			name:               "empty assessment id",
			assessmentID:       "",
			wantResourceID:     "",
			wantResources:      nil,
			wantSubscriptionID: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := buildPayload(assessment{ID: tc.assessmentID, Properties: assessmentProperties{ResourceDetails: tc.details}})
			assert.NilError(t, err)
			assert.Check(t, payload.ResourceID == tc.wantResourceID)
			assert.DeepEqual(t, tc.wantResources, payload.Resources)
			assert.Check(t, payload.SubscriptionID == tc.wantSubscriptionID)
			assert.Check(t, len(payload.AffectedResources) == 1)
		})
	}
}

func TestNewResourceDetails(t *testing.T) {
	tests := []struct {
		name         string
		details      string
		assessmentID string
		want         ResourceDetails
	}{
		{
			name:         "container image",
			details:      `{"ResourceProvider":"acr","ResourceName":"nginx","ResourceType":".containerimage","Source":"Azure","Id":"` + testImageID + `"}`,
			assessmentID: testImageID + "/providers/Microsoft.Security/assessments/abc",
			want: ResourceDetails{
				ID:                 testImageID,
				Source:             "Azure",
				Name:               "nginx",
				Type:               ".containerimage",
				SubscriptionID:     testSubscriptionID,
				ResourceGroup:      "ol-defender-test",
				Provider:           "acr",
				ParentResourceID:   testRegistryID,
				ParentResourceName: "openlane",
				Repository:         "nginx",
				Digest:             testImageDigest,
			},
		},
		{
			name:         "image entity without digest",
			details:      `{"Id":"` + testRegistryID + `/securityentitydata/repositories-nginx"}`,
			assessmentID: "",
			want: ResourceDetails{
				ID:                 testRegistryID + "/securityentitydata/repositories-nginx",
				SubscriptionID:     testSubscriptionID,
				ResourceGroup:      "ol-defender-test",
				ParentResourceID:   testRegistryID,
				ParentResourceName: "openlane",
			},
		},
		{
			name:         "azure resource",
			details:      `{"Source":"Azure","Id":"` + testVMID + `"}`,
			assessmentID: testVMAssessmentID,
			want: ResourceDetails{
				ID:             testVMID,
				Source:         "Azure",
				Name:           "web-01",
				Type:           "Microsoft.Compute/virtualMachines",
				SubscriptionID: testSubscriptionID,
				ResourceGroup:  "prod-rg",
			},
		},
		{
			name:         "resource name overrides parsed name",
			details:      `{"Id":"` + testVMID + `","ResourceName":"web-01-display"}`,
			assessmentID: "",
			want: ResourceDetails{
				ID:             testVMID,
				Name:           "web-01-display",
				Type:           "Microsoft.Compute/virtualMachines",
				SubscriptionID: testSubscriptionID,
				ResourceGroup:  "prod-rg",
			},
		},
		{
			name:         "on premise machine",
			details:      `{"Source":"OnPremise","MachineName":"onprem-01","SourceComputerId":"7c1f2a3b"}`,
			assessmentID: "not-an-arm-id",
			want: ResourceDetails{
				ID:     "7c1f2a3b",
				Source: "OnPremise",
				Name:   "onprem-01",
			},
		},
		{
			name:         "no details falls back to assessment scope",
			details:      `{}`,
			assessmentID: testVMAssessmentID,
			want: ResourceDetails{
				ID:             testVMID,
				Name:           "web-01",
				Type:           "Microsoft.Compute/virtualMachines",
				SubscriptionID: testSubscriptionID,
				ResourceGroup:  "prod-rg",
			},
		},
		{
			name:         "no details and unparsable assessment id",
			details:      `{}`,
			assessmentID: "not-an-arm-id",
			want:         ResourceDetails{},
		},
		{
			name:         "no details and empty assessment id",
			details:      `{}`,
			assessmentID: "",
			want:         ResourceDetails{},
		},
		{
			name:         "null details fields",
			details:      `{"Id":null,"Source":null,"ResourceName":null}`,
			assessmentID: "",
			want:         ResourceDetails{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var details assessmentResourceDetails
			assert.NilError(t, json.Unmarshal([]byte(tc.details), &details))

			assert.DeepEqual(t, tc.want, newResourceDetails(details, tc.assessmentID))
		})
	}
}

func TestSplitImageEntity(t *testing.T) {
	tests := []struct {
		name           string
		entity         string
		wantRepository string
		wantDigest     string
	}{
		{
			name:           "repository and digest",
			entity:         "repositories-nginx-images-sha256:abc",
			wantRepository: "nginx",
			wantDigest:     "sha256:abc",
		},
		{
			name:           "nested repository",
			entity:         "repositories-library/nginx-images-sha256:abc",
			wantRepository: "library/nginx",
			wantDigest:     "sha256:abc",
		},
		{
			name:           "repository containing separator uses last one",
			entity:         "repositories-my-images-app-images-sha256:abc",
			wantRepository: "my-images-app",
			wantDigest:     "sha256:abc",
		},
		{
			name:           "missing repositories prefix",
			entity:         "nginx-images-sha256:abc",
			wantRepository: "nginx",
			wantDigest:     "sha256:abc",
		},
		{
			name:           "empty repository and digest",
			entity:         "repositories--images-",
			wantRepository: "",
			wantDigest:     "",
		},
		{
			name:           "missing separator",
			entity:         "repositories-nginx",
			wantRepository: "",
			wantDigest:     "",
		},
		{
			name:           "empty entity",
			entity:         "",
			wantRepository: "",
			wantDigest:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository, digest := splitImageEntity(tc.entity)
			assert.Check(t, repository == tc.wantRepository)
			assert.Check(t, digest == tc.wantDigest)
		})
	}
}

func TestAssessedResourceID(t *testing.T) {
	tests := []struct {
		name         string
		assessmentID string
		want         string
	}{
		{
			name:         "resource assessment",
			assessmentID: testVMAssessmentID,
			want:         testVMID,
		},
		{
			name:         "sub assessment",
			assessmentID: testVMAssessmentID + "/subAssessments/def",
			want:         testVMID,
		},
		{
			name:         "nested security resources",
			assessmentID: "/subscriptions/" + testSubscriptionID + "/providers/Microsoft.Security/pricings/arm/providers/Microsoft.Security/assessments/abc",
			want:         "/subscriptions/" + testSubscriptionID,
		},
		{
			name:         "subscription assessment",
			assessmentID: "/subscriptions/" + testSubscriptionID + "/providers/Microsoft.Security/assessments/abc",
			want:         "/subscriptions/" + testSubscriptionID,
		},
		{
			name:         "tenant level assessment",
			assessmentID: "/providers/Microsoft.Security/assessments/abc",
			want:         "",
		},
		{
			name:         "unparsable id",
			assessmentID: "not-an-arm-id",
			want:         "",
		},
		{
			name:         "empty id",
			assessmentID: "",
			want:         "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Check(t, assessedResourceID(tc.assessmentID) == tc.want)
		})
	}
}

func TestRecommendationFindings(t *testing.T) {
	tests := []struct {
		name     string
		payloads []AssessmentPayload
		wantIDs  []string
	}{
		{
			name:     "nil payloads",
			payloads: nil,
			wantIDs:  nil,
		},
		{
			name:     "empty payloads",
			payloads: []AssessmentPayload{},
			wantIDs:  nil,
		},
		{
			name:     "missing subscription and key",
			payloads: []AssessmentPayload{{}},
			wantIDs:  []string{"/subscriptions//providers/Microsoft.Security/assessments/"},
		},
		{
			name: "same recommendation merges",
			payloads: []AssessmentPayload{
				{SubscriptionID: "sub", AssessmentKey: "key"},
				{SubscriptionID: "sub", AssessmentKey: "key"},
			},
			wantIDs: []string{"/subscriptions/sub/providers/Microsoft.Security/assessments/key"},
		},
		{
			name: "different subscriptions stay separate",
			payloads: []AssessmentPayload{
				{SubscriptionID: "sub-a", AssessmentKey: "key"},
				{SubscriptionID: "sub-b", AssessmentKey: "key"},
			},
			wantIDs: []string{
				"/subscriptions/sub-a/providers/Microsoft.Security/assessments/key",
				"/subscriptions/sub-b/providers/Microsoft.Security/assessments/key",
			},
		},
		{
			name: "first appearance order kept",
			payloads: []AssessmentPayload{
				{SubscriptionID: "sub", AssessmentKey: "b"},
				{SubscriptionID: "sub", AssessmentKey: "a"},
				{SubscriptionID: "sub", AssessmentKey: "b"},
			},
			wantIDs: []string{
				"/subscriptions/sub/providers/Microsoft.Security/assessments/b",
				"/subscriptions/sub/providers/Microsoft.Security/assessments/a",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotIDs []string
			for _, f := range recommendationFindings(tc.payloads) {
				gotIDs = append(gotIDs, f.ID)
			}

			assert.DeepEqual(t, tc.wantIDs, gotIDs)
		})
	}
}

func TestRecommendationFindingsMergeResources(t *testing.T) {
	early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	late := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	arm := AssessmentPayload{
		SubscriptionID:    "sub",
		AssessmentKey:     "key",
		ResourceID:        "/subscriptions/sub/providers/microsoft.security/pricings/arm",
		Resources:         []string{"/subscriptions/sub/providers/microsoft.security/pricings/arm"},
		AffectedResources: []ResourceDetails{{Name: "arm"}},
		Categories:        []string{"CC7.2"},
		DiscoveredAt:      &late,
		UpdatedAt:         &early,
	}
	keyVaults := AssessmentPayload{
		SubscriptionID:    "sub",
		AssessmentKey:     "key",
		ResourceID:        "/subscriptions/sub/providers/microsoft.security/pricings/keyvaults",
		Resources:         []string{"/subscriptions/sub/providers/microsoft.security/pricings/keyvaults"},
		AffectedResources: []ResourceDetails{{Name: "keyvaults"}},
		Categories:        []string{"CC7.2", "A1.1"},
		GovernanceOwner:   "secops@example.com",
		DiscoveredAt:      &early,
		UpdatedAt:         &late,
	}
	other := AssessmentPayload{
		SubscriptionID: "sub",
		AssessmentKey:  "other",
		Resources:      []string{"/subscriptions/sub"},
	}

	findings := recommendationFindings([]AssessmentPayload{arm, keyVaults, other})
	assert.Check(t, len(findings) == 2)

	merged := findings[0]
	assert.Check(t, merged.ID == "/subscriptions/sub/providers/Microsoft.Security/assessments/key")
	assert.Check(t, merged.ExternalURI == "https://portal.azure.com/#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/key")
	assert.DeepEqual(t, []string{"/subscriptions/sub/providers/microsoft.security/pricings/arm", "/subscriptions/sub/providers/microsoft.security/pricings/keyvaults"}, merged.Resources)
	assert.DeepEqual(t, []string{"CC7.2", "A1.1"}, merged.Categories)
	assert.Check(t, merged.GovernanceOwner == "secops@example.com")
	assert.Check(t, merged.ResourceID == "")
	assert.Check(t, merged.ResourceDetails == nil)
	assert.DeepEqual(t, []ResourceDetails{{Name: "arm"}, {Name: "keyvaults"}}, merged.AffectedResources)
	assert.Check(t, merged.DiscoveredAt.Equal(early))
	assert.Check(t, merged.UpdatedAt.Equal(late))

	assert.Check(t, findings[1].ID == "/subscriptions/sub/providers/Microsoft.Security/assessments/other")
	assert.DeepEqual(t, []string{"/subscriptions/sub"}, findings[1].Resources)
}

func TestMergeRecommendation(t *testing.T) {
	early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	late := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		finding AssessmentPayload
		other   AssessmentPayload
		want    AssessmentPayload
	}{
		{
			name: "merges lists and widens time range",
			finding: AssessmentPayload{
				ResourceID:        testVMID,
				ResourceDetails:   &ResourceDetails{Name: "a"},
				AdditionalData:    map[string]any{"key": "value"},
				Resources:         []string{"a"},
				AffectedResources: []ResourceDetails{{Name: "a"}},
				Categories:        []string{"CC7.2"},
				Standards:         []string{"Azure-SOC-2023"},
				DiscoveredAt:      &late,
				UpdatedAt:         &early,
			},
			other: AssessmentPayload{
				Resources:         []string{"a", "b"},
				AffectedResources: []ResourceDetails{{Name: "b"}},
				Categories:        []string{"CC7.2", "A1.1"},
				Standards:         []string{"Azure-SOC-2023", "Microsoft-cloud-security-benchmark"},
				GovernanceOwner:   "secops@example.com",
				DiscoveredAt:      &early,
				UpdatedAt:         &late,
			},
			want: AssessmentPayload{
				Resources:         []string{"a", "b"},
				AffectedResources: []ResourceDetails{{Name: "a"}, {Name: "b"}},
				Categories:        []string{"CC7.2", "A1.1"},
				Standards:         []string{"Azure-SOC-2023", "Microsoft-cloud-security-benchmark"},
				GovernanceOwner:   "secops@example.com",
				DiscoveredAt:      &early,
				UpdatedAt:         &late,
			},
		},
		{
			name: "keeps existing owner and narrower times",
			finding: AssessmentPayload{
				GovernanceOwner: "first@example.com",
				DiscoveredAt:    &early,
				UpdatedAt:       &late,
			},
			other: AssessmentPayload{
				GovernanceOwner: "second@example.com",
				DiscoveredAt:    &late,
				UpdatedAt:       &early,
			},
			want: AssessmentPayload{
				GovernanceOwner: "first@example.com",
				Resources:       []string{},
				Categories:      []string{},
				Standards:       []string{},
				DiscoveredAt:    &early,
				UpdatedAt:       &late,
			},
		},
		{
			name:    "owner filled when finding has none",
			finding: AssessmentPayload{},
			other: AssessmentPayload{
				GovernanceOwner: "second@example.com",
			},
			want: AssessmentPayload{
				GovernanceOwner: "second@example.com",
				Resources:       []string{},
				Categories:      []string{},
				Standards:       []string{},
			},
		},
		{
			name:    "fills missing times",
			finding: AssessmentPayload{},
			other: AssessmentPayload{
				DiscoveredAt: &early,
				UpdatedAt:    &late,
			},
			want: AssessmentPayload{
				Resources:    []string{},
				Categories:   []string{},
				Standards:    []string{},
				DiscoveredAt: &early,
				UpdatedAt:    &late,
			},
		},
		{
			name: "missing other times keep existing",
			finding: AssessmentPayload{
				DiscoveredAt: &early,
				UpdatedAt:    &late,
			},
			other: AssessmentPayload{},
			want: AssessmentPayload{
				Resources:    []string{},
				Categories:   []string{},
				Standards:    []string{},
				DiscoveredAt: &early,
				UpdatedAt:    &late,
			},
		},
		{
			name:    "both empty",
			finding: AssessmentPayload{},
			other:   AssessmentPayload{},
			want: AssessmentPayload{
				Resources:  []string{},
				Categories: []string{},
				Standards:  []string{},
			},
		},
		{
			name: "empty other clears per-resource fields",
			finding: AssessmentPayload{
				ResourceID:      testVMID,
				ResourceDetails: &ResourceDetails{Name: "a"},
				AdditionalData:  map[string]any{"key": "value"},
				Resources:       []string{testVMID},
			},
			other: AssessmentPayload{},
			want: AssessmentPayload{
				Resources:  []string{testVMID},
				Categories: []string{},
				Standards:  []string{},
			},
		},
		{
			name:    "empty finding takes other lists",
			finding: AssessmentPayload{},
			other: AssessmentPayload{
				Resources:         []string{"b"},
				AffectedResources: []ResourceDetails{{Name: "b"}},
				Categories:        []string{"A1.1"},
				Standards:         []string{"Azure-SOC-2023"},
			},
			want: AssessmentPayload{
				Resources:         []string{"b"},
				AffectedResources: []ResourceDetails{{Name: "b"}},
				Categories:        []string{"A1.1"},
				Standards:         []string{"Azure-SOC-2023"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			finding := tc.finding
			mergeRecommendation(&finding, tc.other)

			assert.DeepEqual(t, tc.want, finding)
		})
	}
}

func TestCvePayloads(t *testing.T) {
	tests := []struct {
		name    string
		payload AssessmentPayload
		want    []AssessmentPayload
	}{
		{
			name: "nil cves keeps recommendation",
			payload: AssessmentPayload{
				ID:       testRecommendationID,
				Severity: "Low",
			},
			want: []AssessmentPayload{{
				ID:       testRecommendationID,
				Severity: "Low",
			}},
		},
		{
			name: "empty cves keeps recommendation",
			payload: AssessmentPayload{
				ID:         testRecommendationID,
				CVEDetails: []cveDetail{},
			},
			want: []AssessmentPayload{{
				ID:         testRecommendationID,
				CVEDetails: []cveDetail{},
			}},
		},
		{
			name:    "empty payload",
			payload: AssessmentPayload{},
			want:    []AssessmentPayload{{}},
		},
		{
			name: "cve values override recommendation",
			payload: AssessmentPayload{
				ID:           testRecommendationID,
				Severity:     "Low",
				FixedVersion: "2.0.0",
				CVEDetails:   []cveDetail{{CveID: "CVE-2024-0001", Severity: "High", FixedVersion: "1.9.0"}},
			},
			want: []AssessmentPayload{{
				ID:           testRecommendationID + "/cves/CVE-2024-0001",
				CveID:        "CVE-2024-0001",
				Severity:     "High",
				FixedVersion: "1.9.0",
				CVEDetails:   []cveDetail{{CveID: "CVE-2024-0001", Severity: "High", FixedVersion: "1.9.0"}},
			}},
		},
		{
			name: "cve without values falls back to recommendation",
			payload: AssessmentPayload{
				ID:           testRecommendationID,
				Severity:     "Low",
				FixedVersion: "2.0.0",
				CVEDetails:   []cveDetail{{CveID: "CVE-2024-0002"}},
			},
			want: []AssessmentPayload{{
				ID:           testRecommendationID + "/cves/CVE-2024-0002",
				CveID:        "CVE-2024-0002",
				Severity:     "Low",
				FixedVersion: "2.0.0",
				CVEDetails:   []cveDetail{{CveID: "CVE-2024-0002"}},
			}},
		},
		{
			name: "cve and recommendation both missing values",
			payload: AssessmentPayload{
				ID:         testRecommendationID,
				CVEDetails: []cveDetail{{CveID: "CVE-2024-0003"}},
			},
			want: []AssessmentPayload{{
				ID:         testRecommendationID + "/cves/CVE-2024-0003",
				CveID:      "CVE-2024-0003",
				CVEDetails: []cveDetail{{CveID: "CVE-2024-0003"}},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.DeepEqual(t, tc.want, cvePayloads(tc.payload))
		})
	}
}

func TestSplitCVEs(t *testing.T) {
	tests := []struct {
		name     string
		payloads []AssessmentPayload
		wantIDs  []string
	}{
		{
			name:     "nil payloads",
			payloads: nil,
			wantIDs:  nil,
		},
		{
			name:     "empty payloads",
			payloads: []AssessmentPayload{},
			wantIDs:  nil,
		},
		{
			name:     "empty cve list keeps recommendation",
			payloads: []AssessmentPayload{{ID: "rec-a", CVEDetails: []cveDetail{}}},
			wantIDs:  []string{"rec-a"},
		},
		{
			name:     "no cves keeps each recommendation",
			payloads: []AssessmentPayload{{ID: "rec-a"}, {ID: "rec-b"}},
			wantIDs:  []string{"rec-a", "rec-b"},
		},
		{
			name: "expands cves in order",
			payloads: []AssessmentPayload{
				{ID: "rec-a", CVEDetails: []cveDetail{{CveID: "CVE-2024-0001"}, {CveID: "CVE-2024-0002"}}},
				{ID: "rec-b"},
			},
			wantIDs: []string{
				"rec-a/cves/CVE-2024-0001",
				"rec-a/cves/CVE-2024-0002",
				"rec-b",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotIDs []string
			for _, p := range splitCVEs(tc.payloads) {
				gotIDs = append(gotIDs, p.ID)
			}

			assert.DeepEqual(t, tc.wantIDs, gotIDs)
		})
	}
}

func TestCveDetails(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  []cveDetail
	}{
		{
			name:  "nil value",
			value: nil,
			want:  nil,
		},
		{
			name:  "json string",
			value: `[{"CveId":"CVE-2020-3810","Severity":"Medium","FixedVersion":"1.4.10"}]`,
			want:  []cveDetail{{CveID: "CVE-2020-3810", Severity: "Medium", FixedVersion: "1.4.10"}},
		},
		{
			name:  "decoded array",
			value: []any{map[string]any{"CveId": "CVE-2024-0727"}},
			want:  []cveDetail{{CveID: "CVE-2024-0727"}},
		},
		{
			name:  "skips empty ids",
			value: `[{"CveId":""},{"CveId":"CVE-2023-5678"}]`,
			want:  []cveDetail{{CveID: "CVE-2023-5678"}},
		},
		{
			name:  "all ids empty or missing",
			value: `[{"CveId":""},{}]`,
			want:  nil,
		},
		{
			name:  "null entry fields",
			value: `[{"CveId":"CVE-2024-0001","Severity":null,"FixedVersion":null}]`,
			want:  []cveDetail{{CveID: "CVE-2024-0001"}},
		},
		{
			name:  "empty string",
			value: "",
			want:  nil,
		},
		{
			name:  "empty json array",
			value: `[]`,
			want:  nil,
		},
		{
			name:  "json null",
			value: `null`,
			want:  nil,
		},
		{
			name:  "empty decoded array",
			value: []any{},
			want:  nil,
		},
		{
			name:  "invalid json",
			value: "not json",
			want:  nil,
		},
		{
			name:  "wrong type",
			value: 42,
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.DeepEqual(t, tc.want, cveDetails(tc.value))
		})
	}
}

func TestDecodeAdditionalScore(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		want   float64
		wantOK bool
	}{
		{
			name:   "json string",
			value:  "5.5",
			want:   5.5,
			wantOK: true,
		},
		{
			name:   "decoded number",
			value:  float64(9.8),
			want:   9.8,
			wantOK: true,
		},
		{
			name:   "nil value",
			value:  nil,
			want:   0,
			wantOK: false,
		},
		{
			name:   "empty string",
			value:  "",
			want:   0,
			wantOK: false,
		},
		{
			name:   "invalid string",
			value:  "high",
			want:   0,
			wantOK: false,
		},
		{
			name:   "wrong type",
			value:  map[string]any{"score": 1},
			want:   0,
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score, ok := decodeAdditional[float64](tc.value)
			assert.Check(t, ok == tc.wantOK)
			assert.Check(t, score == tc.want)
		})
	}
}
