package schema

import "google.golang.org/genai"

var SystemDetailsSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"systemdetails": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type: genai.TypeString,
					},
					"description": {
						Type: genai.TypeString,
					},
					"platformName": {
						Type: genai.TypeString,
					},
					"environmentName": {
						Type: genai.TypeString,
						Enum: []string{"production"},
					},
					"scopeName": {
						Type: genai.TypeString,
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
