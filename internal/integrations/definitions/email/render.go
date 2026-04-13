package email

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/inbucket/html2text"
	"github.com/microcosm-cc/bluemonday"
	"github.com/vanng822/go-premailer/premailer"

	"github.com/theopenlane/core/internal/ent/generated"
)

// renderedEnvelope contains rendered subject and message bodies
type renderedEnvelope struct {
	// Subject is the final rendered email subject
	Subject string
	// Preheader is the rendered preheader text
	Preheader string
	// HTML is the rendered HTML body
	HTML string
	// Text is the rendered plain text body
	Text string
}


// renderDBEnvelope renders a DB-sourced email template using Go template execution.
// Subject and body templates are executed against the data map; the body is optionally
// composed with a system base template when the metadata specifies a base template key.
// When eb is non-nil, a CSS style block derived from the branding colors is applied before
// CSS inlining. Branding template data must be included in the data map by the caller
func renderDBEnvelope(ctx context.Context, db *generated.Client, emailRecord *generated.EmailTemplate, data map[string]any, eb *generated.EmailBranding) (*renderedEnvelope, error) {
	subject, err := executeTextTemplate("subject", emailRecord.SubjectTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("%w: subject: %w", ErrTemplateRenderFailed, err)
	}

	meta := DecodeRenderMetadata(emailRecord.Metadata)

	var htmlBody string

	switch meta.EffectiveRenderMode() {
	case RenderModeRawHTML:
		if meta.BaseTemplateKey != "" {
			baseContent, loadErr := loadBaseTemplateContent(ctx, db, meta.BaseTemplateKey)
			if loadErr != nil {
				return nil, fmt.Errorf("%w: base template: %w", ErrTemplateRenderFailed, loadErr)
			}

			htmlBody, err = composeAndExecute(baseContent, emailRecord.BodyTemplate, data)
		} else {
			htmlBody, err = executeHTMLTemplate("body", emailRecord.BodyTemplate, data)
		}

		if err != nil {
			return nil, fmt.Errorf("%w: body: %w", ErrTemplateRenderFailed, err)
		}

	default:
		return nil, fmt.Errorf("%w: unsupported render mode %q", ErrTemplateRenderFailed, meta.EffectiveRenderMode())
	}

	// Inject branding CSS for automatic color/font application via CSS inlining
	if eb != nil {
		if css := brandingCSS(eb); css != "" {
			htmlBody = injectBrandingCSS(htmlBody, css)
		}
	}

	// CSS inlining is best-effort; use the raw HTML on failure
	if inlined, inlineErr := inlineCSS(htmlBody); inlineErr == nil {
		htmlBody = inlined
	}

	textBody, _ := html2text.FromString(htmlBody, html2text.Options{PrettyTables: true})

	return &renderedEnvelope{Subject: subject, HTML: htmlBody, Text: textBody}, nil
}


// executeTextTemplate parses and executes a text/template string against the data map
func executeTextTemplate(name, tmplStr string, data map[string]any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	tmpl, err := texttemplate.New(name).Funcs(texttemplate.FuncMap(sprig.TxtFuncMap())).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// executeHTMLTemplate parses and executes an html/template string against the data map
func executeHTMLTemplate(name, tmplStr string, data map[string]any) (string, error) {
	if tmplStr == "" {
		return "", nil
	}

	tmpl, err := htmltemplate.New(name).Funcs(htmltemplate.FuncMap(sprig.FuncMap())).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// composeAndExecute combines a base template with a body template and executes the result.
// The base template defines the outer layout with a {{block "content" .}} placeholder;
// the body template fills that block via {{define "content"}}
func composeAndExecute(baseTemplate, bodyTemplate string, data map[string]any) (string, error) {
	tmpl, err := htmltemplate.New(BaseTemplateEntrypoint).
		Funcs(htmltemplate.FuncMap(sprig.FuncMap())).
		Parse(baseTemplate)
	if err != nil {
		return "", err
	}

	contentDef := `{{define "content"}}` + bodyTemplate + `{{end}}`

	tmpl, err = tmpl.Parse(contentDef)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, BaseTemplateEntrypoint, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// injectBrandingCSS prepends a <style> block containing branding CSS rules to the HTML body.
// If a </head> closing tag exists, the block is inserted before it so that the premailer
// CSS inliner can process it; otherwise it is prepended to the entire document
func injectBrandingCSS(html, css string) string {
	styleBlock := "<style>" + css + "</style>"

	if idx := strings.Index(html, "</head>"); idx >= 0 {
		return html[:idx] + styleBlock + html[idx:]
	}

	return styleBlock + html
}

// inlineCSS transforms CSS style blocks into inline style attributes for email client compatibility
func inlineCSS(html string) (string, error) {
	if html == "" {
		return "", nil
	}

	prem, err := premailer.NewPremailerFromString(html, premailer.NewOptions())
	if err != nil {
		return "", err
	}

	return prem.Transform()
}

// emailSanitizePolicy is a bluemonday policy configured for email HTML that preserves
// layout elements and inline styles while stripping scripts and event handlers
var emailSanitizePolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowStyling()
	p.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td", "caption", "center")
	p.AllowAttrs("width", "height", "align", "valign", "bgcolor", "border",
		"cellpadding", "cellspacing", "colspan", "rowspan").Globally()
	p.AllowAttrs("src", "alt", "width", "height").OnElements("img")
	p.AllowURLSchemes("http", "https", "mailto", "cid")

	return p
}()

// renderTimeHTMLSanitize sanitizes HTML for email rendering using a policy that
// preserves email-safe layout elements while stripping dangerous content
func renderTimeHTMLSanitize(html string) string {
	return emailSanitizePolicy.Sanitize(html)
}


