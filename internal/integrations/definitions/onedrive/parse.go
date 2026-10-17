package onedrive

import "strings"

// parseFolderPath normalizes a folder path relative to the drive root, e.g. "Policies" or "/Shared/Policies/"
func parseFolderPath(raw string) string {
	return strings.Trim(strings.TrimSpace(raw), "/")
}
