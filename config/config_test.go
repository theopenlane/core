package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/v2"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func writeConfigFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	assert.NilError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestConfigFilePaths(t *testing.T) {
	assert.Check(t, cmp.DeepEqual(configFilePaths("./config/.config.yaml"), []string{"./config/.config.yaml"}))
	assert.Check(t, cmp.DeepEqual(configFilePaths(" a.yaml , b.yaml ,,"), []string{"a.yaml", "b.yaml"}))
}

func TestLoadConfigFilesMergesInOrder(t *testing.T) {
	base := writeConfigFile(t, "base.yaml", "domain: base.example\nintegrations:\n  gemini:\n    model: base-model\n    location: us-east1\n")
	overlay := writeConfigFile(t, "overlay.yaml", "integrations:\n  gemini:\n    model: overlay-model\n")

	k := koanf.New(".")
	loadConfigFiles(k, []string{base, overlay})

	assert.Check(t, cmp.Equal(k.String("domain"), "base.example"))
	assert.Check(t, cmp.Equal(k.String("integrations.gemini.model"), "overlay-model"))
	assert.Check(t, cmp.Equal(k.String("integrations.gemini.location"), "us-east1"))
}

func TestGeminiPromptsFileLoadsAsConfig(t *testing.T) {
	k := koanf.New(".")
	loadConfigFiles(k, []string{"prompts/gemini.yaml"})

	assert.Check(t, k.String("integrations.gemini.prompts.systeminstruction") != "")
	assert.Check(t, k.String("integrations.gemini.prompts.soc2.sections.assessment") != "")
}
