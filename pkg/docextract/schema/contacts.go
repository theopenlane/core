package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var ContactSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"contacts": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"company": {
						Type: docextract.TypeString,
					},
					"email": {
						Type: docextract.TypeString,
					},
					"entityNames": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
						},
					},
					"fullName": {
						Type: docextract.TypeString,
					},
					"title": {
						Type: docextract.TypeString,
					},
				},
				Required: []string{
					"company",
					"email",
					"fullName",
					"title",
				},
			},
		},
	},
	Required: []string{"contacts"},
}
