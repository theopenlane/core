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

type commandBuilder struct {
	spec CommandSpec
	root *cobra.Command
}

type OperationOutput struct {
	Raw     any
	Records []map[string]any
}

func (b commandBuilder) getColumns(forDelete bool) []ColumnSpec {
	if forDelete && len(b.spec.DeleteColumns) > 0 {
		return b.spec.DeleteColumns
	}

	return b.spec.Columns
}

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

func (b commandBuilder) executeList(ctx context.Context, c *cobra.Command) error {
	if b.spec.List == nil {
		return errors.New("list specification is not defined")
	}

	client, cleanup, err := acquireClient(ctx)
	if err != nil {
		return err
	}

	defer cleanup()

	useFilters := false
	var wherePtr any

	if b.spec.List.Where != nil {
		var err error
		wherePtr, useFilters, err = buildWhereInput(b.spec.List.Where, c)
		if err != nil {
			return err
		}
	}

	var result any

	if useFilters {
		if b.spec.List.FilterMethod == "" {
			return fmt.Errorf("%s list does not support filtering", b.spec.Name)
		}

		result, err = callClientMethod(client, b.spec.List.FilterMethod, ctx, cmdpkg.First, cmdpkg.Last, wherePtr)
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

func (b commandBuilder) executeGet(ctx context.Context, c *cobra.Command, client *openlaneclient.OpenlaneClient, id string) (OperationOutput, error) {
	getSpec := b.spec.Get

	var (
		result any
		err    error
	)

	if getSpec.Where != nil {
		wherePtr, useFilters, buildErr := buildWhereInput(getSpec.Where, c)
		if buildErr != nil {
			return OperationOutput{}, buildErr
		}

		if !useFilters {
			generated, genErr := buildWhereFromID(getSpec.Where, id)
			if genErr != nil {
				return OperationOutput{}, genErr
			}

			wherePtr = generated
		}

		result, err = callClientMethod(client, getSpec.Method, ctx, cmdpkg.First, cmdpkg.Last, wherePtr)
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
				raw := strings.SplitSeq(field.Flag.Default, ",")
				for item := range raw {
					trimmed := strings.TrimSpace(item)
					if trimmed != "" {
						defaults = append(defaults, trimmed)
					}
				}
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

func renderOutput(cmd *cobra.Command, out OperationOutput, columns []ColumnSpec) error {
	if strings.EqualFold(cmdpkg.OutputFormat, cmdpkg.JSONOutput) {
		return PrintJSON(out.Raw)
	}

	if len(columns) == 0 || len(out.Records) == 0 {
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}

	writer := tables.NewTableWriter(cmd.OutOrStdout(), headers...)

	for _, record := range out.Records {
		row := make([]any, len(columns))
		for idx, col := range columns {
			value := extractValue(record, col.Path)
			row[idx] = applyColumnFormatter(col.Format, value)
		}

		writer.AddRow(row...)
	}

	writer.Render()

	return nil
}

func formatValue(v any) any {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case []any:
		strs := make([]string, 0, len(val))
		for _, item := range val {
			strs = append(strs, fmt.Sprint(item))
		}

		return strings.Join(strs, ", ")
	case []string:
		return strings.Join(val, ", ")
	default:
		return val
	}
}

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

	nodes := make([]map[string]any, 0, len(rawEdges))
	for _, edge := range rawEdges {
		edgeMap, err := normalizeToMap(edge)
		if err != nil {
			return nil, err
		}

		node, err := normalizeToMap(edgeMap["node"])
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func extractNode(result any, path []string) (map[string]any, error) {
	value, err := extractValueFromPath(result, path)
	if err != nil {
		return nil, err
	}

	return normalizeToMap(value)
}

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

func buildInput(fields []FieldSpec, inputType reflect.Type, cmd *cobra.Command) (any, error) {
	if inputType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("inputType must be a struct, got %s", inputType.Kind())
	}

	value := reflect.New(inputType).Elem()

	for _, field := range fields {
		var (
			set bool
			raw any
			err error
		)

		switch field.Kind {
		case ValueString:
			val := strings.TrimSpace(cmdpkg.Config.String(field.Flag.Name))
			if val != "" {
				set = true
				if field.Parser != nil {
					raw, err = field.Parser(val)
				} else {
					raw = val
				}
			}
		case ValueStringSlice:
			values := cmdpkg.Config.Strings(field.Flag.Name)
			if len(values) > 0 {
				set = true
				if field.Parser != nil {
					raw, err = field.Parser(values)
				} else {
					raw = values
				}
			}
		case ValueBool:
			rawBool := cmdpkg.Config.Bool(field.Flag.Name)
			if cmd.Flags().Changed(field.Flag.Name) || cmdpkg.Config.Exists(field.Flag.Name) {
				set = true
				if field.Parser != nil {
					raw, err = field.Parser(rawBool)
				} else {
					raw = rawBool
				}
			}
		case ValueDuration:
			duration := cmdpkg.Config.Duration(field.Flag.Name)
			if duration != 0 {
				set = true
				if field.Parser != nil {
					raw, err = field.Parser(duration)
				} else {
					raw = duration
				}
			}
		case ValueInt:
			rawInt := cmdpkg.Config.Int64(field.Flag.Name)
			if cmd.Flags().Changed(field.Flag.Name) || cmdpkg.Config.Exists(field.Flag.Name) {
				set = true
				if field.Parser != nil {
					raw, err = field.Parser(rawInt)
				} else {
					raw = rawInt
				}
			}
		default:
			return nil, fmt.Errorf("unsupported field kind for %s", field.Flag.Name)
		}

		if err != nil {
			return nil, err
		}

		if !set {
			if field.Flag.Required {
				return nil, cmdpkg.NewRequiredFieldMissingError(field.Flag.Name)
			}

			continue
		}

		if field.Field == "" {
			continue
		}

		if err := assignField(value, field.Field, raw); err != nil {
			return nil, err
		}
	}

	return value.Interface(), nil
}

func buildWhereInput(spec *WhereSpec, cmd *cobra.Command) (any, bool, error) {
	if spec == nil {
		return nil, false, nil
	}

	useFilters := false
	for _, field := range spec.Fields {
		if cmd.Flags().Changed(field.Flag.Name) {
			useFilters = true
			break
		}
	}

	if !useFilters {
		return nil, false, nil
	}

	value, err := buildInput(spec.Fields, spec.Type, cmd)
	if err != nil {
		return nil, false, err
	}

	ptr := reflect.New(spec.Type)
	ptr.Elem().Set(reflect.ValueOf(value))

	return ptr.Interface(), true, nil
}

func buildWhereFromID(spec *WhereSpec, id string) (any, error) {
	if spec == nil || len(spec.Fields) == 0 {
		return nil, fmt.Errorf("unable to build where input without field specifications")
	}

	value := reflect.New(spec.Type).Elem()

	field := spec.Fields[0]
	if err := assignField(value, field.Field, id); err != nil {
		return nil, err
	}

	return value.Addr().Interface(), nil
}

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
