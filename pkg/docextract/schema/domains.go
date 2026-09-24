package schema

import "google.golang.org/genai"

var DomainSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"domains": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type: genai.TypeString,
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
