package schema

import "google.golang.org/genai"

var FullImportSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
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
