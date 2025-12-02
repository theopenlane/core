package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/ent/validator"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"ValidName", "JohnDoe", false},
		{"NameWithSpace", "John Doe", false},
		{"NameWithSpecialChar", "John$Doe", true},
		{"NameWithNumbers", "John123", false},
		{"EmptyName", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.SpecialCharValidator(tt.input)
			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
