package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var ProgramSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"programs": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"auditFirm": {
						Type: docextract.TypeString,
					},
					"auditor": {
						Type: docextract.TypeString,
					},
					"auditorEmail": {
						Type: docextract.TypeString,
					},
					"description": {
						Type: docextract.TypeString,
					},
					"frameworkName": {
						Type: docextract.TypeString,
						Enum: []string{"SOC 2"},
					},
					"name": {
						Type: docextract.TypeString,
					},
					"programType": {
						Type: docextract.TypeString,
						Enum: []string{"FRAMEWORK"},
					},
					"status": {
						Type: docextract.TypeString,
						Enum: []string{"NOT_STARTED"},
					},
					"tags": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
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
