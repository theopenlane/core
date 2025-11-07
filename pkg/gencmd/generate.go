package gencmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"
	"unicode"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/fatih/camelcase"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

var (
	//go:embed templates/*
	_templates embed.FS

	// schemaCache caches the loaded schema names to avoid repeated file I/O
	schemaCache       []string
	schemaCacheLoaded bool

	// ErrMainGoNotFound is returned when main.go file cannot be found
	ErrMainGoNotFound = errors.New("could not find main.go file")

	packageOnce sync.Once
	loadedPkgs  []*packages.Package
	loadErr     error
)

const (
	templateSuffix = ".tmpl"
)

// cmd data for the template
type cmd struct {
	// Name is the name of the command
	Name string
	// ListOnly is a flag to indicate if the command is list only
	ListOnly bool
	// SpecDriven indicates we are generating spec files
	SpecDriven       bool
	SpecColumns      []specColumn
	SpecDeleteColumn []specColumn
	SpecCreateFields []specField
	SpecUpdateFields []specField
}

type specColumn struct {
	Header    string
	Path      []string
	Formatter string
}

func (c specColumn) PathJSON() string {
	data, err := json.Marshal(c.Path)
	if err != nil {
		return "[]"
	}

	return string(data)
}

type specField struct {
	FlagName      string
	FlagShorthand string
	FlagUsage     string
	FlagRequired  bool
	FlagDefault   string
	Kind          string
	Field         string
	Parser        string
}

var supportedEnumParsers = map[string]bool{
	"programStatus":  true,
	"taskStatus":     true,
	"standardStatus": true,
}

func templateList(spec bool) ([]string, error) {
	if !spec {
		return nil, fmt.Errorf("legacy CLI generation has been removed; rerun with --spec")
	}

	return []string{"doc.tmpl", "register.tmpl", "spec.json.tmpl", "overrides.tmpl"}, nil
}

func buildSpecColumns(cmdName string) []specColumn {
	candidates := candidateQueryFiles(cmdName)
	var content []byte
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			content = data
			break
		}
	}

	columns := []specColumn{{Header: "ID", Path: []string{"id"}}}

	plural := strings.ToLower(toProperPlural(cmdName))
	fieldPaths := extractNodeFieldPaths(string(content), toProperPlural(cmdName))
	if len(fieldPaths) == 0 {
		fieldPaths = extractNodeFieldPaths(string(content), plural)
	}

	if len(fieldPaths) == 0 {
		return columns
	}

	preferred := []string{"displayID", "name", "title", "status", "shortName", "framework", "refCode"}
	formatterMap := map[string]string{
		"lastUsedAt": "timeOrNever",
		"expiresAt":  "timeOrNever",
		"due":        "timeOrNever",
		"scopes":     "joinedStrings",
		"tags":       "joinedStrings",
	}
	ignore := map[string]struct{}{
		"id":         {},
		"ownerID":    {},
		"createdAt":  {},
		"createdBy":  {},
		"updatedAt":  {},
		"updatedBy":  {},
		"owner":      {},
		"edges":      {},
		"node":       {},
		"__typename": {},
	}

	seen := make(map[string]struct{})
	seen["id"] = struct{}{}

	addField := func(path []string) {
		if len(path) == 0 {
			return
		}
		key := strings.Join(path, ".")
		if _, ok := seen[key]; ok {
			return
		}
		for _, segment := range path {
			if _, skip := ignore[segment]; skip {
				return
			}
		}

		last := path[len(path)-1]
		header := humanizePath(path)
		if header == "" {
			header = humanizeHeader(last)
		}
		formatter := formatterMap[last]

		columns = append(columns, specColumn{Header: header, Path: path, Formatter: formatter})
		seen[key] = struct{}{}
	}

	for _, pref := range preferred {
		for _, path := range fieldPaths {
			if len(path) == 0 {
				continue
			}
			if strings.EqualFold(path[len(path)-1], pref) {
				addField(path)
			}
		}
	}

	for _, path := range fieldPaths {
		if len(columns) >= 6 {
			break
		}
		addField(path)
	}

	return columns
}

func candidateQueryFiles(cmdName string) []string {
	base := strings.ToLower(cmdName)
	candidates := []string{
		fmt.Sprintf("internal/graphapi/query/%s.graphql", base),
		fmt.Sprintf("internal/graphapi/query/%s.graphql", strcase.KebabCase(cmdName)),
		fmt.Sprintf("internal/graphapi/query/%s.graphql", strcase.SnakeCase(cmdName)),
	}

	plural := strings.ToLower(toProperPlural(cmdName))
	candidates = append(candidates,
		fmt.Sprintf("internal/graphapi/query/%s.graphql", plural),
	)

	var withDirs []string
	for _, path := range candidates {
		withDirs = append(withDirs, path)
		withDirs = append(withDirs, filepath.Join(filepath.Dir(path), "simple", filepath.Base(path)))
	}

	return withDirs
}

func humanizeHeader(field string) string {
	parts := camelcase.Split(field)
	if len(parts) == 0 {
		return strings.Title(field)
	}
	for i, part := range parts {
		parts[i] = strings.Title(strings.ToLower(part))
	}
	return strings.Join(parts, "")
}

func humanizePath(path []string) string {
	if len(path) == 0 {
		return ""
	}

	parts := make([]string, len(path))
	for i, segment := range path {
		parts[i] = humanizeHeader(segment)
	}

	return strings.Join(parts, " ")
}

func extractNodeFieldPaths(content string, plural string) [][]string {
	lines := strings.Split(content, "\n")
	var paths [][]string

	inNode := false
	depth := 0
	stack := []string{}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lower := strings.ToLower(line)

		if !inNode {
			if strings.HasPrefix(lower, fmt.Sprintf("%s {", strings.ToLower(plural))) {
				continue
			}
			if strings.HasPrefix(lower, "node {") {
				inNode = true
				depth = 1
				stack = stack[:0]
				continue
			}
			continue
		}

		openCount := strings.Count(line, "{")
		closeCount := strings.Count(line, "}")

		if openCount > 0 {
			field := strings.TrimSpace(strings.Split(line, "{")[0])
			field = strings.TrimSuffix(field, "(")
			field = strings.TrimSpace(field)
			field = strings.TrimSuffix(field, ",")

			if field != "" && !strings.HasPrefix(field, "...") {
				stack = append(stack, field)
			}

			depth += openCount
			continue
		}

		if closeCount > 0 {
			for i := 0; i < closeCount && len(stack) > 0; i++ {
				stack = stack[:len(stack)-1]
			}

			depth -= closeCount
			if depth <= 0 {
				inNode = false
				depth = 0
			}
			continue
		}

		if strings.HasPrefix(line, "...") {
			continue
		}

		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		field := strings.TrimSuffix(tokens[0], ",")
		if field == "" {
			continue
		}

		path := append([]string{}, stack...)
		path = append(path, field)
		paths = append(paths, path)
	}

	return paths
}

func buildSpecFields(resource string, typeName string, update bool) []specField {
	strct, tags, err := loadStructInfo(typeName)
	if err != nil || strct == nil {
		return nil
	}

	fields := make([]specField, 0)
	usedNames := map[string]struct{}{}
	usedNames["id"] = struct{}{}
	shorthands := map[string]struct{}{}

	for i := 0; i < strct.NumFields(); i++ {
		field := strct.Field(i)
		if !field.Exported() {
			continue
		}

		tag := tags[i]
		jsonName, omitempty := parseJSONTag(tag)
		if jsonName == "" {
			jsonName = strcase.SnakeCase(field.Name())
		}
		if jsonName == "-" || jsonName == "" {
			continue
		}

		kind, parser, optional, ok := classifyFieldType(field.Type())
		if !ok {
			continue
		}

		if shouldSkipField(jsonName, kind, update) {
			continue
		}

		if _, exists := usedNames[jsonName]; exists {
			continue
		}

		required := !update && !optional && !omitempty
		usage := defaultUsage(jsonName, resource)
		shorthand := pickShorthand(jsonName, shorthands)

		fields = append(fields, specField{
			FlagName:      jsonName,
			FlagShorthand: shorthand,
			FlagUsage:     usage,
			FlagRequired:  required,
			Kind:          kind,
			Field:         field.Name(),
			Parser:        parser,
		})

		usedNames[jsonName] = struct{}{}

		if len(fields) >= 8 {
			break
		}
	}

	return fields
}

func loadStructInfo(typeName string) (*types.Struct, []string, error) {
	pkgs, err := loadPackages()
	if err != nil {
		return nil, nil, err
	}

	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}

		obj := pkg.Types.Scope().Lookup(typeName)
		if obj == nil {
			continue
		}

		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}

		strct, ok := named.Underlying().(*types.Struct)
		if !ok {
			continue
		}

		tags := make([]string, strct.NumFields())
		for i := 0; i < strct.NumFields(); i++ {
			tags[i] = strct.Tag(i)
		}

		return strct, tags, nil
	}

	return nil, nil, fmt.Errorf("type %s not found in openlaneclient", typeName)
}

func loadPackages() ([]*packages.Package, error) {
	packageOnce.Do(func() {
		cfg := &packages.Config{
			Mode:  packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps,
			Tests: false,
			Env:   os.Environ(),
		}
		loadedPkgs, loadErr = packages.Load(cfg, "github.com/theopenlane/core/pkg/openlaneclient")
	})

	return loadedPkgs, loadErr
}

func classifyFieldType(t types.Type) (kind string, parser string, optional bool, ok bool) {
	switch tt := t.(type) {
	case *types.Pointer:
		kind, parser, _, ok = classifyFieldType(tt.Elem())
		return kind, parser, true, ok
	case *types.Slice:
		elemKind, _, _, ok := classifyFieldType(tt.Elem())
		if !ok {
			return "", "", true, false
		}
		if elemKind != "string" {
			return "", "", true, false
		}
		return "stringSlice", "", true, true
	case *types.Basic:
		switch tt.Kind() {
		case types.String:
			return "string", "", false, true
		case types.Bool:
			return "bool", "", false, true
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
			return "int", "", false, true
		}
	case *types.Named:
		pkgPath := ""
		if tt.Obj() != nil && tt.Obj().Pkg() != nil {
			pkgPath = tt.Obj().Pkg().Path()
		}
		name := tt.Obj().Name()

		switch pkgPath {
		case "github.com/theopenlane/core/pkg/enums":
			enumParser := strcase.LowerCamelCase(name)
			if supportedEnumParsers[enumParser] {
				return "string", enumParser, false, true
			}
			return "string", "", false, true
		case "github.com/theopenlane/core/pkg/models":
			if name == "DateTime" {
				return "string", "dateTime", false, true
			}
		case "time":
			if name == "Time" {
				return "string", "dateTime", false, true
			}
		}

		if basic, ok := tt.Underlying().(*types.Basic); ok {
			return classifyFieldType(basic)
		}
	}

	return "", "", false, false
}

func parseJSONTag(tag string) (name string, omitempty bool) {
	if tag == "" {
		return "", false
	}

	parts := strings.Split(reflect.StructTag(tag).Get("json"), ",")
	if len(parts) == 0 {
		return "", false
	}

	name = parts[0]
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty
}

func defaultUsage(field, resource string) string {
	human := strings.ToLower(strings.TrimSpace(humanizePhrase(field)))
	if human == "" {
		human = field
	}
	return fmt.Sprintf("%s of the %s", human, strings.ToLower(resource))
}

func humanizePhrase(value string) string {
	value = strings.ReplaceAll(value, "_", " ")
	parts := camelcase.Split(value)
	if len(parts) == 0 {
		return value
	}
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
	return strings.Join(parts, " ")
}

func pickShorthand(name string, used map[string]struct{}) string {
	for _, r := range name {
		if !unicode.IsLetter(r) {
			continue
		}
		c := strings.ToLower(string(r))
		if _, exists := used[c]; !exists {
			used[c] = struct{}{}
			return c
		}
	}
	return ""
}

func shouldSkipField(name string, kind string, update bool) bool {
	lower := strings.ToLower(name)

	if lower == "id" || lower == "ownerid" || lower == "createdat" || lower == "createdby" || lower == "updatedat" || lower == "updatedby" {
		return true
	}

	if strings.HasPrefix(lower, "add") || strings.HasPrefix(lower, "remove") || strings.HasPrefix(lower, "clear") || strings.HasPrefix(lower, "append") {
		return true
	}

	if strings.Contains(lower, "internal") || strings.Contains(lower, "system") {
		return true
	}

	if strings.HasSuffix(lower, "ids") && kind != "stringSlice" {
		return true
	}

	if strings.Contains(lower, "avatar") || strings.Contains(lower, "remoteurl") {
		return true
	}

	if strings.Contains(lower, "personalorg") {
		return true
	}

	if strings.HasSuffix(lower, "fileid") {
		return true
	}

	if update && lower == "refcode" {
		return true
	}

	return false
}

// Generate generates the cli command files for the given command name
func Generate(cmdName string, cmdDirName string, readOnly bool, spec bool, force bool) error {
	if !spec {
		return fmt.Errorf("legacy CLI generation has been removed; rerun with --spec")
	}

	// trim any leading/trailing spaces
	cmdName = strings.Trim(cmdName, " ")

	// create new directory first
	if err := os.MkdirAll(getNewCmdDirName(cmdDirName, cmdName), os.ModePerm); err != nil {
		return err
	}

	fmt.Println("----> creating cli cmd for:", cmdName)

	templateNames, err := templateList(spec)
	if err != nil {
		return err
	}

	var specColumns []specColumn
	var specDelete []specColumn
	var specCreate []specField
	var specUpdate []specField

	if spec {
		specColumns = buildSpecColumns(cmdName)
		typeName := toProperCamelCase(cmdName)
		specCreate = buildSpecFields(cmdName, fmt.Sprintf("Create%sInput", typeName), false)
		if !readOnly {
			specDelete = []specColumn{
				{Header: "DeletedID", Path: []string{"deletedID"}},
			}
			specUpdate = buildSpecFields(cmdName, fmt.Sprintf("Update%sInput", typeName), true)
		}
	}

	for _, templateName := range templateNames {
		if err := generateCmdFile(cmdName, cmdDirName, templateName, readOnly, spec, specColumns, specDelete, specCreate, specUpdate, force); err != nil {
			return err
		}
	}

	// Update main.go with the new import
	if err := updateMainImports(cmdName); err != nil {
		// Log the error but don't fail the generation
		fmt.Printf("Warning: Could not update main.go imports: %v\n", err)
	}

	return nil
}

// toProperCamelCase intelligently handles compound words for GraphQL type names
func toProperCamelCase(s string) string {
	// If the string is already properly camelcased, split it and rejoin to preserve capitalization
	words := camelcase.Split(s)

	// If camelcase splitting worked (found boundaries), use the split result
	if len(words) > 1 {
		var result strings.Builder

		for _, word := range words {
			if len(word) == 0 {
				continue
			}
			// Capitalize first letter, keep the rest as-is to preserve internal capitalization
			result.WriteString(strings.ToUpper(string(word[0])) + word[1:])
		}

		return result.String()
	}

	// If no split occurred, try to match against actual schema names
	matched := matchAgainstSchemas(s)
	if matched != "" {
		return matched
	}

	// Fallback to standard case conversion
	return strcase.UpperCamelCase(s)
}

// loadSchemaNames loads and caches schema names from the codebase
func loadSchemaNames() {
	if schemaCacheLoaded {
		return
	}

	// Schema path is always relative to project root
	schemaPath := "../../internal/ent/schema"

	graph, err := entc.LoadGraph(schemaPath, &gen.Config{})
	if err == nil && graph != nil {
		schemaCache = make([]string, 0, len(graph.Schemas))
		for _, schema := range graph.Schemas {
			schemaCache = append(schemaCache, schema.Name)
		}
	}

	schemaCacheLoaded = true
}

// matchAgainstSchemas dynamically matches input against actual schema names in the codebase
func matchAgainstSchemas(input string) string {
	// Load schemas if not already cached
	loadSchemaNames()

	if schemaCache == nil {
		// If we couldn't load schemas, return empty to trigger fallback
		return ""
	}

	// Check each cached schema name for a case-insensitive match
	for _, schemaName := range schemaCache {
		if strings.EqualFold(schemaName, input) {
			// Return the actual schema name with proper capitalization
			return schemaName
		}
	}

	// No direct match found
	return ""
}

// toProperPlural combines proper camelCase with pluralization
func toProperPlural(s string) string {
	// First apply proper camelCase, then pluralize
	camelCased := toProperCamelCase(s)
	pluralized := pluralize.NewClient().Plural(camelCased)

	return pluralized
}

// createCmd creates a new template for generating cli commands
func createCmd(name string) (*template.Template, error) {
	// function map for template
	fm := template.FuncMap{
		"ToUpperCamel": toProperCamelCase, // Use our custom function instead
		"ToLowerCamel": strcase.LowerCamelCase,
		"ToLower":      strings.ToLower,
		"ToPlural":     toProperPlural, // Use our custom pluralization
		"ToKebabCase":  strcase.KebabCase,
	}

	files := []string{fmt.Sprintf("templates/%s", name)}
	if name == "spec.json.tmpl" {
		files = append([]string{"templates/spec_columns.tmpl"}, files...)
	}

	tmpl, err := template.New(name).Funcs(fm).ParseFS(_templates, files...)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// generateCmdFile generates the cmd file for the given command name and template name
func generateCmdFile(cmdName, cmdDirName, templateName string, readOnly bool, spec bool, columns []specColumn, deleteColumns []specColumn, createFields []specField, updateFields []specField, force bool) error {
	// create the template
	tmpl, err := createCmd(templateName)
	if err != nil {
		return err
	}

	// get the file name
	filePath := getFileName(cmdDirName, cmdName, templateName)

	// check if schema already exists, skip generation so we don't overwrite manual changes
	if _, err := os.Stat(filePath); err == nil && !force {
		return nil
	}

	// setup the data required for the template
	c := cmd{
		Name:       cmdName,
		ListOnly:   readOnly, // if read only, set the list only flag
		SpecDriven: spec,
		SpecColumns: func() []specColumn {
			if spec && len(columns) > 0 {
				return columns
			}

			return nil
		}(),
		SpecDeleteColumn: func() []specColumn {
			if spec && len(deleteColumns) > 0 {
				return deleteColumns
			}

			return nil
		}(),
		SpecCreateFields: func() []specField {
			if spec && len(createFields) > 0 {
				return createFields
			}

			return nil
		}(),
		SpecUpdateFields: func() []specField {
			if spec && len(updateFields) > 0 {
				return updateFields
			}

			return nil
		}(),
	}

	fmt.Println("----> executing template:", templateName)

	// execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return err
	}

	if err := writeOutput(filePath, buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func writeOutput(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	if strings.EqualFold(filepath.Ext(path), ".json") {
		return os.WriteFile(path, contents, 0o644)
	}

	out, err := imports.Process(path, contents, nil)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0o644)
}

// getFileName returns the file name for the cmd file to be generated
func getFileName(dir, cmdName, templateName string) string {
	// trim the trailing slash, if any
	dir = strings.TrimRight(dir, "/")

	// trim the suffix to get the file name
	fileName := strings.TrimSuffix(templateName, templateSuffix)
	extension := ".go"

	if strings.HasSuffix(fileName, ".json") {
		extension = ""
	}

	fullPath := fmt.Sprintf("%s/%s/%s%s", dir, strings.ToLower(cmdName), strings.ToLower(fileName), extension)

	return filepath.Clean(fullPath)
}

// getNewCmdDirName returns the directory name for the cmd files to be generated
func getNewCmdDirName(dir, cmdName string) string {
	// trim the trailing slash, if any
	dir = strings.TrimRight(dir, "/")

	fullPath := fmt.Sprintf("%s/%s", dir, strings.ToLower(cmdName))

	return filepath.Clean(fullPath)
}

// updateMainImports updates the main.go file with the new import for the generated command
func updateMainImports(cmdName string) error {
	mainPath := "main.go"
	if _, err := os.Stat(mainPath); err != nil {
		mainPath = "cmd/cli/main.go"
		if _, err := os.Stat(mainPath); err != nil {
			return ErrMainGoNotFound
		}
	}

	content, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	fileContent := string(content)
	packageName := strings.ToLower(cmdName)
	newImport := fmt.Sprintf("\t_ \"github.com/theopenlane/core/cmd/cli/cmd/%s\"", packageName)

	// Check if import already exists
	if strings.Contains(fileContent, newImport) {
		fmt.Printf("----> import for %s already exists in main.go\n", packageName)
		return nil
	}

	// Insert the import before the closing import parenthesis.
	if strings.Contains(fileContent, "\n)") {
		fileContent = strings.Replace(fileContent, "\n)", "\n"+newImport+"\n)", 1)
	} else {
		// Fallback: append the import block if the structure is unexpected.
		fileContent += "\nimport (\n" + newImport + "\n)\n"
	}

	if err := os.WriteFile(mainPath, []byte(fileContent), 0600); err != nil { // nolint:mnd
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	fmt.Printf("----> updated main.go with import for %s\n", packageName)

	return nil
}
