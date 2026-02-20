package enums

import (
	"fmt"
	"io"
	"strings"
)

// ExportType is a custom type representing the various states of ExportType.
type ExportType string

var (
	// ExportTypeAsset indicates the asset.
	ExportTypeAsset ExportType = "ASSET"
	// ExportTypeControl indicates the control.
	ExportTypeControl ExportType = "CONTROL"
	// ExportTypeDirectoryMembership indicates the directorymembership.
	ExportTypeDirectoryMembership ExportType = "DIRECTORY_MEMBERSHIP"
	// ExportTypeEntity indicates the entity.
	ExportTypeEntity ExportType = "ENTITY"
	// ExportTypeEvidence indicates the evidence.
	ExportTypeEvidence ExportType = "EVIDENCE"
	// ExportTypeFinding indicates the finding.
	ExportTypeFinding ExportType = "FINDING"
	// ExportTypeIdentityHolder indicates the identityholder.
	ExportTypeIdentityHolder ExportType = "IDENTITY_HOLDER"
	// ExportTypeInternalPolicy indicates the internalpolicy.
	ExportTypeInternalPolicy ExportType = "INTERNAL_POLICY"
	// ExportTypeProcedure indicates the procedure.
	ExportTypeProcedure ExportType = "PROCEDURE"
	// ExportTypeRemediation indicates the remediation.
	ExportTypeRemediation ExportType = "REMEDIATION"
	// ExportTypeReview indicates the review.
	ExportTypeReview ExportType = "REVIEW"
	// ExportTypeRisk indicates the risk.
	ExportTypeRisk ExportType = "RISK"
	// ExportTypeSubprocessor indicates the subprocessor.
	ExportTypeSubprocessor ExportType = "SUBPROCESSOR"
	// ExportTypeSubscriber indicates the subscriber.
	ExportTypeSubscriber ExportType = "SUBSCRIBER"
	// ExportTypeTask indicates the task.
	ExportTypeTask ExportType = "TASK"
	// ExportTypeTrustCenterFaq indicates the trustcenterfaq.
	ExportTypeTrustCenterFaq ExportType = "TRUST_CENTER_FAQ"
	// ExportTypeTrustCenterSubprocessor indicates the trustcentersubprocessor.
	ExportTypeTrustCenterSubprocessor ExportType = "TRUST_CENTER_SUBPROCESSOR"
	// ExportTypeVulnerability indicates the vulnerability.
	ExportTypeVulnerability ExportType = "VULNERABILITY"
	// ExportTypeInvalid is used when an unknown or unsupported value is provided.
	ExportTypeInvalid ExportType = "EXPORTTYPE_INVALID"
)

// Values returns a slice of strings representing all valid ExportType values.
func (ExportType) Values() []string {
	return []string{
		string(ExportTypeAsset),
		string(ExportTypeControl),
		string(ExportTypeDirectoryMembership),
		string(ExportTypeEntity),
		string(ExportTypeEvidence),
		string(ExportTypeFinding),
		string(ExportTypeIdentityHolder),
		string(ExportTypeInternalPolicy),
		string(ExportTypeProcedure),
		string(ExportTypeRemediation),
		string(ExportTypeReview),
		string(ExportTypeRisk),
		string(ExportTypeSubprocessor),
		string(ExportTypeSubscriber),
		string(ExportTypeTask),
		string(ExportTypeTrustCenterFaq),
		string(ExportTypeTrustCenterSubprocessor),
		string(ExportTypeVulnerability),
	}
}

// String returns the string representation of the ExportType value.
func (r ExportType) String() string {
	return string(r)
}

// ToExportType converts a string to its corresponding ExportType enum value.
func ToExportType(r string) *ExportType {
	switch strings.ToUpper(r) {
	case ExportTypeAsset.String():
		return &ExportTypeAsset
	case ExportTypeControl.String():
		return &ExportTypeControl
	case ExportTypeDirectoryMembership.String():
		return &ExportTypeDirectoryMembership
	case ExportTypeEntity.String():
		return &ExportTypeEntity
	case ExportTypeEvidence.String():
		return &ExportTypeEvidence
	case ExportTypeFinding.String():
		return &ExportTypeFinding
	case ExportTypeIdentityHolder.String():
		return &ExportTypeIdentityHolder
	case ExportTypeInternalPolicy.String():
		return &ExportTypeInternalPolicy
	case ExportTypeProcedure.String():
		return &ExportTypeProcedure
	case ExportTypeRemediation.String():
		return &ExportTypeRemediation
	case ExportTypeReview.String():
		return &ExportTypeReview
	case ExportTypeRisk.String():
		return &ExportTypeRisk
	case ExportTypeSubprocessor.String():
		return &ExportTypeSubprocessor
	case ExportTypeSubscriber.String():
		return &ExportTypeSubscriber
	case ExportTypeTask.String():
		return &ExportTypeTask
	case ExportTypeTrustCenterFaq.String():
		return &ExportTypeTrustCenterFaq
	case ExportTypeTrustCenterSubprocessor.String():
		return &ExportTypeTrustCenterSubprocessor
	case ExportTypeVulnerability.String():
		return &ExportTypeVulnerability
	default:
		return &ExportTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ExportType, got: %T", v) //nolint:err113
	}

	*r = ExportType(str)

	return nil
}
