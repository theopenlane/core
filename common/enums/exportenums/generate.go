package exportenums

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/entx"
)

// ExtensionOption is a function that modifies the Extension configuration.
type ExtensionOption = func(*Extension)

// Config is the configuration for the export enums extension.
type Config struct {
	OutputDir   string
	PackageName string
}

// New creates a new export enums extension
func New(opts ...ExtensionOption) *Extension {
	extension := &Extension{
		// Set configuration defaults that can get overridden with ExtensionOption
		config: &Config{
			OutputDir:   "./common/enums",
			PackageName: "enums",
		},
	}

	for _, opt := range opts {
		opt(extension)
	}

	return extension
}

// Extension implements entc.Extension
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// WithGeneratedDir allows you to set an alternative output directory
// Defaults to "./common/enums"
func WithGeneratedDir(outputDir string) ExtensionOption {
	return func(h *Extension) {
		h.config.OutputDir = outputDir
	}
}

// WithPackageName allows you to set an alternative package name for the generated file
// Defaults to "enums"
func WithPackageName(packageName string) ExtensionOption {
	return func(h *Extension) {
		h.config.PackageName = packageName
	}
}

// Hooks satisfies the entc.Extension interface
func (e Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		e.Hook(),
	}
}

// checkHasExportAnnotation checks if the type has the Export annotation
func checkHasExportAnnotation(node *gen.Type) bool {
	exportAnt := entx.Exportable{}

	if ant, ok := node.Annotations[exportAnt.Name()]; ok {
		if err := exportAnt.Decode(ant); err != nil {
			return false
		}

		return true
	}

	return false
}

func (e Extension) Hook() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			if err := next.Generate(g); err != nil {
				return err
			}

			// set the package name for the generated file
			g.Package = e.config.PackageName

			name := "ExportType"

			lowerToSentence := func(s string) string {
				s = strings.ReplaceAll(s, "_", " ")
				s = strings.ToLower(s)

				return s
			}

			untitle := func(s string) string {
				if s == "" {
					return s
				}

				return strings.ToLower(s[:1]) + s[1:]
			}

			funcMap := template.FuncMap{
				"ToUpper":         strings.ToUpper,
				"ToCamel":         strcase.UpperCamelCase,
				"ToSnake":         strcase.UpperSnakeCase,
				"lowerToSentence": lowerToSentence,
				"untitle":         untitle,
			}

			// loop through all nodes and generate a list of schemas with the Export annotation
			schemas := []string{}

			for _, node := range g.Nodes {
				if checkHasExportAnnotation(node) {
					schemas = append(schemas, node.Name)
				}
			}

			tmpl, err := template.New("enum").Funcs(funcMap).Parse(enumTemplate)
			if err != nil {
				return fmt.Errorf("parsing template: %w", err)
			}

			outputDir := e.config.OutputDir
			if err := os.MkdirAll(outputDir, 0755); err != nil { //nolint:mnd
				return fmt.Errorf("creating output directory: %w", err)
			}

			outputFile := strcase.SnakeCase(strings.ToLower(name)) + ".go"
			outputPath := filepath.Join(outputDir, outputFile)

			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer file.Close()

			data := struct {
				Name   string
				Values []string
			}{
				Name:   name,
				Values: schemas,
			}

			if err := tmpl.Execute(file, data); err != nil {
				return fmt.Errorf("executing template: %w", err)
			}

			log.Info().Str("file", outputPath).Msg("generated enum file")

			return nil
		})
	}
}

var (

	// enumTemplate is the template for generating enum files
	enumTemplate = `package enums

import "io"

// {{ .Name }} is a custom type representing the various states of {{ .Name | ToCamel }}.
type {{ .Name }} string

var (
{{- range .Values }}
	// {{ $.Name }}{{ . | ToCamel }} indicates the {{ lowerToSentence . }}.
	{{ $.Name }}{{ . | ToCamel }} {{ $.Name }} = "{{ . | ToSnake }}"
{{- end }}
	// {{ $.Name }}Invalid is used when an unknown or unsupported value is provided.
	{{ $.Name }}Invalid {{ $.Name }} = "{{ .Name | ToUpper }}_INVALID"
)

var {{ .Name | untitle }}Values = []{{ .Name }}{
{{- range .Values }}
	{{ $.Name }}{{ . | ToCamel }},
{{- end }}
}

// Values returns a slice of strings representing all valid {{ .Name }} values.
func ({{ .Name }}) Values() []string { return stringValues({{ .Name | untitle }}Values) }

// String returns the string representation of the {{ .Name }} value.
func (r {{ .Name }}) String() string { return string(r) }

// To{{ .Name }} converts a string to its corresponding {{ .Name }} enum value.
func To{{ .Name }}(r string) *{{ .Name }} { return parse(r, {{ .Name | untitle }}Values, &{{ .Name }}Invalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r {{ .Name }}) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *{{ .Name }}) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
`
)
