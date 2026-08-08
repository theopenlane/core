package hooks

import (
	"testing"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

func TestAnyRuntimeInterestedMatchesConcernTopics(t *testing.T) {
	workflowRuntime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create workflow runtime: %v", err)
	}

	notificationRuntime, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create notification runtime: %v", err)
	}

	workflowTopic := gala.Topic[entityops.MutationPayload]{
		Name: gala.MutationTopicName(gala.MutationConcernWorkflow, entgen.TypeTask),
	}
	if _, err := gala.Register(workflowRuntime,
		gala.Definition[entityops.MutationPayload]{
			Topic:      workflowTopic,
			Name:       "workflow.listener",
			Operations: []string{ent.OpCreate.String()},
			Handle: func(gala.HandlerContext, entityops.MutationPayload) error {
				return nil
			},
		},
	); err != nil {
		t.Fatalf("failed to register workflow listener: %v", err)
	}

	runtimes := []*gala.Gala{workflowRuntime, notificationRuntime}

	if !anyRuntimeInterested(runtimes, entgen.TypeTask, ent.OpCreate.String()) {
		t.Fatal("expected interest on the workflow concern topic")
	}

	if anyRuntimeInterested(runtimes, entgen.TypeTask, ent.OpDelete.String()) {
		t.Fatal("expected no interest for an unregistered operation")
	}

	if anyRuntimeInterested(runtimes, entgen.TypeControl, ent.OpCreate.String()) {
		t.Fatal("expected no interest for an unregistered schema")
	}
}

func TestMutationConcernTopics(t *testing.T) {
	topics := mutationConcernTopics(entgen.TypeTask)

	if topics[0] != gala.MutationTopicName(gala.MutationConcernDirect, entgen.TypeTask) {
		t.Fatalf("unexpected direct topic %q", topics[0])
	}

	if topics[1] != gala.MutationTopicName(gala.MutationConcernWorkflow, entgen.TypeTask) {
		t.Fatalf("unexpected workflow topic %q", topics[1])
	}

	if topics[2] != gala.MutationTopicName(gala.MutationConcernNotification, entgen.TypeTask) {
		t.Fatalf("unexpected notification topic %q", topics[2])
	}
}
