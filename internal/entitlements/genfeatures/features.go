package genfeatures

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
	"golang.org/x/tools/imports"

	"github.com/theopenlane/core/common/models"
)

const (
	defaultPerm = 0755
)

// ModuleEntry pairs an ent schema name with the module constants returned by its
// Modules() method
type ModuleEntry struct {
	SchemaName string
	Modules    []string
}

// ParseModuleEntries walks through the schema files at schemaPath and returns the
// module requirements declared by each schema's Modules() method
func ParseModuleEntries(schemaPath string) ([]ModuleEntry, error) {
	schemaFiles, err := filepath.Glob(filepath.Join(schemaPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("finding schema files: %w", err)
	}

	var entries []ModuleEntry

	for _, schemaFile := range schemaFiles {
		fileName := filepath.Base(schemaFile)

		if strings.Contains(fileName, "mixin") ||
			strings.Contains(fileName, "default") {
			continue
		}

		schemaName, modules := parseSchemaInfo(schemaFile)
		if schemaName != "" && len(modules) > 0 {
			entries = append(entries, ModuleEntry{
				SchemaName: schemaName,
				Modules:    modules,
			})
		}
	}

	return entries, nil
}

// ParseSchemaModules returns the modules required by each schema, keyed by schema name,
// resolving the constants parsed off the schema into the modules they hold
func ParseSchemaModules(schemaPath string) (map[string][]models.OrgModule, error) {
	entries, err := ParseModuleEntries(schemaPath)
	if err != nil {
		return nil, err
	}

	modules := make(map[string][]models.OrgModule, len(entries))

	for _, entry := range entries {
		for _, constName := range entry.Modules {
			module, err := models.OrgModuleFromConstName(constName)
			if err != nil {
				return nil, err
			}

			modules[entry.SchemaName] = append(modules[entry.SchemaName], module)
		}
	}

	return modules, nil
}

// GenerateModulePerSchema walks through the schema files, parses them
// and generates a static file containing the schema and it's associated modules
func GenerateModulePerSchema(schemaPath, featureMapDir string) error {
	if err := os.MkdirAll(featureMapDir, defaultPerm); err != nil {
		return fmt.Errorf("creating feature map directory: %w", err)
	}

	entries, err := ParseModuleEntries(schemaPath)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"quote": func(s string) string { return fmt.Sprintf("%q", s) },
	}

	tmpl, err := template.New("featuremap").Funcs(funcMap).Parse(featureMapTemplate)
	if err != nil {
		return fmt.Errorf("parsing feature map template: %w", err)
	}

	outputPath := filepath.Join(featureMapDir, "features.go")

	data := struct {
		Entries []ModuleEntry
	}{
		Entries: entries,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing feature map template: %w", err)
	}

	formatted, err := imports.Process(outputPath, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("failed to format file: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating feature map file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(formatted); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	log.Info().Str("file", outputPath).Msg("generated local feature map")

	return nil
}

func parseSchemaInfo(filePath string) (string, []string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Debug().Str("file", filePath).Err(err).Msg("could not read schema file")
		return "", nil
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		log.Debug().Str("file", filePath).Err(err).Msg("could not parse schema file")
		return "", nil
	}

	var (
		schemaName string
		modules    []string
	)

	// retrieve the Name() and Modules() methods

	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
					receiver := ident.Name

					if funcDecl.Name.Name == "Name" && schemaName == "" {
						schemaName = receiver
					}

					if funcDecl.Name.Name == "Modules" && receiver != "" {
						modules = extractModules(funcDecl)
					}
				}
			}
		}

		return true
	})

	return schemaName, modules
}

func extractModules(funcDecl *ast.FuncDecl) []string {
	var modules []string

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if returnStmt, ok := n.(*ast.ReturnStmt); ok {
			if len(returnStmt.Results) > 0 {
				if compositeLit, ok := returnStmt.Results[0].(*ast.CompositeLit); ok {
					for _, elt := range compositeLit.Elts {
						if selectorExpr, ok := elt.(*ast.SelectorExpr); ok {
							if x, ok := selectorExpr.X.(*ast.Ident); ok {
								if x.Name == "models" && strings.HasPrefix(selectorExpr.Sel.Name, "Catalog") {
									modules = append(modules, selectorExpr.Sel.Name)
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	return modules
}

// featureMapTemplate is the template for generating feature map files
var featureMapTemplate = `// code generated by local feature mapping, DO NOT EDIT.
package features

import "github.com/theopenlane/core/common/models"

var FeatureOfType = map[string][]models.OrgModule{
{{- range .Entries }}
	{{ .SchemaName | quote }}: { {{- range $i, $module := .Modules }}{{if $i}}, {{end}}models.{{ $module }}{{- end }} },
{{- end }}
}
`
