package spec

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"

	"github.com/theopenlane/core/pkg/jsonx"
)

// FSLoader reads provider specs from an fs.FS rooted at the configured path
type FSLoader struct {
	// FS is the filesystem used to read provider specs
	FS fs.FS
	// Path is the relative directory containing provider files
	Path string
}

// NewFSLoader builds a loader using the supplied filesystem and relative path
func NewFSLoader(fsys fs.FS, root string) *FSLoader {
	return &FSLoader{
		FS:   fsys,
		Path: root,
	}
}

// Load walks the configured directory and decodes every JSON provider file into a slice of ProviderSpec
func (l *FSLoader) Load() ([]ProviderSpec, error) {
	if l.FS == nil {
		return nil, ErrFSLoaderNotConfigured
	}

	dirEntries, err := fs.ReadDir(l.FS, l.Path)
	if err != nil {
		return nil, &LoaderPathError{Err: ErrReadDirectory, Path: l.Path, Cause: err}
	}

	jsonEntries := lo.Filter(dirEntries, func(e fs.DirEntry, _ int) bool {
		return !e.IsDir() && parserForFile(e.Name()) != nil
	})

	specs := make([]ProviderSpec, 0, len(jsonEntries))

	for _, entry := range jsonEntries {
		fullPath := filepath.Join(l.Path, entry.Name())

		data, readErr := fs.ReadFile(l.FS, fullPath)
		if readErr != nil {
			return nil, &LoaderPathError{Err: ErrReadFile, Path: fullPath, Cause: readErr}
		}

		parser := parserForFile(entry.Name())

		spec, decodeErr := decodeProviderSpec(data, parser)
		if decodeErr != nil {
			return nil, &LoaderPathError{Err: ErrDecodeSpec, Path: fullPath, Cause: decodeErr}
		}

		spec.Name = lo.Ternary(spec.Name != "", spec.Name, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		if spec.SchemaVersion == "" {
			spec.SchemaVersion = DefaultSchemaVersion
		}

		if !spec.supportsSchemaVersion() {
			return nil, &SchemaVersionError{Path: fullPath, Version: spec.SchemaVersion}
		}

		authKind := spec.AuthType.Normalize()
		if !authKind.IsKnown() {
			return nil, &AuthTypeError{Path: fullPath, Value: string(spec.AuthType)}
		}

		spec.AuthType = authKind

		specs = append(specs, spec)
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
			DecodeHook:       defaultMapstructureDecodeHook(),
		},
	}

	if err := k.UnmarshalWithConf("", &spec, conf); err != nil {
		return ProviderSpec{}, err
	}

	return spec, nil
}

// defaultMapstructureDecodeHook composes the decode hooks used by spec decoders
func defaultMapstructureDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
		mapToRawMessageHook(),
	)
}

// mapToRawMessageHook returns a decode hook that converts map[string]any to json.RawMessage
func mapToRawMessageHook() mapstructure.DecodeHookFuncType {
	rawMessageType := reflect.TypeFor[json.RawMessage]()

	return func(_ reflect.Type, to reflect.Type, data any) (any, error) {
		if to != rawMessageType {
			return data, nil
		}

		raw, err := jsonx.ToRawMessage(data)
		if err != nil {
			return nil, err
		}

		return raw, nil
	}
}

// rawBytesProvider implements koanf.Provider for raw byte slices
type rawBytesProvider struct {
	// data is the raw bytes to provide
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
