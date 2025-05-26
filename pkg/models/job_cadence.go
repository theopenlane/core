package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/theopenlane/core/pkg/enums"
)

// MaxRunsInBetween defines how much time each job must have between runs
// Maybe make this configurable or maybe we need to take this down to like
// 5/10 minutes
const MaxRunsInBetween = 30 * time.Minute

type Days []enums.JobWeekday

type JobCadence struct {
	Days      Days                      `json:"days,omitempty"`
	Time      string                    `json:"time,omitempty"`
	Frequency enums.JobCadenceFrequency `json:"frequency,omitempty"`
}

// IsZero checks if the cadence is not set yet
func (c JobCadence) IsZero() bool {
	return c.Days == nil && c.Time == "" && c.Frequency == ""
}

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
	if val := enums.ToJobCadenceFrequency(c.Frequency.String()); *val == enums.JobCadenceFrequencyInvalid {
		return errors.New("invalid frequency")
	}

	if c.Time == "" {
		return errors.New("time is required for cadence configuration")
	}

	if c.Frequency.String() == enums.JobCadenceFrequencyWeekly.String() {
		if len(c.Days) == 0 {
			return errors.New("days must be specified for weekly cadence")
		}
		for _, day := range c.Days {
			if _, ok := validWeekdaysSet[day.String()]; !ok {
				return fmt.Errorf("invalid weekday: %s", day)
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
				return fmt.Errorf("invalid weekday: %s", s)
			}

			if _, dup := seen[s]; dup {
				return fmt.Errorf("duplicate weekday: %s", s)
			}

			seen[s] = struct{}{}
		}

		if len(c.Days) > len(validSet) {
			return fmt.Errorf("too many weekdays: max allowed is %d", len(validSet))
		}
	}

	if _, err := time.Parse("15:04", c.Time); err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}

	return nil
}

func ValidateCronExpression(expr string) error {
	cron, err := cronexpr.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid cron syntax: %w", err)
	}

	// compute the next 5 execution times to cover cases like
	// 0,20,40 * * * * where the user can request to run in the 20th and 40th minute
	// that would break the 30 minute check
	currentTime := time.Now()
	executions := cron.NextN(currentTime, 5)

	for i := 1; i < len(executions); i++ {
		interval := executions[i].Sub(executions[i-1])
		if interval < MaxRunsInBetween {
			return fmt.Errorf("cron runs too frequently: %s between runs, must be at least 30 minutes", interval)
		}
	}

	return nil
}
