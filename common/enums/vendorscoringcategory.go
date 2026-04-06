package enums

import "io"

// VendorScoringCategory represents the taxonomy category for a vendor scoring question
type VendorScoringCategory string

var (
	// VendorScoringCategoryDataAccess covers questions about what data the vendor accesses or processes
	VendorScoringCategoryDataAccess VendorScoringCategory = "DATA_ACCESS"
	// VendorScoringCategorySecurityPractices covers questions about the vendor's security controls and certifications
	VendorScoringCategorySecurityPractices VendorScoringCategory = "SECURITY_PRACTICES"
	// VendorScoringCategoryRegulatoryCompliance covers questions about the vendor's regulatory and standards compliance
	VendorScoringCategoryRegulatoryCompliance VendorScoringCategory = "REGULATORY_COMPLIANCE"
	// VendorScoringCategoryFinancialStability covers questions about the vendor's financial health and viability
	VendorScoringCategoryFinancialStability VendorScoringCategory = "FINANCIAL_STABILITY"
	// VendorScoringCategoryOperationalDependency covers questions about how critical the vendor is to operations
	VendorScoringCategoryOperationalDependency VendorScoringCategory = "OPERATIONAL_DEPENDENCY"
	// VendorScoringCategoryBusinessContinuity covers questions about the vendor's disaster recovery and uptime capabilities
	VendorScoringCategoryBusinessContinuity VendorScoringCategory = "BUSINESS_CONTINUITY"
	// VendorScoringCategorySupplyChainRisk covers questions about the vendor's own third-party risk exposure
	VendorScoringCategorySupplyChainRisk VendorScoringCategory = "SUPPLY_CHAIN_RISK"
	// VendorScoringCategoryIncidentResponse covers questions about the vendor's incident detection and response process
	VendorScoringCategoryIncidentResponse VendorScoringCategory = "INCIDENT_RESPONSE"
	// VendorScoringCategoryDataPrivacy covers questions about the vendor's privacy policies and data subject rights
	VendorScoringCategoryDataPrivacy VendorScoringCategory = "DATA_PRIVACY"
	// VendorScoringCategoryInvalid is returned when parsing an unrecognized value
	VendorScoringCategoryInvalid VendorScoringCategory = "INVALID"
)

var vendorScoringCategoryValues = []VendorScoringCategory{
	VendorScoringCategoryDataAccess,
	VendorScoringCategorySecurityPractices,
	VendorScoringCategoryRegulatoryCompliance,
	VendorScoringCategoryFinancialStability,
	VendorScoringCategoryOperationalDependency,
	VendorScoringCategoryBusinessContinuity,
	VendorScoringCategorySupplyChainRisk,
	VendorScoringCategoryIncidentResponse,
	VendorScoringCategoryDataPrivacy,
}

// Values returns a slice of strings that represents all the possible values of the VendorScoringCategory enum
func (VendorScoringCategory) Values() []string { return stringValues(vendorScoringCategoryValues) }

// String returns the VendorScoringCategory as a string
func (v VendorScoringCategory) String() string { return string(v) }

// ToVendorScoringCategory returns the VendorScoringCategory based on string input
func ToVendorScoringCategory(v string) *VendorScoringCategory {
	return parse(v, vendorScoringCategoryValues, &VendorScoringCategoryInvalid)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorScoringCategory) MarshalGQL(w io.Writer) { marshalGQL(v, w) }

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorScoringCategory) UnmarshalGQL(val any) error { return unmarshalGQL(v, val) }
