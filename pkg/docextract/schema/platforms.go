package schema

import "google.golang.org/genai"

var PlatformSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"platforms": {
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
					"businessPurpose": {
						Type: genai.TypeString,
					},
					"physicalLocation": {
						Type: genai.TypeString,
					},
					"region": {
						Type: genai.TypeString,
					},
					"containsPII": {
						Type: genai.TypeBoolean,
					},
					"source": {
						Type: genai.TypeString,
						Enum: []string{"IMPORTED"},
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
					"businessPurpose",
					"containsPII",
					"source",
					"environmentName",
					"scopeName",
				},
			},
		},
	},
	Required: []string{"platforms"},
}
