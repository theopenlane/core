package schema

import "google.golang.org/genai"

var ProgramSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"programs": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"auditFirm": {
						Type: genai.TypeString,
					},
					"auditor": {
						Type: genai.TypeString,
					},
					"auditorEmail": {
						Type: genai.TypeString,
					},
					"description": {
						Type: genai.TypeString,
					},
					"frameworkName": {
						Type: genai.TypeString,
						Enum: []string{"SOC 2"},
					},
					"name": {
						Type: genai.TypeString,
					},
					"programType": {
						Type: genai.TypeString,
						Enum: []string{"FRAMEWORK"},
					},
					"status": {
						Type: genai.TypeString,
						Enum: []string{"NOT_STARTED"},
					},
					"tags": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{
					"description",
					"frameworkName",
					"name",
					"programType",
					"status",
				},
			},
		},
	},
	Required: []string{"programs"},
}
