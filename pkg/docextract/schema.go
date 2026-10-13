package docextract

// SchemaType is the value type of one response schema node, using declared types for use in the jsonschema
type SchemaType string

const (
	// TypeString is a string value
	TypeString SchemaType = "string"
	// TypeBoolean is a boolean value
	TypeBoolean SchemaType = "boolean"
	// TypeObject is an object with named properties
	TypeObject SchemaType = "object"
	// TypeArray is a list of items
	TypeArray SchemaType = "array"
)

// Schema constrains the shape a model's response must follow, described in provider neutral terms
// so a Kind can declare its sections without depending on any one model sdk
type Schema struct {
	// Type is the value type of this node
	Type SchemaType
	// Description explains the field to the model, omitted when empty
	Description string
	// Properties are the named fields of an object node
	Properties map[string]*Schema
	// Required names the properties the model must supply
	Required []string
	// Items is the element schema of an array node
	Items *Schema
	// Enum restricts a string node to a fixed set of values
	Enum []string
	// Nullable allows the model to return null for the field
	Nullable *bool
}
