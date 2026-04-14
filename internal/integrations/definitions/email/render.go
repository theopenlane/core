package email

import (
	"fmt"

	"github.com/theopenlane/newman/render"

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
// Subject and body templates are executed against the data map.
// When eb is non-nil, a CSS style block derived from the branding colors is applied before
// CSS inlining. Branding template data must be included in the data map by the caller
func renderDBEnvelope(emailRecord *generated.EmailTemplate, data map[string]any, eb *generated.EmailBranding) (*renderedEnvelope, error) {
	subject, err := render.ExecuteTextTemplate("subject", emailRecord.SubjectTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("%w: subject: %w", ErrTemplateRenderFailed, err)
	}

	htmlBody, err := render.ExecuteHTMLTemplate("body", emailRecord.BodyTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("%w: body: %w", ErrTemplateRenderFailed, err)
	}

	// Inject branding CSS for automatic color/font application via CSS inlining
	if eb != nil {
		if css := brandingCSS(eb); css != "" {
			htmlBody = render.InjectStyleBlock(htmlBody, css)
		}
	}

	// CSS inlining is best-effort; use the raw HTML on failure
	if inlined, inlineErr := render.InlineCSS(htmlBody); inlineErr == nil {
		htmlBody = inlined
	}

	textBody, _ := render.HTMLToPlainText(htmlBody)

	return &renderedEnvelope{Subject: subject, HTML: htmlBody, Text: textBody}, nil
}
