package enums

import "io"

// FileBackupStatus is a custom type representing the state of a file's backup replication.
type FileBackupStatus string

var (
	// FileBackupStatusPending indicates the backup has been queued but not yet completed.
	FileBackupStatusPending FileBackupStatus = "PENDING"
	// FileBackupStatusCompleted indicates the backup replication succeeded.
	FileBackupStatusCompleted FileBackupStatus = "COMPLETED"
	// FileBackupStatusFailed indicates the backup replication failed and may be retried.
	FileBackupStatusFailed FileBackupStatus = "FAILED"
	// FileBackupStatusExhausted indicates the backup failed and the max attempts were reached, so it will not be retried.
	FileBackupStatusExhausted FileBackupStatus = "EXHAUSTED"
	// FileBackupStatusInvalid is used when an unknown or unsupported value is provided.
	FileBackupStatusInvalid FileBackupStatus = "FILEBACKUPSTATUS_INVALID"
)

var fileBackupStatusValues = []FileBackupStatus{
	FileBackupStatusPending,
	FileBackupStatusCompleted,
	FileBackupStatusFailed,
	FileBackupStatusExhausted,
}

// Values returns a slice of strings representing all valid FileBackupStatus values.
func (FileBackupStatus) Values() []string { return stringValues(fileBackupStatusValues) }

// String returns the string representation of the FileBackupStatus value.
func (r FileBackupStatus) String() string { return string(r) }

// ToFileBackupStatus converts a string to its corresponding FileBackupStatus enum value.
func ToFileBackupStatus(r string) *FileBackupStatus {
	return parse(r, fileBackupStatusValues, &FileBackupStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r FileBackupStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *FileBackupStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
