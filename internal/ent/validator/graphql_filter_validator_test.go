package validator_test

import (
	"testing"

	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

func TestValidateExportType(t *testing.T) {
	tests := []struct {
		name       string
		filter     string
		exportType enums.ExportType
		wantErr    bool
	}{
		{
			name:       "valid export type",
			filter:     "TASK",
			exportType: enums.ExportTypeTask,
			wantErr:    false,
		},
		{
			name:       "invalid export type",
			filter:     "INVALID_TYPE",
			exportType: enums.ExportTypeTask,
			wantErr:    true,
		},
		{
			name:       "null filter",
			filter:     "",
			exportType: enums.ExportTypeTask,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateExportType(tt.filter, tt.exportType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
