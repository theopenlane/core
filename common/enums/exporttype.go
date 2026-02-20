package enums

import "io"

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

var exportTypeValues = []ExportType{
	ExportTypeAsset,
	ExportTypeControl,
	ExportTypeDirectoryMembership,
	ExportTypeEntity,
	ExportTypeEvidence,
	ExportTypeFinding,
	ExportTypeIdentityHolder,
	ExportTypeInternalPolicy,
	ExportTypeProcedure,
	ExportTypeRemediation,
	ExportTypeReview,
	ExportTypeRisk,
	ExportTypeSubprocessor,
	ExportTypeSubscriber,
	ExportTypeTask,
	ExportTypeTrustCenterFaq,
	ExportTypeTrustCenterSubprocessor,
	ExportTypeVulnerability,
}

// Values returns a slice of strings representing all valid ExportType values.
func (ExportType) Values() []string { return stringValues(exportTypeValues) }

// String returns the string representation of the ExportType value.
func (r ExportType) String() string { return string(r) }

// ToExportType converts a string to its corresponding ExportType enum value.
func ToExportType(r string) *ExportType { return parse(r, exportTypeValues, &ExportTypeInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
