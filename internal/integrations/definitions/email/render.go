package email

import (
	"fmt"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
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

// renderDBEnvelope renders a DB-sourced email template through the theme renderer pipeline.
// Subject, preheader, and body templates are executed against the data map. The rendered body
// is placed into EmailContent.FreeMarkdown and wrapped by the customer theme with branding
// derived from the EmailBranding record
func renderDBEnvelope(emailRecord *generated.EmailTemplate, data map[string]any, eb *generated.EmailBranding) (*renderedEnvelope, error) {
	subject, err := render.ExecuteTextTemplate("subject", emailRecord.SubjectTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("%w: subject: %w", ErrTemplateRenderFailed, err)
	}

	preheader, err := render.ExecuteTextTemplate("preheader", emailRecord.PreheaderTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("%w: preheader: %w", ErrTemplateRenderFailed, err)
	}

	body, err := render.ExecuteTextTemplate("body", emailRecord.BodyTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("%w: body: %w", ErrTemplateRenderFailed, err)
	}

	content := render.EmailContent{
		Body: render.ContentBody{
			FreeMarkdown: render.MarkdownContent(body),
		},
	}

	if eb != nil {
		content.Body.Styles = brandingStyleMap(eb)
	}

	r := render.NewRenderer(
		render.WithTheme(themes.Customer),
		render.WithBranding(brandingFromEB(eb)),
	)

	htmlBody, err := r.GenerateHTML(content)
	if err != nil {
		return nil, fmt.Errorf("%w: html: %w", ErrTemplateRenderFailed, err)
	}

	var textBody string

	switch {
	case emailRecord.TextTemplate != "":
		textBody, err = render.ExecuteTextTemplate("text", emailRecord.TextTemplate, data)
		if err != nil {
			return nil, fmt.Errorf("%w: text: %w", ErrTemplateRenderFailed, err)
		}
	default:
		textBody, err = r.GeneratePlainText(content)
		if err != nil {
			return nil, fmt.Errorf("%w: plaintext: %w", ErrTemplateRenderFailed, err)
		}
	}

	return &renderedEnvelope{
		Subject:   subject,
		Preheader: preheader,
		HTML:      htmlBody,
		Text:      textBody,
	}, nil
}

// brandingFromEB converts an EmailBranding record into a render.Branding for the
// theme renderer. Returns sensible defaults when eb is nil
func brandingFromEB(eb *generated.EmailBranding) render.Branding {
	if eb == nil {
		return render.Branding{}
	}

	var logo string
	if eb.LogoRemoteURL != nil {
		logo = *eb.LogoRemoteURL
	}

	branding := render.Branding{
		Name: eb.BrandName,
		Logo: logo,
	}

	if branding.Name == "" {
		branding.Name = eb.Name
	}

	return branding
}

// brandingStyleMap converts EmailBranding color and font fields into a render.StyleMap
// for merging with the customer theme's base styles
func brandingStyleMap(eb *generated.EmailBranding) render.StyleMap {
	styles := make(render.StyleMap)

	if eb.BackgroundColor != "" {
		styles[".email-wrapper"] = map[string]any{
			"background-color": eb.BackgroundColor,
		}
	}

	if eb.TextColor != "" {
		styles["p"] = map[string]any{
			"color": eb.TextColor,
		}

		styles["body"] = map[string]any{
			"color": eb.TextColor,
		}
	}

	if eb.FontFamily != "" {
		if fontCSS := fontFamilyCSS(eb.FontFamily); fontCSS != "" {
			bodyProps := styles["body"]
			if bodyProps == nil {
				bodyProps = map[string]any{}
			}

			bodyProps["font-family"] = fontCSS
			styles["body"] = bodyProps
		}
	}

	if eb.LinkColor != "" {
		styles["a"] = map[string]any{
			"color": eb.LinkColor,
		}
	}

	buttonProps := map[string]any{}

	if eb.ButtonColor != "" {
		buttonProps["background-color"] = eb.ButtonColor
		buttonProps["border-color"] = eb.ButtonColor
	}

	if eb.ButtonTextColor != "" {
		buttonProps["color"] = eb.ButtonTextColor
	}

	if len(buttonProps) > 0 {
		styles[".button"] = buttonProps
	}

	if len(styles) == 0 {
		return nil
	}

	return styles
}
