package hooks

import (
	"testing"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

func TestResolveGalaRuntimes(t *testing.T) {
	runtimeA, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create first runtime: %v", err)
	}

	runtimeB, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create second runtime: %v", err)
	}

	providers := []func() *gala.Gala{
		nil,
		func() *gala.Gala { return runtimeA },
		func() *gala.Gala { return nil },
		func() *gala.Gala { return runtimeA },
		func() *gala.Gala { return runtimeB },
	}

	runtimes := resolveGalaRuntimes(providers)
	if len(runtimes) != 2 {
		t.Fatalf("expected two runtimes, got %d", len(runtimes))
	}

	if runtimes[0] != runtimeA {
		t.Fatalf("expected first runtime to match runtimeA")
	}

	if runtimes[1] != runtimeB {
		t.Fatalf("expected second runtime to match runtimeB")
	}
}

func TestMutationDispatchTargetsRoutesConcernTopicsToInterestedRuntimes(t *testing.T) {
	workflowRuntime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create workflow runtime: %v", err)
	}

	notificationRuntime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create notification runtime: %v", err)
	}

	workflowTopic := gala.Topic[eventqueue.MutationGalaPayload]{
		Name: eventqueue.MutationTopicName(eventqueue.MutationConcernWorkflow, entgen.TypeTask),
	}
	if _, err := gala.RegisterListeners(workflowRuntime.Registry(),
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      workflowTopic,
			Name:       "workflow.listener",
			Operations: []string{ent.OpCreate.String()},
			Handle: func(gala.HandlerContext, eventqueue.MutationGalaPayload) error {
				return nil
			},
		},
	); err != nil {
		t.Fatalf("failed to register workflow listener: %v", err)
	}

	notificationTopic := gala.Topic[eventqueue.MutationGalaPayload]{
		Name: eventqueue.MutationTopicName(eventqueue.MutationConcernNotification, entgen.TypeTask),
	}
	if _, err := gala.RegisterListeners(notificationRuntime.Registry(),
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      notificationTopic,
			Name:       "notification.listener",
			Operations: []string{ent.OpCreate.String()},
			Handle: func(gala.HandlerContext, eventqueue.MutationGalaPayload) error {
				return nil
			},
		},
	); err != nil {
		t.Fatalf("failed to register notification listener: %v", err)
	}

	targets := mutationDispatchTargets(
		[]*gala.Gala{workflowRuntime, notificationRuntime},
		mutationDispatchTopics(entgen.TypeTask),
		ent.OpCreate.String(),
	)
	if len(targets) != 2 {
		t.Fatalf("expected 2 dispatch targets, got %d", len(targets))
	}

	var sawWorkflow bool
	var sawNotification bool
	for _, target := range targets {
		if target.runtime == workflowRuntime && target.topic == workflowTopic.Name {
			sawWorkflow = true
		}

		if target.runtime == notificationRuntime && target.topic == notificationTopic.Name {
			sawNotification = true
		}
	}

	if !sawWorkflow {
		t.Fatalf("expected workflow runtime dispatch target")
	}

	if !sawNotification {
		t.Fatalf("expected notification runtime dispatch target")
	}
}
