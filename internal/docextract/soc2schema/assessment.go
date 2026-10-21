package soc2schema

import (
	"slices"

	"github.com/theopenlane/precis/docextract"
	"github.com/theopenlane/precis/docextract/kinds/soc2"
	"github.com/theopenlane/precis/docextract/schema"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// from returns the schema property for an entity field, keyed by its input key and typed from its go type
func from(field entityops.FieldDescriptor) schema.Field {
	return schema.From(field.Name, field.Type)
}

// RiskRatingKey is the assessment field holding the risk rating, raised to the opinion floor after extraction
const RiskRatingKey = "risk_rating"

// VendorKey is the assessment field holding the vendor the report covers, keyed by its entity fields
const VendorKey = "vendor"

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
		from(entityops.EntityFields.Name).Required().Describe("The name customers know the company by, the product or brand name when the report names one, not the legal holding company"),
		from(entityops.EntityFields.Domains).Describe("The company's web domains mentioned in the report, lowercase without a scheme or path, e.g. example.com"),
		from(entityops.EntityFields.ProvidedServices).Describe("The services the company provides to customers, from the report's system description, each a short name such as Container orchestration or Managed DNS"),
		from(entityops.EntityFields.Description).Describe("One or two sentences on what the company provides, from the report's system description"),
		from(entityops.EntityFields.Soc2PeriodEnd).Required().Describe("The last day the report covers as YYYY-MM-DD: the end of the review period for a Type 2 report, or the as-of date for a Type 1 report"),
	).Required(),
	from(entityops.ProgramFields.AuditFirm).Required().Describe("The audit firm that performed the examination, as the report names it"),
	schema.Objects("potential_risks",
		schema.Extra("title", schema.TypeString).Required().Describe("A short heading for the risk, such as Shared Responsibility Model"),
		schema.Extra("description", schema.TypeString).Required().Describe("Two to five sentences explaining the risk for the reviewing organization, quoting the report and citing its page where it helps"),
		schema.Extra("customer_responsibilities", schema.TypeArray).Describe("What the reviewing organization keeps responsibility for, each a short noun phrase; empty when the risk carries none"),
	).Required(),
	schema.Extra("positive_observations", schema.TypeArray).Describe("Meaningful evidence of effective security controls, independent assurance, and program maturity, each one sentence"),
	schema.Extra("recommendation_outcome", schema.TypeString).Required().Enum(models.ReviewOutcomes...).Describe("The overall outcome of the review"),
	schema.Extra("recommendation", schema.TypeString).Required().Describe("One or two sentences justifying the outcome, with specific follow-up actions only when warranted"),
	schema.Objects("scoring_answers",
		schema.Extra("question_key", schema.TypeString).Required().Describe("The key of the vendor scoring question answered"),
		schema.Extra("answer", schema.TypeString).Required().Describe("The answer in the question's answer format, true or false for yes or no questions"),
		schema.Extra("likelihood", schema.TypeString).Required().Enum(scoringLikelihoods...).Describe("How likely the risk the question guards against is for this vendor"),
		schema.Extra("notes", schema.TypeString).Required().Describe("The evidence in the report behind the answer, citing its section or page"),
	),
	from(entityops.ReviewFields.Summary).Required().Describe("A short summary of the report for a reader deciding whether to rely on this vendor: what system it covers, the trust services categories in scope, and anything notable"),
	schema.Extra(RiskRatingKey, schema.TypeString).Required().Enum(enums.VendorRiskRating("").Values()...).Describe("The risk of relying on this vendor based on the auditor's opinion, the findings, and the scope of the report"),
	schema.Extra("risk_rationale", schema.TypeString).Required().Describe("The reasons for the risk rating, citing the opinion and any findings that drove it"),
)

// RiskFloor is the least severe risk rating a report with the given opinion can carry
func RiskFloor(opinion docextract.Opinion) enums.VendorRiskRating {
	switch opinion {
	case soc2.OpinionAdverse, soc2.OpinionDisclaimer:
		return enums.VendorRiskRatingHigh
	case soc2.OpinionQualified:
		return enums.VendorRiskRatingMedium
	default:
		return enums.VendorRiskRatingLow
	}
}

// ApplyRiskFloor raises a risk rating to the floor the opinion sets and never lowers it; a rating
// outside the scale is treated as below the floor
func ApplyRiskFloor(rating enums.VendorRiskRating, opinion docextract.Opinion) enums.VendorRiskRating {
	floor := RiskFloor(opinion)
	scale := enums.VendorRiskRating("").Values()

	if slices.Index(scale, rating.String()) < slices.Index(scale, floor.String()) {
		return floor
	}

	return rating
}
