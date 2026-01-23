package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"entgo.io/ent/dialect/sql"

	"entgo.io/ent"
	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrCustomEnumCreationFailed is returned when a custom enum value does not exist but is attempted to be set
	ErrCustomEnumCreationFailed = errors.New("value does not exist")
	// ErrCustomEnumInUse is returned when a custom enum is in use and cannot be deleted
	ErrCustomEnumInUse = errors.New("enum is in use")
)

var (
	// enumRegistry holds schema registration info for custom enums
	enumRegistry = &schemaEnumRegistry{
		objectTypes:  make(map[string]string),
		globalEnums:  make(map[string][]string),
	}
)

// schemaEnumRegistry tracks which schemas use custom enums
type schemaEnumRegistry struct {
	mu           sync.RWMutex
	// objectTypes maps schema name -> table name for non-global enum validation
	objectTypes  map[string]string
	// globalEnums maps field name -> table names for global enum deletion checks
	globalEnums  map[string][]string
}

// RegisterEnumSchema registers a schema as using custom enums
// schemaName is used for object_type validation, tableName for deletion checks
func RegisterEnumSchema(schemaName, tableName string) {
	enumRegistry.mu.Lock()
	defer enumRegistry.mu.Unlock()

	enumRegistry.objectTypes[schemaName] = tableName
}

// RegisterGlobalEnum registers a table as using a global enum field
func RegisterGlobalEnum(fieldName, tableName string) {
	enumRegistry.mu.Lock()
	defer enumRegistry.mu.Unlock()

	for _, t := range enumRegistry.globalEnums[fieldName] {
		if t == tableName {
			return
		}
	}

	enumRegistry.globalEnums[fieldName] = append(enumRegistry.globalEnums[fieldName], tableName)
}

// GetGlobalEnumTables returns all tables that use a global enum field
func GetGlobalEnumTables(fieldName string) []string {
	enumRegistry.mu.RLock()
	defer enumRegistry.mu.RUnlock()

	return enumRegistry.globalEnums[fieldName]
}

// GetTableForObjectType returns the table name for a given object type, or empty if not registered
func GetTableForObjectType(objectType string) string {
	enumRegistry.mu.RLock()
	defer enumRegistry.mu.RUnlock()

	return enumRegistry.objectTypes[objectType]
}

// IsValidObjectType returns true if the object type is registered for custom enums
func IsValidObjectType(objectType string) bool {
	enumRegistry.mu.RLock()
	defer enumRegistry.mu.RUnlock()

	_, ok := enumRegistry.objectTypes[objectType]
	return ok
}

// CustomEnumFilter is used to filter custom enums based on object type and field
type CustomEnumFilter struct {
	// ObjectType is the object type the enum applies to, e.g. "risk", "control", "risk_category"
	ObjectType string
	// Field is the field the enum applies to, e.g. "kind", "category"
	Field string
	// EdgeFieldName is the edge field name the enum applies to that is the foreign key, e.g. "risk_kind_id"
	EdgeFieldName string
	// SchemaFieldName is the schema field name the enum applies to, e.g. "control_kind_name
	SchemaFieldName string
	// AllowGlobal indicates the enum lookup should use global enums with an empty object type
	AllowGlobal bool
}

// HookCustomEnums ensures that a custom enum value exists for the given object type and field
// It looks up the enum by name and sets the corresponding edge field on the mutation
func HookCustomEnums(in CustomEnumFilter) ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// get the value of the enum field from the mutation
			value, ok := m.Field(in.SchemaFieldName)
			if !ok {
				return next.Mutate(ctx, m)
			}

			// if the value is empty, skip the rest of the hook
			enumValue, ok := value.(string)
			if !ok || enumValue == "" {
				return next.Mutate(ctx, m)
			}

			// get the ent client from the mutation
			mut := m.(utils.GenericMutation)
			client := mut.Client()

			// look up the enum by name, object type, and field
			// and ensure it exists
			enumPredicates := []predicate.CustomTypeEnum{
				customtypeenum.NameEqualFold(enumValue),
				customtypeenum.FieldEqualFold(in.Field),
				customtypeenum.DeletedAtIsNil(),
			}

			// lookupEnum fetches a custom enum by object type
			lookupEnum := func(objectType string) (*generated.CustomTypeEnum, error) {
				return client.CustomTypeEnum.Query().
					Where(append(enumPredicates, customtypeenum.ObjectTypeEqualFold(objectType))...).
					Only(ctx)
			}

			var enum *generated.CustomTypeEnum
			var err error

			if in.AllowGlobal {
				enum, err = lookupEnum("")
				if err != nil && generated.IsNotFound(err) {
					enum, err = lookupEnum(in.ObjectType)
				}
			} else {
				enum, err = lookupEnum(in.ObjectType)
			}
			if err != nil {
				// if the enum does not exist, return a custom error
				if generated.IsNotFound(err) {
					return nil, fmt.Errorf("%w: %s is not valid", ErrCustomEnumCreationFailed, enumValue)
				}

				return nil, err
			}

			// set the edge field on the mutation to the enum ID
			if err := m.SetField(in.EdgeFieldName, enum.ID); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}

// HookCustomTypeEnumDelete checks if the enum(s) being deleted is in use by any other object.
// If in use, the deletion cannot proceed
func HookCustomTypeEnumDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CustomTypeEnumFunc(func(ctx context.Context, m *generated.CustomTypeEnumMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			ids := getMutationIDs(ctx, m)
			if len(ids) == 0 {
				return next.Mutate(ctx, m)
			}

			client := m.Client()
			enums, err := client.CustomTypeEnum.Query().
				Where(customtypeenum.IDIn(ids...)).
				Select(
					customtypeenum.FieldID,
					customtypeenum.FieldObjectType,
					customtypeenum.FieldField,
					customtypeenum.FieldName,
				).
				All(ctx)
			if err != nil {
				return nil, err
			}

			var errs []string
			var mu sync.Mutex

			funcs := make([]func(), 0)
			for _, enum := range enums {
				funcs = append(funcs, isEnumInUse(ctx, client, enum.ID, enum.ObjectType, enum.Field, enum.Name, &errs, &mu))
			}

			if len(funcs) == 0 {
				return next.Mutate(ctx, m)
			}

			if err := client.Pool.SubmitMultipleAndWait(funcs); err != nil {
				return nil, err
			}

			if len(errs) > 0 {
				logx.FromContext(ctx).Error().
					Int("error_count", len(errs)).
					Strs("errors", errs).
					Msg("custom enum deletion failed: enums are in use")
				return nil, fmt.Errorf("%w: %d enum(s) are in use and cannot be deleted", ErrCustomEnumInUse, len(errs)) //nolint:err113
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdateOne|ent.OpUpdate)
}

// isEnumInUse returns a closure that checks whether a custom enum is referenced by any records
func isEnumInUse(ctx context.Context, client *generated.Client, enumID, objectType, enumField, name string, allErrors *[]string, mu *sync.Mutex) func() {
	ctrlCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if enumField == "" {
		enumField = "kind"
	}

	// handle global enums (empty object type)
	if objectType == "" {
		return isGlobalEnumInUse(ctrlCtx, ctx, client, enumID, enumField, name, allErrors, mu)
	}

	return isNonGlobalEnumInUse(ctrlCtx, ctx, client, enumID, objectType, enumField, name, allErrors, mu)
}

// isGlobalEnumInUse checks if a global enum is in use across all registered tables
func isGlobalEnumInUse(ctrlCtx, logCtx context.Context, client *generated.Client, enumID, enumField, name string, allErrors *[]string, mu *sync.Mutex) func() {
	tables := GetGlobalEnumTables(enumField)
	if len(tables) == 0 {
		return func() {}
	}

	columnName := fmt.Sprintf("%s_id", strcase.SnakeCase(enumField))

	return func() {
		var unionParts []string
		for _, table := range tables {
			unionParts = append(unionParts, fmt.Sprintf("SELECT count(id) as cnt FROM %s WHERE %s = $1 AND deleted_at IS NULL", table, columnName))
		}

		query := fmt.Sprintf("SELECT SUM(cnt) FROM (%s) combined", strings.Join(unionParts, " UNION ALL "))

		var rows sql.Rows
		if err := client.Driver().Query(ctrlCtx, query, lo.ToAnySlice([]string{enumID}), &rows); err != nil {
			mu.Lock()
			logx.FromContext(logCtx).Error().Err(err).Str("enum_field", enumField).Str("enum_id", enumID).Strs("tables", tables).Msg("failed to query global enum usage")
			*allErrors = append(*allErrors, fmt.Sprintf("failed to check if global %s enum %s is in use: %v", enumField, name, err))
			mu.Unlock()

			return
		}
		defer rows.Close()

		var count int
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				mu.Lock()
				logx.FromContext(logCtx).Error().Err(err).Str("enum_field", enumField).Str("enum_id", enumID).Msg("failed to scan global enum count")
				*allErrors = append(*allErrors, fmt.Sprintf("failed to check if global %s enum %s is in use: %v", enumField, name, err))
				mu.Unlock()

				return
			}
		}

		if count > 0 {
			mu.Lock()

			label := "record"
			if count != 1 {
				label = "records"
			}

			*allErrors = append(*allErrors, fmt.Sprintf("the global %s value of %s is in use by %d %s and cannot be deleted until those are updated", enumField, name, count, label))
			mu.Unlock()
		}
	}
}

// isNonGlobalEnumInUse checks if a non-global enum is in use in its specific table
func isNonGlobalEnumInUse(ctrlCtx, logCtx context.Context, client *generated.Client, enumID, objectType, enumField, name string, allErrors *[]string, mu *sync.Mutex) func() {
	edgeName := strings.ToLower(objectType)
	table := pluralize.NewClient().Plural(edgeName)
	columnName := fmt.Sprintf("%s_%s_id", strcase.SnakeCase(edgeName), strcase.SnakeCase(enumField))
	label := strings.ReplaceAll(edgeName, "_", " ")

	return func() {
		query := fmt.Sprintf("SELECT count(id) FROM %s WHERE %s = $1 AND deleted_at IS NULL", table, columnName)

		var rows sql.Rows
		if err := client.Driver().Query(ctrlCtx, query, lo.ToAnySlice([]string{enumID}), &rows); err != nil {
			mu.Lock()
			logx.FromContext(logCtx).Error().Err(err).Str("table", table).Str("field", columnName).Str("enum_id", enumID).Msg("failed to query enum edges")
			*allErrors = append(*allErrors, fmt.Sprintf("failed to check if %s enum %s is in use: %v", objectType, name, err))
			mu.Unlock()

			return
		}
		defer rows.Close()

		var count int
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				mu.Lock()
				logx.FromContext(logCtx).Error().Err(err).Str("table", table).Str("field", columnName).Str("enum_id", enumID).Msg("failed to scan enum edge count")
				*allErrors = append(*allErrors, fmt.Sprintf("failed to check if %s enum %s is in use: %v", objectType, name, err))
				mu.Unlock()

				return
			}
		}

		if count > 0 {
			mu.Lock()

			displayLabel := label
			if count != 1 {
				displayLabel = pluralize.NewClient().Plural(label)
			}

			*allErrors = append(*allErrors, fmt.Sprintf("the %s value of %s is in use by %d %s and cannot be deleted until those are updated", objectType, name, count, displayLabel))
			mu.Unlock()
		}
	}
}
