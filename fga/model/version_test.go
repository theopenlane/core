package model

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// markerHashLength is the number of hex characters the version marker keeps, it must match the
// cut in task fga:generate:version
const markerHashLength = 12

// versionFile is the generated file holding the marker, excluded from the hash it carries
const versionFile = "generated/version.fga"

// modFileEntry matches a model file entry in fga.mod
var modFileEntry = regexp.MustCompile(`(?m)^\s+- (\S+)$`)

func TestVersionMarker(t *testing.T) {
	marker, err := VersionMarker()
	assert.NilError(t, err)
	assert.Check(t, strings.HasPrefix(marker, "model_version_"))
}

// TestVersionMarkerIsCurrent fails when a model file changed without the marker being regenerated
// a stale marker means the deployed model would never be recognised as out of date, so the model
// change would silently never reach the store
func TestVersionMarkerIsCurrent(t *testing.T) {
	marker, err := VersionMarker()
	assert.NilError(t, err)

	assert.Check(t, is.Equal("model_version_"+modelFilesHash(t), marker),
		"fga model files changed without regenerating the marker, run: task fga:generate:version")
}

// modelFilesHash mirrors task fga:generate:version, hashing every model file listed in fga.mod
// except the generated version file itself
func modelFilesHash(t *testing.T) string {
	t.Helper()

	mod, err := os.ReadFile("fga.mod")
	assert.NilError(t, err)

	var files []string

	for _, match := range modFileEntry.FindAllStringSubmatch(string(mod), -1) {
		if match[1] == versionFile {
			continue
		}

		files = append(files, match[1])
	}

	assert.Assert(t, len(files) > 0, "no model files found in fga.mod")

	sort.Strings(files)

	digest := sha256.New()

	for _, file := range files {
		contents, err := os.ReadFile(filepath.Clean(file))
		assert.NilError(t, err)

		_, err = digest.Write(contents)
		assert.NilError(t, err)
	}

	return hex.EncodeToString(digest.Sum(nil))[:markerHashLength]
}
