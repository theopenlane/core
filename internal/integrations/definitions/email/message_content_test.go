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

func TestBrandColorKeysFromSchema(t *testing.T) {
	assert.Check(t, is.DeepEqual([]string{
		"primaryColor",
		"bodyBackgroundColor",
		"cardBackgroundColor",
		"heroBackgroundColor",
		"buttonColor",
		"buttonTextColor",
		"textColor",
		"footerTextColor",
	}, brandColorKeys))
}

func TestValidateBrandColors(t *testing.T) {
	assert.Check(t, ValidateBrandColors(map[string]any{"buttonColor": "#000000", "title": "red"}))
	assert.Check(t, ValidateBrandColors(map[string]any{"buttonColor": ""}))
	assert.Check(t, is.ErrorIs(ValidateBrandColors(map[string]any{"buttonColor": "red"}), ErrInvalidBrandColor))
}

func TestValidateTemplateDefaults(t *testing.T) {
	questionnaireKey := QuestionnaireAccessOp.Name()
	brandedKey := BrandedMessageOp.Name()

	tests := []struct {
		name     string
		key      string
		defaults map[string]any
		wantErr  error
	}{
		{
			name:     "questionnaire with button text",
			key:      questionnaireKey,
			defaults: map[string]any{"buttonText": "Start", "buttonColor": "#000000"},
		},
		{
			name:     "questionnaire missing button text",
			key:      questionnaireKey,
			defaults: map[string]any{"title": "Hi"},
			wantErr:  ErrQuestionnaireButtonTextRequired,
		},
		{
			name:     "questionnaire empty button text",
			key:      questionnaireKey,
			defaults: map[string]any{"buttonText": ""},
			wantErr:  ErrQuestionnaireButtonTextRequired,
		},
		{
			name:     "questionnaire nil defaults",
			key:      questionnaireKey,
			defaults: nil,
			wantErr:  ErrQuestionnaireButtonTextRequired,
		},
		{
			name:     "questionnaire invalid color",
			key:      questionnaireKey,
			defaults: map[string]any{"buttonText": "Start", "buttonColor": "red"},
			wantErr:  ErrInvalidBrandColor,
		},
		{
			name:     "branded message without button text",
			key:      brandedKey,
			defaults: map[string]any{"title": "Hi"},
		},
		{
			name:     "branded message invalid color",
			key:      brandedKey,
			defaults: map[string]any{"primaryColor": "#12345g"},
			wantErr:  ErrInvalidBrandColor,
		},
		{
			name:     "unknown key nil defaults",
			key:      "",
			defaults: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTemplateDefaults(tc.key, tc.defaults)

			if tc.wantErr != nil {
				assert.Check(t, is.ErrorIs(err, tc.wantErr))
				return
			}

			assert.Check(t, err)
		})
	}
}
