package gala

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

const (
	migrationOldTopic  = TopicName("GalaMigrationTest")
	migrationNewTopic  = TopicName(TopicPrefixMutation + "GalaMigrationTest")
	migrationSkipTopic = TopicName("gala.test.migration.skip")
)

// emitLegacyJob inserts a pre-kind envelope the way an old binary did: straight through
// the dispatcher with no kind header
func emitLegacyJob(t *testing.T, g *Gala, topic TopicName, message string, scheduledAt *time.Time) {
	t.Helper()

	snapshot, err := g.contextManager.Capture(context.Background())
	if err != nil {
		t.Fatalf("failed to capture snapshot: %v", err)
	}

	err = g.dispatcher.Dispatch(context.Background(), Envelope{
		ID:              NewEventID(),
		Topic:           topic,
		OccurredAt:      time.Now().UTC(),
		Payload:         []byte(`{"message":"` + message + `"}`),
		Headers:         Headers{ScheduledAt: scheduledAt},
		ContextSnapshot: snapshot,
	})
	if err != nil {
		t.Fatalf("failed to dispatch legacy job: %v", err)
	}
}

// TestMigrateJobsEndToEnd verifies the full transition: topic renames, kind resolution,
// header enrichment, schedule preservation, idempotency, and dispatch-time rename fallback
func TestMigrateJobsEndToEnd(t *testing.T) {
	ctx := context.Background()

	legacy := NewTestGala(t, WithTestStartWorkers(false))

	oldTopic := Topic[runtimeTestPayload]{Name: migrationOldTopic}
	skipTopic := Topic[runtimeTestPayload]{Name: migrationSkipTopic}

	if err := registerTopic(legacy.Gala.registry, oldTopic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register topic: %v", err)
	}

	if err := registerTopic(legacy.Gala.registry, skipTopic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register skip topic: %v", err)
	}

	futureRun := time.Now().Add(2 * time.Minute).UTC()

	emitLegacyJob(t, legacy.Gala, oldTopic.Name, "immediate", nil)
	emitLegacyJob(t, legacy.Gala, oldTopic.Name, "scheduled", &futureRun)
	emitLegacyJob(t, legacy.Gala, skipTopic.Name, "unresolved", nil)

	renames := map[TopicName]TopicName{migrationOldTopic: migrationNewTopic}

	upgraded, err := NewGala(ctx, Config{
		ConnectionURI: legacy.ConnectionURI,
		QueueName:     defaultTestQueueName,
		WorkerCount:   defaultTestWorkerCount,
		TopicRenames:  renames,
	})
	if err != nil {
		t.Fatalf("failed to create upgraded runtime: %v", err)
	}

	t.Cleanup(func() { _ = upgraded.Close() })

	// the upgraded runtime knows only the designated topic
	newTopic := Topic[runtimeTestPayload]{Name: migrationNewTopic, Kind: JobKindMutation}
	if err := registerTopic(upgraded.registry, newTopic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register renamed topic: %v", err)
	}

	if err := registerTopic(upgraded.registry, skipTopic, JSONCodec[runtimeTestPayload]{}); err != nil {
		t.Fatalf("failed to register skip topic on upgraded runtime: %v", err)
	}

	var handled atomic.Int32

	if _, err := attachListener(upgraded, Definition[runtimeTestPayload]{
		Topic:  newTopic,
		Name:   "gala.test.migration.listener",
		Handle: func(HandlerContext, runtimeTestPayload) error { handled.Add(1); return nil },
	}); err != nil {
		t.Fatalf("failed to attach listener: %v", err)
	}

	transform := func(_ string, envelope Envelope) (Envelope, bool) {
		if renamed, ok := renames[envelope.Topic]; ok {
			envelope.Topic = renamed
		}

		if envelope.Headers.Properties == nil {
			envelope.Headers.Properties = map[string]string{}
		}

		envelope.Headers.Properties["migrated_from"] = string(migrationOldTopic)

		return envelope, true
	}

	migrated, err := upgraded.MigrateJobs(ctx, transform)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if migrated != 1 {
		t.Fatalf("expected only the scheduled job to migrate, got %d", migrated)
	}

	rc := upgraded.jobClient.GetRiverClient()

	kindRows := listJobs(t, rc, river.NewJobListParams().
		Kinds(JobKindMutation).
		States(rivertype.JobStateAvailable, rivertype.JobStateScheduled))
	if len(kindRows) != 1 {
		t.Fatalf("expected 1 migrated row, got %d", len(kindRows))
	}

	row := kindRows[0]

	if row.Queue != QueueNameForKind(JobKindMutation) {
		t.Fatalf("expected kind queue, got %q", row.Queue)
	}

	if !strings.Contains(string(row.Metadata), string(migrationNewTopic)) {
		t.Fatalf("expected renamed topic in metadata, got %s", row.Metadata)
	}

	if !strings.Contains(string(row.Metadata), "migrated_from") {
		t.Fatalf("expected enrichment property in metadata, got %s", row.Metadata)
	}

	if row.State != rivertype.JobStateScheduled {
		t.Fatalf("expected migrated job to stay scheduled, got %s", row.State)
	}

	if drift := row.ScheduledAt.Sub(futureRun); drift < -5*time.Second || drift > 5*time.Second {
		t.Fatalf("expected preserved schedule near %v, got %v", futureRun, row.ScheduledAt)
	}

	legacyActive := listJobs(t, rc, river.NewJobListParams().
		Kinds(riverDispatchJobKind).
		States(rivertype.JobStateAvailable, rivertype.JobStateScheduled))
	if len(legacyActive) != 2 {
		t.Fatalf("expected the claimable and unresolved jobs to remain on the legacy kind, got %d rows", len(legacyActive))
	}

	cancelled := listJobs(t, rc, river.NewJobListParams().
		Kinds(riverDispatchJobKind).
		States(rivertype.JobStateCancelled))
	if len(cancelled) != 1 {
		t.Fatalf("expected 1 cancelled original, got %d", len(cancelled))
	}

	again, err := upgraded.MigrateJobs(ctx, transform)
	if err != nil {
		t.Fatalf("second migration failed: %v", err)
	}

	if again != 0 {
		t.Fatalf("expected idempotent second migration, got %d", again)
	}

	// the claimable old-topic job executes under the legacy worker through the
	// dispatch-time rename fallback
	if err := upgraded.StartWorkers(ctx); err != nil {
		t.Fatalf("failed to start workers: %v", err)
	}

	t.Cleanup(func() { _ = upgraded.StopWorkers(context.Background()) })

	deadline := time.Now().Add(15 * time.Second)
	for handled.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}

	if got := handled.Load(); got != 1 {
		t.Fatalf("expected the claimable job to execute via the rename fallback, got %d", got)
	}
}

// jobLister is the job listing capability the assertions need
type jobLister interface {
	JobList(context.Context, *river.JobListParams) (*river.JobListResult, error)
}

// listJobs pages a full job list result set
func listJobs(t *testing.T, rc jobLister, params *river.JobListParams) []*rivertype.JobRow {
	t.Helper()

	var jobs []*rivertype.JobRow

	params = params.First(activeJobsPageSize)

	for {
		result, err := rc.JobList(context.Background(), params)
		if err != nil {
			t.Fatalf("job list failed: %v", err)
		}

		jobs = append(jobs, result.Jobs...)

		if len(result.Jobs) < activeJobsPageSize || result.LastCursor == nil {
			return jobs
		}

		params = params.After(result.LastCursor)
	}
}
