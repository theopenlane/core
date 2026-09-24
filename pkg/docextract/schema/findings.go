package schema

import "google.golang.org/genai"

var FindingSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"findings": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"open": {
						Type: genai.TypeBoolean,
					},
					"description": {
						Type: genai.TypeString,
					},
					"severity": {
						Type: genai.TypeString,
						Enum: []string{"high", "medium", "low"},
					},
					"source": {
						Type: genai.TypeString,
					},
					"reported_at": {
						Type: genai.TypeString,
					},
					"reporter": {
						Type: genai.TypeString,
					},
					"refCodes": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"environmentName": {
						Type: genai.TypeString,
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
