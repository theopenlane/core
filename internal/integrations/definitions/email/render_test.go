package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/newman/render"

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
