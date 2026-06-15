package onedrive

import (
	"net/url"
	"strings"
)

// parseFolderID extracts an OneDrive folder ID or path from raw input.
// Accepts forms like:
//
//	01ABC123DEF456...          (raw OneDrive item ID)
//	Policies         (path relative to drive root)
//	https://onedrive.live.com/...  (OneDrive web URL — path extracted)
func parseFolderID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !strings.Contains(raw, "/") {
		return raw
	}

	if !strings.HasPrefix(raw, "http") {
		return strings.Trim(raw, "/")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, seg := range segments {
		if seg == "items" && i+1 < len(segments) {
			return segments[i+1]
		}
	}

	return raw
}
