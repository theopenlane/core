package themes

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"github.com/theopenlane/newman/render"
)

//go:embed templates/standard.tpl.html templates/standard.tpl.txt templates/standard.css templates/trustcenter.tpl.html templates/trustcenter.css templates/questionnaire.tpl.html templates/questionnaire.css
var themeFS embed.FS

// themeName constants for the three Openlane email themes
const (
	// StandardThemeName is the name for the standard Openlane email theme
	StandardThemeName = "openlane-standard"
	// TrustCenterThemeName is the name for the trust center email theme
	TrustCenterThemeName = "openlane-trustcenter"
	// QuestionnaireThemeName is the name for the questionnaire email theme
	QuestionnaireThemeName = "openlane-questionnaire"
)

// readThemeFile reads a file from the embedded theme filesystem
func readThemeFile(name string) (string, error) {
	data, err := fs.ReadFile(themeFS, name)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %v", render.ErrTemplateReadFailed, name, err)
	}

	return string(data), nil
}

// StandardTheme implements render.Theme for the primary Openlane branded email
// layout with logo header, card-based content, help section, signature, and
// full footer with social links
type StandardTheme struct{}

// Name returns the theme identifier
func (t StandardTheme) Name() string { return StandardThemeName }

// Styles returns the base CSS property map parsed from the embedded standard.css
func (t StandardTheme) Styles() render.StyleMap {
	return standardStyles
}

// HTMLTemplate returns the raw Go HTML template string for the standard layout
func (t StandardTheme) HTMLTemplate() string {
	s, _ := readThemeFile("templates/standard.tpl.html")
	return s
}

// PlainTextTemplate returns the raw Go text template string for the standard layout
func (t StandardTheme) PlainTextTemplate() string {
	s, _ := readThemeFile("templates/standard.tpl.txt")
	return s
}

// ParsedHTMLTemplate returns the pre-parsed HTML template, cached after first parse
func (t StandardTheme) ParsedHTMLTemplate() (*template.Template, error) {
	return standardHTMLOnce()
}

// ParsedPlainTextTemplate returns the pre-parsed plain text template, cached after first parse
func (t StandardTheme) ParsedPlainTextTemplate() (*template.Template, error) {
	return standardTextOnce()
}

// TrustCenterTheme implements render.Theme for Openlane trust center emails
// with a simpler layout, no signature/help section, and a trust-center-specific footer
type TrustCenterTheme struct{}

// Name returns the theme identifier
func (t TrustCenterTheme) Name() string { return TrustCenterThemeName }

// Styles returns the base CSS property map parsed from the embedded trustcenter.css
func (t TrustCenterTheme) Styles() render.StyleMap {
	return trustCenterStyles
}

// HTMLTemplate returns the raw Go HTML template string for the trust center layout
func (t TrustCenterTheme) HTMLTemplate() string {
	s, _ := readThemeFile("templates/trustcenter.tpl.html")
	return s
}

// PlainTextTemplate returns an empty string; trust center emails use HTML-to-text conversion
func (t TrustCenterTheme) PlainTextTemplate() string {
	return t.HTMLTemplate()
}

// ParsedHTMLTemplate returns the pre-parsed HTML template, cached after first parse
func (t TrustCenterTheme) ParsedHTMLTemplate() (*template.Template, error) {
	return trustCenterHTMLOnce()
}

// QuestionnaireTheme implements render.Theme for Openlane questionnaire emails
// with a simple layout and questionnaire-specific footer
type QuestionnaireTheme struct{}

// Name returns the theme identifier
func (t QuestionnaireTheme) Name() string { return QuestionnaireThemeName }

// Styles returns the base CSS property map parsed from the embedded questionnaire.css
func (t QuestionnaireTheme) Styles() render.StyleMap {
	return questionnaireStyles
}

// HTMLTemplate returns the raw Go HTML template string for the questionnaire layout
func (t QuestionnaireTheme) HTMLTemplate() string {
	s, _ := readThemeFile("templates/questionnaire.tpl.html")
	return s
}

// PlainTextTemplate returns an empty string; questionnaire emails use HTML-to-text conversion
func (t QuestionnaireTheme) PlainTextTemplate() string {
	return t.HTMLTemplate()
}

// ParsedHTMLTemplate returns the pre-parsed HTML template, cached after first parse
func (t QuestionnaireTheme) ParsedHTMLTemplate() (*template.Template, error) {
	return questionnaireHTMLOnce()
}

// package-level CSS parsed once at init
var (
	standardStyles      render.StyleMap
	trustCenterStyles   render.StyleMap
	questionnaireStyles render.StyleMap
)

func init() {
	if css, err := readThemeFile("templates/standard.css"); err == nil {
		standardStyles = render.ParseStyleMap(css)
	}

	if css, err := readThemeFile("templates/trustcenter.css"); err == nil {
		trustCenterStyles = render.ParseStyleMap(css)
	}

	if css, err := readThemeFile("templates/questionnaire.css"); err == nil {
		questionnaireStyles = render.ParseStyleMap(css)
	}
}

// templateBase returns a base template pre-loaded with sprig functions and the
// same custom helpers used by the render engine (url, css, safe)
func templateBase() *template.Template {
	return template.New("").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
		"url": func(s string) template.URL {
			return template.URL(s) //nolint:gosec
		},
		"css": func(in any) template.CSS {
			s, ok := in.(string)
			if !ok {
				return ""
			}

			return template.CSS(s) //nolint:gosec
		},
		"safe": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	})
}

// sync.Once wrappers for lazy template parsing

var (
	standardHTMLTmpl *template.Template
	standardHTMLErr  error
	standardHTMLSync sync.Once

	standardTextTmpl *template.Template
	standardTextErr  error
	standardTextSync sync.Once

	trustCenterHTMLTmpl *template.Template
	trustCenterHTMLErr  error
	trustCenterHTMLSync sync.Once

	questionnaireHTMLTmpl *template.Template
	questionnaireHTMLErr  error
	questionnaireHTMLSync sync.Once
)

func standardHTMLOnce() (*template.Template, error) {
	standardHTMLSync.Do(func() {
		raw := StandardTheme{}.HTMLTemplate()
		standardHTMLTmpl, standardHTMLErr = templateBase().Parse(raw)
	})

	return standardHTMLTmpl, standardHTMLErr
}

func standardTextOnce() (*template.Template, error) {
	standardTextSync.Do(func() {
		raw := StandardTheme{}.PlainTextTemplate()
		standardTextTmpl, standardTextErr = templateBase().Parse(raw)
	})

	return standardTextTmpl, standardTextErr
}

func trustCenterHTMLOnce() (*template.Template, error) {
	trustCenterHTMLSync.Do(func() {
		raw := TrustCenterTheme{}.HTMLTemplate()
		trustCenterHTMLTmpl, trustCenterHTMLErr = templateBase().Parse(raw)
	})

	return trustCenterHTMLTmpl, trustCenterHTMLErr
}

func questionnaireHTMLOnce() (*template.Template, error) {
	questionnaireHTMLSync.Do(func() {
		raw := QuestionnaireTheme{}.HTMLTemplate()
		questionnaireHTMLTmpl, questionnaireHTMLErr = templateBase().Parse(raw)
	})

	return questionnaireHTMLTmpl, questionnaireHTMLErr
}

// Interface compliance assertions
var (
	_ render.Theme              = StandardTheme{}
	_ render.ParsedHTMLTheme    = StandardTheme{}
	_ render.ParsedPlainTextTheme = StandardTheme{}
	_ render.Theme              = TrustCenterTheme{}
	_ render.ParsedHTMLTheme    = TrustCenterTheme{}
	_ render.Theme              = QuestionnaireTheme{}
	_ render.ParsedHTMLTheme    = QuestionnaireTheme{}
)
