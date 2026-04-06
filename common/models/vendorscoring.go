package models

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/theopenlane/core/common/enums"
)

// VendorScoringQuestionDef defines a single vendor scoring question.
// Impact and likelihood are not stored here — both are per-vendor on VendorRiskScore.
type VendorScoringQuestionDef struct {
	// Key is the stable identifier used in VendorRiskScore.question_key; never changes after initial use.
	// For CAIQ-sourced questions this is the CAIQ question ID (e.g. "IAM-14.1").
	Key string `json:"key"`
	// Name is the human-readable label for this question
	Name string `json:"name"`
	// Description explains what the question is evaluating
	Description string `json:"description,omitempty"`
	// Category is the taxonomy grouping for this question
	Category enums.VendorScoringCategory `json:"category"`
	// AnswerType defines the expected input format for the answer field
	AnswerType enums.VendorScoringAnswerType `json:"answerType"`
	// AnswerOptions lists valid values for SINGLE_SELECT questions; empty for all other types
	AnswerOptions []string `json:"answerOptions,omitempty"`
	// SuggestedImpact is the default impact pre-populated on VendorRiskScore at creation;
	// assessors override per vendor based on the vendor's specific risk context
	SuggestedImpact enums.VendorRiskImpact `json:"suggestedImpact"`
	// Enabled controls whether this question is active; set to false to retire a question
	// without removing it (removing a key orphans existing VendorRiskScore rows)
	Enabled bool `json:"enabled"`
}

// VendorScoringQuestionsConfig is stored as a JSON field on VendorScoringConfig.
// Only org-custom questions are persisted; system defaults always come from DefaultVendorScoringQuestions.
// Custom entries with the same Key as a system default replace the default entry,
// allowing per-org wording changes, impact adjustments, or disabling of system defaults.
type VendorScoringQuestionsConfig struct {
	// Custom holds org-specific question additions and overrides of system defaults
	Custom []VendorScoringQuestionDef `json:"custom"`
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v VendorScoringQuestionsConfig) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, v)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *VendorScoringQuestionsConfig) UnmarshalGQL(val any) error {
	return unmarshalGQLJSON(val, v)
}

// All returns the merged set of system defaults and org-custom questions.
// Custom entries with the same Key as a system default replace the default entry.
// Pure custom-key entries (not in defaults) are appended after the defaults.
func (v VendorScoringQuestionsConfig) All() []VendorScoringQuestionDef {
	if len(v.Custom) == 0 {
		return DefaultVendorScoringQuestions
	}

	overrides := make(map[string]VendorScoringQuestionDef, len(v.Custom))
	for _, c := range v.Custom {
		overrides[c.Key] = c
	}

	merged := make([]VendorScoringQuestionDef, 0, len(DefaultVendorScoringQuestions)+len(v.Custom))

	for _, d := range DefaultVendorScoringQuestions {
		if override, ok := overrides[d.Key]; ok {
			merged = append(merged, override)
			delete(overrides, d.Key)
		} else {
			merged = append(merged, d)
		}
	}

	// remaining entries in overrides are custom-only keys not present in defaults
	for _, c := range v.Custom {
		if _, remaining := overrides[c.Key]; remaining {
			merged = append(merged, c)
		}
	}

	return merged
}

// RiskThreshold maps a VendorRiskRating to its upper score bound
type RiskThreshold struct {
	// Rating is the risk rating tier
	Rating enums.VendorRiskRating `json:"rating"`
	// MaxScore is the upper bound (inclusive) for this tier
	MaxScore float64 `json:"maxScore"`
}

// DefaultRiskThresholds defines the system-default risk rating bands
var DefaultRiskThresholds = []RiskThreshold{
	{Rating: enums.VendorRiskRatingNone, MaxScore: 0},
	{Rating: enums.VendorRiskRatingVeryLow, MaxScore: 3},
	{Rating: enums.VendorRiskRatingLow, MaxScore: 5},
	{Rating: enums.VendorRiskRatingMedium, MaxScore: 11},
	{Rating: enums.VendorRiskRatingHigh, MaxScore: 15},
	{Rating: enums.VendorRiskRatingCritical, MaxScore: 20},
}

// RiskThresholdsConfig is stored as a JSON field on VendorScoringConfig.
// Only org-custom overrides are persisted; system defaults come from DefaultRiskThresholds.
// Custom entries with the same Rating as a default replace the default's MaxScore.
type RiskThresholdsConfig struct {
	// Custom holds org-specific threshold overrides keyed by Rating
	Custom []RiskThreshold `json:"custom"`
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (v RiskThresholdsConfig) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, v)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (v *RiskThresholdsConfig) UnmarshalGQL(val any) error {
	return unmarshalGQLJSON(val, v)
}

// All returns the merged set of default and custom thresholds sorted by MaxScore ascending.
// Custom entries with the same Rating as a default replace the default entry.
func (v RiskThresholdsConfig) All() []RiskThreshold {
	if len(v.Custom) == 0 {
		return DefaultRiskThresholds
	}

	overrides := make(map[enums.VendorRiskRating]RiskThreshold, len(v.Custom))
	for _, c := range v.Custom {
		overrides[c.Rating] = c
	}

	merged := make([]RiskThreshold, 0, len(DefaultRiskThresholds))

	for _, d := range DefaultRiskThresholds {
		if override, ok := overrides[d.Rating]; ok {
			merged = append(merged, override)
		} else {
			merged = append(merged, d)
		}
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].MaxScore < merged[j].MaxScore
	})

	return merged
}

// Resolve returns the risk rating for a given score by finding the first threshold
// where score <= MaxScore. If the score exceeds all thresholds, the highest tier is returned.
func (v RiskThresholdsConfig) Resolve(score float64) string {
	thresholds := v.All()

	for _, t := range thresholds {
		if score <= t.MaxScore {
			return t.Rating.String()
		}
	}

	// Score exceeds all configured thresholds; return the highest tier
	return thresholds[len(thresholds)-1].Rating.String()
}

// customCategoryPrefix maps each VendorScoringCategory to a short domain prefix
// used when generating keys for org-custom questions
var customCategoryPrefix = map[enums.VendorScoringCategory]string{
	enums.VendorScoringCategorySecurityPractices:     "CUST-SP",
	enums.VendorScoringCategoryDataAccess:            "CUST-DA",
	enums.VendorScoringCategoryDataPrivacy:           "CUST-DP",
	enums.VendorScoringCategoryBusinessContinuity:    "CUST-BC",
	enums.VendorScoringCategoryIncidentResponse:      "CUST-IR",
	enums.VendorScoringCategoryRegulatoryCompliance:  "CUST-RC",
	enums.VendorScoringCategorySupplyChainRisk:       "CUST-SC",
	enums.VendorScoringCategoryFinancialStability:    "CUST-FS",
	enums.VendorScoringCategoryOperationalDependency: "CUST-OD",
}

// customKeyPattern matches keys generated by AssignCustomKeys (e.g. CUST-SP-01.01)
var customKeyPattern = regexp.MustCompile(`^(CUST-[A-Z]{2})-(\d{2})\.\d{2}$`)

// AssignCustomKeys generates stable keys for custom questions that have an empty Key field
// Keys follow the format {CUST-prefix}-{nn}.01 where nn is zero-padded and scoped
// to the category prefix. Keys that match a system default are preserved as intentional
// overrides; all other non-CUST keys are reassigned to prevent collisions
func (v *VendorScoringQuestionsConfig) AssignCustomKeys() {
	if len(v.Custom) == 0 {
		return
	}

	// Collect the highest existing sequence per prefix across all custom entries
	maxSeq := make(map[string]int)

	for _, c := range v.Custom {
		matches := customKeyPattern.FindStringSubmatch(c.Key)
		if matches == nil {
			continue
		}

		prefix := matches[1]

		seq, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		if seq > maxSeq[prefix] {
			maxSeq[prefix] = seq
		}
	}

	// Assign keys to entries that need them; reassign any non-CUST keys
	// but preserve keys that match a system default (intentional overrides)
	for i := range v.Custom {
		if isCustomGeneratedKey(v.Custom[i].Key) || isDefaultQuestionKey(v.Custom[i].Key) {
			continue
		}

		prefix := customCategoryPrefix[v.Custom[i].Category]
		if prefix == "" {
			prefix = "CUST-XX"
		}

		maxSeq[prefix]++

		v.Custom[i].Key = fmt.Sprintf("%s-%02d.01", prefix, maxSeq[prefix])
	}
}

// isCustomGeneratedKey reports whether a key was generated by AssignCustomKeys
func isCustomGeneratedKey(key string) bool {
	return strings.HasPrefix(key, "CUST-") && customKeyPattern.MatchString(key)
}

// defaultQuestionKeys is the set of keys from DefaultVendorScoringQuestions, built once at init
var defaultQuestionKeys = func() map[string]struct{} {
	keys := make(map[string]struct{}, len(DefaultVendorScoringQuestions))
	for _, q := range DefaultVendorScoringQuestions {
		keys[q.Key] = struct{}{}
	}

	return keys
}()

// isDefaultQuestionKey reports whether a key matches a system default question
func isDefaultQuestionKey(key string) bool {
	_, ok := defaultQuestionKeys[key]

	return ok
}

// IsCustomKey reports whether a key belongs to the custom question namespace
func IsCustomKey(key string) bool {
	return strings.HasPrefix(key, "CUST-")
}

// DefaultVendorScoringQuestions is a curated subset of CAIQ v4.0.3 questions selected for
// TPRM scoring. Only outcome-oriented questions are included — policy-existence questions
// ("are policies established and documented?") are excluded as they do not differentiate vendors.
//
// Keys are CAIQ question IDs and are stable permanent identifiers.
//
// IMPORTANT: never change or remove a Key once it has been used in production — VendorRiskScore
// rows reference questions by Key, and removing a Key orphans those records. To retire a question,
// set Enabled: false so the key still resolves in All() and existing answers remain displayable.
var DefaultVendorScoringQuestions = []VendorScoringQuestionDef{
	// -------------------------------------------------------------------------
	// IDENTITY & ACCESS MANAGEMENT
	// -------------------------------------------------------------------------
	{
		Key:             "IAM-05.1",
		Name:            "Is the least privilege principle employed when implementing information system access?",
		Description:     "Users and systems are granted only the minimum access required to perform their function.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactMedium,
		Enabled:         true,
	},
	{
		Key:             "IAM-07.1",
		Name:            "Is a process in place to de-provision or modify access in a timely manner for movers and leavers?",
		Description:     "Access is revoked or adjusted promptly when employees change roles or leave the organization.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "IAM-08.1",
		Name:            "Are reviews and revalidation of user access for least privilege and separation of duties completed with a frequency commensurate with organizational risk tolerance?",
		Description:     "Periodic access reviews ensure permissions remain appropriate over time.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "IAM-09.1",
		Name:            "Are processes for the segregation of privileged access roles defined and implemented such that administrative data access, encryption, key management, and logging capabilities are distinct and separate?",
		Description:     "Privileged access is segregated to prevent any single account from controlling all critical functions.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "IAM-14.1",
		Name:            "Are processes for authenticating access to systems, applications, and data assets including multi-factor authentication for least-privileged users and sensitive data access defined, implemented, and evaluated?",
		Description:     "MFA is enforced for access to systems and sensitive data, not just administrative accounts.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactCritical,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// CRYPTOGRAPHY & ENCRYPTION
	// -------------------------------------------------------------------------
	{
		Key:             "CEK-03.1",
		Name:            "Are data at-rest and in-transit cryptographically protected using cryptographic libraries certified to approved standards?",
		Description:     "All data is encrypted both when stored and when transmitted using industry-approved cryptographic standards.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactCritical,
		Enabled:         true,
	},
	{
		Key:             "CEK-12.1",
		Name:            "Are cryptographic keys rotated based on a cryptoperiod calculated while considering information disclosure risks and legal and regulatory requirements?",
		Description:     "Cryptographic keys are rotated on a defined schedule to limit exposure from key compromise.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// DATA SECURITY & ACCESS
	// -------------------------------------------------------------------------
	{
		Key:             "DSP-02.1",
		Name:            "Are industry-accepted methods applied for secure data disposal from storage media so information is not recoverable by any forensic means?",
		Description:     "Data is securely and irrecoverably destroyed when no longer needed.",
		Category:        enums.VendorScoringCategoryDataAccess,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "DSP-03.1",
		Name:            "Is a data inventory created and maintained for sensitive and personal information?",
		Description:     "The vendor maintains an up-to-date inventory of where sensitive and personal data is stored and processed.",
		Category:        enums.VendorScoringCategoryDataAccess,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "DSP-16.1",
		Name:            "Do data retention, archiving, and deletion practices follow business requirements, applicable laws, and regulations?",
		Description:     "Data is retained only for the period required and deleted in accordance with legal obligations.",
		Category:        enums.VendorScoringCategoryDataAccess,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactMedium,
		Enabled:         true,
	},
	{
		Key:             "DSP-19.1",
		Name:            "Are processes defined and implemented to specify and document physical data locations, including locales where data is processed or backed up?",
		Description:     "The vendor can identify and document all geographic locations where data is stored or processed.",
		Category:        enums.VendorScoringCategoryDataAccess,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// DATA PRIVACY
	// -------------------------------------------------------------------------
	{
		Key:             "DSP-08.1",
		Name:            "Are systems, products, and business practices based on privacy principles by design and according to industry best practices?",
		Description:     "Privacy is built into systems and processes from the outset rather than added as an afterthought.",
		Category:        enums.VendorScoringCategoryDataPrivacy,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactMedium,
		Enabled:         true,
	},
	{
		Key:             "DSP-09.1",
		Name:            "Is a data protection impact assessment (DPIA) conducted when processing personal data and evaluating the origin, nature, particularity, and severity of risks?",
		Description:     "Formal DPIAs are conducted before processing personal data to identify and mitigate privacy risks.",
		Category:        enums.VendorScoringCategoryDataPrivacy,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "DSP-10.1",
		Name:            "Are processes defined to ensure any transfer of personal or sensitive data is protected from unauthorized access and only processed within scope?",
		Description:     "Personal data transfers are controlled, authorized, and protected in transit.",
		Category:        enums.VendorScoringCategoryDataPrivacy,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "DSP-11.1",
		Name:            "Are processes defined to enable data subjects to request access to, modify, or delete personal data per applicable laws and regulations?",
		Description:     "The vendor supports data subject rights including access, rectification, and erasure requests.",
		Category:        enums.VendorScoringCategoryDataPrivacy,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// BUSINESS CONTINUITY & RESILIENCE
	// -------------------------------------------------------------------------
	{
		Key:             "BCR-06.1",
		Name:            "Are the business continuity and operational resilience plans exercised and tested at least annually and when significant changes occur?",
		Description:     "BCP is tested regularly to verify it is effective and that staff know how to execute it.",
		Category:        enums.VendorScoringCategoryBusinessContinuity,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "BCR-08.1",
		Name:            "Is cloud data periodically backed up?",
		Description:     "Data is backed up on a regular schedule to enable recovery from data loss events.",
		Category:        enums.VendorScoringCategoryBusinessContinuity,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "BCR-09.1",
		Name:            "Is a disaster response plan established, documented, approved, and maintained to ensure recovery from natural and man-made disasters?",
		Description:     "A formal disaster recovery plan exists and is kept current.",
		Category:        enums.VendorScoringCategoryBusinessContinuity,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "BCR-10.1",
		Name:            "Is the disaster response plan exercised annually or when significant changes occur?",
		Description:     "DR plan effectiveness is validated through regular exercises.",
		Category:        enums.VendorScoringCategoryBusinessContinuity,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// THREAT & VULNERABILITY MANAGEMENT
	// -------------------------------------------------------------------------
	{
		Key:             "TVM-03.1",
		Name:            "Are processes defined, implemented, and evaluated to enable scheduled and emergency responses to vulnerability identifications based on identified risk?",
		Description:     "Vulnerabilities are remediated on a risk-based schedule with a defined process for emergency response.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "TVM-06.1",
		Name:            "Are processes defined, implemented, and evaluated for periodic, independent, third-party penetration testing?",
		Description:     "Independent penetration tests are conducted regularly by qualified third parties.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "TVM-07.1",
		Name:            "Are processes defined, implemented, and evaluated for vulnerability detection on organizationally managed assets at least monthly?",
		Description:     "Vulnerability scanning is performed at least monthly across managed assets.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// SECURITY INCIDENT MANAGEMENT
	// -------------------------------------------------------------------------
	{
		Key:             "SEF-03.1",
		Name:            "Is a security incident response plan that includes relevant internal departments, impacted customers, and supply-chain relationships established, documented, and maintained?",
		Description:     "A comprehensive incident response plan covers internal teams, customers, and supply chain partners.",
		Category:        enums.VendorScoringCategoryIncidentResponse,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "SEF-04.1",
		Name:            "Is the security incident response plan tested and updated for effectiveness at planned intervals or upon significant organizational or environmental changes?",
		Description:     "The incident response plan is validated through testing and kept current.",
		Category:        enums.VendorScoringCategoryIncidentResponse,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "SEF-07.1",
		Name:            "Are processes, procedures, and technical measures for security breach notifications defined and implemented?",
		Description:     "A defined breach notification process exists with clear procedures for notifying affected parties.",
		Category:        enums.VendorScoringCategoryIncidentResponse,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactCritical,
		Enabled:         true,
	},
	{
		Key:             "SEF-07.2",
		Name:            "Are security breaches and assumed security breaches reported, including any relevant supply chain breaches, per applicable SLAs, laws, and regulations?",
		Description:     "Breaches are reported to affected parties and regulators within legally required timeframes.",
		Category:        enums.VendorScoringCategoryIncidentResponse,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactCritical,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// AUDIT & ASSURANCE
	// -------------------------------------------------------------------------
	{
		Key:             "A&A-02.1",
		Name:            "Are independent audit and assurance assessments conducted according to relevant standards at least annually?",
		Description:     "Third-party audits (e.g. SOC 2, ISO 27001) are conducted at least annually by qualified independent assessors.",
		Category:        enums.VendorScoringCategoryRegulatoryCompliance,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// SUPPLY CHAIN MANAGEMENT
	// -------------------------------------------------------------------------
	{
		Key:             "STA-07.1",
		Name:            "Is an inventory of all supply chain relationships developed and maintained?",
		Description:     "The vendor maintains a current inventory of their own subprocessors and supply chain partners.",
		Category:        enums.VendorScoringCategorySupplyChainRisk,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactMedium,
		Enabled:         true,
	},
	{
		Key:             "STA-09.1",
		Name:            "Do service agreements between the vendor and their customers incorporate security requirements, incident management, right to audit, and data privacy provisions?",
		Description:     "Contractual agreements with customers include substantive security and privacy obligations.",
		Category:        enums.VendorScoringCategorySupplyChainRisk,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "STA-14.1",
		Name:            "Is a process to conduct periodic security assessments for all supply chain organizations defined and implemented?",
		Description:     "The vendor actively assesses security risk across their own supply chain, not just their direct controls.",
		Category:        enums.VendorScoringCategorySupplyChainRisk,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// GOVERNANCE, RISK & COMPLIANCE
	// -------------------------------------------------------------------------
	{
		Key:             "GRC-05.1",
		Name:            "Has an information security program been developed and implemented?",
		Description:     "A formal, operational information security program exists covering all relevant control domains.",
		Category:        enums.VendorScoringCategoryRegulatoryCompliance,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// HUMAN RESOURCES SECURITY
	// -------------------------------------------------------------------------
	{
		Key:             "HRS-01.1",
		Name:            "Are background verification policies and procedures established for all new employees, contractors, and third parties?",
		Description:     "Background checks are conducted for all personnel with access to organizational systems or data.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactMedium,
		Enabled:         true,
	},
	{
		Key:             "HRS-11.1",
		Name:            "Is a security awareness training program established, documented, and maintained for all employees?",
		Description:     "All employees receive regular security awareness training.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactMedium,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// APPLICATION SECURITY
	// -------------------------------------------------------------------------
	{
		Key:             "AIS-04.1",
		Name:            "Is an SDLC process defined and implemented for application design, development, deployment, and operation per organizationally designed security requirements?",
		Description:     "Security is integrated throughout the software development lifecycle, not applied only at release.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
	{
		Key:             "AIS-07.1",
		Name:            "Are application security vulnerabilities remediated following defined processes?",
		Description:     "A defined process exists for tracking and remediating application security vulnerabilities to closure.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// INFRASTRUCTURE & NETWORK SECURITY
	// -------------------------------------------------------------------------
	{
		Key:             "IVS-03.2",
		Name:            "Are communications between environments encrypted?",
		Description:     "All inter-environment communications (prod, staging, cloud, on-prem) are encrypted in transit.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactCritical,
		Enabled:         true,
	},

	// -------------------------------------------------------------------------
	// LOGGING & MONITORING
	// -------------------------------------------------------------------------
	{
		Key:             "LOG-03.1",
		Name:            "Are security-related events identified and monitored within applications and the underlying infrastructure?",
		Description:     "Security events are actively monitored across applications and infrastructure with alerting in place.",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	},
}
