package scim

// DirectorySync is a no-op placeholder result for push-based SCIM directory sync
type DirectorySync struct {
	// Message describes the push-based sync state
	Message string `json:"message"`
}

const directorySyncAckMessage = "scim is push-based; sync is triggered by the external identity provider"
