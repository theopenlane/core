package schema

import "google.golang.org/genai"

var ContactSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"contacts": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"company": {
						Type: genai.TypeString,
					},
					"email": {
						Type: genai.TypeString,
					},
					"entityNames": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"fullName": {
						Type: genai.TypeString,
					},
					"title": {
						Type: genai.TypeString,
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
