package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"
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

		spec, decodeErr := decodeProviderSpec(bytes, lookup)
		if decodeErr != nil {
			return nil, fmt.Errorf("integrations/config: decode %q: %w", fullPath, decodeErr)
		}

		if !spec.Active {
			continue
		}

		spec.Name = lo.Ternary(spec.Name != "", spec.Name, strings.TrimSuffix(entry.Name(), ".json"))
		if spec.SchemaVersion == "" {
			spec.SchemaVersion = DefaultSchemaVersion
		}

		if !spec.supportsSchemaVersion() {
			return nil, fmt.Errorf("%w: %s (declared %q)", integrations.ErrSchemaVersionUnsupported, fullPath, spec.SchemaVersion)
		}

		pt := spec.ProviderType()

		specs[pt] = spec
	}

	return specs, nil
}

func decodeProviderSpec(data []byte, lookup EnvLookup) (ProviderSpec, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return ProviderSpec{}, err
	}

	k := koanf.New(".")
	if err := k.Load(mapProvider{data: raw}, nil); err != nil {
		return ProviderSpec{}, err
	}

	var spec ProviderSpec
	conf := koanf.UnmarshalConf{
		Tag: "json",
		DecoderConfig: &mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				envInterpolationHook(lookup),
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
			),
		},
	}

	if err := k.UnmarshalWithConf("", &spec, conf); err != nil {
		return ProviderSpec{}, err
	}

	if err := interpolateSchemaMaps(&spec, lookup); err != nil {
		return ProviderSpec{}, err
	}

	return spec, nil
}

type mapProvider struct {
	data map[string]any
}

func (p mapProvider) Read() (map[string]any, error) {
	return p.data, nil
}

func (p mapProvider) ReadBytes() ([]byte, error) {
	return json.Marshal(p.data)
}

func envInterpolationHook(lookup EnvLookup) mapstructure.DecodeHookFuncType {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	return func(from reflect.Type, to reflect.Type, value any) (any, error) {
		str, ok := value.(string)
		if !ok || !strings.Contains(str, "${") {
			return value, nil
		}

		replaced, err := interpolateEnvString(str, lookup)
		if err != nil {
			return nil, err
		}

		return replaced, nil
	}
}

var envPlaceholderExpr = regexp.MustCompile(`\$\{[A-Za-z0-9_]+\}`)

func interpolateEnvString(value string, lookup EnvLookup) (string, error) {
	if value == "" || lookup == nil || !strings.Contains(value, "${") {
		return value, nil
	}

	var interpolationErr error
	replaced := envPlaceholderExpr.ReplaceAllStringFunc(value, func(segment string) string {
		if interpolationErr != nil {
			return segment
		}

		key := strings.TrimSuffix(strings.TrimPrefix(segment, "${"), "}")
		if key == "" {
			interpolationErr = fmt.Errorf("integrations/config: empty env placeholder in %q", value)
			return segment
		}

		resolved, ok := lookup(key)
		if !ok {
			interpolationErr = fmt.Errorf("%w: %s", integrations.ErrEnvVarNotDefined, key)
			return segment
		}

		return resolved
	})

	if interpolationErr != nil {
		return "", interpolationErr
	}

	return replaced, nil
}

func interpolateSchemaMaps(spec *ProviderSpec, lookup EnvLookup) error {
	if spec == nil || lookup == nil {
		return nil
	}
	if err := interpolateAnyMap(spec.Metadata, lookup); err != nil {
		return err
	}
	if err := interpolateAnyMap(spec.Defaults, lookup); err != nil {
		return err
	}
	if err := interpolateAnyMap(spec.CredentialsSchema, lookup); err != nil {
		return err
	}
	return nil
}

func interpolateAnyMap(values map[string]any, lookup EnvLookup) error {
	if len(values) == 0 || lookup == nil {
		return nil
	}
	for key, raw := range values {
		interpolated, err := interpolateAnyValue(raw, lookup)
		if err != nil {
			return err
		}
		values[key] = interpolated
	}
	return nil
}

func interpolateAnySlice(values []any, lookup EnvLookup) error {
	if len(values) == 0 || lookup == nil {
		return nil
	}
	for idx, raw := range values {
		interpolated, err := interpolateAnyValue(raw, lookup)
		if err != nil {
			return err
		}
		values[idx] = interpolated
	}
	return nil
}

func interpolateAnyValue(value any, lookup EnvLookup) (any, error) {
	switch typed := value.(type) {
	case string:
		return interpolateEnvString(typed, lookup)
	case map[string]any:
		if err := interpolateAnyMap(typed, lookup); err != nil {
			return nil, err
		}
		return typed, nil
	case []any:
		if err := interpolateAnySlice(typed, lookup); err != nil {
			return nil, err
		}
		return typed, nil
	default:
		return value, nil
	}
}
