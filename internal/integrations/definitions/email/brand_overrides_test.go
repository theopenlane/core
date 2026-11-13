package email

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestBrandOverridesApplySetsValidValues(t *testing.T) {
	base := RuntimeEmailConfig{LogoURL: "https://openlane.io/logo.png", ButtonColor: "#000000", ButtonTextColor: "#ffffff"}

	out := BrandOverrides{
		LogoURL:         "https://acme.com/logo.png",
		PrimaryColor:    "#112233",
		TextColor:       "#abc",
		ButtonColor:     "#ffffff",
		ButtonTextColor: "#000000",
	}.apply(base)

	assert.Check(t, is.Equal("https://acme.com/logo.png", out.LogoURL))
	assert.Check(t, is.Equal("#112233", out.HeadingColor))
	assert.Check(t, is.Equal("#abc", out.TextColor))
	assert.Check(t, is.Equal("#ffffff", out.ButtonColor))
	assert.Check(t, is.Equal("#000000", out.ButtonTextColor))
}

func TestBrandOverridesApplyKeepsConfigWhenUnset(t *testing.T) {
	base := RuntimeEmailConfig{LogoURL: "https://openlane.io/logo.png", HeaderLogoURL: "https://openlane.io/icon.png", HeadingColor: "#0f3d3a"}

	out := BrandOverrides{}.apply(base)

	assert.Check(t, is.Equal(base.LogoURL, out.LogoURL))
	assert.Check(t, is.Equal(base.HeaderLogoURL, out.HeaderLogoURL))
	assert.Check(t, is.Equal(base.HeadingColor, out.HeadingColor))
}

func TestBrandOverridesApplyIgnoresInvalidColors(t *testing.T) {
	base := RuntimeEmailConfig{HeadingColor: "#0f3d3a", CardBackgroundColor: "#ffffff"}

	out := BrandOverrides{PrimaryColor: "red", CardBackgroundColor: "#12345g"}.apply(base)

	assert.Check(t, is.Equal("#0f3d3a", out.HeadingColor))
	assert.Check(t, is.Equal("#ffffff", out.CardBackgroundColor))
}
