package email

// MetadataKey identifies keys used in EmailTemplate.Metadata
type MetadataKey string

// String returns the metadata key as a string
func (k MetadataKey) String() string {
	return string(k)
}

const (
	// MetadataKeyRenderMode identifies template rendering mode in EmailTemplate.Metadata
	MetadataKeyRenderMode MetadataKey = "render_mode"
	// MetadataKeyBaseTemplate identifies a system base email template key to use for RAW_HTML assembly.
	// When set, the customer body_template is composed with the named system base template at render time
	MetadataKeyBaseTemplate MetadataKey = "base_template_key"
)

// RenderMode controls how an email template body is rendered
type RenderMode string

// String returns the render mode as a string
func (m RenderMode) String() string {
	return string(m)
}

// IsValid returns true when the mode is supported by the email renderer
func (m RenderMode) IsValid() bool {
	return m == RenderModeRawHTML
}

const (
	// RenderModeRawHTML renders body_template as a single HTML template.
	// If MetadataKeyBaseTemplate is set, the body_template is composed with a system base template
	RenderModeRawHTML RenderMode = "RAW_HTML"
	// BaseTemplateEntrypoint is the named template executed when assembling RAW_HTML with a base template.
	// The base template must define a block named "content" that the customer body_template overrides
	BaseTemplateEntrypoint = "base"
)

// RenderMetadata holds typed render configuration decoded from EmailTemplate.Metadata
type RenderMetadata struct {
	// Mode controls html/body rendering behavior
	Mode RenderMode
	// BaseTemplateKey defines the system base email template key used for raw-html assembly
	BaseTemplateKey string
}

// EffectiveRenderMode applies defaults when Mode is unset or invalid
func (m RenderMetadata) EffectiveRenderMode() RenderMode {
	if m.Mode.IsValid() {
		return m.Mode
	}

	return RenderModeRawHTML
}

// DecodeRenderMetadata decodes typed renderer options from EmailTemplate.Metadata
func DecodeRenderMetadata(metadata map[string]any) RenderMetadata {
	if len(metadata) == 0 {
		return RenderMetadata{}
	}

	return RenderMetadata{
		Mode:            RenderMode(readMetadataValue(metadata, MetadataKeyRenderMode)),
		BaseTemplateKey: readMetadataValue(metadata, MetadataKeyBaseTemplate),
	}
}

// readMetadataValue reads a string value from metadata by typed key
func readMetadataValue(metadata map[string]any, key MetadataKey) string {
	value, ok := metadata[key.String()]
	if !ok {
		return ""
	}

	typed, ok := value.(string)
	if !ok {
		return ""
	}

	return typed
}
