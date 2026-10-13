package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

type Review struct {
	Title           string   `json:"title,omitempty"`
	Summary         string   `json:"summary,omitempty"`
	Details         string   `json:"details,omitempty"`
	Source          string   `json:"source,omitempty"`
	ReportedAt      string   `json:"reportedAt,omitempty"`
	Reporter        string   `json:"reporter,omitempty"`
	ApprovedAt      string   `json:"approvedAt,omitempty"`
	Approved        bool     `json:"approved,omitempty"`
	RefCodes        []string `json:"refCodes,omitempty"`
	ExternalID      string   `json:"externalID,omitempty"`
	EnvironmentName string   `json:"environmentName,omitempty"`
	ScopeName       string   `json:"scopeName,omitempty"`
	State           string   `json:"state,omitempty"`
}

type Reviews struct {
	Reviews []Review `json:"reviews"`
}

var ReviewSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"reviews": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"title": {
						Type: docextract.TypeString,
					},
					"summary": {
						Type: docextract.TypeString,
					},
					"details": {
						Type: docextract.TypeString,
					},
					"source": {
						Type: docextract.TypeString,
					},
					"reportedAt": {
						Type: docextract.TypeString,
					},
					"reviewedAt": {
						Type: docextract.TypeString,
					},
					"reporter": {
						Type: docextract.TypeString,
					},
					"approvedAt": {
						Type: docextract.TypeString,
					},
					"approved": {
						Type: docextract.TypeBoolean,
					},
					"refCodes": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
						},
					},
					"externalID": {
						Type: docextract.TypeString,
					},
					"environmentName": {
						Type: docextract.TypeString,
						Enum: []string{"production"},
					},
					"scopeName": {
						Type: docextract.TypeString,
						Enum: []string{"in-scope"},
					},
					"state": {
						Type: docextract.TypeString,
						Enum: []string{"COMPLETED"},
					},
					"category": {
						Type: docextract.TypeString,
						Enum: []string{"audit-review"},
					},
				},
				Required: []string{
					"title",
					"summary",
					"details",
					"approved",
					"environmentName",
					"scopeName",
					"state",
					"category",
				},
			},
		},
	},
	Required: []string{"reviews"},
}
