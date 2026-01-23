package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"

	"entgo.io/ent"
	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/migrate"
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
	// ErrInvalidGlobalEnumField is returned when creating a global enum with an invalid field
	ErrInvalidGlobalEnumField = errors.New("invalid global enum field")
)

// tableInfo holds table metadata for enum checks
type tableInfo struct {
	name          string
	hasSoftDelete bool
}

// tableHasColumn checks if a table has a column with the given name
func tableHasColumn(table *schema.Table, columnName string) bool {
	for _, col := range table.Columns {
		if col.Name == columnName {
			return true
		}
	}

	return false
}

// tableHasSoftDelete checks if a table has soft delete by looking for deleted_at column
func tableHasSoftDelete(table *schema.Table) bool {
	return tableHasColumn(table, "deleted_at")
}

// findTablesWithColumn returns all tables that have a column with the given name
func findTablesWithColumn(columnName string) []tableInfo {
	var tables []tableInfo

	for _, table := range migrate.Tables {
		if tableHasColumn(table, columnName) {
			tables = append(tables, tableInfo{
				name:          table.Name,
				hasSoftDelete: tableHasSoftDelete(table),
			})
		}
	}

	return tables
}

// findTableWithColumn returns the first table that has a column with the given name
func findTableWithColumn(columnName string) *tableInfo {
	for _, table := range migrate.Tables {
		if tableHasColumn(table, columnName) {
			return &tableInfo{
				name:          table.Name,
				hasSoftDelete: tableHasSoftDelete(table),
			}
		}
	}

	return nil
}

// IsValidObjectType returns true if any table has a column matching the object type enum pattern
func IsValidObjectType(objectType string) bool {
	columnName := fmt.Sprintf("%s_kind_id", strcase.SnakeCase(objectType))
	return findTableWithColumn(columnName) != nil
}

// IsValidGlobalEnumField returns true if any table has a column matching the global enum field pattern
func IsValidGlobalEnumField(fieldName string) bool {
	columnName := fmt.Sprintf("%s_id", strcase.SnakeCase(fieldName))
	tables := findTablesWithColumn(columnName)

	return len(tables) > 0
}

// HookCustomTypeEnumCreate validates that global enums have a valid field
func HookCustomTypeEnumCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CustomTypeEnumFunc(func(ctx context.Context, m *generated.CustomTypeEnumMutation) (generated.Value, error) {
			objectType, _ := m.ObjectType()
			if objectType != "" {
				return next.Mutate(ctx, m)
			}

			fieldName, ok := m.GetField()
			if !ok || fieldName == "" {
				fieldName = "kind"
			}

			if !IsValidGlobalEnumField(fieldName) {
				return nil, fmt.Errorf("%w: %s is not a valid global enum field", ErrInvalidGlobalEnumField, fieldName)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
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

// isGlobalEnumInUse checks if a global enum is in use across all tables with the enum column
func isGlobalEnumInUse(ctrlCtx, logCtx context.Context, client *generated.Client, enumID, enumField, name string, allErrors *[]string, mu *sync.Mutex) func() {
	columnName := fmt.Sprintf("%s_id", strcase.SnakeCase(enumField))
	tables := findTablesWithColumn(columnName)

	if len(tables) == 0 {
		return func() {}
	}

	return func() {
		var unionParts []string
		var tableNames []string

		for _, table := range tables {
			tableNames = append(tableNames, table.name)

			if table.hasSoftDelete {
				unionParts = append(unionParts, fmt.Sprintf("SELECT count(id) as cnt FROM %s WHERE %s = $1 AND deleted_at IS NULL", table.name, columnName))
			} else {
				unionParts = append(unionParts, fmt.Sprintf("SELECT count(id) as cnt FROM %s WHERE %s = $1", table.name, columnName))
			}
		}

		query := fmt.Sprintf("SELECT SUM(cnt) FROM (%s) combined", strings.Join(unionParts, " UNION ALL "))

		var rows sql.Rows
		if err := client.Driver().Query(ctrlCtx, query, lo.ToAnySlice([]string{enumID}), &rows); err != nil {
			mu.Lock()
			logx.FromContext(logCtx).Error().Err(err).Str("enum_field", enumField).Str("enum_id", enumID).Strs("tables", tableNames).Msg("failed to query global enum usage")
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
	columnName := fmt.Sprintf("%s_%s_id", strcase.SnakeCase(edgeName), strcase.SnakeCase(enumField))
	label := strings.ReplaceAll(edgeName, "_", " ")

	tblInfo := findTableWithColumn(columnName)
	if tblInfo == nil {
		// fallback to convention if no table found
		tableName := pluralize.NewClient().Plural(strcase.SnakeCase(edgeName))
		tblInfo = &tableInfo{name: tableName, hasSoftDelete: true}
	}

	return func() {
		var query string
		if tblInfo.hasSoftDelete {
			query = fmt.Sprintf("SELECT count(id) FROM %s WHERE %s = $1 AND deleted_at IS NULL", tblInfo.name, columnName)
		} else {
			query = fmt.Sprintf("SELECT count(id) FROM %s WHERE %s = $1", tblInfo.name, columnName)
		}

		var rows sql.Rows
		if err := client.Driver().Query(ctrlCtx, query, lo.ToAnySlice([]string{enumID}), &rows); err != nil {
			mu.Lock()
			logx.FromContext(logCtx).Error().Err(err).Str("table", tblInfo.name).Str("field", columnName).Str("enum_id", enumID).Msg("failed to query enum edges")
			*allErrors = append(*allErrors, fmt.Sprintf("failed to check if %s enum %s is in use: %v", objectType, name, err))
			mu.Unlock()

			return
		}
		defer rows.Close()

		var count int
		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				mu.Lock()
				logx.FromContext(logCtx).Error().Err(err).Str("table", tblInfo.name).Str("field", columnName).Str("enum_id", enumID).Msg("failed to scan enum edge count")
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
