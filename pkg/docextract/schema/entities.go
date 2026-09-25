package schema

import (
	"github.com/samber/lo"
	"google.golang.org/genai"
)

var EntitySchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"entities": {
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
					"status": {
						Type: genai.TypeString,
						Enum: []string{"ACTIVE"},
					},
					"approvedForUse": {
						Type: genai.TypeBoolean,
					},
					"hasSOC2": {
						Type:     genai.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"domains": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"ssoEnforced": {
						Type:     genai.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"mfaSupport": {
						Type:     genai.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"mfaEnforced": {
						Type:     genai.TypeBoolean,
						Nullable: lo.ToPtr(true),
					},
					"statusPageURL": {
						Type:     genai.TypeString,
						Nullable: lo.ToPtr(true),
					},
					"providedServices": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
					"entityTypeName": {
						Type: genai.TypeString,
						Enum: []string{"vendor"},
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
