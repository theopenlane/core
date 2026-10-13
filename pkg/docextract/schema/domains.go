package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var DomainSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"domains": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"name": {
						Type: docextract.TypeString,
					},
				},
				Required: []string{
					"name",
				},
			},
		},
	},
	Required: []string{"domains"},
}
