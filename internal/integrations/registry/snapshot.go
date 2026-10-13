package registry

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// surfacesDir is the directory within a snapshot filesystem holding one snapshot per definition
	surfacesDir = "surfaces"
	// snapshotExtension is the file extension of every committed snapshot
	snapshotExtension = ".json"
)

// Surfaces is the committed surface snapshot of every catalog definition, embedded under surfaces/
//
//go:embed surfaces/*.json
var Surfaces embed.FS

// committedVersion returns the version of def's committed snapshot in fsys, requiring its hash to match def's current surface
func committedVersion(fsys fs.FS, def types.Definition) (string, error) {
	hash, err := SurfaceHash(def)
	if err != nil {
		return "", err
	}

	data, err := fs.ReadFile(fsys, path.Join(surfacesDir, def.ID+snapshotExtension))

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "", fmt.Errorf("%w: definition %s has no committed snapshot", ErrSnapshotStale, def.ID)
	case err != nil:
		return "", err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return "", fmt.Errorf("decode snapshot %s: %w", def.ID, err)
	}

	if snapshot.Hash != hash {
		return "", fmt.Errorf("%w: definition %s", ErrSnapshotStale, def.ID)
	}

	return snapshot.Version, nil
}
