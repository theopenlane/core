package models_test

import (
	"testing"
	"time"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"gotest.tools/v3/assert"
)

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

func TestJobCadence_Next(t *testing.T) {
	location := time.UTC
	baseTime := time.Date(2025, 5, 27, 14, 30, 0, 0, location) // May 27, 2025 14:30 UTC

	tests := []struct {
		name        string
		cadence     models.JobCadence
		from        time.Time
		want        time.Time
		wantErr     bool
		errContains string
	}{
		{
			name: "daily - next run is today",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
				// 15:00 is after 14:30, so next run should be today
				Time: "15:00",
			},
			from: baseTime,
			want: time.Date(2025, 5, 27, 15, 0, 0, 0, location),
		},
		{
			name: "daily - next run is tomorrow",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
				// 14:00 is before 14:30, so next run should be tomorrow
				// as we have gone past the run time at 14:00
				Time: "14:00",
			},
			from: baseTime,
			want: time.Date(2025, 5, 28, 14, 0, 0, 0, location),
		},
		{
			name: "daily - exactly same time",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
				Time:      "14:30",
			},
			from: baseTime,
			want: time.Date(2025, 5, 27, 14, 30, 0, 0, location),
		},
		{
			name: "weekly - next run this week",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "15:00",
				Days:      []enums.JobWeekday{enums.JobWeekdayTuesday},
			},
			from: baseTime,
			want: time.Date(2025, 5, 27, 15, 0, 0, 0, location),
		},
		{
			name: "weekly - next run is next week even though it is same day but time is past",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "14:50",
				Days:      []enums.JobWeekday{enums.JobWeekdayThursday},
			},
			from: baseTime,
			want: time.Date(2025, 5, 29, 14, 50, 0, 0, location),
		},
		{
			name: "weekly - next run is next week",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "14:00",
				Days:      []enums.JobWeekday{enums.JobWeekdayTuesday},
			},
			from: baseTime,
			want: time.Date(2025, 6, 3, 14, 0, 0, 0, location),
		},
		{
			name: "weekly - multiple days, next available day",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "15:00",
				Days:      []enums.JobWeekday{enums.JobWeekdayMonday, enums.JobWeekdayWednesday, enums.JobWeekdayFriday},
			},
			from: baseTime, // base time is tuesday, so it should be the next Wednesday in the same week
			want: time.Date(2025, 5, 28, 15, 0, 0, 0, location),
		},
		{
			name: "weekly - multiple days, current day but past time",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "14:00", // earlier than current time
				Days:      []enums.JobWeekday{enums.JobWeekdayMonday, enums.JobWeekdayTuesday, enums.JobWeekdayFriday},
			},
			from: baseTime, // base time is tuesday, but it is past time, next run time should be Friday
			want: time.Date(2025, 5, 30, 14, 0, 0, 0, location),
		},
		{
			name: "weekly - days wrapping to next week",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyWeekly,
				Time:      "14:00",
				Days:      []enums.JobWeekday{enums.JobWeekdayMonday},
			},
			from: baseTime,
			want: time.Date(2025, 6, 2, 14, 0, 0, 0, location),
		},
		{
			name: "monthly - next run is this month",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyMonthly,
				Time:      "15:00",
			},
			from: baseTime,
			want: time.Date(2025, 5, 27, 15, 0, 0, 0, location),
		},
		{
			name: "monthly - next run is next month",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyMonthly,
				Time:      "14:00",
			},
			from: baseTime,
			want: time.Date(2025, 6, 27, 14, 0, 0, 0, location),
		},
		{
			name: "monthly - month rollover",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyMonthly,
				Time:      "14:00",
			},
			from: time.Date(2025, 12, 31, 14, 30, 0, 0, location),
			want: time.Date(2026, 1, 31, 14, 0, 0, 0, location),
		},
		{
			name: "invalid frequency",
			cadence: models.JobCadence{
				Frequency: "INVALID",
				Time:      "15:00",
			},
			from:        baseTime,
			wantErr:     true,
			errContains: "unsupported cadence frequency",
		},
		{
			name: "invalid time format",
			cadence: models.JobCadence{
				Frequency: enums.JobCadenceFrequencyDaily,
				Time:      "25:00",
			},
			from:        baseTime,
			wantErr:     true,
			errContains: "invalid time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cadence.Next(tt.from)
			if tt.wantErr {
				assert.Assert(t, err != nil)
				assert.ErrorContains(t, err, tt.errContains)
				return
			}
			assert.NilError(t, err)
			assert.DeepEqual(t, tt.want, got)
		})
	}
}
