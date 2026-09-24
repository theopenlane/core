package schema

import (
	"github.com/samber/lo"
	"google.golang.org/genai"
)

var AssetSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"assets": {
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
					"hasPII": {
						Type: genai.TypeBoolean,
					},
					"assetType": {
						Type: genai.TypeString,
						Enum: []string{"TECHNOLOGY", "DOMAIN", "DEVICE", "TELEPHONE"},
					},
					"physicalLocation": {
						Type:     genai.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"region": {
						Type:     genai.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"source": {
						Type: genai.TypeString,
						Enum: []string{"IMPORTED"},
					},
					"categories": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"entityName": {
						Type:     genai.TypeString,
						Nullable: lo.ToPtr(true),
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
					"displayName",
					"description",
					"hasPII",
					"assetType",
					"source",
					"environmentName",
					"scopeName",
				},
			},
		},
	},
	Required: []string{"assets"},
}
