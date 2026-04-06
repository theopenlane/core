package enums

import "io"

// ExportType is a custom type representing the various states of ExportType.
type ExportType string

var (
	// ExportTypeAssessment indicates the assessment.
	ExportTypeAssessment ExportType = "ASSESSMENT"
	// ExportTypeAsset indicates the asset.
	ExportTypeAsset ExportType = "ASSET"
	// ExportTypeCampaign indicates the campaign.
	ExportTypeCampaign ExportType = "CAMPAIGN"
	// ExportTypeContact indicates the contact.
	ExportTypeContact ExportType = "CONTACT"
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
	// ExportTypeSystemDetail indicates the systemdetail.
	ExportTypeSystemDetail ExportType = "SYSTEM_DETAIL"
	// ExportTypeTask indicates the task.
	ExportTypeTask ExportType = "TASK"
	// ExportTypeTrustCenterFaq indicates the trustcenterfaq.
	ExportTypeTrustCenterFaq ExportType = "TRUST_CENTER_FAQ"
	// ExportTypeTrustCenterSubprocessor indicates the trustcentersubprocessor.
	ExportTypeTrustCenterSubprocessor ExportType = "TRUST_CENTER_SUBPROCESSOR"
	// ExportTypeVendorRiskScore indicates the vendorriskscore.
	ExportTypeVendorRiskScore ExportType = "VENDOR_RISK_SCORE"
	// ExportTypeVendorScoringConfig indicates the vendorscoringconfig.
	ExportTypeVendorScoringConfig ExportType = "VENDOR_SCORING_CONFIG"
	// ExportTypeVulnerability indicates the vulnerability.
	ExportTypeVulnerability ExportType = "VULNERABILITY"
	// ExportTypeInvalid is used when an unknown or unsupported value is provided.
	ExportTypeInvalid ExportType = "EXPORTTYPE_INVALID"
)

var exportTypeValues = []ExportType{
	ExportTypeAssessment,
	ExportTypeAsset,
	ExportTypeCampaign,
	ExportTypeContact,
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
	ExportTypeSystemDetail,
	ExportTypeTask,
	ExportTypeTrustCenterFaq,
	ExportTypeTrustCenterSubprocessor,
	ExportTypeVendorRiskScore,
	ExportTypeVendorScoringConfig,
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
