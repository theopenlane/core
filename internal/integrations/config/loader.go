package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

// FSLoader reads provider specs from an fs.FS rooted at the configured path
type FSLoader struct {
	// FS is the filesystem used to read provider specs
	FS fs.FS
	// Path is the relative directory containing provider files
	Path string
}

// NewFSLoader builds a loader using the supplied filesystem and relative path
func NewFSLoader(fsys fs.FS, path string) *FSLoader {
	return &FSLoader{
		FS:   fsys,
		Path: path,
	}
}

// Load walks the configured directory and decodes every JSON provider file
func (l *FSLoader) Load(_ context.Context) (map[types.ProviderType]ProviderSpec, error) {
	if l == nil || l.FS == nil {
		return nil, fmt.Errorf("%w: %w", integrations.ErrLoaderRequired, ErrFSLoaderNotConfigured)
	}

	dirEntries, err := fs.ReadDir(l.FS, l.Path)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %w", ErrReadDirectory, l.Path, err)
	}

	specs := make(map[types.ProviderType]ProviderSpec, len(dirEntries))

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		parser := parserForFile(entry.Name())
		if parser == nil {
			continue
		}

		fullPath := filepath.Join(l.Path, entry.Name())
		bytes, readErr := fs.ReadFile(l.FS, fullPath)
		if readErr != nil {
			return nil, fmt.Errorf("%w %q: %w", ErrReadFile, fullPath, readErr)
		}

		spec, decodeErr := decodeProviderSpec(bytes, parser)
		if decodeErr != nil {
			return nil, fmt.Errorf("%w %q: %w", ErrDecodeSpec, fullPath, decodeErr)
		}

		if !spec.Active {
			continue
		}

		spec.Name = lo.Ternary(spec.Name != "", spec.Name, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
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

// decodeProviderSpec unmarshals provider spec data using the specified parser
func decodeProviderSpec(data []byte, parser koanf.Parser) (ProviderSpec, error) {
	k := koanf.New(".")
	if parser == nil {
		parser = jsonParser{}
	}

	if err := k.Load(rawBytesProvider{data: data}, parser); err != nil {
		return ProviderSpec{}, err
	}

	var spec ProviderSpec
	conf := koanf.UnmarshalConf{
		Tag: "json",
		DecoderConfig: &mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
			),
		},
	}

	if err := k.UnmarshalWithConf("", &spec, conf); err != nil {
		return ProviderSpec{}, err
	}

	return spec, nil
}

// rawBytesProvider implements koanf.Provider for raw byte slices
type rawBytesProvider struct {
	data []byte
}

// Read is not supported by rawBytesProvider
func (p rawBytesProvider) Read() (map[string]any, error) {
	return nil, ErrRawBytesProviderRead
}

// ReadBytes returns the raw byte data for parsing
func (p rawBytesProvider) ReadBytes() ([]byte, error) {
	return p.data, nil
}

// jsonParser implements koanf.Parser for JSON data
type jsonParser struct{}

// Unmarshal decodes JSON bytes into a map
func (jsonParser) Unmarshal(bytes []byte) (map[string]any, error) {
	var out map[string]any
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Marshal encodes a map into JSON bytes
func (jsonParser) Marshal(value map[string]any) ([]byte, error) {
	return json.Marshal(value)
}

// parserForFile selects the appropriate parser based on file extension
func parserForFile(name string) koanf.Parser {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".yaml", ".yml":
		return yaml.Parser()
	case ".json":
		return jsonParser{}
	default:
		return nil
	}
}
