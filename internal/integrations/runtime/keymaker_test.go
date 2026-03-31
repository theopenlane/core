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

	_, err := resolveKeymakerInstallation(context.Background(), "", func(context.Context, IntegrationLookup) (*ent.Integration, error) {
		t.Fatal("unexpected resolver call")

		return nil, nil
	})
	if !errors.Is(err, keymaker.ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}
}

func TestResolveKeymakerInstallationMapsNotFound(t *testing.T) {
	t.Parallel()

	_, err := resolveKeymakerInstallation(context.Background(), "install-1", func(context.Context, IntegrationLookup) (*ent.Integration, error) {
		return nil, ErrInstallationNotFound
	})
	if !errors.Is(err, keymaker.ErrInstallationNotFound) {
		t.Fatalf("expected ErrInstallationNotFound, got %v", err)
	}
}

func TestResolveKeymakerInstallationPassesThroughUnexpectedErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("db unavailable")

	_, err := resolveKeymakerInstallation(context.Background(), "install-1", func(context.Context, IntegrationLookup) (*ent.Integration, error) {
		return nil, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestResolveKeymakerInstallationReturnsRecord(t *testing.T) {
	t.Parallel()

	record, err := resolveKeymakerInstallation(context.Background(), "install-1", func(_ context.Context, lookup IntegrationLookup) (*ent.Integration, error) {
		if lookup.OwnerID != "" {
			t.Fatalf("expected empty OwnerID, got %q", lookup.OwnerID)
		}
		if lookup.IntegrationID != "install-1" {
			t.Fatalf("expected IntegrationID install-1, got %q", lookup.IntegrationID)
		}
		if lookup.DefinitionID != "" {
			t.Fatalf("expected empty DefinitionID, got %q", lookup.DefinitionID)
		}

		return &ent.Integration{
			ID:           lookup.IntegrationID,
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
