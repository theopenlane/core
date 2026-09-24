package schema

import "google.golang.org/genai"

var GroupSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"groups": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type: genai.TypeString,
					},
					"displayName": {
						Type: genai.TypeString,
					},
					"description": {
						Type: genai.TypeString,
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
