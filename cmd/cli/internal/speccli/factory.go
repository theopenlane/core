//go:build cli

package speccli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/cli/tables"
)

// Register wires a resource command using the provided specification.
func Register(spec CommandSpec) {
	root := &cobra.Command{
		Use:     spec.Use,
		Short:   spec.Short,
		Aliases: spec.Aliases,
	}

	cmdpkg.RootCmd.AddCommand(root)

	builder := commandBuilder{spec: spec, root: root}

	if spec.Primary != nil {
		if len(spec.Primary.Flags) > 0 {
			registerFieldFlags(root, spec.Primary.Flags)
		}

		if spec.Primary.PreHook != nil {
			root.RunE = func(c *cobra.Command, _ []string) error {
				return spec.Primary.PreHook(c.Context(), c)
			}
		}
	}

	if spec.Get != nil {
		root.AddCommand(builder.getCommand())
	}

	if spec.Create != nil {
		root.AddCommand(builder.createCommand())
	}

	if spec.Update != nil {
		root.AddCommand(builder.updateCommand())
	}

	if spec.Delete != nil {
		root.AddCommand(builder.deleteCommand())
	}
}

// commandBuilder assembles cobra commands for a single resource spec.
type commandBuilder struct {
	spec CommandSpec
	root *cobra.Command
}

type OperationOutput struct {
	// Raw holds the original API response for JSON output mode.
	Raw any
	// Records enumerates the rows rendered in table mode.
	Records []map[string]any
}

// getColumns selects either the standard or delete-specific column set.
func (b commandBuilder) getColumns(forDelete bool) []ColumnSpec {
	if forDelete && len(b.spec.DeleteColumns) > 0 {
		return b.spec.DeleteColumns
	}

	return b.spec.Columns
}

// getCommand builds the cobra command that powers `resource get`.
func (b commandBuilder) getCommand() *cobra.Command {
	getSpec := b.spec.Get

	cmd := &cobra.Command{
		Use:   "get",
		Short: fmt.Sprintf("get an existing %s", b.spec.Name),
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()

			var (
				client  *openlaneclient.OpenlaneClient
				cleanup func()
				err     error
			)

			if getSpec.PreHook != nil {
				client, cleanup, err = acquireClient(ctx)
				if err != nil {
					return err
				}

				defer cleanup()

				handled, out, hookErr := getSpec.PreHook(ctx, c, client)
				if hookErr != nil {
					return hookErr
				}
				if handled {
					return renderOutput(c, out, b.getColumns(false))
				}
			}

			id := strings.TrimSpace(cmdpkg.Config.String(getSpec.IDFlag.Name))
			if id == "" {
				if !getSpec.FallbackList || b.spec.List == nil {
					return cmdpkg.NewRequiredFieldMissingError(fmt.Sprintf("%s id", b.spec.Name))
				}

				return b.executeList(ctx, c)
			}

			if client == nil {
				client, cleanup, err = acquireClient(ctx)
				if err != nil {
					return err
				}
				defer cleanup()
			}

			out, err := b.executeGet(ctx, c, client, id)
			if err != nil {
				return err
			}

			return renderOutput(c, out, b.getColumns(false))
		},
	}

	if getSpec.IDFlag.Shorthand != "" {
		cmd.Flags().StringP(getSpec.IDFlag.Name, getSpec.IDFlag.Shorthand, "", getSpec.IDFlag.Usage)
	} else {
		cmd.Flags().String(getSpec.IDFlag.Name, "", getSpec.IDFlag.Usage)
	}

	if len(getSpec.Flags) > 0 {
		registerFieldFlags(cmd, getSpec.Flags)
	}

	return cmd
}

// createCommand builds the cobra command that powers `resource create`.
func (b commandBuilder) createCommand() *cobra.Command {
	createSpec := b.spec.Create

	cmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("create a new %s", b.spec.Name),
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			client, cleanup, err := acquireClient(ctx)
			if err != nil {
				return err
			}

			defer cleanup()

			if createSpec.PreHook != nil {
				handled, out, hookErr := createSpec.PreHook(ctx, c, client)
				if hookErr != nil {
					return hookErr
				}
				if handled {
					return renderOutput(c, out, b.getColumns(false))
				}
			}

			payload, err := buildInput(createSpec.Fields, createSpec.InputType, c)
			if err != nil {
				return err
			}

			result, err := callClientMethod(client, createSpec.Method, ctx, payload)
			if err != nil {
				return err
			}

			out, err := transformSingle(result, createSpec.ResultPath)
			if err != nil {
				return err
			}

			return renderOutput(c, out, b.getColumns(false))
		},
	}

	registerFieldFlags(cmd, createSpec.Fields)

	return cmd
}

// updateCommand builds the cobra command that powers `resource update`.
func (b commandBuilder) updateCommand() *cobra.Command {
	updateSpec := b.spec.Update

	cmd := &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("update an existing %s", b.spec.Name),
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			client, cleanup, err := acquireClient(ctx)
			if err != nil {
				return err
			}

			defer cleanup()

			if updateSpec.PreHook != nil {
				handled, out, hookErr := updateSpec.PreHook(ctx, c, client)
				if hookErr != nil {
					return hookErr
				}
				if handled {
					return renderOutput(c, out, b.getColumns(false))
				}
			}

			id := cmdpkg.Config.String(updateSpec.IDFlag.Name)
			if strings.TrimSpace(id) == "" {
				return cmdpkg.NewRequiredFieldMissingError(fmt.Sprintf("%s id", b.spec.Name))
			}

			payload, err := buildInput(updateSpec.Fields, updateSpec.InputType, c)
			if err != nil {
				return err
			}

			result, err := callClientMethod(client, updateSpec.Method, ctx, id, payload)
			if err != nil {
				return err
			}

			out, err := transformSingle(result, updateSpec.ResultPath)
			if err != nil {
				return err
			}

			return renderOutput(c, out, b.getColumns(false))
		},
	}

	if updateSpec.IDFlag.Shorthand != "" {
		cmd.Flags().StringP(updateSpec.IDFlag.Name, updateSpec.IDFlag.Shorthand, "", updateSpec.IDFlag.Usage)
	} else {
		cmd.Flags().String(updateSpec.IDFlag.Name, "", updateSpec.IDFlag.Usage)
	}

	registerFieldFlags(cmd, updateSpec.Fields)

	return cmd
}

// deleteCommand builds the cobra command that powers `resource delete`.
func (b commandBuilder) deleteCommand() *cobra.Command {
	deleteSpec := b.spec.Delete

	cmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("delete an existing %s", b.spec.Name),
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			client, cleanup, err := acquireClient(ctx)
			if err != nil {
				return err
			}

			defer cleanup()

			if deleteSpec.PreHook != nil {
				handled, out, hookErr := deleteSpec.PreHook(ctx, c, client)
				if hookErr != nil {
					return hookErr
				}
				if handled {
					return renderOutput(c, out, b.getColumns(true))
				}
			}

			id := cmdpkg.Config.String(deleteSpec.IDFlag.Name)
			if strings.TrimSpace(id) == "" {
				return cmdpkg.NewRequiredFieldMissingError(fmt.Sprintf("%s id", b.spec.Name))
			}

			result, err := callClientMethod(client, deleteSpec.Method, ctx, id)
			if err != nil {
				return err
			}

			out, err := transformDelete(result, deleteSpec.ResultPath, deleteSpec.ResultField)
			if err != nil {
				return err
			}

			return renderOutput(c, out, b.getColumns(true))
		},
	}

	if deleteSpec.IDFlag.Shorthand != "" {
		cmd.Flags().StringP(deleteSpec.IDFlag.Name, deleteSpec.IDFlag.Shorthand, "", deleteSpec.IDFlag.Usage)
	} else {
		cmd.Flags().String(deleteSpec.IDFlag.Name, "", deleteSpec.IDFlag.Usage)
	}

	return cmd
}

// executeList runs the list client call, handling optional filters and output formatting.
func (b commandBuilder) executeList(ctx context.Context, c *cobra.Command) error {
	if b.spec.List == nil {
		return errors.New("list specification is not defined")
	}

	client, cleanup, err := acquireClient(ctx)
	if err != nil {
		return err
	}

	defer cleanup()

	whereInput := mo.None[any]()
	if b.spec.List.Where != nil {
		var err error
		whereInput, err = buildWhereInput(b.spec.List.Where, c)
		if err != nil {
			return err
		}
	}

	var result any

	if whereValue, ok := whereInput.Get(); ok {
		if b.spec.List.FilterMethod == "" {
			return fmt.Errorf("%s list does not support filtering", b.spec.Name)
		}

		result, err = callClientMethod(client, b.spec.List.FilterMethod, ctx, cmdpkg.First, cmdpkg.Last, whereValue)
	} else {
		result, err = callClientMethod(client, b.spec.List.Method, ctx)
	}

	if err != nil {
		return err
	}

	out, err := transformList(result, b.spec.List.Root)
	if err != nil {
		return err
	}

	return renderOutput(c, out, b.getColumns(false))
}

// executeGet runs the get client call, falling back to where filters when configured.
func (b commandBuilder) executeGet(ctx context.Context, c *cobra.Command, client *openlaneclient.OpenlaneClient, id string) (OperationOutput, error) {
	getSpec := b.spec.Get

	var (
		result any
		err    error
	)

	if getSpec.Where != nil {
		whereInput, buildErr := buildWhereInput(getSpec.Where, c)
		if buildErr != nil {
			return OperationOutput{}, buildErr
		}

		if whereInput.IsAbsent() {
			generated, genErr := buildWhereFromID(getSpec.Where, getSpec.IDFlag.Name, id)
			if genErr != nil {
				return OperationOutput{}, genErr
			}

			whereInput = mo.Some[any](generated)
		}

		result, err = callClientMethod(client, getSpec.Method, ctx, cmdpkg.First, cmdpkg.Last, whereInput.MustGet())
	} else {
		result, err = callClientMethod(client, getSpec.Method, ctx, id)
	}
	if err != nil {
		return OperationOutput{}, err
	}

	if getSpec.ListRoot != "" {
		return transformList(result, getSpec.ListRoot)
	}

	return transformSingle(result, getSpec.ResultPath)
}

// registerFieldFlags registers Cobra flags for the provided FieldSpecs.
func registerFieldFlags(cmd *cobra.Command, fields []FieldSpec) {
	for _, field := range fields {
		switch field.Kind {
		case ValueString:
			if field.Flag.Shorthand != "" {
				cmd.Flags().StringP(field.Flag.Name, field.Flag.Shorthand, field.Flag.Default, field.Flag.Usage)
			} else {
				cmd.Flags().String(field.Flag.Name, field.Flag.Default, field.Flag.Usage)
			}
		case ValueStringSlice:
			defaults := []string{}
			if field.Flag.Default != "" {
				defaults = lo.FilterMap(strings.Split(field.Flag.Default, ","), func(item string, _ int) (string, bool) {
					trimmed := strings.TrimSpace(item)
					if trimmed == "" {
						return "", false
					}

					return trimmed, true
				})
			}

			if field.Flag.Shorthand != "" {
				cmd.Flags().StringSliceP(field.Flag.Name, field.Flag.Shorthand, defaults, field.Flag.Usage)
			} else {
				cmd.Flags().StringSlice(field.Flag.Name, defaults, field.Flag.Usage)
			}
		case ValueBool:
			defaultBool := false
			if field.Flag.Default != "" {
				parsed, err := strconv.ParseBool(field.Flag.Default)
				if err != nil {
					panic(fmt.Sprintf("invalid boolean default for flag %s: %v", field.Flag.Name, err))
				}
				defaultBool = parsed
			}

			if field.Flag.Shorthand != "" {
				cmd.Flags().BoolP(field.Flag.Name, field.Flag.Shorthand, defaultBool, field.Flag.Usage)
			} else {
				cmd.Flags().Bool(field.Flag.Name, defaultBool, field.Flag.Usage)
			}
		case ValueDuration:
			defaultDuration := time.Duration(0)
			if field.Flag.Default != "" {
				parsed, err := time.ParseDuration(field.Flag.Default)
				if err != nil {
					panic(fmt.Sprintf("invalid duration default for flag %s: %v", field.Flag.Name, err))
				}
				defaultDuration = parsed
			}

			if field.Flag.Shorthand != "" {
				cmd.Flags().DurationP(field.Flag.Name, field.Flag.Shorthand, defaultDuration, field.Flag.Usage)
			} else {
				cmd.Flags().Duration(field.Flag.Name, defaultDuration, field.Flag.Usage)
			}
		case ValueInt:
			defaultInt := int64(0)
			if field.Flag.Default != "" {
				parsed, err := strconv.ParseInt(field.Flag.Default, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("invalid integer default for flag %s: %v", field.Flag.Name, err))
				}
				defaultInt = parsed
			}

			if field.Flag.Shorthand != "" {
				cmd.Flags().Int64P(field.Flag.Name, field.Flag.Shorthand, defaultInt, field.Flag.Usage)
			} else {
				cmd.Flags().Int64(field.Flag.Name, defaultInt, field.Flag.Usage)
			}
		}
	}
}

// renderOutput writes either JSON or table output for a command invocation.
func renderOutput(cmd *cobra.Command, out OperationOutput, columns []ColumnSpec) error {
	if strings.EqualFold(cmdpkg.OutputFormat, cmdpkg.JSONOutput) {
		return PrintJSON(out.Raw)
	}

	if len(columns) == 0 || len(out.Records) == 0 {
		return nil
	}

	headers := lo.Map(columns, func(col ColumnSpec, _ int) string {
		return col.Header
	})

	writer := tables.NewTableWriter(cmd.OutOrStdout(), headers...)

	rows := lo.Map(out.Records, func(record map[string]any, _ int) []any {
		return lo.Map(columns, func(col ColumnSpec, _ int) any {
			value := extractValue(record, col.Path)
			return applyColumnFormatter(col.Format, value)
		})
	})

	for _, row := range rows {
		writer.AddRow(row...)
	}

	writer.Render()

	return nil
}

// formatValue provides human readable defaults for common value kinds.
func formatValue(v any) any {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case []any:
		strs := lo.Map(val, func(item any, _ int) string {
			return fmt.Sprint(item)
		})

		return strings.Join(strs, ", ")
	case []string:
		return strings.Join(val, ", ")
	default:
		return val
	}
}

// applyColumnFormatter looks up an optional formatter before falling back to default formatting.
func applyColumnFormatter(format string, value any) any {
	if format == "" {
		return formatValue(value)
	}

	if formatter, ok := columnFormatters[format]; ok {
		formatted, err := formatter(value)
		if err == nil {
			return formatted
		}
	}

	return formatValue(value)
}

// extractValue walks a flattened path within a record.
func extractValue(record map[string]any, path []string) any {
	if len(path) == 0 {
		return record
	}

	current := any(record)

	for _, segment := range path {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[segment]
			if !ok {
				return nil
			}

			current = next
		default:
			return nil
		}
	}

	return current
}

// transformList converts a GraphQL list response into OperationOutput.
func transformList(result any, root string) (OperationOutput, error) {
	nodes, err := extractEdgeNodes(result, root)
	if err != nil {
		return OperationOutput{}, err
	}

	return OperationOutput{
		Raw:     result,
		Records: nodes,
	}, nil
}

// transformSingle converts a GraphQL object response into OperationOutput.
func transformSingle(result any, path []string) (OperationOutput, error) {
	node, err := extractNode(result, path)
	if err != nil {
		return OperationOutput{}, err
	}

	return OperationOutput{
		Raw:     result,
		Records: []map[string]any{node},
	}, nil
}

// transformDelete extracts the requested field from a delete payload for display.
func transformDelete(result any, path []string, field string) (OperationOutput, error) {
	value, err := extractValueFromPath(result, path)
	if err != nil {
		return OperationOutput{}, err
	}

	record := map[string]any{
		field: value,
	}

	return OperationOutput{
		Raw:     result,
		Records: []map[string]any{record},
	}, nil
}

// WrapListResult converts a GraphQL response with edges into OperationOutput using the provided root key.
func WrapListResult(result any, root string) (OperationOutput, error) {
	return transformList(result, root)
}

// WrapSingleResult converts a GraphQL response into OperationOutput using the provided path.
func WrapSingleResult(result any, path []string) (OperationOutput, error) {
	return transformSingle(result, path)
}

// extractEdgeNodes normalizes GraphQL edge responses into plain node maps.
func extractEdgeNodes(result any, root string) ([]map[string]any, error) {
	data, err := normalizeToMap(result)
	if err != nil {
		return nil, err
	}

	rootNode, ok := data[root].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing root node %q in list response", root)
	}

	rawEdgesValue, ok := rootNode["edges"]
	if !ok || rawEdgesValue == nil {
		return []map[string]any{}, nil
	}

	rawEdges, ok := rawEdgesValue.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected edges format in list response for %q", root)
	}

	return edgeNodesFromEdges(rawEdges)
}

// extractNode plucks a nested node and normalizes it into a map.
func extractNode(result any, path []string) (map[string]any, error) {
	value, err := extractValueFromPath(result, path)
	if err != nil {
		return nil, err
	}

	return normalizeToMap(value)
}

// extractValueFromPath walks a nested structure following the provided path.
func extractValueFromPath(result any, path []string) (any, error) {
	current := any(result)

	for _, segment := range path {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[segment]
			if !ok {
				return nil, fmt.Errorf("missing path segment %q", segment)
			}

			current = next
		default:
			nextMap, err := normalizeToMap(node)
			if err != nil {
				return nil, err
			}

			next, ok := nextMap[segment]
			if !ok {
				return nil, fmt.Errorf("missing path segment %q", segment)
			}

			current = next
		}
	}

	return current, nil
}

// normalizeToMap converts supported value shapes into map[string]any for downstream processing.
func normalizeToMap(value any) (map[string]any, error) {
	switch v := value.(type) {
	case nil:
		return nil, errors.New("cannot convert nil to map")
	case map[string]any:
		return v, nil
	case *map[string]any:
		if v == nil {
			return nil, errors.New("cannot convert nil pointer to map")
		}

		return *v, nil
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}

		return out, nil
	}
}

// edgeNodesFromEdges normalizes a slice of GraphQL edges into plain node maps,
// stopping early if any edge cannot be coerced into the expected shape.
func edgeNodesFromEdges(edges []any) ([]map[string]any, error) {
	var firstErr error

	nodes := lo.FilterMap(edges, func(edge any, _ int) (map[string]any, bool) {
		if firstErr != nil {
			return nil, false
		}

		edgeMap, err := normalizeToMap(edge)
		if err != nil {
			firstErr = err
			return nil, false
		}

		node, err := normalizeToMap(edgeMap["node"])
		if err != nil {
			firstErr = err
			return nil, false
		}

		return node, true
	})

	if firstErr != nil {
		return nil, firstErr
	}

	return nodes, nil
}

// buildInput hydrates a GraphQL input struct from CLI flag values.
func buildInput(fields []FieldSpec, inputType reflect.Type, cmd *cobra.Command) (any, error) {
	if inputType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("inputType must be a struct, got %s", inputType.Kind())
	}

	value := reflect.New(inputType).Elem()

	for _, field := range fields {
		valueResult := extractFieldValue(field, cmd)
		if valueResult.IsError() {
			return nil, valueResult.Error()
		}

		valueOption := valueResult.MustGet()
		if valueOption.IsAbsent() {
			if field.Flag.Required {
				return nil, cmdpkg.NewRequiredFieldMissingError(field.Flag.Name)
			}

			continue
		}

		if field.Field == "" {
			continue
		}

		if err := assignField(value, field.Field, valueOption.MustGet()); err != nil {
			return nil, err
		}
	}

	return value.Interface(), nil
}

// extractFieldValue reads a CLI flag according to the FieldSpec metadata and
// returns a typed option indicating whether the user provided a value.
func extractFieldValue(field FieldSpec, cmd *cobra.Command) mo.Result[mo.Option[any]] {
	switch field.Kind {
	case ValueString:
		val := strings.TrimSpace(cmdpkg.Config.String(field.Flag.Name))
		if val == "" {
			return mo.Ok(mo.None[any]())
		}

		return parseFieldValue(field, val)
	case ValueStringSlice:
		provided := cmd.Flags().Changed(field.Flag.Name) || cmdpkg.Config.Exists(field.Flag.Name)
		values := cmdpkg.Config.Strings(field.Flag.Name)
		if !provided {
			return mo.Ok(mo.None[any]())
		}

		return parseFieldValue(field, values)
	case ValueBool:
		provided := cmd.Flags().Changed(field.Flag.Name) || cmdpkg.Config.Exists(field.Flag.Name)
		if !provided {
			return mo.Ok(mo.None[any]())
		}

		return parseFieldValue(field, cmdpkg.Config.Bool(field.Flag.Name))
	case ValueDuration:
		provided := cmd.Flags().Changed(field.Flag.Name) || cmdpkg.Config.Exists(field.Flag.Name)
		if !provided {
			return mo.Ok(mo.None[any]())
		}

		duration := cmdpkg.Config.Duration(field.Flag.Name)

		return parseFieldValue(field, duration)
	case ValueInt:
		provided := cmd.Flags().Changed(field.Flag.Name) || cmdpkg.Config.Exists(field.Flag.Name)
		if !provided {
			return mo.Ok(mo.None[any]())
		}

		return parseFieldValue(field, cmdpkg.Config.Int64(field.Flag.Name))
	default:
		return mo.Err[mo.Option[any]](fmt.Errorf("unsupported field kind for %s", field.Flag.Name))
	}
}

// parseFieldValue applies an optional parser to the raw flag input and wraps it
// in an Option so callers can distinguish between unset and zero values.
func parseFieldValue(field FieldSpec, value any) mo.Result[mo.Option[any]] {
	if field.Parser == nil {
		return mo.Ok(mo.Some[any](value))
	}

	parsed, err := field.Parser(value)
	if err != nil {
		return mo.Err[mo.Option[any]](err)
	}

	return mo.Ok(mo.Some[any](parsed))
}

// buildWhereInput inspects list/get filter flags, constructing the corresponding
// GraphQL where input when at least one filter flag is set.
func buildWhereInput(spec *WhereSpec, cmd *cobra.Command) (mo.Option[any], error) {
	if spec == nil {
		return mo.None[any](), nil
	}

	hasFilters := lo.SomeBy(spec.Fields, func(field FieldSpec) bool {
		return cmd.Flags().Changed(field.Flag.Name)
	})

	if !hasFilters {
		return mo.None[any](), nil
	}

	value, err := buildInput(spec.Fields, spec.Type, cmd)
	if err != nil {
		return mo.None[any](), err
	}

	ptr := reflect.New(spec.Type)
	ptr.Elem().Set(reflect.ValueOf(value))

	return mo.Some[any](ptr.Interface()), nil
}

// buildWhereFromID fabricates a where input using the supplied ID flag so get
// commands can fall back to filter-based queries when an explicit where clause
// is required by the API.
func buildWhereFromID(spec *WhereSpec, idFlagName string, id string) (any, error) {
	if spec == nil || len(spec.Fields) == 0 {
		return nil, fmt.Errorf("unable to build where input without field specifications")
	}

	targetField, ok := lo.Find(spec.Fields, func(field FieldSpec) bool {
		return strings.EqualFold(field.Flag.Name, idFlagName)
	})
	if !ok {
		targetField = spec.Fields[0]
	}

	value := reflect.New(spec.Type).Elem()

	if err := assignField(value, targetField.Field, id); err != nil {
		return nil, err
	}

	return value.Addr().Interface(), nil
}

// assignField sets a struct field, handling pointer and slice conversions safely.
func assignField(target reflect.Value, fieldName string, value any) error {
	field := target.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %q not found on input struct", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("cannot set field %q", fieldName)
	}

	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)

	switch field.Kind() {
	case reflect.Ptr:
		elemType := field.Type().Elem()
		ptr := reflect.New(elemType)

		if val.Type() != elemType && val.Type().ConvertibleTo(elemType) {
			val = val.Convert(elemType)
		}

		if val.Type() != elemType {
			return fmt.Errorf("cannot assign value of type %s to pointer field %s", val.Type(), fieldName)
		}

		ptr.Elem().Set(val)
		field.Set(ptr)
	case reflect.Slice:
		if val.Type() != field.Type() && val.Type().ConvertibleTo(field.Type()) {
			val = val.Convert(field.Type())
		}

		if val.Type() != field.Type() {
			return fmt.Errorf("cannot assign value of type %s to slice field %s", val.Type(), fieldName)
		}

		field.Set(val)
	default:
		if val.Type() != field.Type() && val.Type().ConvertibleTo(field.Type()) {
			val = val.Convert(field.Type())
		}

		if val.Type() != field.Type() {
			return fmt.Errorf("cannot assign value of type %s to field %s", val.Type(), fieldName)
		}

		field.Set(val)
	}

	return nil
}

// acquireClient obtains an authenticated Openlane client, returning a cleanup callback.
func acquireClient(ctx context.Context) (*openlaneclient.OpenlaneClient, func(), error) {
	client, err := cmdpkg.TokenAuth(ctx, cmdpkg.Config)
	if err != nil || client == nil {
		client, err = cmdpkg.SetupClientWithAuth(ctx)
		if err != nil {
			return nil, func() {}, err
		}
		return client, func() { cmdpkg.StoreSessionCookies(client) }, nil
	}

	return client, func() {}, nil
}

// callClientMethod uses reflection to invoke an Openlane client method and returns the results.
func callClientMethod(client *openlaneclient.OpenlaneClient, method string, args ...any) (any, error) {
	value := reflect.ValueOf(client)
	call := value.MethodByName(method)
	if !call.IsValid() {
		return nil, fmt.Errorf("openlaneclient method %q not found", method)
	}

	if call.Type().NumOut() != 2 {
		return nil, fmt.Errorf("method %q must return (value, error)", method)
	}

	inputs := make([]reflect.Value, len(args))
	for i, arg := range args {
		inputs[i] = reflect.ValueOf(arg)
	}

	results := call.Call(inputs)

	if errVal := results[1]; !errVal.IsNil() {
		err, ok := errVal.Interface().(error)
		if !ok {
			return nil, fmt.Errorf("method %q returned non-error type", method)
		}

		return nil, err
	}

	return results[0].Interface(), nil
}
