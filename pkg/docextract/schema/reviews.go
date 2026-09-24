package schema

import "google.golang.org/genai"

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

var ReviewSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"reviews": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type: genai.TypeString,
					},
					"summary": {
						Type: genai.TypeString,
					},
					"details": {
						Type: genai.TypeString,
					},
					"source": {
						Type: genai.TypeString,
					},
					"reportedAt": {
						Type: genai.TypeString,
					},
					"reviewedAt": {
						Type: genai.TypeString,
					},
					"reporter": {
						Type: genai.TypeString,
					},
					"approvedAt": {
						Type: genai.TypeString,
					},
					"approved": {
						Type: genai.TypeBoolean,
					},
					"refCodes": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"externalID": {
						Type: genai.TypeString,
					},
					"environmentName": {
						Type: genai.TypeString,
						Enum: []string{"production"},
					},
					"scopeName": {
						Type: genai.TypeString,
						Enum: []string{"in-scope"},
					},
					"state": {
						Type: genai.TypeString,
						Enum: []string{"COMPLETED"},
					},
					"category": {
						Type: genai.TypeString,
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
