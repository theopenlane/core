package schema

import (
	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

var EntitySchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"entities": {
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
					"status": {
						Type: docextract.TypeString,
						Enum: []string{"ACTIVE"},
					},
					"approvedForUse": {
						Type: docextract.TypeBoolean,
					},
					"hasSOC2": {
						Type:     docextract.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"domains": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
						},
					},
					"ssoEnforced": {
						Type:     docextract.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"mfaSupport": {
						Type:     docextract.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"mfaEnforced": {
						Type:     docextract.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"statusPageURL": {
						Type:     docextract.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"providedServices": {
						Type: docextract.TypeArray,
						Items: &docextract.Schema{
							Type: docextract.TypeString,
						},
					},
					"entityTypeName": {
						Type: docextract.TypeString,
						Enum: []string{"vendor"},
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
					"status",
					"approvedForUse",
					"entityTypeName",
					"environmentName",
					"scopeName",
				},
			},
		},
	},
	Required: []string{"entities"},
}
