package emailruntime

import (
	"context"
	htmltemplate "html/template"
	"regexp"
	"strings"
	texttemplate "text/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/logx"
)

var goTemplateBareIdentifierPattern = regexp.MustCompile(`{{\s*([A-Za-z_][A-Za-z0-9_.]*)\s*}}`)

// goTemplateSubmatchCount is the expected number of submatches from goTemplateBareIdentifierPattern:
// index 0 is the full match, index 1 is the captured identifier
const goTemplateSubmatchCount = 2

// renderedTemplateEnvelope contains rendered subject and message bodies
type renderedTemplateEnvelope struct {
	// Subject is the final rendered email subject
	Subject string
	// Preheader is the rendered preheader text
	Preheader string
	// HTML is the rendered HTML body
	HTML string
	// Text is the rendered plain text body
	Text string
}

// renderTemplateEnvelope renders notification/email templates into final message content
func renderTemplateEnvelope(ctx context.Context, client *generated.Client, notificationRecord *generated.NotificationTemplate, emailRecord *generated.EmailTemplate, data map[string]any) (*renderedTemplateEnvelope, error) {
	renderMetadata := DecodeRenderMetadata(emailRecord.Metadata)
	renderMode := renderMetadata.EffectiveRenderMode()

	subjectTemplate := emailRecord.SubjectTemplate
	if subjectTemplate == "" {
		subjectTemplate = notificationRecord.SubjectTemplate
	}

	preheaderTemplate := emailRecord.PreheaderTemplate
	if preheaderTemplate == "" {
		preheaderTemplate = notificationRecord.TitleTemplate
	}

	htmlEntrypoint := renderMetadata.HTMLEntrypoint
	textEntrypoint := renderMetadata.TextEntrypoint
	baseTemplateKey := renderMetadata.BaseTemplateKey

	var baseTemplateContent string

	if renderMode == RenderModeRawHTML && baseTemplateKey != "" {
		baseContent, err := loadBaseTemplateContent(ctx, client, baseTemplateKey)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("base_template_key", baseTemplateKey).Msg("failed loading base template for RAW_HTML assembly")

			return nil, ErrTemplateRenderFailed
		}

		baseTemplateContent = baseContent
	}

	renderedSubject, err := renderTextTemplate("subject", subjectTemplate, "", data)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", emailRecord.Key).Msg("failed rendering subject template")

		return nil, ErrTemplateRenderFailed
	}

	renderedPreheader, err := renderTextTemplate("preheader", preheaderTemplate, "", data)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", emailRecord.Key).Msg("failed rendering preheader template")

		return nil, ErrTemplateRenderFailed
	}

	renderedHTML, err := renderHTMLTemplate(renderMode, emailRecord.Key, emailRecord.BodyTemplate, baseTemplateContent, htmlEntrypoint, data)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", emailRecord.Key).Msg("failed rendering html template")

		return nil, ErrTemplateRenderFailed
	}

	renderedText, err := renderTextTemplate(emailRecord.Key, emailRecord.TextTemplate, textEntrypoint, data)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", emailRecord.Key).Msg("failed rendering text template")

		return nil, ErrTemplateRenderFailed
	}

	if renderedText == "" {
		renderedText = bluemonday.StrictPolicy().Sanitize(renderedHTML)
	}

	if renderedSubject == "" {
		renderedSubject = notificationRecord.Name
	}

	return &renderedTemplateEnvelope{
		Subject:   renderedSubject,
		Preheader: renderedPreheader,
		HTML:      renderedHTML,
		Text:      renderedText,
	}, nil
}

// renderHTMLTemplate renders an HTML payload according to configured mode.
// When baseContent is non-empty and renderMode is RenderModeRawHTML, the base template is
// parsed first and the customer content (which must define named blocks) is parsed into the
// same template set. Execution then begins at the BaseTemplateEntrypoint named template.
// The base template is expected to define a {{block "content" .}}{{end}} injection point
func renderHTMLTemplate(renderMode RenderMode, templateName string, content string, baseContent string, entrypoint string, data map[string]any) (string, error) {
	if content == "" {
		return "", nil
	}

	content = normalizeGoTemplateShorthand(content)

	if renderMode == RenderModeRawHTML && baseContent != "" {
		return renderHTMLWithBase(templateName, baseContent, content, data)
	}

	tmpl, err := htmltemplate.New(templateName).
		Funcs(templateFuncMap()).
		Parse(content)
	if err != nil {
		return "", err
	}

	buffer := &strings.Builder{}

	// RenderModeRawHTML always executes the root template regardless of entrypoint.
	if renderMode != RenderModeRawHTML && entrypoint != "" {
		if err := tmpl.ExecuteTemplate(buffer, entrypoint, data); err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// renderHTMLWithBase assembles a customer body template with a system base template.
// The base template is parsed first, then the customer content (containing {{define}} blocks)
// is added to the same template set. Execution starts at the BaseTemplateEntrypoint.
// Customer templates must use {{define "content"}}...{{end}} to inject content into the base
func renderHTMLWithBase(templateName string, baseContent string, customerContent string, data map[string]any) (string, error) {
	baseContent = normalizeGoTemplateShorthand(baseContent)

	tmpl, err := htmltemplate.New(BaseTemplateEntrypoint).
		Funcs(templateFuncMap()).
		Parse(baseContent)
	if err != nil {
		return "", err
	}

	if _, err = tmpl.New(templateName).Parse(customerContent); err != nil {
		return "", err
	}

	buffer := &strings.Builder{}
	if err := tmpl.ExecuteTemplate(buffer, BaseTemplateEntrypoint, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// renderTextTemplate renders a text/template payload with optional entrypoint
func renderTextTemplate(templateName string, content string, entrypoint string, data map[string]any) (string, error) {
	if content == "" {
		return "", nil
	}

	content = normalizeGoTemplateShorthand(content)

	tmpl, err := texttemplate.New(templateName).
		Funcs(texttemplate.FuncMap(templateFuncMap())).
		Parse(content)
	if err != nil {
		return "", err
	}

	buffer := &strings.Builder{}
	if entrypoint != "" {
		if err := tmpl.ExecuteTemplate(buffer, entrypoint, data); err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// templateFuncMap returns rendering helper functions used by migrated template content
func templateFuncMap() htmltemplate.FuncMap {
	return htmltemplate.FuncMap{
		"ToUpper": strcase.UpperCamelCase,
	}
}

// normalizeGoTemplateShorthand converts bare identifier placeholders into dot-access placeholders.
// Example: {{ FirstName }} -> {{ .FirstName }}
func normalizeGoTemplateShorthand(content string) string {
	if !strings.Contains(content, "{{") {
		return content
	}

	return goTemplateBareIdentifierPattern.ReplaceAllStringFunc(content, func(segment string) string {
		matches := goTemplateBareIdentifierPattern.FindStringSubmatch(segment)
		if len(matches) != goTemplateSubmatchCount {
			return segment
		}

		expr := matches[1]
		if expr == "" || strings.HasPrefix(expr, ".") || isGoTemplateKeyword(expr) {
			return segment
		}

		return "{{ ." + expr + " }}"
	})
}

// renderTimeHTMLScrubPolicy returns a bluemonday policy for sanitizing fully-rendered
// HTML email content. At render time all Go template expressions have been resolved into
// real values, so the policy allows standard http/https/mailto/tel href schemes without
// the Go template expression allowance used at storage time
func renderTimeHTMLScrubPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("style").OnElements("table", "tr", "td", "th", "div", "span", "p", "a", "img")
	p.AllowAttrs("width", "height", "border", "cellpadding", "cellspacing", "align", "valign", "bgcolor").Globally()
	p.AllowElements("table", "thead", "tbody", "tfoot", "tr", "th", "td", "caption", "colgroup", "col")

	return p
}

// renderTimeHTMLSanitize sanitizes fully-rendered HTML email content. This is the
// render-time safety net that catches any XSS injected via template data values
// after Go template substitution has completed
func renderTimeHTMLSanitize(html string) string {
	return renderTimeHTMLScrubPolicy().Sanitize(html)
}

// isGoTemplateKeyword reports whether expr is a go template directive keyword
func isGoTemplateKeyword(expr string) bool {
	switch strings.ToLower(expr) {
	case "if", "else", "end", "range", "with", "template", "define", "block":
		return true
	default:
		return false
	}
}
