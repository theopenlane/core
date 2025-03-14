package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToEvidenceStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *EvidenceStatus
	}{
		{"READY", &EvidenceReady},
		{"APPROVED", &EvidenceApproved},
		{"MISSING_ARTIFACT", &EvidenceMissingArtifact},
		{"REJECTED", &EvidenceRejected},
		{"NEEDS_RENEWAL", &EvidenceNeedsRenewal},
		{"ready", &EvidenceReady},
		{"approved", &EvidenceApproved},
		{"missing_artifact", &EvidenceMissingArtifact},
		{"rejected", &EvidenceRejected},
		{"needs_renewal", &EvidenceNeedsRenewal},
		{"Ready", &EvidenceReady},
		{"Approved", &EvidenceApproved},
		{"Missing_Artifact", &EvidenceMissingArtifact},
		{"Rejected", &EvidenceRejected},
		{"Needs_Renewal", &EvidenceNeedsRenewal},
		{"invalid", nil},
		{"", nil},
	}

	for _, test := range tests {
		result := ToEvidenceStatus(test.input)
		assert.Equal(t, test.expected, result, "ToEvidenceStatus(%s)", test.input)
	}
}
