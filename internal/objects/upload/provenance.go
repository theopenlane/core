package upload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
)

const (
	claimedSHA256Key = "artifact_sha256"
	// maxProvenanceBytes is the maximum JSON-encoded size of a file's provenance claims (16KB)
	maxProvenanceBytes = 16 * 1024
)

// validateProvenance checks a file's client claims before they are stored immutably, rejecting any
// claimed artifact_sha256 that does not match the received bytes; files without claims always pass
func validateProvenance(f pkgobjects.File) error {
	if f.ProvenanceClaims == nil {
		return nil
	}

	// trust center files are readable anonymously, so client capture records must not be attached to them
	if f.CorrelatedObjectType != ent.TypeEvidence {
		return ErrProvenanceNotSupported
	}

	encoded, err := json.Marshal(f.ProvenanceClaims)
	if err != nil {
		return err
	}

	if len(encoded) > maxProvenanceBytes {
		return fmt.Errorf("%w: %d bytes exceeds the %d byte limit", ErrProvenanceTooLarge, len(encoded), maxProvenanceBytes)
	}

	if len(f.SHA256) == 0 {
		return ErrProvenanceWithoutContent
	}

	return verifyClaimedSHA256(f.ProvenanceClaims, string(f.SHA256))
}

// verifyClaimedSHA256 rejects a claimed hash that is malformed or differs from the received bytes,
// so a self-contradicting record is never stored
func verifyClaimedSHA256(claims map[string]any, serverSHA256 string) error {
	raw, ok := claims[claimedSHA256Key]
	if !ok || raw == nil {
		return nil
	}

	claimed, ok := raw.(string)
	if !ok || !isSHA256Hex(claimed) {
		return ErrInvalidProvenanceHash
	}

	if !strings.EqualFold(claimed, serverSHA256) {
		return ErrProvenanceHashMismatch
	}

	return nil
}

func isSHA256Hex(s string) bool {
	if len(s) != hex.EncodedLen(sha256.Size) {
		return false
	}

	_, err := hex.DecodeString(s)

	return err == nil
}
