package googledrive

import (
	"net/url"
	"strings"
)

// parseFolderID extracts a Google Drive folder ID from a raw ID or full URL.
// Accepts forms like:
//
//	0AOV4Cj9uyA4SUk9PVA
//	https://drive.google.com/drive/folders/0AOV4Cj9uyA4SUk9PVA
//	https://drive.google.com/drive/u/0/folders/0AOV4Cj9uyA4SUk9PVA
func parseFolderID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !strings.Contains(raw, "/") {
		return raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, seg := range segments {
		if seg == "folders" && i+1 < len(segments) {
			return segments[i+1]
		}
	}

	return raw
}
