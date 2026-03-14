package emailruntime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/newman/compose"
)

func TestNewTemplateData_EmptyBuildSucceeds(t *testing.T) {
	data, err := NewTemplateData().Build(compose.Config{}, compose.Recipient{})
	require.NoError(t, err)
	require.NotNil(t, data)
}

func TestTemplateData_WithField_SetsKey(t *testing.T) {
	data, err := NewTemplateData().
		WithField("Name", "Ada").
		Build(compose.Config{}, compose.Recipient{})

	require.NoError(t, err)
	assert.Equal(t, "Ada", data["Name"])
}

func TestTemplateData_WithFields_MergesKeys(t *testing.T) {
	data, err := NewTemplateData().
		WithFields(map[string]any{
			"Alpha": "one",
			"Beta":  42,
		}).
		Build(compose.Config{}, compose.Recipient{})

	require.NoError(t, err)
	assert.Equal(t, "one", data["Alpha"])
	assert.Equal(t, 42, data["Beta"])
}

func TestTemplateData_WithField_OverwritesPreviousValue(t *testing.T) {
	data, err := NewTemplateData().
		WithField("Key", "first").
		WithField("Key", "second").
		Build(compose.Config{}, compose.Recipient{})

	require.NoError(t, err)
	assert.Equal(t, "second", data["Key"])
}

func TestTemplateData_WithURL_BuildsWithoutError(t *testing.T) {
	_, err := NewTemplateData().
		WithURL(TemplateURLVerify, "https://example.com/verify").
		Build(compose.Config{}, compose.Recipient{})

	require.NoError(t, err)
}

func TestTemplateData_WithTokenURL_UnknownKeyReturnsError(t *testing.T) {
	_, err := NewTemplateData().
		WithTokenURL(TemplateURLKey("unknown_custom_key_xyz"), "token123").
		Build(compose.Config{}, compose.Recipient{})

	require.ErrorIs(t, err, ErrUnsupportedTemplateURLKey)
}

func TestTemplateData_WithFieldsAndURL_BothPresent(t *testing.T) {
	data, err := NewTemplateData().
		WithField("Greeting", "Hello").
		WithURL(TemplateURLVerify, "https://example.com/verify/abc").
		Build(compose.Config{}, compose.Recipient{})

	require.NoError(t, err)
	assert.Equal(t, "Hello", data["Greeting"])
}
