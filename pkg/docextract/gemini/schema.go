package gemini

import (
	"google.golang.org/genai"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

// schemaTypes maps the neutral schema types onto the genai ones
var schemaTypes = map[docextract.SchemaType]genai.Type{
	docextract.TypeString:  genai.TypeString,
	docextract.TypeBoolean: genai.TypeBoolean,
	docextract.TypeObject:  genai.TypeObject,
	docextract.TypeArray:   genai.TypeArray,
}

// responseSchema converts a neutral response schema into the genai shape the sdk requires
func responseSchema(schema *docextract.Schema) *genai.Schema {
	if schema == nil {
		return nil
	}

	out := &genai.Schema{
		Type:        schemaTypes[schema.Type],
		Description: schema.Description,
		Required:    schema.Required,
		Enum:        schema.Enum,
		Nullable:    schema.Nullable,
		Items:       responseSchema(schema.Items),
	}

	if len(schema.Properties) == 0 {
		return out
	}

	out.Properties = make(map[string]*genai.Schema, len(schema.Properties))

	for name, property := range schema.Properties {
		out.Properties[name] = responseSchema(property)
	}

	return out
}
