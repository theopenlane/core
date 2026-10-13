package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

// RiskRatings are the assessment risk ratings ordered from least to most severe
var RiskRatings = []string{"low", "medium", "high", "critical"}

var AssessmentSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"assessment": {
			Type: docextract.TypeObject,
			Properties: map[string]*docextract.Schema{
				"summary": {
					Type:        docextract.TypeString,
					Description: "A short summary of the report for a reader deciding whether to rely on this vendor: what system it covers, the trust services categories in scope, and anything notable",
				},
				"riskRating": {
					Type:        docextract.TypeString,
					Enum:        RiskRatings,
					Description: "The risk of relying on this vendor based on the auditor's opinion, the findings, and the scope of the report",
				},
				"riskRationale": {
					Type:        docextract.TypeString,
					Description: "The reasons for the risk rating, citing the opinion and any findings that drove it",
				},
				"periodEnd": {
					Type:        docextract.TypeString,
					Description: "The last day the report covers as YYYY-MM-DD: the end of the review period for a Type 2 report, or the as-of date for a Type 1 report",
				},
				"vendorName": {
					Type:        docextract.TypeString,
					Description: "The name customers know the company by, the product or brand name when the report names one, not the legal holding company",
				},
				"vendorDomains": {
					Type:        docextract.TypeArray,
					Items:       &docextract.Schema{Type: docextract.TypeString},
					Description: "The company's web domains mentioned in the report, lowercase without a scheme or path, e.g. example.com",
				},
				"vendorServices": {
					Type:        docextract.TypeArray,
					Items:       &docextract.Schema{Type: docextract.TypeString},
					Description: "The services the company provides to customers, from the report's system description, each a short name such as Container orchestration or Managed DNS",
				},
				"vendorDescription": {
					Type:        docextract.TypeString,
					Description: "One or two sentences on what the company provides, from the report's system description",
				},
			},
			Required: []string{
				"summary",
				"riskRating",
				"riskRationale",
				"periodEnd",
				"vendorName",
			},
		},
	},
	Required: []string{"assessment"},
}
