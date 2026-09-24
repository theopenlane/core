package schema

import "google.golang.org/genai"

var ProcedureSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"procedures": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type: genai.TypeString,
					},
					"details": {
						Type: genai.TypeString,
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
