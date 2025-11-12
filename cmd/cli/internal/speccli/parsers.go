//go:build cli

package speccli

import (
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// DefaultParsers returns the built-in parser catalog for spec fields.
func DefaultParsers() map[string]ValueParser {
	return map[string]ValueParser{
		"programStatus":   parseProgramStatus,
		"durationFromNow": parseDurationFromNow,
		"taskStatus":      parseTaskStatus,
		"standardStatus":  parseStandardStatus,
		"dateTime":        parseDateTime,
	}
}

// parseProgramStatus converts CLI input into an enums.ProgramStatus.
func parseProgramStatus(input any) (any, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("programStatus parser expects string input, got %T", input)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("program status cannot be empty")
	}

	status := enums.ToProgramStatus(value)
	if status == nil || status.String() == enums.ProgramStatusInvalid.String() {
		return nil, fmt.Errorf("invalid program status %q", value)
	}

	return *status, nil
}

// parseDurationFromNow interprets durations as offsets from now.
func parseDurationFromNow(input any) (any, error) {
	duration, ok := input.(time.Duration)
	if !ok {
		return nil, fmt.Errorf("durationFromNow parser expects time.Duration input, got %T", input)
	}

	if duration == 0 {
		return nil, fmt.Errorf("duration must be non-zero")
	}

	t := time.Now().Add(duration)
	return t, nil
}

// parseTaskStatus converts CLI input into an enums.TaskStatus.
func parseTaskStatus(input any) (any, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("taskStatus parser expects string input, got %T", input)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("task status cannot be empty")
	}

	status := enums.ToTaskStatus(value)
	if status == nil || status.String() == enums.TaskStatusInvalid.String() {
		return nil, fmt.Errorf("invalid task status %q", value)
	}

	return *status, nil
}

// parseStandardStatus converts CLI input into an enums.StandardStatus.
func parseStandardStatus(input any) (any, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("standardStatus parser expects string input, got %T", input)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("standard status cannot be empty")
	}

	status := enums.ToStandardStatus(value)
	if status == nil {
		return nil, fmt.Errorf("invalid standard status %q", value)
	}

	return *status, nil
}

// parseDateTime parses a user-provided string into the platform DateTime type.
func parseDateTime(input any) (any, error) {
	value, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("dateTime parser expects string input, got %T", input)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("datetime cannot be empty")
	}

	parsed, err := models.ToDateTime(value)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}
