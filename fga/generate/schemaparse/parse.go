package schemaparse

import (
	"fmt"
	"slices"
	"sort"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/load"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/v2/fga/generate/schemainfo"
)

// parsedSchema adds the annotation and policy parsing to the schema snapshot
type parsedSchema struct {
	schemainfo.SchemaInfo
}

// GetSchemas returns the settings of the loaded ent schemas based on annotations and policies
func GetSchemas(loaded []*load.Schema) []schemainfo.SchemaInfo {
	schemas := []schemainfo.SchemaInfo{}

	allSchemas := sortSchemas(loaded)

	fmt.Printf("found %d schemas\n", len(allSchemas))

	for _, schema := range allSchemas {
		si, ok := checkSchemaAnnotations(schema)
		if !ok {
			continue
		}

		si.OnlyAllowCreate = si.HasCreate && !si.HasUpdate && !si.HasDelete

		if body := parsePolicyBody(schema); body != nil {
			si.functionContainsCheckCreateAccess(body)
		}

		schemas = append(schemas, si.SchemaInfo)
	}

	return schemas
}

// sortSchemas returns a copy of schemas sorted alphabetically by name, leaving the graph order untouched
func sortSchemas(schemas []*load.Schema) []*load.Schema {
	sorted := slices.Clone(schemas)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// checkSchemaAnnotations checks annotations on schemas, eliminating not-needed schemas before policy checks
func checkSchemaAnnotations(schema *load.Schema) (parsedSchema, bool) {
	si := parsedSchema{
		Name: schema.Name}

	si.checkEntGQLAnnotation(schema)

	// run this after anything that touches CRUD settings so it takes precedence over other parsing
	si.checkSkipAnnotation(schema)

	// additional settings
	si.checkGroupPermissions(schema)
	si.getParentSchemas(schema)

	si.OnlyAllowCreate = si.HasCreate && !si.HasUpdate

	return si, true
}

// checkEntGQLAnnotation checks for the Skip annotation on the schema
// SkipAll implies that there shouldn't be any additional crud settings
// It uses the MutationInputs settings to see if the schema allows create and update
func (si *parsedSchema) checkEntGQLAnnotation(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	entgqlAnt, ok := entx.GetAnnotation[*entgql.Annotation](schema)
	if !ok {
		return
	}

	// skip means there is no external crud operations
	// so this does't need to be included in crud scopes
	// Is matches any flag, so check the type skip itself rather than SkipAll, partial skips like mutation inputs still have crud
	if entgqlAnt.Skip.Is(entgql.SkipType) {
		si.ExcludeFromGeneration = true
		return
	}

	for _, mu := range entgqlAnt.MutationInputs {
		if mu.IsCreate {
			si.HasCreate = true
		} else {
			si.HasUpdate = true
		}
	}
}

// checkSkipAnnotation checks the optional FGA Crud annotation that is used as precedence
// over any setting that is determined based on annotations + policies
func (si *parsedSchema) checkSkipAnnotation(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	entFGACrudAnt, ok := entx.GetAnnotation[*entx.FGACrudAnnotation](schema)
	if !ok {
		return
	}

	if entFGACrudAnt.Skip.Has(entx.SkipAll) {
		si.ExcludeFromGeneration = true
	} else {
		if entFGACrudAnt.Skip.Has(entx.SkipCreate) {
			si.HasCreate = false
		}
		if entFGACrudAnt.Skip.Has(entx.SkipUpdate) {
			si.HasUpdate = false
		}
		if entFGACrudAnt.Skip.Has(entx.SkipDelete) {
			si.HasDelete = false
		}

		si.ViewOnly = entFGACrudAnt.Skip.Has(entx.SkipCreate | entx.SkipUpdate | entx.SkipDelete)
	}
}

// checkGroupPermissions looks for the group permissions annotation
// that will enable setting the `_creator` settings
func (si *parsedSchema) checkGroupPermissions(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	if ok := entx.HasAnnotation[entx.GroupPermissionsEnabled](schema); ok {
		si.HasGroupCreator = true
	}
}

// getParentSchemas looks for the parent fga annotation to set
// inheritance of crud in the model
func (si *parsedSchema) getParentSchemas(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	ant, ok := entx.GetAnnotation[*entx.FGAParentCrudAnnotation](schema)
	if !ok {
		return
	}

	si.Parents = ant.ParentSchemas
}
