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
		{"SUBMITTED", &EvidenceStatusSubmitted},
		{"READY_FOR_AUDITOR", &EvidenceStatusReadyForAuditor},
		{"AUDITOR_APPROVED", &EvidenceStatusAuditorApproved},
		{"IN_REVIEW", &EvidenceStatusInReview},
		{"MISSING_ARTIFACT", &EvidenceStatusMissingArtifact},
		{"NEEDS_RENEWAL", &EvidenceStatusNeedsRenewal},
		{"REJECTED", &EvidenceStatusRejected},
		{"submitted", &EvidenceStatusSubmitted},
		{"ready_for_auditor", &EvidenceStatusReadyForAuditor},
		{"auditor_approved", &EvidenceStatusAuditorApproved},
		{"in_review", &EvidenceStatusInReview},
		{"missing_artifact", &EvidenceStatusMissingArtifact},
		{"needs_renewal", &EvidenceStatusNeedsRenewal},
		{"rejected", &EvidenceStatusRejected},
		{"Submitted", &EvidenceStatusSubmitted},
		{"Ready_For_Auditor", &EvidenceStatusReadyForAuditor},
		{"Auditor_Approved", &EvidenceStatusAuditorApproved},
		{"In_Review", &EvidenceStatusInReview},
		{"Missing_Artifact", &EvidenceStatusMissingArtifact},
		{"Needs_Renewal", &EvidenceStatusNeedsRenewal},
		{"Rejected", &EvidenceStatusRejected},
		{"invalid", &EvidenceStatusInvalid},
		{"", &EvidenceStatusInvalid},
	}

	for _, test := range tests {
		result := ToEvidenceStatus(test.input)
		assert.Equal(t, test.expected, result, "ToEvidenceStatus(%s)", test.input)
	}
}
