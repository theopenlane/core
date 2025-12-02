package models_test

import (
	"testing"
	"time"

	"github.com/theopenlane/shared/models"
	"gotest.tools/v3/assert"
)

func TestCron_Validate(t *testing.T) {

	tt := []struct {
		name    string
		cron    string
		wantErr bool
	}{
		{
			name:    "valid hourly runs",
			cron:    "0 0 */1 * * *",
			wantErr: false,
		},
		{
			name:    "valid run every 30 minutes",
			cron:    "0 */30 * * * *",
			wantErr: false,
		},
		{
			name:    "valid",
			cron:    "0 0 12 * * *",
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
			name:    "every 3 hours",
			cron:    "0 0 */3 ? * *",
			wantErr: false,
		},
		{
			name:    "every weekday",
			cron:    "0 0 12 * * MON-FRI",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			cron:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			err := models.Cron(tt.cron).Validate()

			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err, "expected no error for cron: %s", tt.cron)
		})
	}
}

func TestCron_Next(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tt := []struct {
		name    string
		cron    string
		from    time.Time
		want    time.Time
		wantErr bool
	}{
		{
			name:    "hourly at minute zero",
			cron:    "0 * * * *",
			from:    baseTime,
			want:    time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "every 30 minutes",
			cron:    "0,30 * * * *",
			from:    baseTime,
			want:    time.Date(2025, 1, 1, 0, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "daily at midnight",
			cron:    "0 0 * * *",
			from:    baseTime,
			want:    time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			cron:    "invalid",
			from:    baseTime,
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			next, err := models.Cron(tt.cron).Next(tt.from)

			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, next, tt.want)
		})
	}
}
