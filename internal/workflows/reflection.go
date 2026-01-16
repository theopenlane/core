package workflows

import "reflect"

// StringField extracts a string field by name from a struct or pointer to struct.
func StringField(node any, field string) string {
	if node == nil {
		return ""
	}

	val := reflect.ValueOf(node)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ""
	}

	fieldVal := val.FieldByName(field)
	if !fieldVal.IsValid() || fieldVal.Kind() != reflect.String {
		return ""
	}

	return fieldVal.String()
}
