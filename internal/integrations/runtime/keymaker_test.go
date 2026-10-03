package runtime

import (
	"errors"
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/keymaker"
)

func TestLookupKeymakerInstallationRequiresInstallationID(t *testing.T) {
	t.Parallel()

	_, err := NewForTesting(registry.New()).lookupKeymakerInstallation(t.Context(), "")
	if !errors.Is(err, keymaker.ErrInstallationIDRequired) {
		t.Fatalf("expected ErrInstallationIDRequired, got %v", err)
	}
}

func TestKeymakerRecordMapsNotFound(t *testing.T) {
	t.Parallel()

	_, err := keymakerRecord(nil, &ent.NotFoundError{})
	if !errors.Is(err, keymaker.ErrInstallationNotFound) {
		t.Fatalf("expected ErrInstallationNotFound, got %v", err)
	}
}

func TestKeymakerRecordPassesThroughUnexpectedErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("db unavailable")

	_, err := keymakerRecord(nil, expectedErr)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestKeymakerRecordReturnsRecord(t *testing.T) {
	t.Parallel()

	record, err := keymakerRecord(&ent.Integration{ID: "install-1", OwnerID: "org-1", DefinitionID: "github-oauth"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if record.ID != "install-1" || record.OwnerID != "org-1" || record.DefinitionID != "github-oauth" {
		t.Fatalf("unexpected record %+v", record)
	}
}
