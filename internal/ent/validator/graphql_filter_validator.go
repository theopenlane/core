package validator

import (
	"fmt"
	"strings"

	"github.com/theopenlane/core/pkg/enums"
)

// ValidateExportType checks if the provided export_type is valid.
func ValidateExportType(filter string, exportType enums.ExportType) error {
	if filter == "" {
		return nil
	}

	// Get all allowed values for the enum
	validValues := exportType.Values()

	for _, v := range validValues {
		if strings.ToUpper(filter) == v {
			return nil // valid
		}
	}

	return fmt.Errorf("invalid export_type '%s', must be one of %v", filter, validValues)
}
