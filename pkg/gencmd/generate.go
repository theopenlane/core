//go:build cligen

package gencmd

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/fatih/camelcase"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
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
	// HistoryCmd is a flag to indicate if the command is for history
	HistoryCmd bool
}

var (
	mutationTemplates = []string{"create.tmpl", "update.tmpl", "delete.tmpl"}
)

// Generate generates the cli command files for the given command name
func Generate(cmdName string, cmdDirName string, readOnly bool, force bool) error {
	// trim any leading/trailing spaces
	cmdName = strings.Trim(cmdName, " ")

	// create new directory first
	if err := os.MkdirAll(getNewCmdDirName(cmdDirName, cmdName), os.ModePerm); err != nil {
		return err
	}

	fmt.Println("----> creating cli cmd for:", cmdName)

	templates, err := _templates.ReadDir("templates")
	if err != nil {
		return err
	}

	// generate all the cmd files
	for _, t := range templates {
		// if read only, skip the mutation templates
		if readOnly && slices.Contains(mutationTemplates, t.Name()) {
			continue
		}

		if err := generateCmdFile(cmdName, cmdDirName, t.Name(), readOnly, force); err != nil {
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

	// create schema template
	tmpl, err := template.New(name).Funcs(fm).ParseFS(_templates, fmt.Sprintf("templates/%s", name))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// generateCmdFile generates the cmd file for the given command name and template name
func generateCmdFile(cmdName, cmdDirName, templateName string, readOnly bool, force bool) error {
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

	isHistory := strings.Contains(cmdName, "History")

	// setup the data required for the template
	c := cmd{
		Name:       cmdName,
		ListOnly:   readOnly, // if read only, set the list only flag
		HistoryCmd: isHistory,
	}

	fmt.Println("----> executing template:", templateName)

	// execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		return err
	}

	out, err := imports.Process(filePath, buf.Bytes(), nil)
	if err != nil {
		return err
	}

	// create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	if _, err := file.Write(out); err != nil {
		return err
	}

	return nil
}

// getFileName returns the file name for the cmd file to be generated
func getFileName(dir, cmdName, templateName string) string {
	// trim the trailing slash, if any
	dir = strings.TrimRight(dir, "/")

	// trim the suffix to get the file name
	fileName := strings.TrimSuffix(templateName, templateSuffix)

	fullPath := fmt.Sprintf("%s/%s/%s.go", dir, strings.ToLower(cmdName), strings.ToLower(fileName))

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

	// Insert the import in the right section
	isHistory := strings.HasSuffix(strings.ToLower(cmdName), "history")

	switch {
	case isHistory && strings.Contains(fileContent, "// history commands\n"):
		// Insert after "// history commands" for history commands
		fileContent = strings.Replace(fileContent, "// history commands\n", "// history commands\n"+newImport+"\n", 1)
	case !isHistory && strings.Contains(fileContent, "\n\t// history commands"):
		// Insert before "// history commands" for regular commands
		fileContent = strings.Replace(fileContent, "\n\t// history commands", newImport+"\n\t// history commands", 1)
	default:
		// Fallback: insert before closing import parenthesis
		fileContent = strings.Replace(fileContent, "\n)", "\n"+newImport+"\n)", 1)
	}

	if err := os.WriteFile(mainPath, []byte(fileContent), 0600); err != nil { // nolint:mnd
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	fmt.Printf("----> updated main.go with import for %s\n", packageName)

	return nil
}
