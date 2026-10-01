package schema

import "github.com/theopenlane/core/v2/pkg/docextract"

var FullImportSchema = &docextract.Schema{
	Type: docextract.TypeObject,
	Properties: map[string]*docextract.Schema{
		"entities":      EntitySchema.Properties["entities"],
		"assets":        AssetSchema.Properties["assets"],
		"groups":        GroupSchema.Properties["groups"],
		"domains":       DomainSchema.Properties["domains"],
		"controls":      ControlSchema.Properties["controls"],
		"reviews":       ReviewSchema.Properties["reviews"],
		"findings":      FindingSchema.Properties["findings"],
		"procedures":    ProcedureSchema.Properties["procedures"],
		"platforms":     PlatformSchema.Properties["platforms"],
		"systemdetails": SystemDetailsSchema.Properties["systemdetails"],
		"contacts":      ContactSchema.Properties["contacts"],
		"programs":      ProgramSchema.Properties["programs"],
	},
}
