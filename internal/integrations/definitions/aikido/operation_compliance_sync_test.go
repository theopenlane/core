package aikido

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/mappingtest"
)

func TestCollectCompliance(t *testing.T) {
	examples := map[string]string{
		fmt.Sprintf(pathComplianceOverviewFormat, "soc2"): "checks/soc2_checks.json",
		fmt.Sprintf(pathComplianceOverviewFormat, "iso"):  "checks/iso27001_checks.json",
		fmt.Sprintf(pathComplianceOverviewFormat, "pci"):  "checks/pci_checks.json",
		fmt.Sprintf(pathComplianceOverviewFormat, "gdpr"): "checks/gdpr_checks.json",
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		example, ok := examples[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		writeJSON(w, string(mappingtest.LoadExample(t, "examples", example)))
	})

	payloads, err := CollectCompliance(context.Background(), client)
	assert.NilError(t, err)

	envelopes := payloadSet(payloads, entityops.SchemaCheckResult.Name)
	assert.Check(t, is.Len(envelopes, 22))

	refCodes := map[string]int{}

	for _, envelope := range envelopes {
		var check complianceCheck
		assert.NilError(t, json.Unmarshal(envelope.Payload, &check))

		refCodes[check.Standard+" "+check.RefCode]++
	}

	assert.Check(t, is.DeepEqual(map[string]int{"SOC 2 CC3.3": 16, "ISO 27001 A.8.2": 3, "PCI DSS 1.2": 2, "GDPR 2.1": 1}, refCodes))
}

func TestCollectComplianceFailsWithoutScope(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, err := CollectCompliance(context.Background(), client)
	assert.Check(t, is.ErrorIs(err, ErrComplianceFetchFailed))
	assert.Check(t, is.ErrorIs(err, ErrUnauthorized))
}

func TestComplianceRefCode(t *testing.T) {
	tests := []struct {
		name     string
		control  complianceControl
		expected string
	}{
		{name: "soc 2 code", control: complianceControl{Code: "cc3.3", Title: "CC3.3: Consider the potential for fraud"}, expected: "CC3.3"},
		{name: "iso title reference", control: complianceControl{Code: "privileged_access_rights", Title: "A.8.2 - Privileged access rights"}, expected: "A.8.2"},
		{name: "pci numbered code", control: complianceControl{Code: "1.2_network_security_controls_are_configured", Title: "1.2 Network security controls (NSCs) are configured and maintained. "}, expected: "1.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.expected, complianceRefCode(tt.control)))
		})
	}
}
