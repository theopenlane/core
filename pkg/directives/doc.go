// Package directives contains the implementations for the custom GraphQL directives used in the schema
// as well as an entc extension to modify the generated schema to add the directives to the appropriate fields.
//
// The following directives are implemented:
//
//	@hidden(if: Boolean) - Hides a field or type from the GraphQL schema and from non-system admin users.
//	@readOnly - Marks a field as read-only, preventing it from being set in create or update mutations by non-system admin users.
//
// The directives are added to the GraphQL schema using the entgql package and are applied to fields via annotations in the ent schema.
//
// Additionally, the extension modifies input object types to ensure that fields marked with @hidden are also marked as @readOnly
// in input types, preventing them from being set in mutations.
//
// Usage:
//
//	// In your ent schema field definitions, use the annotations to add directives
//	field.String("example").
//		Annotations(
//			directives.HiddenDirectiveAnnotation, // Adds @hidden(if: true) to the query field and @readOnly to the input fields
//		)
//
// To add additional directives, extend the Extension struct and modify the hook function accordingly and add the directive
// to the graphql schema `internal/graphapi/schema/directives.graphql` file
package directives
