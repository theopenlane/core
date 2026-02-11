package operations

import "testing"

type schemaSample struct {
	Name string `json:"name"`
}

func TestSchemaFrom(t *testing.T) {
	schema := SchemaFrom[schemaSample]()
	if schema == nil {
		t.Fatalf("expected schema map")
	}
	props := findSchemaProperties(schema)
	if props == nil {
		t.Fatalf("expected properties map")
	}
	if _, ok := props["name"]; !ok {
		t.Fatalf("expected name property")
	}
}

func findSchemaProperties(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		if props, ok := typed["properties"].(map[string]any); ok {
			return props
		}
		for _, item := range typed {
			if props := findSchemaProperties(item); props != nil {
				return props
			}
		}
	case []any:
		for _, item := range typed {
			if props := findSchemaProperties(item); props != nil {
				return props
			}
		}
	}
	return nil
}
