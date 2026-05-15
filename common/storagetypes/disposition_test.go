package storagetypes

import "testing"

func TestDispositionFor(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{"pdf exact", "application/pdf", DispositionInline},
		{"pdf with charset", "application/pdf; charset=binary", DispositionInline},
		{"pdf uppercase", "APPLICATION/PDF", DispositionInline},
		{"pdf with whitespace and params", "  application/pdf ; foo=bar  ", DispositionInline},
		{"png", "image/png", DispositionInline},
		{"jpeg", "image/jpeg", DispositionInline},
		{"gif", "image/gif", DispositionInline},
		{"webp", "image/webp", DispositionInline},

		{"empty", "", DispositionAttachment},
		{"malformed", "not a mime type", DispositionAttachment},
		{"trailing semicolon only", ";", DispositionAttachment},
		{"params without type", "; charset=utf-8", DispositionAttachment},

		// Security-relevant exclusions — these MUST NOT render inline.
		{"svg excluded", "image/svg+xml", DispositionAttachment},
		{"html excluded", "text/html", DispositionAttachment},
		{"javascript excluded", "application/javascript", DispositionAttachment},
		{"javascript text excluded", "text/javascript", DispositionAttachment},
		{"xhtml excluded", "application/xhtml+xml", DispositionAttachment},

		// Other common types — attachment is correct behavior.
		{"octet-stream", "application/octet-stream", DispositionAttachment},
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", DispositionAttachment},
		{"zip", "application/zip", DispositionAttachment},
		{"csv", "text/csv", DispositionAttachment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DispositionFor(tt.contentType); got != tt.want {
				t.Errorf("DispositionFor(%q) = %q, want %q", tt.contentType, got, tt.want)
			}
		})
	}
}
