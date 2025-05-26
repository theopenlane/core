package models_test

import (
	"testing"

	"github.com/theopenlane/core/pkg/models"
	"gotest.tools/v3/assert"
)

func TestValidateCronExpression(t *testing.T) {

	tt := []struct {
		name    string
		cron    string
		wantErr bool
	}{
		{
			name:    "valid hourly runs",
			cron:    "0 * * * *",
			wantErr: false,
		},
		{
			name:    "valid run every 30 minutes",
			cron:    "0,30 * * * *",
			wantErr: false,
		},
		{
			name:    "too frequent runs (5 minutes)",
			cron:    "*/5 * * * *",
			wantErr: true,
		},
		{
			name:    "too frequent uneven gaps (0,20,50)",
			cron:    "0,20,50 * * * *",
			wantErr: true,
		},
		{
			name:    "invalid syntax",
			cron:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateCronExpression(tt.cron)

			assert.Assert(t, (err != nil) == tt.wantErr)
		})
	}
}
