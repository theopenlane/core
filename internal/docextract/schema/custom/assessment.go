package custom

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/docextract/schema"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

// RiskRatings are the assessment risk ratings ordered from least to most severe
var RiskRatings = []enums.VendorRiskRating{
	enums.VendorRiskRatingLow,
	enums.VendorRiskRatingMedium,
	enums.VendorRiskRatingHigh,
	enums.VendorRiskRatingCritical,
}

// RiskRatingKey is the assessment field holding the risk rating, raised to the opinion floor after extraction
const RiskRatingKey = "risk_rating"

// VendorKey is the assessment field holding the vendor the report covers, keyed by its entity fields
const VendorKey = "vendor"

// OutcomePostures are how each review outcome reads as an overall risk posture
var OutcomePostures = map[string]string{
	models.ReviewOutcomeLowRisk:      "low",
	models.ReviewOutcomeModerateRisk: "moderate",
	models.ReviewOutcomeConditional:  "elevated",
	models.ReviewOutcomeNotApproved:  "high",
}

// OutcomeRatings are the vendor risk ratings that agree with each review outcome; a reviewed vendor is never
// unassessed, so none agrees with no outcome
var OutcomeRatings = map[string][]enums.VendorRiskRating{
	models.ReviewOutcomeLowRisk:      {enums.VendorRiskRatingVeryLow, enums.VendorRiskRatingLow},
	models.ReviewOutcomeModerateRisk: {enums.VendorRiskRatingMedium},
	models.ReviewOutcomeConditional:  {enums.VendorRiskRatingHigh},
	models.ReviewOutcomeNotApproved:  {enums.VendorRiskRatingHigh, enums.VendorRiskRatingCritical},
}

// OutcomeScore is the impact and likelihood a review outcome is scored with on the review outcome question
type OutcomeScore struct {
	// Impact is the impact of the outcome
	Impact enums.VendorRiskImpact
	// Likelihood is the likelihood of the outcome
	Likelihood enums.VendorRiskLikelihood
}

// OutcomeScores set each review outcome's baseline score inside the rating band it agrees with: low 4, medium 6,
// high 12, and critical 20
var OutcomeScores = map[string]OutcomeScore{
	models.ReviewOutcomeLowRisk:      {Impact: enums.VendorRiskImpactLow, Likelihood: enums.VendorRiskLikelihoodMedium},
	models.ReviewOutcomeModerateRisk: {Impact: enums.VendorRiskImpactMedium, Likelihood: enums.VendorRiskLikelihoodMedium},
	models.ReviewOutcomeConditional:  {Impact: enums.VendorRiskImpactHigh, Likelihood: enums.VendorRiskLikelihoodHigh},
	models.ReviewOutcomeNotApproved:  {Impact: enums.VendorRiskImpactCritical, Likelihood: enums.VendorRiskLikelihoodVeryHigh},
}

// scoringLikelihoods are the likelihoods a scoring answer may carry, the vendor scoring scale
var scoringLikelihoods = []string{
	enums.VendorRiskLikelihoodVeryLow.String(),
	enums.VendorRiskLikelihoodLow.String(),
	enums.VendorRiskLikelihoodMedium.String(),
	enums.VendorRiskLikelihoodHigh.String(),
	enums.VendorRiskLikelihoodVeryHigh.String(),
}

// AssessmentSchema asks for a vendor risk read of a report: its summary, risk rating, the risks the reviewing
// organization keeps, its answers to the vendor scoring questions, and the vendor it covers
var AssessmentSchema = schema.Object("assessment",
	schema.Nested(VendorKey,
		schema.From(entityops.EntityFields.Name).Required().Describe("The name customers know the company by, the product or brand name when the report names one, not the legal holding company"),
		schema.From(entityops.EntityFields.Domains).Describe("The company's web domains mentioned in the report, lowercase without a scheme or path, e.g. example.com"),
		schema.From(entityops.EntityFields.ProvidedServices).Describe("The services the company provides to customers, from the report's system description, each a short name such as Container orchestration or Managed DNS"),
		schema.From(entityops.EntityFields.Description).Describe("One or two sentences on what the company provides, from the report's system description"),
		schema.From(entityops.EntityFields.Soc2PeriodEnd).Required().Describe("The last day the report covers as YYYY-MM-DD: the end of the review period for a Type 2 report, or the as-of date for a Type 1 report"),
	).Required(),
	schema.From(entityops.ProgramFields.AuditFirm).Required().Describe("The audit firm that performed the examination, as the report names it"),
	schema.Objects("potential_risks",
		schema.Extra("title", docextract.TypeString).Required().Describe("A short heading for the risk, such as Shared Responsibility Model"),
		schema.Extra("description", docextract.TypeString).Required().Describe("Two to five sentences explaining the risk for the reviewing organization, quoting the report and citing its page where it helps"),
		schema.Extra("customer_responsibilities", docextract.TypeArray).Describe("What the reviewing organization keeps responsibility for, each a short noun phrase; empty when the risk carries none"),
	).Required(),
	schema.Extra("positive_observations", docextract.TypeArray).Describe("Meaningful evidence of effective security controls, independent assurance, and program maturity, each one sentence"),
	schema.Extra("recommendation_outcome", docextract.TypeString).Required().Enum(models.ReviewOutcomes...).Describe("The overall outcome of the review"),
	schema.Extra("recommendation", docextract.TypeString).Required().Describe("One or two sentences justifying the outcome, with specific follow-up actions only when warranted"),
	schema.Objects("scoring_answers",
		schema.Extra("question_key", docextract.TypeString).Required().Describe("The key of the vendor scoring question answered"),
		schema.Extra("answer", docextract.TypeString).Required().Describe("The answer in the question's answer format, true or false for yes or no questions"),
		schema.Extra("likelihood", docextract.TypeString).Required().Enum(scoringLikelihoods...).Describe("How likely the risk the question guards against is for this vendor"),
		schema.Extra("notes", docextract.TypeString).Required().Describe("The evidence in the report behind the answer, citing its section or page"),
	),
	schema.From(entityops.ReviewFields.Summary).Required().Describe("A short summary of the report for a reader deciding whether to rely on this vendor: what system it covers, the trust services categories in scope, and anything notable"),
	schema.Extra(RiskRatingKey, docextract.TypeString).Required().Enum(lo.Map(RiskRatings, enumString)...).Describe("The risk of relying on this vendor based on the auditor's opinion, the findings, and the scope of the report"),
	schema.Extra("risk_rationale", docextract.TypeString).Required().Describe("The reasons for the risk rating, citing the opinion and any findings that drove it"),
)

// enumString is a vendor risk rating's value, for listing ratings as enum options
func enumString(rating enums.VendorRiskRating, _ int) string {
	return rating.String()
}
