package keystore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// LoadProviderSpecs loads integration provider specifications from a YAML file
func LoadProviderSpecs(path string) (map[string]ProviderSpec, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return loadProviderSpecsFromDir(path)
	}
	return loadProviderSpecsFromFile(path)
}

func loadProviderSpecsFromFile(path string) (map[string]ProviderSpec, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, err
	}
	var specs map[string]ProviderSpec
	if err := k.Unmarshal("integrations.providers", &specs); err != nil {
		return nil, err
	}
	return normalizeSpecNames(specs), nil
}

func loadProviderSpecsFromDir(dir string) (map[string]ProviderSpec, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	specs := make(map[string]ProviderSpec)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(dir, name)
		switch strings.ToLower(filepath.Ext(name)) {
		case ".json":
			spec, err := decodeSingleProviderSpec(path)
			if err != nil {
				return nil, fmt.Errorf("unmarshal provider spec %s: %w", path, err)
			}
			specs[strings.ToLower(spec.Name)] = spec
		case ".yaml", ".yml":
			fileSpecs, err := loadProviderSpecsFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("load provider spec %s: %w", path, err)
			}
			for key, spec := range fileSpecs {
				specs[strings.ToLower(key)] = spec
			}
		default:
			continue
		}
	}

	return normalizeSpecNames(specs), nil
}

func normalizeSpecNames(specs map[string]ProviderSpec) map[string]ProviderSpec {
	if specs == nil {
		return nil
	}
	for key, spec := range specs {
		if spec.Name == "" {
			spec.Name = key
		}
		specs[strings.ToLower(spec.Name)] = spec
	}
	return specs
}

func decodeSingleProviderSpec(path string) (ProviderSpec, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return ProviderSpec{}, err
	}
	var spec ProviderSpec
	if err := k.Unmarshal("", &spec); err != nil {
		return ProviderSpec{}, err
	}
	if spec.Name == "" {
		base := filepath.Base(path)
		spec.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return spec, nil
}
