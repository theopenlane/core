package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

// EnvLookup resolves an environment variable placeholder.
type EnvLookup func(string) (string, bool)

// FSLoader reads provider specs from an fs.FS rooted at the configured path
type FSLoader struct {
	// FS is the filesystem used to read provider specs
	FS fs.FS
	// Path is the relative directory containing provider files
	Path string
	// EnvLookup resolves environment variable placeholders found in specs
	EnvLookup EnvLookup
}

// NewFSLoader builds a loader using the supplied filesystem and relative path
func NewFSLoader(fsys fs.FS, path string) *FSLoader {
	return &FSLoader{
		FS:        fsys,
		Path:      path,
		EnvLookup: os.LookupEnv,
	}
}

func (l *FSLoader) lookupFn() EnvLookup {
	if l == nil || l.EnvLookup == nil {
		return os.LookupEnv
	}
	return l.EnvLookup
}

// Load walks the configured directory and decodes every JSON provider file
func (l *FSLoader) Load(ctx context.Context) (map[types.ProviderType]ProviderSpec, error) {
	if l == nil || l.FS == nil {
		return nil, fmt.Errorf("%w: fs loader not configured", integrations.ErrLoaderRequired)
	}

	dirEntries, err := fs.ReadDir(l.FS, l.Path)
	if err != nil {
		return nil, fmt.Errorf("integrations/config: read dir %q: %w", l.Path, err)
	}

	specs := make(map[types.ProviderType]ProviderSpec, len(dirEntries))
	lookup := l.lookupFn()

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fullPath := filepath.Join(l.Path, entry.Name())
		bytes, readErr := fs.ReadFile(l.FS, fullPath)
		if readErr != nil {
			return nil, fmt.Errorf("integrations/config: read %q: %w", fullPath, readErr)
		}

		var spec ProviderSpec
		if decodeErr := json.Unmarshal(bytes, &spec); decodeErr != nil {
			return nil, fmt.Errorf("integrations/config: decode %q: %w", fullPath, decodeErr)
		}

		if !spec.Active {
			continue
		}

		spec.Name = lo.Ternary(spec.Name != "", spec.Name, strings.TrimSuffix(entry.Name(), ".json"))
		if spec.SchemaVersion == "" {
			spec.SchemaVersion = DefaultSchemaVersion
		}

		if err := spec.interpolatePlaceholders(lookup); err != nil {
			return nil, fmt.Errorf("integrations/config: interpolate %q: %w", fullPath, err)
		}
		if !spec.supportsSchemaVersion() {
			return nil, fmt.Errorf("%w: %s (declared %q)", integrations.ErrSchemaVersionUnsupported, fullPath, spec.SchemaVersion)
		}

		pt := spec.ProviderType()

		specs[pt] = spec
	}

	return specs, nil
}
