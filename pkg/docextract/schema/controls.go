package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var ControlSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"controls": {
			Type: docextract.TypeArray,
			Items: &docextract.Schema{
				Type: docextract.TypeObject,
				Properties: map[string]*docextract.Schema{
					"refCode": {
						Type: docextract.TypeString,
					},
					"title": {
						Type: docextract.TypeString,
					},
					"description": {
						Type: docextract.TypeString,
					},
					"auditorReferenceID": {
						Type: docextract.TypeString,
					},
					"status": {
						Type: docextract.TypeString,
						Enum: []string{"APPROVED"},
					},
					"source": {
						Type: docextract.TypeString,
						Enum: []string{"IMPORTED"},
					},
					"category": {
						Type: docextract.TypeString,
					},
					"subcategory": {
						Type: docextract.TypeString,
					},
					"soc2Mapping": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
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
