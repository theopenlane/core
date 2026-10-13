package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var ProcedureSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"procedures": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"name": {
						Type: docextract.TypeString,
					},
					"details": {
						Type: docextract.TypeString,
					},
				},
				Required: []string{
					"name",
					"details",
				},
			},
		},
	},
	Required: []string{"procedures"},
}
