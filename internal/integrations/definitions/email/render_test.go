package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/email/themes"
)

// TestBrandingButtonColorInRenderedHTML verifies that EmailBranding button colors
// are applied to rendered HTML button elements via CSS inlining. This exercises
// the full pipeline: brandingStyleMap → MergeWith(theme) → GenerateHTML (with
// premailer CSS inlining) → inlined style on <a class="button"> elements
func TestBrandingButtonColorInRenderedHTML(t *testing.T) {
	eb := &generated.EmailBranding{
		ButtonColor:     "#ff5500",
		ButtonTextColor: "#ffffff",
		BackgroundColor: "#f0f0f0",
		TextColor:       "#333333",
		BrandName:       "TestBrand",
	}

	styles := brandingStyleMap(eb)

	content := render.EmailContent{
		Body: render.ContentBody{
			Intros: []string{"Click below to get started."},
			Actions: []render.Action{
				{
					Instructions: "Press the button to confirm",
					Button: render.Button{
						Text: "Confirm",
						Link: "https://example.com/confirm",
					},
				},
			},
			Styles: styles,
		},
	}

	r := render.NewRenderer(
		render.WithTheme(themes.Customer),
		render.WithBranding(brandingFromEB(eb)),
	)

	html, err := r.GenerateHTML(content)
	require.NoError(t, err)

	t.Run("button color inlined on button element", func(t *testing.T) {
		assert.Contains(t, html, "#ff5500",
			"expected branding button color in rendered HTML")
	})

	t.Run("button text color inlined on button element", func(t *testing.T) {
		assert.Contains(t, html, "Confirm")
		// The button text color should appear as an inline style on the button element
		// Find the <a> with class="button" and verify it has the button text color
		idx := strings.Index(html, `class="button"`)
		require.NotEqual(t, -1, idx, "expected an element with class=button")

		// Extract a window around the button element to check its inline styles
		start := idx - 200
		if start < 0 {
			start = 0
		}

		end := idx + 200
		if end > len(html) {
			end = len(html)
		}

		buttonContext := html[start:end]
		assert.Contains(t, buttonContext, "#ffffff",
			"expected button text color in button element styles")
	})

	t.Run("background color applied", func(t *testing.T) {
		assert.Contains(t, html, "#f0f0f0",
			"expected branding background color in rendered HTML")
	})

	t.Run("text color applied", func(t *testing.T) {
		assert.Contains(t, html, "#333333",
			"expected branding text color in rendered HTML")
	})

	t.Run("brand name present", func(t *testing.T) {
		assert.Contains(t, html, "TestBrand",
			"expected brand name in rendered HTML")
	})
}

// TestBrandingStyleMapButtonColors verifies brandingStyleMap produces the
// correct CSS properties for button color fields
func TestBrandingStyleMapButtonColors(t *testing.T) {
	eb := &generated.EmailBranding{
		ButtonColor:     "#aa0000",
		ButtonTextColor: "#eeeeee",
	}

	styles := brandingStyleMap(eb)

	require.Contains(t, styles, ".button")
	assert.Equal(t, "#aa0000", styles[".button"]["background-color"])
	assert.Equal(t, "#aa0000", styles[".button"]["border-color"])
	assert.Equal(t, "#eeeeee", styles[".button"]["color"])
}

// TestBrandingStyleMapNilBranding verifies brandingStyleMap returns nil for
// an empty branding record
func TestBrandingStyleMapNilBranding(t *testing.T) {
	eb := &generated.EmailBranding{}

	styles := brandingStyleMap(eb)

	assert.Nil(t, styles)
}

// TestBrandingStyleMapPartialFields verifies brandingStyleMap only produces
// styles for fields that are set
func TestBrandingStyleMapPartialFields(t *testing.T) {
	eb := &generated.EmailBranding{
		BackgroundColor: "#fafafa",
	}

	styles := brandingStyleMap(eb)

	require.Contains(t, styles, ".email-wrapper")
	assert.Equal(t, "#fafafa", styles[".email-wrapper"]["background-color"])
	assert.NotContains(t, styles, ".button",
		"button styles should not be present when button colors are empty")
	assert.NotContains(t, styles, "p",
		"text styles should not be present when text color is empty")
}

// TestBrandingStyleMapFontFamily verifies font family CSS is applied
func TestBrandingStyleMapFontFamily(t *testing.T) {
	eb := &generated.EmailBranding{
		FontFamily: enums.FontCourier,
	}

	styles := brandingStyleMap(eb)

	require.Contains(t, styles, "body")
	assert.Equal(t, "'Courier New', Courier, monospace", styles["body"]["font-family"])
}

// TestBrandingStyleMapLinkColor verifies link color styles
func TestBrandingStyleMapLinkColor(t *testing.T) {
	eb := &generated.EmailBranding{
		LinkColor: "#0066cc",
	}

	styles := brandingStyleMap(eb)

	require.Contains(t, styles, "a")
	assert.Equal(t, "#0066cc", styles["a"]["color"])
}

// TestBrandingStyleMapTextColorSetsBodyAndP verifies text color sets both body and p styles
func TestBrandingStyleMapTextColorSetsBodyAndP(t *testing.T) {
	eb := &generated.EmailBranding{
		TextColor: "#444444",
	}

	styles := brandingStyleMap(eb)

	require.Contains(t, styles, "p")
	assert.Equal(t, "#444444", styles["p"]["color"])
	require.Contains(t, styles, "body")
	assert.Equal(t, "#444444", styles["body"]["color"])
}

// TestBrandingFromEB_Nil verifies nil branding returns empty struct
func TestBrandingFromEB_Nil(t *testing.T) {
	b := brandingFromEB(nil)

	assert.Empty(t, b.Name)
	assert.Empty(t, b.Logo)
}

// TestBrandingFromEB_BrandNameUsed verifies BrandName is preferred over Name
func TestBrandingFromEB_BrandNameUsed(t *testing.T) {
	eb := &generated.EmailBranding{
		Name:      "Internal Name",
		BrandName: "Public Brand",
	}

	b := brandingFromEB(eb)

	assert.Equal(t, "Public Brand", b.Name)
}

// TestBrandingFromEB_FallbackToName verifies Name is used when BrandName is empty
func TestBrandingFromEB_FallbackToName(t *testing.T) {
	eb := &generated.EmailBranding{
		Name: "Fallback Name",
	}

	b := brandingFromEB(eb)

	assert.Equal(t, "Fallback Name", b.Name)
}

// TestBrandingFromEB_LogoFromRemoteURL verifies LogoRemoteURL is used when set
func TestBrandingFromEB_LogoFromRemoteURL(t *testing.T) {
	logoURL := "https://cdn.example.com/logo.png"
	eb := &generated.EmailBranding{
		BrandName:     "LogoCo",
		LogoRemoteURL: &logoURL,
	}

	b := brandingFromEB(eb)

	assert.Equal(t, "https://cdn.example.com/logo.png", b.Logo)
}

// TestBrandingFromEB_NilLogo verifies no logo when LogoRemoteURL is nil
func TestBrandingFromEB_NilLogo(t *testing.T) {
	eb := &generated.EmailBranding{
		BrandName: "NoLogoCo",
	}

	b := brandingFromEB(eb)

	assert.Empty(t, b.Logo)
}

// TestRenderDBEnvelope_BasicRendering verifies the full rendering pipeline
func TestRenderDBEnvelope_BasicRendering(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate:   "Hello {{ .recipientFirstName }}",
		PreheaderTemplate: "Welcome to {{ .companyName }}",
		BodyTemplate:      "# Hi {{ .recipientFirstName }}\n\nWelcome!",
	}

	data := map[string]any{
		"recipientFirstName": "Alice",
		"companyName":        "TestCo",
	}

	env, err := renderDBEnvelope(emailRecord, data, nil)
	require.NoError(t, err)

	assert.Equal(t, "Hello Alice", env.Subject)
	assert.Equal(t, "Welcome to TestCo", env.Preheader)
	assert.Contains(t, env.HTML, "Alice")
	assert.Contains(t, env.HTML, "Welcome!")
	assert.NotEmpty(t, env.Text)
}

// TestRenderDBEnvelope_WithBranding verifies branding is applied to rendered HTML
func TestRenderDBEnvelope_WithBranding(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate: "Test",
		BodyTemplate:    "Content here",
	}

	eb := &generated.EmailBranding{
		BrandName:       "BrandedCo",
		BackgroundColor: "#abcdef",
	}

	env, err := renderDBEnvelope(emailRecord, map[string]any{}, eb)
	require.NoError(t, err)

	assert.Contains(t, env.HTML, "BrandedCo")
	assert.Contains(t, env.HTML, "#abcdef")
}

// TestRenderDBEnvelope_CustomTextTemplate verifies custom text template is used when present
func TestRenderDBEnvelope_CustomTextTemplate(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate: "Test",
		BodyTemplate:    "# HTML Body",
		TextTemplate:    "Plain text for {{ .recipientFirstName }}",
	}

	data := map[string]any{
		"recipientFirstName": "Bob",
	}

	env, err := renderDBEnvelope(emailRecord, data, nil)
	require.NoError(t, err)

	assert.Equal(t, "Plain text for Bob", env.Text)
}

// TestRenderDBEnvelope_FallbackPlainText verifies generated plain text when TextTemplate is empty
func TestRenderDBEnvelope_FallbackPlainText(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate: "Test",
		BodyTemplate:    "Some content here",
	}

	env, err := renderDBEnvelope(emailRecord, map[string]any{}, nil)
	require.NoError(t, err)

	assert.NotEmpty(t, env.Text)
	assert.Contains(t, env.Text, "Some content here")
}

// TestRenderDBEnvelope_BadSubjectTemplate verifies error on invalid subject template
func TestRenderDBEnvelope_BadSubjectTemplate(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate: "{{ .bad | nonExistentFunc }}",
		BodyTemplate:    "body",
	}

	_, err := renderDBEnvelope(emailRecord, map[string]any{}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTemplateRenderFailed)
}

// TestRenderDBEnvelope_BadBodyTemplate verifies error on invalid body template
func TestRenderDBEnvelope_BadBodyTemplate(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate: "valid subject",
		BodyTemplate:    "{{ .bad | nonExistentFunc }}",
	}

	_, err := renderDBEnvelope(emailRecord, map[string]any{}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTemplateRenderFailed)
}

// TestRenderDBEnvelope_EmptyPreheader verifies empty preheader is handled
func TestRenderDBEnvelope_EmptyPreheader(t *testing.T) {
	emailRecord := &generated.EmailTemplate{
		SubjectTemplate: "Subject",
		BodyTemplate:    "Body",
	}

	env, err := renderDBEnvelope(emailRecord, map[string]any{}, nil)
	require.NoError(t, err)

	assert.Empty(t, env.Preheader)
}
