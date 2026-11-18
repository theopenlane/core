// Package schemainfo holds the schema snapshot read by the fga generator, it must not import ent
// so the generator does not compile the schema package and everything it depends on
package schemainfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const snapshotFilePermissions = 0o644

// ErrNoSchemas is returned when the snapshot has no schemas
var ErrNoSchemas = errors.New("schema snapshot has no schemas")

// SchemaInfo holds information about an ent schema that is relevant for generating FGA permissions, such as the name of the schema and whether it has create access rules defined in its Policy function.
type SchemaInfo struct {
	// Name is the name of the schema
	Name string `json:"name"`
	// ExcludeFromGeneration excludes this type from any crud generation, only allowing default access
	// and no api token access or additional permissions for users
	ExcludeFromGeneration bool `json:"excludeFromGeneration,omitempty"`
	// CanCreateServiceOnly is used for objects that only service (api tokens) are allowed to create and not users/groups
	CanCreateServiceOnly bool `json:"canCreateServiceOnly,omitempty"`
	// OnlyAllowCreate checks settings of create and update and delete and determines if only creation is allowed
	OnlyAllowCreate bool `json:"onlyAllowCreate,omitempty"`
	// HasCreate indicates the schema should have crud create operations
	HasCreate bool `json:"hasCreate,omitempty"`
	// HasUpdate indicates the schema should have crud update operations
	HasUpdate bool `json:"hasUpdate,omitempty"`
	// HasDelete indicates the schema should have crud delete operations
	HasDelete bool `json:"hasDelete,omitempty"`
	// ViewOnly limits the schema to a view relation, with no create, edit, or delete relations or scopes
	ViewOnly bool `json:"viewOnly,omitempty"`
	// HasGroupCreator indicates the schema should have group creator permissions
	HasGroupCreator bool `json:"hasGroupCreator,omitempty"`
	// Parents are the parent types the schema can also inherit crud access from
	Parents []string `json:"parents,omitempty"`
}

// Write stores the schema snapshot at path
func Write(path string, schemas []SchemaInfo) error {
	contents, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding schema snapshot: %w", err)
	}

	return os.WriteFile(path, append(contents, '\n'), snapshotFilePermissions)
}

// Read loads the schema snapshot from path
func Read(path string) ([]SchemaInfo, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading schema snapshot: %w", err)
	}

	var schemas []SchemaInfo
	if err := json.Unmarshal(contents, &schemas); err != nil {
		return nil, fmt.Errorf("decoding schema snapshot: %w", err)
	}

	if len(schemas) == 0 {
		return nil, ErrNoSchemas
	}

	return schemas, nil
}
