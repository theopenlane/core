package runtime

import (
	"context"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/keymaker"
)

func TestResolveKeymakerInstallationRequiresInstallationID(t *testing.T) {
	t.Parallel()

	_, err := resolveKeymakerInstallation(context.Background(), "", func(context.Context, string, string, string) (*ent.Integration, error) {
		t.Fatal("unexpected resolver call")

		return nil, nil
	})
	if !errors.Is(err, keymaker.ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}
}

func TestResolveKeymakerInstallationMapsNotFound(t *testing.T) {
	t.Parallel()

	_, err := resolveKeymakerInstallation(context.Background(), "install-1", func(context.Context, string, string, string) (*ent.Integration, error) {
		return nil, ErrInstallationNotFound
	})
	if !errors.Is(err, keymaker.ErrInstallationNotFound) {
		t.Fatalf("expected ErrInstallationNotFound, got %v", err)
	}
}

func TestResolveKeymakerInstallationPassesThroughUnexpectedErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("db unavailable")

	_, err := resolveKeymakerInstallation(context.Background(), "install-1", func(context.Context, string, string, string) (*ent.Integration, error) {
		return nil, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestResolveKeymakerInstallationReturnsRecord(t *testing.T) {
	t.Parallel()

	record, err := resolveKeymakerInstallation(context.Background(), "install-1", func(_ context.Context, ownerID, integrationID, definitionID string) (*ent.Integration, error) {
		if ownerID != "" {
			t.Fatalf("expected empty ownerID, got %q", ownerID)
		}
		if integrationID != "install-1" {
			t.Fatalf("expected integrationID install-1, got %q", integrationID)
		}
		if definitionID != "" {
			t.Fatalf("expected empty definitionID, got %q", definitionID)
		}

		return &ent.Integration{
			ID:           integrationID,
			OwnerID:      "org-1",
			DefinitionID: "github-oauth",
		}, nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if record.ID != "install-1" {
		t.Fatalf("expected ID install-1, got %q", record.ID)
	}
	if record.OwnerID != "org-1" {
		t.Fatalf("expected OwnerID org-1, got %q", record.OwnerID)
	}
	if record.DefinitionID != "github-oauth" {
		t.Fatalf("expected DefinitionID github-oauth, got %q", record.DefinitionID)
	}
}
