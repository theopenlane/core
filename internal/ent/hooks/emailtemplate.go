package hooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// tmplURLExpr matches Go template expressions of the form {{ .URLS.<Name> }}.
// Only .URLS.* references are matched; arbitrary template expressions are left
// for the sanitizer to handle normally (which strips them from href attributes).
var tmplURLExpr = regexp.MustCompile(`\{\{\s*\.URLS\.[A-Za-z][A-Za-z0-9_]*\s*\}\}`)

// tmplURLPlaceholderBase is the https base used to temporarily stand in for template
// URL expressions during sanitization. The .invalid TLD is reserved by RFC 2606
// and is not resolvable, making collisions with real content extremely unlikely.
const tmplURLPlaceholderBase = "https://templ-url-placeholder.invalid/"

// EmailTemplateSanitizePolicy returns a bluemonday policy suitable for storing
// HTML email template source. It extends the UGC policy (which already covers tables,
// images, links, and standard elements) with inline style support required for HTML
// email layout. Go template URL expressions ({{ .URLS.* }}) in href and src attributes
// are handled by SanitizeBodyHTML, which preprocesses them before passing content to this policy.
func EmailTemplateSanitizePolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("style").OnElements("table", "tr", "td", "th", "div", "span", "p", "a", "img")

	return p
}

// SanitizeBodyHTML sanitizes HTML email template body content using the provided policy
// while preserving Go template URL expressions ({{ .URLS.<Name> }}) in attribute values.
// Template expressions matching {{ .URLS.<Name> }} are replaced with indexed https placeholder
// URLs before sanitization, then restored afterward. This is necessary because bluemonday's
// URL scheme validation rejects non-URL strings regardless of any custom regex policy.
// All other template expressions in href/src attributes are subject to normal policy enforcement
// and will be stripped by bluemonday's URL scheme check.
func SanitizeBodyHTML(p *bluemonday.Policy, content string) string {
	var originals []string

	processed := tmplURLExpr.ReplaceAllStringFunc(content, func(match string) string {
		placeholder := fmt.Sprintf("%s%d", tmplURLPlaceholderBase, len(originals))
		originals = append(originals, match)

		return placeholder
	})

	sanitized := p.Sanitize(processed)

	for i, original := range originals {
		placeholder := fmt.Sprintf("%s%d", tmplURLPlaceholderBase, i)
		sanitized = strings.Replace(sanitized, placeholder, original, 1)
	}

	return sanitized
}

// HookEmailTemplateSanitize sanitizes HTML template content fields on create and update
// for non-system-owned email templates. System-owned templates are loaded via harmonize
// and are trusted; user-composed templates are sanitized to prevent stored XSS.
// Body content preserves {{ .URLS.<Name> }} template expressions; subject, preheader,
// and text fields are stripped of all HTML.
func HookEmailTemplateSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EmailTemplateFunc(func(ctx context.Context, m *generated.EmailTemplateMutation) (generated.Value, error) {
			// system-owned templates are loaded via harmonize and are pre-trusted
			if owned, exists := m.SystemOwned(); exists && owned {
				return next.Mutate(ctx, m)
			}

			p := EmailTemplateSanitizePolicy()

			if v, exists := m.SubjectTemplate(); exists {
				m.SetSubjectTemplate(bluemonday.StrictPolicy().Sanitize(v))
			}

			if v, exists := m.PreheaderTemplate(); exists {
				m.SetPreheaderTemplate(bluemonday.StrictPolicy().Sanitize(v))
			}

			if v, exists := m.BodyTemplate(); exists {
				m.SetBodyTemplate(SanitizeBodyHTML(p, v))
			}

			if v, exists := m.TextTemplate(); exists {
				m.SetTextTemplate(bluemonday.StrictPolicy().Sanitize(v))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}
