package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var SystemDetailsSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"systemdetails": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"name": {
						Type: docextract.TypeString,
					},
					"description": {
						Type: docextract.TypeString,
					},
					"platformName": {
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
				},
				Required: []string{
					"name",
					"description",
					"platformName",
					"environmentName",
					"scopeName",
				},
			},
		},
	},
	Required: []string{"systemdetails"},
}
