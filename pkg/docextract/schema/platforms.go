package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var PlatformSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"platforms": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"name": {
						Type: docextract.TypeString,
					},
					"description": {
						Type: docextract.TypeString,
					},
					"businessPurpose": {
						Type: docextract.TypeString,
					},
					"physicalLocation": {
						Type: docextract.TypeString,
					},
					"region": {
						Type: docextract.TypeString,
					},
					"containsPII": {
						Type: docextract.TypeBoolean,
					},
					"source": {
						Type: docextract.TypeString,
						Enum: []string{"IMPORTED"},
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
