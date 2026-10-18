# custom

Hand-written extraction schemas. These are the exception: almost every section a document kind extracts should be built from entityops field descriptors instead.

## The default: build from entityops

A section that becomes records, such as controls, vendors, assets, or findings, is declared next to its kind (for SOC 2, `internal/docextract/soc2/sections.go`) with the builder in `internal/docextract/schema`:

```go
var ControlSchema = schema.Section(ControlsPart,
	schema.From(entityops.ControlFields.RefCode).Required(),
	schema.From(entityops.ControlFields.Title).Required(),
	schema.Extra(MappingKey, docextract.TypeArray),
)
```

`schema.From` keys and types each property from the entity field, so the extracted keys match what ingest expects and stay in step with the schema when a field changes. A few values the entity has no field for, such as the keys an item links by, go on the same section with `schema.Extra`.

## When a schema belongs here

Add a schema here only when it is not a set of records, so there are no entity fields to build it from. For example:

- A judgment or summary of the whole document, such as `AssessmentSchema`: a vendor risk read with its rating, risks, recommendation, and scoring answers.
- A section returned as a single object rather than a list (`schema.Object`).
- Values that only feed code, never a stored entity, such as answers that are translated into other records.

If most of a section's fields are entity fields, keep it with its kind and use `schema.Extra` for the rest rather than writing it here.

## Writing one

- Use the builder: `schema.Object` or `schema.Section`, with `schema.Extra`, `schema.Objects`, and `schema.Nested` for each property. Do not write `docextract.Schema` literals by hand.
- Build any property an entity has a field for with `schema.From`, even here. Group fields that describe one record under `schema.Nested`, as `AssessmentSchema` does for its `vendor`, so their keys match that entity's input and never collide with the other entities' fields.
- Use snake_case keys, matching every other section.
- Give each property a `Describe` that says exactly what the model should return. The section's prompt in config can add guidance, but the schema is what the model must follow.
- Export a constant for any key code reads back out of the extracted map, and any shared values such as enum options, so readers do not repeat the strings.
- Reference the schema from the kind's section map (for SOC 2, `sectionSchemas` in `internal/docextract/soc2/parts.go`) and add the section's prompt to the kind's prompt config.
