package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var FindingSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"findings": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"open": {
						Type: docextract.TypeBoolean,
					},
					"description": {
						Type: docextract.TypeString,
					},
					"severity": {
						Type: docextract.TypeString,
						Enum: []string{"high", "medium", "low"},
					},
					"source": {
						Type: docextract.TypeString,
					},
					"reported_at": {
						Type: docextract.TypeString,
					},
					"reporter": {
						Type: docextract.TypeString,
					},
					"refCodes": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
						},
					},
					"testDetails": {
						Type:        docextract.TypeString,
						Description: "The verbatim text of the test procedure printed in the same table row as this exception, copied exactly and not summarized",
					},
					"environmentName": {
						Type: docextract.TypeString,
						Enum: []string{"production"},
					},
				},
				Required: []string{
					"open",
					"description",
					"severity",
					"refCodes",
					"environmentName",
				},
			},
		},
	},
	Required: []string{"findings"},
}
