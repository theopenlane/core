package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
)

type testPayload struct {
	ID string `json:"id"`
}

func TestScopeEndRecordsMetrics(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	opName := OperationName("test_op_success")
	origin := Origin("test_origin")
	trigger := "test_trigger"

	before := testutil.ToFloat64(metrics.WorkflowOperationsTotal.WithLabelValues(string(opName), string(origin), trigger, metrics.LabelSuccess))
	scope := observer.begin(ctx, Operation{
		Name:         opName,
		Origin:       origin,
		TriggerEvent: trigger,
	}, nil)
	scope.End(nil, nil)
	after := testutil.ToFloat64(metrics.WorkflowOperationsTotal.WithLabelValues(string(opName), string(origin), trigger, metrics.LabelSuccess))

	if after != before+1 {
		t.Fatalf("expected success count to increment by 1, got before=%v after=%v", before, after)
	}
}

func TestScopeEndRecordsFailure(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	opName := OperationName("test_op_failure")
	origin := Origin("test_origin")
	trigger := "test_trigger"

	before := testutil.ToFloat64(metrics.WorkflowOperationsTotal.WithLabelValues(string(opName), string(origin), trigger, metrics.LabelFailure))
	scope := observer.begin(ctx, Operation{
		Name:         opName,
		Origin:       origin,
		TriggerEvent: trigger,
	}, nil)
	scope.RecordError(errors.New("boom"), nil)
	scope.End(nil, nil)
	after := testutil.ToFloat64(metrics.WorkflowOperationsTotal.WithLabelValues(string(opName), string(origin), trigger, metrics.LabelFailure))

	if after != before+1 {
		t.Fatalf("expected failure count to increment by 1, got before=%v after=%v", before, after)
	}
}

func TestHandleEmitRecordsError(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	op := Operation{Name: OperationName("emit_op"), Origin: Origin("emit_origin")}
	topic := "workflow.emit.test"

	before := testutil.ToFloat64(metrics.WorkflowEmitErrorsTotal.WithLabelValues(topic, string(op.Origin)))

	observer.handleEmitError(ctx, op, Fields{"k": "v"}, topic, errors.New("emit failed"))

	after := testutil.ToFloat64(metrics.WorkflowEmitErrorsTotal.WithLabelValues(topic, string(op.Origin)))
	require.Equal(t, before+1, after)
}

func TestBeginListenerTopicAppliesSpec(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)
	oldLogger := log.Logger
	log.Logger = logger
	defer func() { log.Logger = oldLogger }()

	payload := testPayload{ID: "payload-1"}
	topic := gala.TopicName("workflow.listener.test")
	handlerCtx := gala.HandlerContext{
		Context: context.Background(),
		Envelope: gala.Envelope{
			Topic: "custom_trigger",
		},
	}

	scope := BeginListenerTopic(handlerCtx, New(), topic, payload, Fields{"extra": "value"})

	if scope.op.Name != OperationName(topic) {
		t.Fatalf("expected operation %q, got %v", topic, scope.op.Name)
	}
	if scope.op.Origin != OriginListeners {
		t.Fatalf("expected origin %q, got %v", OriginListeners, scope.op.Origin)
	}
	if scope.op.TriggerEvent != "custom_trigger" {
		t.Fatalf("expected trigger custom_trigger, got %v", scope.op.TriggerEvent)
	}
}

func TestEmitTypedNoRuntime(t *testing.T) {
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	err := emitTyped(ctx, New(), nil, gala.TopicName("workflow.emit.typed"), testPayload{ID: "payload-2"}, Operation{
		Name:   OperationName("emit_op"),
		Origin: OriginEngine,
	}, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func findLogEntry(t *testing.T, buf *bytes.Buffer, msg string) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry["message"] == msg {
			return entry
		}
	}

	t.Fatalf("log entry with message %q not found", msg)
	return nil
}

func TestScopeSkipMarksSkippedAndLogsDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)
	ctx := logger.WithContext(context.Background())
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}

	scope := observer.begin(ctx, Operation{
		Name:   OperationName("skip_op"),
		Origin: Origin("skip_origin"),
	}, nil)
	scope.Skip("not applicable", Fields{"reason_detail": "missing config"})

	if !scope.skipped {
		t.Fatal("expected scope.skipped to be true")
	}

	scope.End(nil, nil)

	entry := findLogEntry(t, &buf, msgOpSkipped)
	if entry[FieldOperation] != "skip_op" {
		t.Fatalf("expected operation skip_op, got %v", entry[FieldOperation])
	}
}

func TestScopeFailReturnsErrorAndRecords(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	opName := OperationName("fail_op")
	origin := Origin("fail_origin")

	scope := observer.begin(ctx, Operation{
		Name:   opName,
		Origin: origin,
	}, nil)

	testErr := errors.New("operation failed")
	returned := scope.Fail(testErr, Fields{"action": "test"})

	if returned != testErr {
		t.Fatalf("expected Fail to return the error, got %v", returned)
	}

	if scope.recordErr != testErr {
		t.Fatalf("expected recordErr to be set, got %v", scope.recordErr)
	}
}

func TestScopeFailWithNilReturnsNil(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	scope := observer.begin(ctx, Operation{
		Name:   OperationName("fail_nil_op"),
		Origin: Origin("fail_origin"),
	}, nil)

	returned := scope.Fail(nil, Fields{"action": "test"})
	if returned != nil {
		t.Fatalf("expected Fail(nil) to return nil, got %v", returned)
	}

	if scope.recordErr != nil {
		t.Fatalf("expected recordErr to remain nil, got %v", scope.recordErr)
	}
}

func TestScopeContextReturnsEnrichedContext(t *testing.T) {
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}

	scope := observer.begin(ctx, Operation{
		Name:   OperationName("fields_op"),
		Origin: Origin("fields_origin"),
	}, Fields{"initial": "value"})

	// Verify scope returns a context (for downstream use)
	scopeCtx := scope.Context()
	if scopeCtx == nil {
		t.Fatal("expected non-nil context from scope")
	}

	// Verify the context has a logger attached
	logger := logx.FromContext(scopeCtx)
	if logger == nil {
		t.Fatal("expected logger in scope context")
	}
}

func TestRecordErrorKeepsFirstError(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	scope := observer.begin(ctx, Operation{
		Name:   OperationName("multi_err_op"),
		Origin: Origin("multi_err_origin"),
	}, nil)

	firstErr := errors.New("first error")
	secondErr := errors.New("second error")

	scope.RecordError(firstErr, nil)
	scope.RecordError(secondErr, nil)

	if scope.recordErr != firstErr {
		t.Fatalf("expected first error to be retained, got %v", scope.recordErr)
	}
}

func TestRecordErrorIgnoresNil(t *testing.T) {
	observer := &Observer{now: func() time.Time { return time.Unix(0, 0) }}
	ctx := zerolog.New(io.Discard).With().Logger().WithContext(context.Background())

	scope := observer.begin(ctx, Operation{
		Name:   OperationName("nil_err_op"),
		Origin: Origin("nil_err_origin"),
	}, nil)

	scope.RecordError(nil, Fields{"ignored": "field"})

	if scope.recordErr != nil {
		t.Fatalf("expected recordErr to remain nil, got %v", scope.recordErr)
	}
}

func TestWarnEngineLogsWarning(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.WarnLevel)
	ctx := logger.WithContext(context.Background())

	testErr := errors.New("engine warning")
	WarnEngine(ctx, OpTriggerWorkflow, "test.event", Fields{"detail": "info"}, testErr)

	entry := findLogEntry(t, &buf, msgOpWarning)
	if entry[FieldOperation] != string(OpTriggerWorkflow) {
		t.Fatalf("expected operation %s, got %v", OpTriggerWorkflow, entry[FieldOperation])
	}
	if entry[FieldOrigin] != string(OriginEngine) {
		t.Fatalf("expected origin %s, got %v", OriginEngine, entry[FieldOrigin])
	}
	if entry[FieldTriggerEvent] != "test.event" {
		t.Fatalf("expected trigger event test.event, got %v", entry[FieldTriggerEvent])
	}
	if entry["error"] != "engine warning" {
		t.Fatalf("expected error field, got %v", entry["error"])
	}
}

func TestWarnListenerLogsWarning(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.WarnLevel)
	ctx := logger.WithContext(context.Background())

	WarnListener(ctx, OpHandleAssignmentCompleted, "assignment.completed", Fields{"key": "val"}, nil)

	entry := findLogEntry(t, &buf, msgOpWarning)
	if entry[FieldOrigin] != string(OriginListeners) {
		t.Fatalf("expected origin %s, got %v", OriginListeners, entry[FieldOrigin])
	}
	if entry[FieldOperation] != string(OpHandleAssignmentCompleted) {
		t.Fatalf("expected operation %s, got %v", OpHandleAssignmentCompleted, entry[FieldOperation])
	}
}

func TestWarnWithNilErrorOmitsErrorField(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.WarnLevel)
	ctx := logger.WithContext(context.Background())

	WarnEngine(ctx, OpExecuteAction, "action.trigger", nil, nil)

	entry := findLogEntry(t, &buf, msgOpWarning)
	if _, hasErr := entry["error"]; hasErr {
		t.Fatal("expected no error field when err is nil")
	}
}
