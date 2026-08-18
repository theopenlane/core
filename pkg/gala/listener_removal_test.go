package gala

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type listenerRemovalPayload struct {
	Operation string `json:"operation"`
}

func (p listenerRemovalPayload) PayloadOperation() string {
	return p.Operation
}

type listenerRemovalJobController struct {
	jobs              map[int64]*rivertype.JobRow
	cancelled         []int64
	deleted           []int64
	listErr           error
	successorOnDelete *rivertype.JobRow
}

func (c *listenerRemovalJobController) JobList(context.Context, *river.JobListParams) (*river.JobListResult, error) {
	if c.listErr != nil {
		err := c.listErr
		c.listErr = nil

		return nil, err
	}

	jobs := make([]*rivertype.JobRow, 0, len(c.jobs))
	for _, job := range c.jobs {
		if slices.Contains(uniqueLiveStates(), job.State) {
			jobs = append(jobs, job)
		}
	}

	return &river.JobListResult{Jobs: jobs}, nil
}

func (c *listenerRemovalJobController) JobCancel(_ context.Context, id int64) (*rivertype.JobRow, error) {
	job, ok := c.jobs[id]
	if !ok {
		return nil, rivertype.ErrNotFound
	}

	c.cancelled = append(c.cancelled, id)
	if job.State != rivertype.JobStateRunning {
		job.State = rivertype.JobStateCancelled
	}

	return job, nil
}

func (c *listenerRemovalJobController) JobDelete(_ context.Context, id int64) (*rivertype.JobRow, error) {
	job, ok := c.jobs[id]
	if !ok {
		return nil, rivertype.ErrNotFound
	}

	if job.State == rivertype.JobStateRunning {
		// Simulate a running job reaching a deletable terminal state
		job.State = rivertype.JobStateCancelled

		return nil, rivertype.ErrJobRunning
	}

	delete(c.jobs, id)
	c.deleted = append(c.deleted, id)
	if c.successorOnDelete != nil {
		c.jobs[c.successorOnDelete.ID] = c.successorOnDelete
		c.successorOnDelete = nil
	}

	return job, nil
}

func TestPurgeActiveJobsWithMetadataUsesSharedDurableCleanup(t *testing.T) {
	runtime := newTestGala(t, nil)
	controller := &listenerRemovalJobController{jobs: map[int64]*rivertype.JobRow{
		1: {ID: 1, State: rivertype.JobStateScheduled},
		2: {ID: 2, State: rivertype.JobStateRunning},
	}, successorOnDelete: &rivertype.JobRow{ID: 3, State: rivertype.JobStateScheduled}}
	runtime.jobController = controller

	purged, err := runtime.PurgeActiveJobsWithMetadata(context.Background(), `{"properties":{"entityId":"integration-1"}}`)
	if err != nil {
		t.Fatalf("failed to purge active jobs: %v", err)
	}
	if purged != 3 {
		t.Fatalf("expected 3 purged jobs including the late successor, got %d", purged)
	}
	if len(controller.jobs) != 0 {
		t.Fatalf("expected every selected job to be deleted, got %d", len(controller.jobs))
	}
	if len(controller.cancelled) != 3 {
		t.Fatalf("expected 3 cancellation requests, got %d", len(controller.cancelled))
	}
}

func TestRemoveListenersPurgesEveryLiveRiverState(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[listenerRemovalPayload]{Name: "listener.removal.live_states"}

	ids, err := Register(runtime, Definition[listenerRemovalPayload]{
		Topic:      topic,
		Operations: []string{"create"},
		Handle:     func(HandlerContext, listenerRemovalPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	controller := &listenerRemovalJobController{jobs: map[int64]*rivertype.JobRow{}}
	for index, state := range uniqueLiveStates() {
		id := int64(index + 1)
		controller.jobs[id] = newListenerRemovalJob(t, id, topic.Name, "create", state)
	}
	runtime.jobController = controller

	if err := runtime.RemoveListeners(context.Background(), ids...); err != nil {
		t.Fatalf("failed to remove listener: %v", err)
	}

	if len(controller.jobs) != 0 {
		t.Fatalf("expected every live River job to be purged, got %d", len(controller.jobs))
	}
	if len(controller.cancelled) != len(uniqueLiveStates()) {
		t.Fatalf("expected %d cancelled jobs, got %d", len(uniqueLiveStates()), len(controller.cancelled))
	}
	if got := len(runtime.registry.registeredListeners(topic.Name)); got != 0 {
		t.Fatalf("expected listener to be detached, got %d", got)
	}
}

func TestRemoveListenersPreservesJobsNeededByAnotherListener(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[listenerRemovalPayload]{Name: "listener.removal.shared_topic"}

	createIDs, err := Register(runtime, Definition[listenerRemovalPayload]{
		Topic:      topic,
		Operations: []string{"create"},
		Handle:     func(HandlerContext, listenerRemovalPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to register create listener: %v", err)
	}

	updateIDs, err := Register(runtime, Definition[listenerRemovalPayload]{
		Topic:      topic,
		Operations: []string{"update"},
		Handle:     func(HandlerContext, listenerRemovalPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to register update listener: %v", err)
	}

	unrelatedTopic := TopicName("listener.removal.unrelated")
	controller := &listenerRemovalJobController{jobs: map[int64]*rivertype.JobRow{
		1: newListenerRemovalJob(t, 1, topic.Name, "create", rivertype.JobStateScheduled),
		2: newListenerRemovalJob(t, 2, topic.Name, "update", rivertype.JobStateScheduled),
		3: newListenerRemovalJob(t, 3, unrelatedTopic, "create", rivertype.JobStateScheduled),
	}}
	runtime.jobController = controller

	if err := runtime.RemoveListeners(context.Background(), createIDs...); err != nil {
		t.Fatalf("failed to remove create listener: %v", err)
	}

	if _, ok := controller.jobs[1]; ok {
		t.Fatal("expected the removed create listener's job to be purged")
	}
	if _, ok := controller.jobs[2]; !ok {
		t.Fatal("expected the update job to remain for its matching listener")
	}
	if _, ok := controller.jobs[3]; !ok {
		t.Fatal("expected the unrelated topic job to remain")
	}

	if err := runtime.RemoveListeners(context.Background(), updateIDs...); err != nil {
		t.Fatalf("failed to remove update listener: %v", err)
	}
	if _, ok := controller.jobs[2]; ok {
		t.Fatal("expected the update job to be purged after its listener was removed")
	}
}

func TestRemoveListenersRetriesDurableCleanup(t *testing.T) {
	runtime := newTestGala(t, nil)
	topic := Topic[listenerRemovalPayload]{Name: "listener.removal.retry"}

	ids, err := Register(runtime, Definition[listenerRemovalPayload]{
		Topic:  topic,
		Handle: func(HandlerContext, listenerRemovalPayload) error { return nil },
	})
	if err != nil {
		t.Fatalf("failed to register listener: %v", err)
	}

	controller := &listenerRemovalJobController{
		jobs: map[int64]*rivertype.JobRow{
			1: newListenerRemovalJob(t, 1, topic.Name, "", rivertype.JobStateScheduled),
		},
		listErr: errors.New("temporary list failure"),
	}
	runtime.jobController = controller

	err = runtime.RemoveListeners(context.Background(), ids...)
	if !errors.Is(err, ErrRiverListenerCleanupFailed) {
		t.Fatalf("expected listener cleanup error, got %v", err)
	}
	if got := len(runtime.registry.registeredListeners(topic.Name)); got != 0 {
		t.Fatalf("expected listener to remain detached after cleanup failure, got %d", got)
	}

	if err := runtime.RemoveListeners(context.Background(), ids...); err != nil {
		t.Fatalf("failed to retry listener cleanup: %v", err)
	}
	if len(controller.jobs) != 0 {
		t.Fatalf("expected retry to purge the scheduled job, got %d", len(controller.jobs))
	}
}

func TestListenerTopicMetadataFragmentOnlyFiltersByTopic(t *testing.T) {
	fragment, err := listenerTopicMetadataFragment("listener.removal.metadata")
	if err != nil {
		t.Fatalf("failed to build topic metadata fragment: %v", err)
	}
	if fragment != `{"topic":"listener.removal.metadata"}` {
		t.Fatalf("unexpected topic metadata fragment: %s", fragment)
	}
}

func newListenerRemovalJob(t *testing.T, id int64, topic TopicName, operation string, state rivertype.JobState) *rivertype.JobRow {
	t.Helper()

	payload, err := json.Marshal(listenerRemovalPayload{Operation: operation})
	if err != nil {
		t.Fatalf("failed to encode removal payload: %v", err)
	}

	args, err := newRiverDispatchArgs(Envelope{
		ID:      EventID("listener-removal-event"),
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("failed to build removal args: %v", err)
	}

	encodedArgs, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("failed to encode removal args: %v", err)
	}

	return &rivertype.JobRow{
		ID:          id,
		EncodedArgs: encodedArgs,
		State:       state,
	}
}
