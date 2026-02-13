package eventqueue

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const (
	testSubjectID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testOrgID     = "01ARZ3NDEKTSV4RRFFQ69G5FAW"
)

func TestNewMutationDispatchArgs(t *testing.T) {
	t.Parallel()

	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:          testSubjectID,
		OrganizationID:     testOrgID,
		OrganizationIDs:    []string{testOrgID},
		AuthenticationType: auth.JWTAuthentication,
		OrganizationRole:   auth.OwnerRole,
	})

	props := soiree.NewProperties()
	props.Set("ID", "entity_123")
	props.Set("field_name", "display_name")
	props.Set("active", true)
	props.Set("count", 7)

	payload := &events.MutationPayload{
		MutationType: "organization",
		Operation:    "UPDATE",
		EntityID:     "entity_123",
		ChangedFields: []string{
			"display_name",
		},
		ClearedFields: []string{
			"description",
		},
		ChangedEdges: []string{
			"delegate",
		},
		AddedIDs: map[string][]string{
			"delegate": []string{"user_1"},
		},
		RemovedIDs: map[string][]string{
			"delegate": []string{"user_2"},
		},
		ProposedChanges: map[string]any{
			"display_name": "Acme",
			"description":  nil,
		},
	}

	args := NewMutationDispatchArgs(ctx, "organization", payload, props)

	require.Equal(t, "organization", args.Topic)
	require.Equal(t, "UPDATE", args.Operation)
	require.Equal(t, "entity_123", args.EntityID)
	require.Equal(t, "organization", args.MutationType)
	require.Equal(t, []string{"display_name"}, args.ChangedFields)
	require.Equal(t, []string{"description"}, args.ClearedFields)
	require.Equal(t, []string{"delegate"}, args.ChangedEdges)
	require.Equal(t, []string{"user_1"}, args.AddedIDs["delegate"])
	require.Equal(t, []string{"user_2"}, args.RemovedIDs["delegate"])
	require.Equal(t, "Acme", args.ProposedChanges["display_name"])
	require.Contains(t, args.ProposedChanges, "description")
	require.Nil(t, args.ProposedChanges["description"])
	require.NotEmpty(t, args.EventID)
	require.NotNil(t, args.Auth)
	require.Equal(t, testSubjectID, args.Auth.SubjectID)
	require.Equal(t, testOrgID, args.Auth.OrganizationID)
	require.Equal(t, "true", args.Properties["active"])
	require.Equal(t, "7", args.Properties["count"])
	require.Equal(t, args.EventID, args.Properties[soiree.PropertyEventID])
	require.WithinDuration(t, time.Now().UTC(), args.OccurredAt, time.Minute)
}

func TestMutationDispatchWorkerWork(t *testing.T) {
	t.Parallel()

	bus := soiree.New()
	t.Cleanup(func() {
		_ = bus.Close()
	})

	called := false
	_, err := bus.On("organization", func(ctx *soiree.EventContext) error {
		called = true

		payload, ok := ctx.Payload().(*events.MutationPayload)
		if !ok {
			return errors.New("unexpected payload type")
		}

		if payload.Operation != "UPDATE" {
			return fmt.Errorf("unexpected operation %q", payload.Operation)
		}

		if payload.EntityID != "entity_123" {
			return fmt.Errorf("unexpected entity id %q", payload.EntityID)
		}
		if payload.MutationType != "organization" {
			return fmt.Errorf("unexpected mutation type %q", payload.MutationType)
		}
		if !reflect.DeepEqual([]string{"display_name"}, payload.ChangedFields) {
			return fmt.Errorf("unexpected changed fields %#v", payload.ChangedFields)
		}
		if !reflect.DeepEqual([]string{"description"}, payload.ClearedFields) {
			return fmt.Errorf("unexpected cleared fields %#v", payload.ClearedFields)
		}
		if payload.ProposedChanges["display_name"] != "Acme" {
			return fmt.Errorf("unexpected proposed change %#v", payload.ProposedChanges["display_name"])
		}
		if _, ok := payload.ProposedChanges["description"]; !ok {
			return errors.New("missing cleared field in proposed changes")
		}
		if payload.ProposedChanges["description"] != nil {
			return fmt.Errorf("unexpected description proposed value %#v", payload.ProposedChanges["description"])
		}

		subjectID, err := auth.GetSubjectIDFromContext(ctx.Context())
		if err != nil {
			return fmt.Errorf("subject id from context: %w", err)
		}

		if subjectID != testSubjectID {
			return fmt.Errorf("unexpected subject id %q", subjectID)
		}

		orgID, err := auth.GetOrganizationIDFromContext(ctx.Context())
		if err != nil {
			return fmt.Errorf("organization id from context: %w", err)
		}

		if orgID != testOrgID {
			return fmt.Errorf("unexpected organization id %q", orgID)
		}

		entityID, ok := ctx.PropertyString("ID")
		if !ok {
			return errors.New("missing ID property")
		}

		if entityID != "entity_123" {
			return fmt.Errorf("unexpected property ID %q", entityID)
		}

		return nil
	})
	require.NoError(t, err)

	worker := NewMutationDispatchWorker(func() *soiree.EventBus {
		return bus
	})

	job := &river.Job[MutationDispatchArgs]{
		Args: MutationDispatchArgs{
			Topic:        "organization",
			Operation:    "UPDATE",
			EntityID:     "entity_123",
			MutationType: "organization",
			ChangedFields: []string{
				"display_name",
			},
			ClearedFields: []string{
				"description",
			},
			ChangedEdges: []string{
				"delegate",
			},
			AddedIDs: map[string][]string{
				"delegate": []string{"user_1"},
			},
			RemovedIDs: map[string][]string{
				"delegate": []string{"user_2"},
			},
			ProposedChanges: map[string]any{
				"display_name": "Acme",
				"description":  nil,
			},
			EventID: "evt_123",
			Properties: map[string]string{
				"ID":                         "entity_123",
				soiree.PropertyEventID:       "evt_123",
				"changed_field_display_name": "true",
			},
			Auth: &MutationAuthContext{
				SubjectID:          testSubjectID,
				OrganizationID:     testOrgID,
				OrganizationIDs:    []string{testOrgID},
				AuthenticationType: string(auth.JWTAuthentication),
				OrganizationRole:   string(auth.OwnerRole),
			},
			OccurredAt: time.Now().UTC(),
		},
	}

	require.NoError(t, worker.Work(context.Background(), job))
	require.True(t, called)
}
