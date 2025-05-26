package models_test

import (
	"testing"

	"github.com/theopenlane/core/pkg/enums"
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

func TestJobCadence_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cadence models.JobCadence
		wantErr bool
	}{
		{
			name: "valid daily cadence",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
				Time:      "15:00",
			},
			wantErr: false,
		},
		{
			name: "valid weekly cadence",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "09:30",
				Days:      models.Days{enums.JobWeekdayMonday, enums.JobWeekdayWednesday},
			},
			wantErr: false,
		},
		{
			name: "invalid frequency",
			cadence: models.JobCadence{
				Frequency: "invalid",
				Time:      "15:00",
			},
			wantErr: true,
		},
		{
			name: "missing time",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
			},
			wantErr: true,
		},
		{
			name: "invalid time format",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
				Time:      "25:00", // invalid hour
			},
			wantErr: true,
		},
		{
			name: "weekly cadence without days",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "09:00",
			},
			wantErr: true,
		},
		{
			name: "weekly cadence with duplicate days",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "09:00",
				Days:      models.Days{enums.JobWeekdayMonday, enums.JobWeekdayMonday},
			},
			wantErr: true,
		},
		{
			name: "weekly cadence with invalid weekday",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "09:00",
				Days:      models.Days{"invalidday"},
			},
			wantErr: true,
		},
		{
			name: "weekly cadence with too many days",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "09:00",
				Days: models.Days{
					enums.JobWeekdayMonday,
					enums.JobWeekdayTuesday,
					enums.JobWeekdayWednesday,
					enums.JobWeekdayThursday,
					enums.JobWeekdayFriday,
					enums.JobWeekdaySaturday,
					enums.JobWeekdaySunday,
					"extraday",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cadence.Validate()
			assert.Assert(t, (err != nil) == tt.wantErr)
		})
	}
}
