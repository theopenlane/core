package validator_test

import (
	"testing"

	"github.com/theopenlane/ent/validator"
	"github.com/theopenlane/shared/enums"
)

func TestValidateFilter(t *testing.T) {
	validFilter := `statusIn: [OPEN, IN_PROGRESS]`
	invalidFilter := `banana: [POTATO]`

	if err := validator.ValidateFilter(validFilter, enums.ExportTypeTask); err != nil {
		t.Errorf("expected valid filter, got error: %v", err)
	}

	if err := validator.ValidateFilter(invalidFilter, enums.ExportTypeTask); err == nil {
		t.Errorf("expected invalid filter to fail")
	}
}
