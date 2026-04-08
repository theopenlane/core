package operations

import (
	"context"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
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

func TestPersistUpsert_CreatePath(t *testing.T) {
	t.Parallel()

	type input struct {
		Name string
	}

	created := false

	err := persistUpsert(
		context.Background(),
		input{Name: "new"},
		func(in input) (input, error) { return in, nil },
		func(context.Context) (any, error) {
			return nil, &ent.NotFoundError{}
		},
		func(_ context.Context, in input) error {
			created = true
			if in.Name != "new" {
				t.Fatalf("create input Name=%q, want %q", in.Name, "new")
			}
			return nil
		},
		func(context.Context, any, input) error {
			t.Fatal("update should not be called on create path")
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected create to be called")
	}
}

func TestPersistUpsert_UpdatePath(t *testing.T) {
	t.Parallel()

	type input struct {
		Name string
	}

	updated := false

	err := persistUpsert(
		context.Background(),
		input{Name: "existing"},
		func(in input) (input, error) { return in, nil },
		func(context.Context) (string, error) {
			return "existing-id", nil
		},
		func(context.Context, input) error {
			t.Fatal("create should not be called on update path")
			return nil
		},
		func(_ context.Context, existing string, in input) error {
			updated = true
			if existing != "existing-id" {
				t.Fatalf("existing=%q, want %q", existing, "existing-id")
			}
			if in.Name != "existing" {
				t.Fatalf("update input Name=%q, want %q", in.Name, "existing")
			}
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Fatal("expected update to be called")
	}
}

func TestPersistUpsert_FindExistingError(t *testing.T) {
	t.Parallel()

	type input struct{ Name string }

	dbErr := errors.New("database connection lost")

	err := persistUpsert(
		context.Background(),
		input{Name: "test"},
		func(in input) (input, error) { return in, nil },
		func(context.Context) (any, error) {
			return nil, dbErr
		},
		func(context.Context, input) error {
			t.Fatal("create should not be called")
			return nil
		},
		func(context.Context, any, input) error {
			t.Fatal("update should not be called")
			return nil
		},
	)

	if !errors.Is(err, ErrIngestPersistFailed) {
		t.Fatalf("expected ErrIngestPersistFailed, got %v", err)
	}
}

func TestPersistUpsert_ToUpdateError(t *testing.T) {
	t.Parallel()

	type input struct{ Name string }

	toUpdateErr := errors.New("conversion failed")

	err := persistUpsert(
		context.Background(),
		input{Name: "test"},
		func(input) (input, error) { return input{}, toUpdateErr },
		func(context.Context) (string, error) {
			return "exists", nil
		},
		func(context.Context, input) error {
			t.Fatal("create should not be called")
			return nil
		},
		func(context.Context, string, input) error {
			t.Fatal("update should not be called")
			return nil
		},
	)

	if !errors.Is(err, toUpdateErr) {
		t.Fatalf("expected toUpdateErr, got %v", err)
	}
}

func TestPersistUpsert_CreateError(t *testing.T) {
	t.Parallel()

	type input struct{ Name string }

	err := persistUpsert(
		context.Background(),
		input{Name: "test"},
		func(in input) (input, error) { return in, nil },
		func(context.Context) (any, error) {
			return nil, &ent.NotFoundError{}
		},
		func(context.Context, input) error {
			return &ent.ValidationError{Name: "name"}
		},
		func(context.Context, any, input) error {
			t.Fatal("update should not be called")
			return nil
		},
	)

	if !errors.Is(err, ErrIngestMappedDocumentInvalid) {
		t.Fatalf("expected ErrIngestMappedDocumentInvalid, got %v", err)
	}
}

func TestPersistRoundTripUpsert_CreatePath(t *testing.T) {
	t.Parallel()

	type createInput struct {
		Name string `json:"name"`
	}
	type updateInput struct {
		Name string `json:"name"`
	}

	created := false

	err := persistRoundTripUpsert(
		context.Background(),
		createInput{Name: "new"},
		func(context.Context) (any, error) {
			return nil, &ent.NotFoundError{}
		},
		func(_ context.Context, in createInput) error {
			created = true
			if in.Name != "new" {
				t.Fatalf("input Name=%q, want %q", in.Name, "new")
			}
			return nil
		},
		func(context.Context, any, updateInput) error {
			t.Fatal("update should not be called")
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected create to be called")
	}
}

func TestPersistRoundTripUpsert_UpdatePath(t *testing.T) {
	t.Parallel()

	type createInput struct {
		Name string `json:"name"`
	}
	type updateInput struct {
		Name string `json:"name"`
	}

	updated := false

	err := persistRoundTripUpsert(
		context.Background(),
		createInput{Name: "existing"},
		func(context.Context) (string, error) {
			return "existing-id", nil
		},
		func(context.Context, createInput) error {
			t.Fatal("create should not be called")
			return nil
		},
		func(_ context.Context, existing string, in updateInput) error {
			updated = true
			if in.Name != "existing" {
				t.Fatalf("update input Name=%q, want %q", in.Name, "existing")
			}
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated {
		t.Fatal("expected update to be called")
	}
}

func TestPersistUpsert_UpdateError(t *testing.T) {
	t.Parallel()

	type input struct{ Name string }

	err := persistUpsert(
		context.Background(),
		input{Name: "test"},
		func(in input) (input, error) { return in, nil },
		func(context.Context) (string, error) {
			return "exists", nil
		},
		func(context.Context, input) error {
			t.Fatal("create should not be called")
			return nil
		},
		func(_ context.Context, _ string, _ input) error {
			return &ent.ConstraintError{}
		},
	)

	if !errors.Is(err, ErrIngestUpsertConflict) {
		t.Fatalf("expected ErrIngestUpsertConflict, got %v", err)
	}
}
