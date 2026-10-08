package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var GroupSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"groups": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"name": {
						Type: docextract.TypeString,
					},
					"displayName": {
						Type: docextract.TypeString,
					},
					"description": {
						Type: docextract.TypeString,
					},
				},
				Required: []string{
					"name",
					"displayName",
					"description",
				},
			},
		},
	},
	Required: []string{"groups"},
}
