package schema

import "google.golang.org/genai"

var ControlSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"controls": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"refCode": {
						Type: genai.TypeString,
					},
					"title": {
						Type: genai.TypeString,
					},
					"description": {
						Type: genai.TypeString,
					},
					"auditorReferenceID": {
						Type: genai.TypeString,
					},
					"status": {
						Type: genai.TypeString,
						Enum: []string{"APPROVED"},
					},
					"source": {
						Type: genai.TypeString,
						Enum: []string{"IMPORTED"},
					},
					"category": {
						Type: genai.TypeString,
					},
					"subcategory": {
						Type: genai.TypeString,
					},
					"soc2Mapping": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{
					"refCode",
					"title",
					"description",
					"status",
					"source",
					"category",
					"subcategory",
					"soc2Mapping",
				},
			},
		},
	},
	Required: []string{"controls"},
}
