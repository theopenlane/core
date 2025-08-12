//go:build ignore

// See Upstream docs for more details: https://entgo.io/docs/code-gen/#use-entc-as-a-package

package main

import (
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	_ "github.com/jackc/pgx/v5"
	"gocloud.dev/secrets"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/genhooks"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/core/pkg/windmill"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/internal/genhelpers"
)

const (
	graphDir            = "./internal/graphapi/"
	graphSchemaDir      = graphDir + "schema/"
	graphQueryDir       = graphDir + "query/"
	graphSimpleQueryDir = graphDir + "query/simple/"
	schemaPath          = "./internal/ent/schema"
	templateDir         = "./internal/ent/generate/templates/ent"
	featureMapDir       = "./internal/entitlements/features/"
)

func main() {
	// setup logging
	genhelpers.SetupLogging()

	// change to the root of the repo so that the config hierarchy is correct
	genhelpers.ChangeToRootDir("../../../")

	if err := os.Mkdir("schema", 0755); err != nil && !os.IsExist(err) {
		log.Fatal().Err(err).Msg("creating schema directory")
	}

	// generate the history schemas
	historyExt, entfgaExt := preRun()

	xExt, err := entx.NewExtension(entx.WithJSONScalar())
	if err != nil {
		log.Fatal().Err(err).Msg("creating entx extension")
	}

	gqlExt, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath(graphSchemaDir+"ent.graphql"),
		entgql.WithConfigPath(graphDir+"/generate/.gqlgen.yml"),
		entgql.WithWhereInputs(true),
		entgql.WithSchemaHook(xExt.GQLSchemaHooks()...),
		WithGqlWithTemplates(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("creating entgql extension")
	}

	accessMapExt := accessmap.New(
		accessmap.WithSchemaPath(schemaPath),
	)

	log.Info().Msg("running ent codegen with extensions")

	if err := entc.Generate(schemaPath, &gen.Config{
		Target: "./internal/ent/generated",
		Hooks: []gen.Hook{
			genhooks.GenSchema(graphSchemaDir),
			genhooks.GenQuery(graphQueryDir),
			genhooks.GenQuery(graphSimpleQueryDir),
			genhooks.GenSearchSchema(graphSchemaDir, graphQueryDir),
			accessMapExt.Hook(),
		},
		Package: "github.com/theopenlane/core/internal/ent/generated",
		Features: []gen.Feature{
			gen.FeatureVersionedMigration,
			gen.FeaturePrivacy,
			gen.FeatureEntQL,
			gen.FeatureNamedEdges,
			gen.FeatureSchemaConfig,
			gen.FeatureIntercept,
			gen.FeatureModifier,
			// this is disabled because it is not compatible with the entcache driver
			// gen.FeatureExecQuery,
		},
	},
		entc.Dependency(
			entc.DependencyName("EntConfig"),
			entc.DependencyType(&entconfig.Config{}),
		),
		entc.Dependency(
			entc.DependencyName("Secrets"),
			entc.DependencyType(&secrets.Keeper{}),
		),
		entc.Dependency(
			entc.DependencyName("Authz"),
			entc.DependencyType(fgax.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("TokenManager"),
			entc.DependencyType(&tokens.TokenManager{}),
		),
		entc.Dependency(
			entc.DependencyName("SessionConfig"),
			entc.DependencyType(&sessions.SessionConfig{}),
		),
		entc.Dependency(
			entc.DependencyName("Emailer"),
			entc.DependencyType(&emailtemplates.Config{}),
		),
		entc.Dependency(
			entc.DependencyName("TOTP"),
			entc.DependencyType(&totp.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("EntitlementManager"),
			entc.DependencyType(&entitlements.StripeClient{}),
		),
		entc.Dependency(
			entc.DependencyName("ObjectManager"),
			entc.DependencyType(&objects.Objects{}),
		),
		entc.Dependency(
			entc.DependencyName("Summarizer"),
			entc.DependencyType(&summarizer.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("Windmill"),
			entc.DependencyType(&windmill.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("PondPool"),
			entc.DependencyType(&soiree.PondPool{}),
		),
		entc.TemplateDir(templateDir),
		entc.Extensions(
			gqlExt,
			historyExt,
			entfgaExt,
		)); err != nil {
		log.Fatal().Err(err).Msg("running ent codegen")
	}
}

// preRun runs before the ent codegen to generate the history schemas and authz checks
// and returns the history and fga extensions to be used in the ent codegen
func preRun() (*history.Extension, *entfga.AuthzExtension) {
	// generate the history schemas
	log.Info().Msg("pre-run: generating the history schemas")

	historyExt := history.New(
		history.WithAuditing(),
		history.WithImmutableFields(),
		history.WithHistoryTimeIndex(),
		history.WithNillableFields(),
		history.WithGQLQuery(),
		history.WithAuthzPolicy(),
		history.WithSchemaPath(schemaPath),
		history.WithFirstRun(true),
		history.WithAllowedRelation("audit_log_viewer"),
		history.WithUpdatedByFromSchema(history.ValueTypeString, false),
	)

	if err := historyExt.GenerateSchemas(); err != nil {
		log.Fatal().Err(err).Msg("generating history schema")
	}

	log.Info().Msg("pre-run: generating the authz checks")

	// initialize the entfga extension
	entfgaExt := entfga.New(
		entfga.WithSoftDeletes(),
		entfga.WithSchemaPath(schemaPath),
	)

	// generate authz checks, this needs to happen before we regenerate the schema
	// because the authz checks are generated based on the schema
	if err := entfgaExt.GenerateAuthzChecks(); err != nil {
		log.Fatal().Err(err).Msg("generating authz checks")
	}

	log.Info().Msg("pre-run: generating the history schemas with authz checks")

	// run again with policy
	historyExt.SetFirstRun(false)

	// generate the updated history schemas with authz checks
	if err := historyExt.GenerateSchemas(); err != nil {
		log.Fatal().Err(err).Msg("generating history schema")
	}

	log.Info().Msg("pre-run: generating exportable validation and enum")

	// generate exportable schemas validation using existing entx method
	exportableGen := entx.NewExportableGenerator(schemaPath, "internal/ent/hooks").
		WithPackage("hooks")

	if err := exportableGen.Generate(); err != nil {
		log.Fatal().Err(err).Msg("generating exportable validation")
	}

	// extract exportable schemas and generate ExportType enum here
	exportableSchemas, err := extractExportableSchemasFromGenerated()
	if err != nil {
		log.Fatal().Err(err).Msg("extracting exportable schemas")
	}

	if err := generateExportTypeEnum(exportableSchemas); err != nil {
		log.Fatal().Err(err).Msg("generating ExportType enum")
	}

	log.Info().Msg("pre-run: generating feature modules")

	if err := generateModulePerSchema(); err != nil {
		log.Fatal().Err(err).Msg("generating feature modules")
	}

	return historyExt, entfgaExt
}

// WithGqlWithTemplates is a schema hook to replace entgql default template used for pagination
// The only change to the template is the function used to get the totalCount field uses
// CountIDs(ctx) instead of `Count(ctx)`. The rest is a direct copy of the default template from:
// https://github.com/ent/contrib/tree/master/entgql/template
func WithGqlWithTemplates() entgql.ExtensionOption {
	paginationTmpl := gen.MustParse(gen.NewTemplate("node").
		Funcs(entgql.TemplateFuncs).ParseFS(_entqlTemplates, "templates/entgql/gql_where.tmpl", "templates/entgql/pagination.tmpl"))
	return entgql.WithTemplates(append(entgql.AllTemplates, paginationTmpl)...)
}

// extractExportableSchemasFromGenerated reads the generated exportable validation file
// and extracts the schema names from the ExportableSchemas map
func extractExportableSchemasFromGenerated() ([]string, error) {
	filePath := "internal/ent/hooks/exportable_generated.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading exportable generated file: %w", err)
	}

	var schemas []string
	lines := strings.Split(string(content), "\n")
	inMap := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "var ExportableSchemas = map[string]bool{") {
			inMap = true
			continue
		}
		if inMap && line == "}" {
			break
		}
		if inMap && strings.Contains(line, ":") {
			// extract schema name from lines like: "control": true,
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				schemaName := strings.Trim(strings.TrimSpace(parts[0]), `"`)
				if schemaName != "" {
					schemas = append(schemas, schemaName)
				}
			}
		}
	}

	return schemas, nil
}

func generateExportTypeEnum(schemas []string) error {
	if len(schemas) == 0 {
		return nil
	}

	var enumValues []string
	for _, schema := range schemas {
		enumValues = append(enumValues, strings.ToUpper(schema))
	}

	log.Info().Str("values", strings.Join(enumValues, ",")).Msg("generating ExportType enum")

	return generateEnum("ExportType", enumValues)
}

func generateEnum(name string, values []string) error {
	lowerToSentence := func(s string) string {
		s = strings.ReplaceAll(s, "_", " ")
		s = strings.ToLower(s)
		return s
	}

	funcMap := template.FuncMap{
		"ToCamel":         strcase.UpperCamelCase,
		"ToUpper":         strings.ToUpper,
		"lowerToSentence": lowerToSentence,
	}

	tmpl, err := template.New("enum").Funcs(funcMap).Parse(enumTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	outputDir := "pkg/enums"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	outputFile := strcase.SnakeCase(strings.ToLower(name)) + ".go"
	outputPath := filepath.Join(outputDir, outputFile)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	seen := map[string]struct{}{}
	uniqueValues := []string{}

	for _, v := range values {
		val := strings.ToUpper(v)
		if _, exists := seen[val]; !exists {
			seen[val] = struct{}{}
			uniqueValues = append(uniqueValues, val)
		}
	}

	data := struct {
		Name   string
		Values []string
	}{
		Name:   name,
		Values: uniqueValues,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	log.Info().Str("file", outputPath).Msg("generated enum file")
	return nil
}

// generateModulePerSchema walks through the schema files, parses them
// and generates a static file containing the schema and it's associated modules
func generateModulePerSchema() error {
	if err := os.MkdirAll(featureMapDir, 0755); err != nil {
		return fmt.Errorf("creating feature map directory: %w", err)
	}

	schemaFiles, err := filepath.Glob(filepath.Join(schemaPath, "*.go"))
	if err != nil {
		return fmt.Errorf("finding schema files: %w", err)
	}

	type moduleEntry struct {
		SchemaName string
		Modules    []string
	}

	var entries []moduleEntry

	for _, schemaFile := range schemaFiles {
		fileName := filepath.Base(schemaFile)

		if strings.Contains(fileName, "history") ||
			strings.Contains(fileName, "mixin") ||
			strings.Contains(fileName, "default") {
			continue
		}

		schemaName, modules := parseSchemaInfo(schemaFile)
		if schemaName != "" && len(modules) > 0 {
			entries = append(entries, moduleEntry{
				SchemaName: strcase.UpperCamelCase(schemaName),
				Modules:    modules,
			})
		}
	}

	funcMap := template.FuncMap{
		"quote": func(s string) string { return fmt.Sprintf("%q", s) },
	}

	tmpl, err := template.New("featuremap").Funcs(funcMap).Parse(featureMapTemplate)
	if err != nil {
		return fmt.Errorf("parsing feature map template: %w", err)
	}

	outputPath := filepath.Join(featureMapDir, "features.go")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating feature map file: %w", err)
	}

	defer file.Close()

	data := struct {
		Entries []moduleEntry
	}{
		Entries: entries,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("executing feature map template: %w", err)
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

	var schemaName string
	var modules []string

	// retrieve the Name() and Modules() methods
	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
					receiverType := ident.Name

					if funcDecl.Name.Name == "Name" && schemaName == "" {
						schemaName = extractNameFromMethod(funcDecl, file)
					}

					if funcDecl.Name.Name == "Modules" && receiverType != "" {
						modules = extractModules(funcDecl)
					}
				}
			}
		}
		return true
	})

	return schemaName, modules
}

// extractNameFromMethod extracts the schema name from the Name() method
func extractNameFromMethod(funcDecl *ast.FuncDecl, file *ast.File) string {
	var name string

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if returnStmt, ok := n.(*ast.ReturnStmt); ok {
			if len(returnStmt.Results) > 0 {
				switch result := returnStmt.Results[0].(type) {
				case *ast.BasicLit:
					name = strings.Trim(result.Value, `"`)
				case *ast.Ident:
					name = retrieveConstantValue(result.Name, file)
				}
			}
		}
		return name == ""
	})

	return name
}

// retrieveConstantValue resolves a constant value from the file
func retrieveConstantValue(constName string, file *ast.File) string {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for i, name := range valueSpec.Names {
						if name.Name == constName && i < len(valueSpec.Values) {
							if basicLit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
								return strings.Trim(basicLit.Value, `"`)
							}
						}
					}
				}
			}
		}
	}
	return ""
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

var (
	//go:embed templates/entgql/*.tmpl
	_entqlTemplates embed.FS

	// enumTemplate is the template for generating enum files
	enumTemplate = `package enums

import (
	"fmt"
	"io"
	"strings"
)

// {{ .Name }} is a custom type representing the various states of {{ .Name | ToCamel }}.
type {{ .Name }} string

var (
{{- range .Values }}
	// {{ $.Name }}{{ . | ToCamel }} indicates the {{ lowerToSentence . }}.
	{{ $.Name }}{{ . | ToCamel }} {{ $.Name }} = "{{ . }}"
{{- end }}
	// {{ $.Name }}Invalid is used when an unknown or unsupported value is provided.
	{{ $.Name }}Invalid {{ $.Name }} = "{{ .Name | ToUpper }}_INVALID"
)

// Values returns a slice of strings representing all valid {{ .Name }} values.
func ({{ .Name }}) Values() []string {
	return []string{
	{{- range .Values }}
		string({{ $.Name }}{{ . | ToCamel }}),
	{{- end }}
	}
}

// String returns the string representation of the {{ .Name }} value.
func (r {{ .Name }}) String() string {
	return string(r)
}

// To{{ .Name }} converts a string to its corresponding {{ .Name }} enum value.
func To{{ .Name }}(r string) *{{ .Name }} {
	switch strings.ToUpper(r) {
	{{- range .Values }}
	case {{ $.Name }}{{ . | ToCamel }}.String():
		return &{{ $.Name }}{{ . | ToCamel }}
	{{- end }}
	default:
		return &{{ $.Name }}Invalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r {{ .Name }}) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(` + "`" + `"` + "`" + ` + r.String() + ` + "`" + `"` + "`" + `))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *{{ .Name }}) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for {{ .Name }}, got: %T", v) //nolint:err113
	}

	*r = {{ .Name }}(str)

	return nil
}
`

	// featureMapTemplate is the template for generating feature map files
	featureMapTemplate = `// code generated by local feature mapping, DO NOT EDIT.
package ent

import "github.com/theopenlane/core/pkg/models"

var FeatureOfType = map[string][]models.OrgModule{
{{- range .Entries }}
	{{ .SchemaName | quote }}: { {{- range $i, $module := .Modules }}{{if $i}}, {{end}}models.{{ $module }}{{- end }} },
{{- end }}
}
`
)
