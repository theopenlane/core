package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/theopenlane/shared/enums"
)

var (
	// ErrComputeNextRunInvalid is used to define an error when a weekly run cannot be
	// computed
	ErrComputeNextRunInvalid = errors.New("could not compute next run time in weekly cadence")
)

// Days is used to provide a human readable version of weekdays
type Days []enums.JobWeekday

// JobCadence defines the logic for the execution of a job
type JobCadence struct {
	Days      Days                      `json:"days,omitempty"`
	Time      string                    `json:"time,omitempty"`
	Frequency enums.JobCadenceFrequency `json:"frequency,omitempty"`
}

// IsZero checks if the cadence is not set yet
func (c JobCadence) IsZero() bool {
	return c.Days == nil && c.Time == "" && c.Frequency == ""
}

// String marshals the cadence into a human readable version
func (c JobCadence) String() string {
	var b bytes.Buffer

	if err := json.NewEncoder(&b).Encode(c); err != nil {
		return ""
	}

	return b.String()
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (c JobCadence) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, c)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (c *JobCadence) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, c)
}

// simple cache
var validWeekdaysSet = func() map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range enums.JobWeekdayFriday.Values() {
		m[v] = struct{}{}
	}
	return m
}()

// Validate makes sure we have a usable job cadence setting
func (c *JobCadence) Validate() error {
	if c.IsZero() {
		return nil
	}

	if val := enums.ToJobCadenceFrequency(c.Frequency.String()); *val == enums.JobCadenceFrequencyInvalid {
		return errors.New("invalid frequency") // nolint:err113
	}

	// time of the day must be required
	if c.Time == "" {
		return errors.New("time is required for cadence configuration") // nolint:err113
	}

	if c.Frequency.String() == enums.JobCadenceFrequencyWeekly.String() {
		if len(c.Days) == 0 {
			return errors.New("days must be specified for weekly cadence") // nolint:err113
		}

		// make sure we cannot have days like ["sunday", "sunday"]
		for _, day := range c.Days {
			if _, ok := validWeekdaysSet[day.String()]; !ok {
				return fmt.Errorf("invalid weekday: %s", day) // nolint:err113
			}
		}

		valid := enums.JobWeekdaySaturday.Values()
		validSet := make(map[string]struct{}, len(valid))

		for _, v := range valid {
			validSet[v] = struct{}{}
		}

		seen := make(map[string]struct{}, len(c.Days))

		for _, day := range c.Days {
			s := day.String()

			if _, ok := validSet[s]; !ok {
				return fmt.Errorf("invalid weekday: %s", s) // nolint:err113
			}

			if _, dup := seen[s]; dup {
				return fmt.Errorf("duplicate weekday: %s", s) // nolint:err113
			}

			seen[s] = struct{}{}
		}

		if len(c.Days) > len(validSet) {
			return fmt.Errorf("too many weekdays: max allowed is %d", len(validSet)) // nolint:err113
		}
	}

	if _, err := time.Parse("15:04", c.Time); err != nil {
		return fmt.Errorf("invalid time format: %w", err) // nolint:err113
	}

	return nil
}

// Next calculates the next execution time for a JobCadence
func (c JobCadence) Next(from time.Time) (time.Time, error) {
	// we do not call Validate again as the db hook
	// already does that
	expectedRunTime, err := time.Parse("15:04", c.Time)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format in cadence: %w", err)
	}

	expectedTargetHour := expectedRunTime.Hour()
	expectedTargetMinute := expectedRunTime.Minute()

	switch c.Frequency {
	case enums.JobCadenceFrequencyDaily:
		// if it's past the expected time today, then set the next run time to the next 24 hours
		expectedNextRun := time.Date(from.Year(), from.Month(), from.Day(), expectedTargetHour,
			expectedTargetMinute, 0, 0, from.Location())

		if expectedNextRun.Before(from) {
			const next24hrs = 24 * time.Hour

			expectedNextRun = expectedNextRun.Add(next24hrs)
		}

		return expectedNextRun, nil

	case enums.JobCadenceFrequencyWeekly:
		targetWeekdays := make([]time.Weekday, 0, len(c.Days))
		for _, day := range c.Days {
			targetWeekdays = append(targetWeekdays, enums.ToTimeWeekday(day))
		}

		// peek into the next 2 weeks
		for i := range 14 {
			next := from.AddDate(0, 0, i)
			for _, d := range targetWeekdays {
				if next.Weekday() == d {
					currentCandidateCheck := time.Date(next.Year(), next.Month(), next.Day(), expectedTargetHour,
						expectedTargetMinute, 0, 0, from.Location())

					if currentCandidateCheck.After(from) {
						return currentCandidateCheck, nil
					}
				}
			}
		}

		return time.Time{}, ErrComputeNextRunInvalid

	case enums.JobCadenceFrequencyMonthly:
		// initial run time should be set to the target time on the same day of the current month
		expectedNextRun := time.Date(from.Year(), from.Month(), from.Day(), expectedTargetHour,
			expectedTargetMinute, 0, 0, from.Location())

		// past time today? move to the same day next month
		if expectedNextRun.Before(from) {
			expectedNextRun = expectedNextRun.AddDate(0, 1, 0)
		}

		return expectedNextRun, nil

	default:
		return time.Time{}, fmt.Errorf("unsupported cadence frequency: %s", c.Frequency) // nolint:err113
	}
}
