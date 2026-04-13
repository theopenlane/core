package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeRenderMetadata_Defaults(t *testing.T) {
	cfg := DecodeRenderMetadata(nil)

	assert.Equal(t, RenderModeRawHTML, cfg.EffectiveRenderMode())
	assert.Equal(t, "", cfg.BaseTemplateKey)
}

func TestDecodeRenderMetadata_ExplicitValues(t *testing.T) {
	cfg := DecodeRenderMetadata(map[string]any{
		MetadataKeyRenderMode.String():   RenderModeRawHTML.String(),
		MetadataKeyBaseTemplate.String(): "base-template",
	})

	assert.Equal(t, RenderModeRawHTML, cfg.EffectiveRenderMode())
	assert.Equal(t, "base-template", cfg.BaseTemplateKey)
}

func TestDecodeRenderMetadata_NonStringValueIgnored(t *testing.T) {
	cfg := DecodeRenderMetadata(map[string]any{
		MetadataKeyRenderMode.String(): 42,
	})

	assert.Equal(t, RenderModeRawHTML, cfg.EffectiveRenderMode())
}

func TestRenderModeString_ReturnsRawValue(t *testing.T) {
	assert.Equal(t, "RAW_HTML", RenderModeRawHTML.String())
}

func TestRenderModeIsValid_KnownModes(t *testing.T) {
	assert.True(t, RenderModeRawHTML.IsValid())
}

func TestRenderModeIsValid_Unknown(t *testing.T) {
	assert.False(t, RenderMode("UNKNOWN").IsValid())
	assert.False(t, RenderMode("").IsValid())
}

func TestMetadataKeyString_ReturnsRawValue(t *testing.T) {
	assert.Equal(t, "render_mode", MetadataKeyRenderMode.String())
}

func TestEffectiveRenderMode_ExplicitModeRespected(t *testing.T) {
	cfg := RenderMetadata{
		Mode: RenderModeRawHTML,
	}
	assert.Equal(t, RenderModeRawHTML, cfg.EffectiveRenderMode())
}
