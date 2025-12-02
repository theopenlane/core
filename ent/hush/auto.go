package hush

import (
	"reflect"

	"entgo.io/ent"

	hooks "github.com/theopenlane/ent/encrypt"
)

// AutoEncryptionHook automatically creates encryption hooks for all fields
// annotated with hush.EncryptField() in a schema.
//
// This function should be called automatically by the code generation system
// or by a schema that wants to enable automatic encryption detection.
//
// Usage in schema (if needed manually):
//
//	func (MySchema) Hooks() []ent.Hook {
//	    return hush.AutoEncryptionHook(MySchema{})
//	}
func AutoEncryptionHook(schema ent.Interface) []ent.Hook {
	encryptedFields := detectEncryptedFields(schema)
	if len(encryptedFields) == 0 {
		// Return a no-op hook to satisfy the generated runtime
		return []ent.Hook{
			func(next ent.Mutator) ent.Mutator {
				return next
			},
		}
	}

	// Import the hooks package functions
	// Note: This creates a dependency, but it's cleaner than duplicating code
	return []ent.Hook{
		createMultiFieldEncryptionHook(encryptedFields),
	}
}

// AutoDecryptionInterceptor automatically creates decryption interceptors
// for all fields annotated with hush.EncryptField() in a schema.
func AutoDecryptionInterceptor(schema ent.Interface) []ent.Interceptor {
	encryptedFields := detectEncryptedFields(schema)
	if len(encryptedFields) == 0 {
		// Return a no-op interceptor to satisfy the generated runtime
		return []ent.Interceptor{
			ent.InterceptFunc(func(next ent.Querier) ent.Querier {
				return next
			}),
		}
	}

	return []ent.Interceptor{
		createMultiFieldDecryptionInterceptor(encryptedFields),
	}
}

// detectEncryptedFields scans a schema and returns field names that have
// the hush.EncryptField() annotation.
func detectEncryptedFields(schema ent.Interface) []string {
	encryptedFields := []string{}

	// Handle nil schema
	if schema == nil {
		return encryptedFields
	}

	// Use reflection to get the Fields() method
	schemaValue := reflect.ValueOf(schema)
	if schemaValue.Kind() == reflect.Ptr {
		if schemaValue.IsNil() {
			return encryptedFields
		}

		schemaValue = schemaValue.Elem()
	}

	// Get the Fields method
	fieldsMethod := schemaValue.MethodByName("Fields")
	if !fieldsMethod.IsValid() {
		return encryptedFields
	}

	// Call Fields() to get field definitions
	results := fieldsMethod.Call([]reflect.Value{})
	if len(results) != 1 {
		return encryptedFields
	}

	fieldsValue := results[0]
	if fieldsValue.Kind() != reflect.Slice {
		return encryptedFields
	}

	// Iterate through fields to find encrypted ones
	for i := 0; i < fieldsValue.Len(); i++ {
		fieldValue := fieldsValue.Index(i)

		field, ok := fieldValue.Interface().(ent.Field)
		if !ok {
			continue
		}

		// Check if field has encryption annotation
		if hasEncryptionAnnotation(field) {
			fieldName := extractFieldName(field)
			if fieldName != "" {
				encryptedFields = append(encryptedFields, fieldName)
			}
		}
	}

	return encryptedFields
}

// hasEncryptionAnnotation checks if a field has the hush encryption annotation
func hasEncryptionAnnotation(field ent.Field) bool {
	// Use reflection to access the field's descriptor and annotations
	fieldValue := reflect.ValueOf(field)
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	if fieldValue.Kind() != reflect.Struct {
		return false
	}

	// Try to access the descriptor field
	descField := fieldValue.FieldByName("desc")
	if !descField.IsValid() || descField.IsNil() {
		return false
	}

	descValue := descField.Elem()
	if descValue.Kind() != reflect.Struct {
		return false
	}

	// Try to access annotations
	annotationsField := descValue.FieldByName("Annotations")
	if !annotationsField.IsValid() || annotationsField.Kind() != reflect.Map {
		return false
	}

	// Check if our encryption annotation exists
	for _, key := range annotationsField.MapKeys() {
		keyStr := key.String()
		if keyStr == "HushEncryption" || keyStr == "*hush.EncryptionAnnotation" {
			return true
		}

		// Also check if the map value is our annotation type
		value := annotationsField.MapIndex(key)
		if value.IsValid() {
			valueType := value.Type()
			if valueType.String() == "hush.EncryptionAnnotation" ||
				valueType.String() == "*hush.EncryptionAnnotation" {
				return true
			}
		}
	}

	return false
}

// extractFieldName extracts the field name from an ent.Field
func extractFieldName(field ent.Field) string {
	// Use reflection to access the field name
	fieldValue := reflect.ValueOf(field)
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	if fieldValue.Kind() != reflect.Struct {
		return ""
	}

	// Try to access descriptor field first
	descField := fieldValue.FieldByName("desc")
	if descField.IsValid() && !descField.IsNil() {
		descValue := descField.Elem()
		if descValue.Kind() == reflect.Struct {
			nameField := descValue.FieldByName("Name")
			if nameField.IsValid() && nameField.Kind() == reflect.String {
				return nameField.String()
			}
		}
	}

	// Fallback: try to access name field directly
	nameField := fieldValue.FieldByName("name")
	if nameField.IsValid() && nameField.Kind() == reflect.String {
		return nameField.String()
	}

	return ""
}

// createMultiFieldEncryptionHook creates a hook that encrypts multiple fields
func createMultiFieldEncryptionHook(fieldNames []string) ent.Hook {
	// Import the hooks package to use the actual encryption functions
	return hooks.HookEncryption(fieldNames...)
}

// createMultiFieldDecryptionInterceptor creates an interceptor that decrypts multiple fields
func createMultiFieldDecryptionInterceptor(fieldNames []string) ent.Interceptor {
	// Import the interceptors package to use the actual decryption functions
	return interceptorEncryption(fieldNames...)
}
