package schema

import (
	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

var AssetSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"assets": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"name": {
						Type: docextract.TypeString,
					},
					"displayName": {
						Type: docextract.TypeString,
					},
					"description": {
						Type: docextract.TypeString,
					},
					"hasPII": {
						Type: docextract.TypeBoolean,
					},
					"assetType": {
						Type: docextract.TypeString,
						Enum: []string{"TECHNOLOGY", "DOMAIN", "DEVICE", "TELEPHONE"},
					},
					"physicalLocation": {
						Type:     docextract.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"region": {
						Type:     docextract.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"source": {
						Type: docextract.TypeString,
						Enum: []string{"IMPORTED"},
					},
					"categories": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
						},
					},
					"entityName": {
						Type:     docextract.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"environmentName": {
						Type: docextract.TypeString,
						Enum: []string{"production"},
					},
					"scopeName": {
						Type: docextract.TypeString,
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
