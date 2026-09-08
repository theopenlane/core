package operations

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestRoundTripUpdateInput(t *testing.T) {
	t.Parallel()

	type createInput struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type updateInput struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("successful round trip", func(t *testing.T) {
		t.Parallel()

		create := createInput{Name: "test", Value: 42}
		got, err := roundTripUpdateInput[createInput, updateInput](create)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "test" || got.Value != 42 {
			t.Fatalf("got %+v, want Name=test Value=42", got)
		}
	})

	t.Run("partial field overlap", func(t *testing.T) {
		t.Parallel()

		type partialUpdate struct {
			Name string `json:"name"`
		}

		create := createInput{Name: "test", Value: 42}
		got, err := roundTripUpdateInput[createInput, partialUpdate](create)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "test" {
			t.Fatalf("got Name=%q, want %q", got.Name, "test")
		}
	})
}

func TestPersistLookupUpsert(t *testing.T) {
	t.Parallel()

	type fakeRow struct {
		ID string `json:"id"`
	}

	type input struct {
		Name string `json:"name"`
	}

	integration := &ent.Integration{OwnerID: "org-1"}

	t.Run("absent rows route through the catalog create", func(t *testing.T) {
		t.Parallel()

		schema := &entityops.Schema{
			SchemaDescriptor: entityops.SchemaDescriptor{Name: "Widget", Snake: "widget"},
			Create: func(_ context.Context, _ *ent.Client, payload json.RawMessage) (string, error) {
				if !strings.Contains(string(payload), `"name":"x"`) {
					t.Fatalf("expected the create payload, got %s", payload)
				}
				return "new-id", nil
			},
		}

		id, _, _, err := persistLookupUpsert(context.Background(), nil, schema, integration, input{Name: "x"}, func(context.Context) (fakeRow, error) {
			return fakeRow{}, &ent.NotFoundError{}
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "new-id" {
			t.Fatalf("expected id=new-id, got %q", id)
		}
	})

	t.Run("lookup failures wrap as persist errors", func(t *testing.T) {
		t.Parallel()

		schema := &entityops.Schema{SchemaDescriptor: entityops.SchemaDescriptor{Name: "Widget", Snake: "widget"}}

		_, _, _, err := persistLookupUpsert(context.Background(), nil, schema, integration, input{Name: "x"}, func(context.Context) (fakeRow, error) {
			return fakeRow{}, errors.New("connection lost")
		})
		if !errors.Is(err, ErrIngestPersistFailed) {
			t.Fatalf("expected ErrIngestPersistFailed, got %v", err)
		}
	})
}
