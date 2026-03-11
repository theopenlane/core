package emailruntime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/newman"
)

func TestWithTags_AppendsToRequest(t *testing.T) {
	req := &ComposeRequest{}
	tag := newman.Tag{Name: "campaign_id", Value: "abc"}

	WithTags(tag)(req)

	assert.Len(t, req.Tags, 1)
	assert.Equal(t, tag, req.Tags[0])
}

func TestWithTags_AppendsMultiple(t *testing.T) {
	req := &ComposeRequest{
		Tags: []newman.Tag{{Name: "existing", Value: "x"}},
	}
	t1 := newman.Tag{Name: "a", Value: "1"}
	t2 := newman.Tag{Name: "b", Value: "2"}

	WithTags(t1, t2)(req)

	assert.Len(t, req.Tags, 3)
}

func TestWithReplyTo_SetsAddress(t *testing.T) {
	req := &ComposeRequest{}

	WithReplyTo("reply@example.com")(req)

	assert.Equal(t, "reply@example.com", req.ReplyTo)
}

func TestWithAttachments_AppendsToRequest(t *testing.T) {
	req := &ComposeRequest{}
	a := newman.NewAttachment("file.pdf", []byte("content"))

	WithAttachments(a)(req)

	assert.Len(t, req.Attachments, 1)
	assert.Equal(t, a, req.Attachments[0])
}

func TestWithAttachments_AppendsToExisting(t *testing.T) {
	existing := newman.NewAttachment("existing.pdf", []byte("x"))
	req := &ComposeRequest{Attachments: []*newman.Attachment{existing}}
	a := newman.NewAttachment("new.pdf", []byte("y"))

	WithAttachments(a)(req)

	assert.Len(t, req.Attachments, 2)
}

func TestWithHeaders_SetsHeaders(t *testing.T) {
	req := &ComposeRequest{}

	WithHeaders(map[string]string{"X-Custom": "value"})(req)

	assert.Equal(t, "value", req.Headers["X-Custom"])
}

func TestWithHeaders_MergesIntoExisting(t *testing.T) {
	req := &ComposeRequest{Headers: map[string]string{"X-A": "1"}}

	WithHeaders(map[string]string{"X-B": "2"})(req)

	assert.Equal(t, "1", req.Headers["X-A"])
	assert.Equal(t, "2", req.Headers["X-B"])
}

func TestWithHeaders_InitializesNilMap(t *testing.T) {
	req := &ComposeRequest{}

	WithHeaders(map[string]string{"X-Init": "yes"})(req)

	assert.NotNil(t, req.Headers)
	assert.Equal(t, "yes", req.Headers["X-Init"])
}

func TestWithOwnerOnly_SetsFlag(t *testing.T) {
	req := &ComposeRequest{}

	WithOwnerOnly()(req)

	assert.True(t, req.OwnerOnly)
}

func TestTemplateRefValidate_Missing(t *testing.T) {
	ref := TemplateRef{}

	err := ref.Validate()
	assert.ErrorIs(t, err, ErrMissingTemplateReference)
}

func TestTemplateRefValidate_Conflict(t *testing.T) {
	ref := TemplateRef{
		ID:  "template-id",
		Key: TemplateKeyVerifyEmail,
	}

	err := ref.Validate()
	assert.ErrorIs(t, err, ErrTemplateReferenceConflict)
}

func TestTemplateRefValidate_KeyOnly(t *testing.T) {
	ref := TemplateRef{
		Key: ParseTemplateKey(" verify_email "),
	}

	err := ref.Validate()
	assert.NoError(t, err)
}

func TestDecodeRenderMetadata_Defaults(t *testing.T) {
	cfg := DecodeRenderMetadata(nil)

	assert.Equal(t, RenderModeRawHTML, cfg.EffectiveRenderMode())
	assert.Equal(t, "", cfg.HTMLEntrypoint)
	assert.Equal(t, "", cfg.TextEntrypoint)
	assert.Equal(t, "", cfg.BaseTemplateKey)
}

func TestDecodeRenderMetadata_ExplicitValues(t *testing.T) {
	cfg := DecodeRenderMetadata(map[string]any{
		MetadataKeyRenderMode.String():     RenderModeGoTemplateBundle.String(),
		MetadataKeyHTMLEntrypoint.String(): "main",
		MetadataKeyTextEntrypoint.String(): "text_main",
		MetadataKeyBaseTemplate.String():   "base-template",
	})

	assert.Equal(t, RenderModeGoTemplateBundle, cfg.EffectiveRenderMode())
	assert.Equal(t, "main", cfg.HTMLEntrypoint)
	assert.Equal(t, "text_main", cfg.TextEntrypoint)
	assert.Equal(t, "base-template", cfg.BaseTemplateKey)
}
