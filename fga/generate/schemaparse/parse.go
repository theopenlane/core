package schemaparse

import (
	"fmt"
	"sort"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/theopenlane/entx"
)

// SchemaInfo holds information about an ent schema that is relevant for generating FGA permissions, such as the name of the schema and whether it has create access rules defined in its Policy function.
type SchemaInfo struct {
	// Name is the name of the schema
	Name string
	// ExcludeFromGeneration excludes this type from any crud generation, only allowing default access
	// and no api token access or additional permissions for users
	ExcludeFromGeneration bool
	// CanCreateServiceOnly is used for objects that only service (api tokens) are allowed to create and not users/groups
	CanCreateServiceOnly bool
	// OnlyAllowCreate checks settings of create and update and delete and determines if only creation is allowed
	OnlyAllowCreate bool
	// HasCreate indicates the schema should have crud create operations
	HasCreate bool
	// HasUpdate indicates the schema should have crud update operations
	HasUpdate bool
	// HasDelete indicates the schema should have crud delete operations
	HasDelete bool
	// HasGroupCreator indicates the schema should have group creator permissions
	HasGroupCreator bool
	// Parents are the parent types the schema can also inherit crud access from
	Parents []string
}

// GetSchemas gets all ent schemas and returns settings based on annotations and policies
func GetSchemas(schemaDir string) ([]SchemaInfo, error) {
	schemas := []SchemaInfo{}

	graph, err := entc.LoadGraph(schemaDir, &gen.Config{})
	if err != nil {
		return nil, err
	}

	allSchemas := sortSchemas(graph.Schemas)

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

		schemas = append(schemas, si)
	}

	return schemas, nil
}

// sortSchemas sorts schemas alphabetically by name
func sortSchemas(schemas []*load.Schema) []*load.Schema {
	sort.SliceStable(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})

	return schemas
}

// checkSchemaAnnotations checks annotations on schemas, eliminating not-needed schemas before policy checks
func checkSchemaAnnotations(schema *load.Schema) (SchemaInfo, bool) {
	si := SchemaInfo{
		Name: schema.Name,
	}

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
func (si *SchemaInfo) checkEntGQLAnnotation(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	entgqlAnt := entgql.Annotation{}

	if raw, ok := schema.Annotations[entgqlAnt.Name()]; ok {
		if err := entgqlAnt.Decode(raw); err == nil {
			// skip means there is no external crud operations
			// so this does't need to be included in crud scopes
			if entgqlAnt.Skip.Is(entgql.SkipAll) {
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
	}
}

// checkSkipAnnotation checks the optional FGA Crud annotation that is used as precedence
// over any setting that is determined based on annotations + policies
func (si *SchemaInfo) checkSkipAnnotation(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	entFGACrudAnt := entx.FGACrudAnnotation{}

	if raw, ok := schema.Annotations[entFGACrudAnt.Name()]; ok {
		if err := entFGACrudAnt.Decode(raw); err == nil {
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
			}
		}
	}
}

// checkGroupPermissions looks for the group permissions annotation
// that will enable setting the `_creator` settings
func (si *SchemaInfo) checkGroupPermissions(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	entFGACrudAnt := entx.GroupPermissionsEnabled{}

	if _, ok := schema.Annotations[entFGACrudAnt.Name()]; ok {
		si.HasGroupCreator = true
	}
}

// getParentSchemas looks for the parent fga annotation to set
// inheritance of crud in the model
func (si *SchemaInfo) getParentSchemas(schema *load.Schema) {
	if si.ExcludeFromGeneration {
		return
	}

	entFGAParentCrudAnt := entx.FGAParentCrudAnnotation{}

	if raw, ok := schema.Annotations[entFGAParentCrudAnt.Name()]; ok {
		if err := entFGAParentCrudAnt.Decode(raw); err == nil {
			si.Parents = entFGAParentCrudAnt.ParentSchemas
		}
	}
}
