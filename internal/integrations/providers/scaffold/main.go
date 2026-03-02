// Command new-provider scaffolds a new integration provider package under
// internal/integrations/providers/<name>/. Run via:
//
//	task new-provider NAME=<provider> AUTH=<oauth|apikey|custom>
//
// or directly:
//
//	go run ./internal/integrations/providers/scaffold -name <provider> -auth <oauth|apikey|custom>
package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

const (
	authOAuth  = "oauth"
	authAPIKey = "apikey"
	authCustom = "custom"

	goModule = "github.com/theopenlane/core"
)

// providerData holds the template variables for all rendered files
type providerData struct {
	// Provider is the lowercase provider name used as the Go package name (e.g. "datadog")
	Provider string
	// ProviderPascal is the PascalCase provider name used for type and function identifiers (e.g. "Datadog")
	ProviderPascal string
	// AuthType is one of: oauth, apikey, custom
	AuthType string
	// Module is the Go module path for import statements
	Module string
}

func main() {
	name := flag.String("name", "", "provider name (lowercase, e.g. datadog)")
	auth := flag.String("auth", authOAuth, "auth type: oauth, apikey, or custom")
	flag.Parse()

	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := validateAuthType(*auth); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	data := providerData{
		Provider:       *name,
		ProviderPascal: toPascalCase(*name),
		AuthType:       *auth,
		Module:         goModule,
	}

	targetDir := filepath.Join(repoRoot, "internal", "integrations", "providers", data.Provider)
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Fprintf(os.Stderr, "error: directory %s already exists\n", targetDir)
		os.Exit(1)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "error creating directory:", err)
		os.Exit(1)
	}

	if err := renderFiles(data, targetDir); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// renderFiles renders all template files into targetDir for the given provider data
func renderFiles(data providerData, targetDir string) error {
	files := staticFiles(data)
	for outName, tmplName := range files {
		if err := renderFile(tmplName, outName, targetDir, data); err != nil {
			return fmt.Errorf("rendering %s: %w", outName, err)
		}
	}

	providerTmpl := fmt.Sprintf("templates/provider_%s.go.tmpl", data.AuthType)
	return renderFile(providerTmpl, data.Provider+".go", targetDir, data)
}

// staticFiles returns the output filename -> template filename mapping for files
// that do not vary by auth type
func staticFiles(data providerData) map[string]string {
	return map[string]string{
		"doc.go":                   "templates/doc.go.tmpl",
		"errors.go":                "templates/errors.go.tmpl",
		"client.go":                "templates/client.go.tmpl",
		"operations.go":            "templates/operations.go.tmpl",
		data.Provider + "_test.go": "templates/provider_test.go.tmpl",
	}
}

// renderFile parses and executes a single template, writing output to targetDir/outName
func renderFile(tmplName, outName, targetDir string, data providerData) error {
	tmplContent, err := templateFS.ReadFile(tmplName)
	if err != nil {
		return err
	}

	tmpl, err := template.New(outName).Parse(string(tmplContent))
	if err != nil {
		return err
	}

	outPath := filepath.Join(targetDir, outName)
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// validateAuthType returns an error if authType is not one of the supported values
func validateAuthType(authType string) error {
	switch authType {
	case authOAuth, authAPIKey, authCustom:
		return nil
	default:
		return fmt.Errorf("auth type must be one of: %s, %s, %s (got: %s)", authOAuth, authAPIKey, authCustom, authType)
	}
}

// toPascalCase converts a snake_case or lowercase string to PascalCase.
// Examples: "datadog" -> "Datadog", "linear_app" -> "LinearApp"
func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' })
	var b strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		runes := []rune(part)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// findRepoRoot walks up from the current working directory to locate go.work or go.mod
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.work or go.mod; run from within the repository")
		}
		dir = parent
	}
}
