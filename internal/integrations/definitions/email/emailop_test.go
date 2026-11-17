package email

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

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
