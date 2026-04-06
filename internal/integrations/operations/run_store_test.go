package operations

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMarkRunRunning_EmptyRunID(t *testing.T) {
	t.Parallel()

	err := MarkRunRunning(context.Background(), nil, "")
	if !errors.Is(err, ErrRunIDRequired) {
		t.Fatalf("expected ErrRunIDRequired, got %v", err)
	}
}

func TestCompleteRun_EmptyRunID(t *testing.T) {
	t.Parallel()

	err := CompleteRun(context.Background(), nil, "", time.Now(), RunResult{})
	if !errors.Is(err, ErrRunIDRequired) {
		t.Fatalf("expected ErrRunIDRequired, got %v", err)
	}
}

func TestCreatePendingRun_NilInstallation(t *testing.T) {
	t.Parallel()

	_, err := CreatePendingRun(context.Background(), nil, nil, DispatchRequest{})
	if !errors.Is(err, ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}
}

func TestRunResult_DefaultStatus(t *testing.T) {
	t.Parallel()

	result := RunResult{}
	if result.Status != "" {
		t.Fatalf("expected empty default status, got %q", result.Status)
	}
}
