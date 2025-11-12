//go:build cli

package speccli

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"reflect"
	"sort"
	"strings"

	"github.com/samber/lo"
)

// LoaderOptions controls how specs are resolved from an embedded filesystem.
type LoaderOptions struct {
	// Pattern determines which files are considered specs.
	Pattern string
	// TypeResolver resolves type strings (e.g. openlaneclient inputs) to reflect types.
	TypeResolver TypeResolver
	// Parsers maps parser names to implementations.
	Parsers map[string]ValueParser
	// Overrides are spec mutations applied after loading.
	Overrides map[string]SpecOverride
	// CreateHooks enumerates available Create hooks keyed by name.
	CreateHooks map[string]CreateHookFactory
	// UpdateHooks enumerates available Update hooks keyed by name.
	UpdateHooks map[string]UpdateHookFactory
	// GetHooks enumerates available Get hooks keyed by name.
	GetHooks map[string]GetHookFactory
	// PrimaryHooks enumerates available Primary hooks keyed by name.
	PrimaryHooks map[string]PrimaryHookFactory
	// DeleteHooks enumerates available Delete hooks keyed by name.
	DeleteHooks map[string]DeleteHookFactory
}

// SpecOverride mutates a spec after it has been loaded.
type SpecOverride func(*CommandSpec) error

// TypeResolver returns the reflect.Type for a given identifier.
type TypeResolver func(string) (reflect.Type, error)

// StaticTypeResolver returns a resolver backed by the provided map.
func StaticTypeResolver(types map[string]reflect.Type) TypeResolver {
	return func(name string) (reflect.Type, error) {
		typ, ok := types[name]
		if !ok {
			return nil, fmt.Errorf("unknown type reference %q", name)
		}

		return typ, nil
	}
}

// LoadSpecsFromFS reads command specs from the provided filesystem without registering them.
func LoadSpecsFromFS(fsys fs.FS, opts LoaderOptions) ([]CommandSpec, error) {
	pattern := opts.Pattern
	if pattern == "" {
		pattern = "*.json"
	}

	files, err := fs.Glob(fsys, pattern)
	if err != nil {
		return nil, err
	}

	// Sort for determinism
	sort.Strings(files)

	specs := make([]CommandSpec, 0, len(files))

	for _, file := range files {
		f, err := fsys.Open(file)
		if err != nil {
			return nil, fmt.Errorf("open spec file %q: %w", file, err)
		}

		spec, err := decodeSpec(f, opts)
		cErr := f.Close()

		if err == nil && cErr != nil {
			err = cErr
		}

		if err != nil {
			return nil, fmt.Errorf("load spec %q: %w", file, err)
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// RegisterFromFS loads specs from fsys and registers each command with the root CLI.
func RegisterFromFS(fsys fs.FS, opts LoaderOptions) error {
	if opts.Parsers == nil {
		opts.Parsers = DefaultParsers()
	}
	if opts.CreateHooks == nil {
		opts.CreateHooks = map[string]CreateHookFactory{}
	}
	if opts.UpdateHooks == nil {
		opts.UpdateHooks = map[string]UpdateHookFactory{}
	}
	if opts.GetHooks == nil {
		opts.GetHooks = map[string]GetHookFactory{}
	}
	if opts.PrimaryHooks == nil {
		opts.PrimaryHooks = map[string]PrimaryHookFactory{}
	}
	if opts.DeleteHooks == nil {
		opts.DeleteHooks = map[string]DeleteHookFactory{}
	}

	specs, err := LoadSpecsFromFS(fsys, opts)
	if err != nil {
		return err
	}

	for _, spec := range specs {
		Register(spec)
	}

	return nil
}

// fileCommandSpec mirrors the JSON structure stored on disk for a command.
type fileCommandSpec struct {
	Name    string   `json:"name"`
	Use     string   `json:"use"`
	Short   string   `json:"short"`
	Aliases []string `json:"aliases"`

	List          *fileListSpec    `json:"list"`
	Get           *fileGetSpec     `json:"get"`
	Create        *fileCreateSpec  `json:"create"`
	Update        *fileUpdateSpec  `json:"update"`
	Delete        *fileDeleteSpec  `json:"delete"`
	Columns       []fileColumnSpec `json:"columns"`
	DeleteColumns []fileColumnSpec `json:"deleteColumns"`
	Primary       *filePrimarySpec `json:"primary"`
}

// fileListSpec is the on-disk representation of ListSpec.
type fileListSpec struct {
	Method       string         `json:"method"`
	Root         string         `json:"root"`
	FilterMethod string         `json:"filterMethod"`
	Where        *fileWhereSpec `json:"where"`
}

// fileWhereSpec stores where configuration in JSON.
type fileWhereSpec struct {
	Type   string          `json:"type"`
	Fields []fileFieldSpec `json:"fields"`
}

// fileGetSpec stores get configuration in JSON.
type fileGetSpec struct {
	Method       string          `json:"method"`
	IDFlag       FlagSpec        `json:"idFlag"`
	ResultPath   []string        `json:"resultPath"`
	FallbackList bool            `json:"fallbackList"`
	Where        *fileWhereSpec  `json:"where"`
	ListRoot     string          `json:"listRoot"`
	Fields       []fileFieldSpec `json:"fields"`
	PreHook      string          `json:"preHook"`
}

// fileCreateSpec stores create configuration in JSON.
type fileCreateSpec struct {
	Method     string          `json:"method"`
	InputType  string          `json:"inputType"`
	Fields     []fileFieldSpec `json:"fields"`
	ResultPath []string        `json:"resultPath"`
	PreHook    string          `json:"preHook"`
}

// fileUpdateSpec stores update configuration in JSON.
type fileUpdateSpec struct {
	Method     string          `json:"method"`
	IDFlag     FlagSpec        `json:"idFlag"`
	InputType  string          `json:"inputType"`
	Fields     []fileFieldSpec `json:"fields"`
	ResultPath []string        `json:"resultPath"`
	PreHook    string          `json:"preHook"`
}

// fileDeleteSpec stores delete configuration in JSON.
type fileDeleteSpec struct {
	Method      string   `json:"method"`
	IDFlag      FlagSpec `json:"idFlag"`
	ResultPath  []string `json:"resultPath"`
	ResultField string   `json:"resultField"`
	PreHook     string   `json:"preHook"`
}

// fileFieldSpec stores FieldSpec metadata in JSON.
type fileFieldSpec struct {
	Flag       FlagSpec `json:"flag"`
	Kind       string   `json:"kind"`
	Field      string   `json:"field"`
	ParserName string   `json:"parser"`
}

// fileColumnSpec stores ColumnSpec metadata in JSON.
type fileColumnSpec struct {
	Header    string   `json:"header"`
	Path      []string `json:"path"`
	Formatter string   `json:"formatter"`
}

// filePrimarySpec stores primary command configuration in JSON.
type filePrimarySpec struct {
	Fields  []fileFieldSpec `json:"fields"`
	PreHook string          `json:"preHook"`
}

// decodeSpec parses a single spec file into a hydrated CommandSpec.
func decodeSpec(r io.Reader, opts LoaderOptions) (CommandSpec, error) {
	var fileSpec fileCommandSpec

	if err := json.NewDecoder(r).Decode(&fileSpec); err != nil {
		return CommandSpec{}, err
	}

	spec := CommandSpec{
		Name:    fileSpec.Name,
		Use:     fileSpec.Use,
		Short:   fileSpec.Short,
		Aliases: fileSpec.Aliases,
		Columns: lo.Map(fileSpec.Columns, func(column fileColumnSpec, _ int) ColumnSpec {
			return ColumnSpec{
				Header: column.Header,
				Path:   column.Path,
				Format: column.Formatter,
			}
		}),
		DeleteColumns: lo.Map(fileSpec.DeleteColumns, func(column fileColumnSpec, _ int) ColumnSpec {
			return ColumnSpec{
				Header: column.Header,
				Path:   column.Path,
				Format: column.Formatter,
			}
		}),
	}

	if fileSpec.List != nil {
		list, err := hydrateListSpec(*fileSpec.List, opts)
		if err != nil {
			return CommandSpec{}, err
		}

		spec.List = list
	}

	if fileSpec.Get != nil {
		spec.Get = &GetSpec{
			Method:       fileSpec.Get.Method,
			IDFlag:       fileSpec.Get.IDFlag,
			ResultPath:   fileSpec.Get.ResultPath,
			FallbackList: fileSpec.Get.FallbackList,
			ListRoot:     fileSpec.Get.ListRoot,
		}

		if fileSpec.Get.Where != nil {
			where, err := hydrateWhereSpec(*fileSpec.Get.Where, opts)
			if err != nil {
				return CommandSpec{}, err
			}

			spec.Get.Where = where
		}

		if len(fileSpec.Get.Fields) > 0 {
			fields, err := convertFields(fileSpec.Get.Fields, opts.Parsers)
			if err != nil {
				return CommandSpec{}, err
			}

			spec.Get.Flags = fields
		}

		if fileSpec.Get.PreHook != "" {
			factory, ok := opts.GetHooks[fileSpec.Get.PreHook]
			if !ok {
				return CommandSpec{}, fmt.Errorf("unknown get preHook %q", fileSpec.Get.PreHook)
			}

			spec.Get.PreHook = factory(spec.Get)
		}
	}

	if fileSpec.Create != nil {
		create, err := hydrateMutationSpec(*fileSpec.Create, opts)
		if err != nil {
			return CommandSpec{}, err
		}
		spec.Create = &create
	}

	if fileSpec.Update != nil {
		if fileSpec.Update.IDFlag.Name == "" {
			return CommandSpec{}, fmt.Errorf("update spec for %s missing id flag name", fileSpec.Name)
		}

		update, err := hydrateUpdateSpec(*fileSpec.Update, opts)
		if err != nil {
			return CommandSpec{}, err
		}
		spec.Update = &update
	}

	if fileSpec.Delete != nil {
		if fileSpec.Delete.IDFlag.Name == "" {
			return CommandSpec{}, fmt.Errorf("delete spec for %s missing id flag name", fileSpec.Name)
		}

		deleteSpec := &DeleteSpec{
			Method:      fileSpec.Delete.Method,
			IDFlag:      fileSpec.Delete.IDFlag,
			ResultPath:  fileSpec.Delete.ResultPath,
			ResultField: fileSpec.Delete.ResultField,
		}

		if fileSpec.Delete.PreHook != "" {
			factory, ok := opts.DeleteHooks[fileSpec.Delete.PreHook]
			if !ok {
				return CommandSpec{}, fmt.Errorf("unknown delete preHook %q", fileSpec.Delete.PreHook)
			}
			deleteSpec.PreHook = factory(deleteSpec)
		}

		spec.Delete = deleteSpec
	}

	if fileSpec.Primary != nil {
		primary := &PrimarySpec{}

		if len(fileSpec.Primary.Fields) > 0 {
			fields, err := convertFields(fileSpec.Primary.Fields, opts.Parsers)
			if err != nil {
				return CommandSpec{}, err
			}

			primary.Flags = fields
		}

		if fileSpec.Primary.PreHook != "" {
			factory, ok := opts.PrimaryHooks[fileSpec.Primary.PreHook]
			if !ok {
				return CommandSpec{}, fmt.Errorf("unknown primary preHook %q", fileSpec.Primary.PreHook)
			}

			primary.PreHook = factory(primary)
		}

		spec.Primary = primary
	}

	applied := map[string]struct{}{}

	applyLoaderOverride := func(key string) error {
		if key == "" || opts.Overrides == nil {
			return nil
		}

		norm := normalizeOverrideKey(key)
		if _, seen := applied[norm]; seen {
			return nil
		}

		if override, ok := opts.Overrides[key]; ok {
			applied[norm] = struct{}{}
			if err := override(&spec); err != nil {
				return fmt.Errorf("override %q: %w", key, err)
			}
			return nil
		}

		for candidate, override := range opts.Overrides {
			if normalizeOverrideKey(candidate) == norm {
				applied[norm] = struct{}{}
				if err := override(&spec); err != nil {
					return fmt.Errorf("override %q: %w", candidate, err)
				}
				return nil
			}
		}

		return nil
	}

	if err := applyLoaderOverride(spec.Name); err != nil {
		return CommandSpec{}, err
	}
	if err := applyLoaderOverride(spec.Use); err != nil {
		return CommandSpec{}, err
	}

	applyGlobalOverride := func(key string) error {
		if key == "" {
			return nil
		}

		norm := normalizeOverrideKey(key)
		if _, seen := applied[norm]; seen {
			return nil
		}

		if override, ok := lookupOverride(key); ok {
			applied[norm] = struct{}{}
			if err := override(&spec); err != nil {
				return fmt.Errorf("override %q: %w", key, err)
			}
		}

		return nil
	}

	if err := applyGlobalOverride(spec.Name); err != nil {
		return CommandSpec{}, err
	}
	if err := applyGlobalOverride(spec.Use); err != nil {
		return CommandSpec{}, err
	}

	return spec, nil
}

// hydrateListSpec converts the file representation of a list spec into runtime metadata.
func hydrateListSpec(in fileListSpec, opts LoaderOptions) (*ListSpec, error) {
	spec := &ListSpec{
		Method:       in.Method,
		Root:         in.Root,
		FilterMethod: in.FilterMethod,
	}

	if in.Where != nil {
		where, err := hydrateWhereSpec(*in.Where, opts)
		if err != nil {
			return nil, err
		}

		spec.Where = where
	}

	return spec, nil
}

// hydrateWhereSpec converts the file representation of where filters into runtime metadata.
func hydrateWhereSpec(in fileWhereSpec, opts LoaderOptions) (*WhereSpec, error) {
	if in.Type == "" {
		return nil, fmt.Errorf("where type must be provided when filters are defined")
	}

	typ, err := resolveType(opts.TypeResolver, in.Type)
	if err != nil {
		return nil, err
	}

	fields, err := convertFields(in.Fields, opts.Parsers)
	if err != nil {
		return nil, err
	}

	return &WhereSpec{
		Type:   typ,
		Fields: fields,
	}, nil
}

// hydrateMutationSpec converts file create specs into runtime create metadata.
func hydrateMutationSpec(in fileCreateSpec, opts LoaderOptions) (CreateSpec, error) {
	inputType, err := resolveType(opts.TypeResolver, in.InputType)
	if err != nil {
		return CreateSpec{}, err
	}

	fields, err := convertFields(in.Fields, opts.Parsers)
	if err != nil {
		return CreateSpec{}, err
	}

	spec := CreateSpec{
		Method:     in.Method,
		InputType:  inputType,
		Fields:     fields,
		ResultPath: in.ResultPath,
	}

	if in.PreHook != "" {
		factory, ok := opts.CreateHooks[in.PreHook]
		if !ok {
			return CreateSpec{}, fmt.Errorf("unknown create preHook %q", in.PreHook)
		}
		spec.PreHook = factory(&spec)
	}

	return spec, nil
}

// hydrateUpdateSpec converts file update specs into runtime update metadata.
func hydrateUpdateSpec(in fileUpdateSpec, opts LoaderOptions) (UpdateSpec, error) {
	inputType, err := resolveType(opts.TypeResolver, in.InputType)
	if err != nil {
		return UpdateSpec{}, err
	}

	fields, err := convertFields(in.Fields, opts.Parsers)
	if err != nil {
		return UpdateSpec{}, err
	}

	spec := UpdateSpec{
		Method:     in.Method,
		IDFlag:     in.IDFlag,
		InputType:  inputType,
		Fields:     fields,
		ResultPath: in.ResultPath,
	}

	if in.PreHook != "" {
		factory, ok := opts.UpdateHooks[in.PreHook]
		if !ok {
			return UpdateSpec{}, fmt.Errorf("unknown update preHook %q", in.PreHook)
		}
		spec.PreHook = factory(&spec)
	}

	return spec, nil
}

// convertFields normalizes field specs from disk into FieldSpec structs.
func convertFields(in []fileFieldSpec, parsers map[string]ValueParser) ([]FieldSpec, error) {
	fields := make([]FieldSpec, len(in))

	for i, field := range in {
		kind, err := parseValueKind(field.Kind)
		if err != nil {
			return nil, err
		}

		fields[i] = FieldSpec{
			Flag:  field.Flag,
			Kind:  kind,
			Field: field.Field,
		}

		if field.ParserName != "" {
			parser, ok := parsers[field.ParserName]
			if !ok {
				return nil, fmt.Errorf("unknown parser %q", field.ParserName)
			}

			fields[i].Parser = parser
		}
	}

	return fields, nil
}

// parseValueKind maps string representations to ValueKind enums.
func parseValueKind(kind string) (ValueKind, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "string":
		return ValueString, nil
	case "stringslice", "string_slice", "string-slice", "slice":
		return ValueStringSlice, nil
	case "bool", "boolean":
		return ValueBool, nil
	case "duration":
		return ValueDuration, nil
	case "int", "int64", "integer":
		return ValueInt, nil
	default:
		return ValueString, fmt.Errorf("unknown value kind %q", kind)
	}
}

// resolveType resolves a named type using the configured resolver.
func resolveType(resolver TypeResolver, name string) (reflect.Type, error) {
	if name == "" {
		return nil, fmt.Errorf("missing type reference")
	}

	if resolver == nil {
		return nil, fmt.Errorf("no type resolver configured for %q", name)
	}

	return resolver(name)
}
