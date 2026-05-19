package storage

import "mime"

// DispositionInline is the Content-Disposition value used when a file's MIME
// type is safe for the browser to render inline (PDFs, raster images).
const DispositionInline = "inline"

// DispositionAttachment is the Content-Disposition value used to force the
// browser to download the file instead of rendering it.
const DispositionAttachment = "attachment"

// inlineSafeMIMEs enumerates the MIME types that are safe to serve with an
// inline Content-Disposition. The set is intentionally narrow: only formats
// that the browser renders in a sandboxed viewer with no script execution.
//
// SVG is deliberately excluded — image/svg+xml can carry embedded JavaScript.
// HTML, JavaScript, and CSS are excluded for the same reason. Adding a new
// entry here is a security decision and should be reviewed accordingly.
var inlineSafeMIMEs = map[string]struct{}{
	"application/pdf": {},
	"image/png":       {},
	"image/jpeg":      {},
	"image/gif":       {},
	"image/webp":      {},
}

// DispositionFor returns the Content-Disposition value that is appropriate
// for serving a file with the given content type. Files whose MIME type is in
// the inline-safe allowlist receive "inline" so they render in-place when
// requested from an <iframe> or <img>; everything else receives "attachment"
// so the browser triggers a download rather than guessing what to do with an
// unknown payload.
//
// Parameters (e.g. "; charset=utf-8") and casing are normalized via
// mime.ParseMediaType. Malformed media types fall through to "attachment".
func DispositionFor(contentType string) string {
	if isInlineSafeMIME(contentType) {
		return DispositionInline
	}

	return DispositionAttachment
}

// isInlineSafeMIME reports whether contentType, after parsing as a media type
// (which lowercases the type and strips parameters), is in the inline-safe
// allowlist. Returns false for empty or malformed input.
func isInlineSafeMIME(contentType string) bool {
	if contentType == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	_, ok := inlineSafeMIMEs[mediaType]

	return ok
}
