package controls

import "errors"

var (
	// ErrStandardNotFound is the error message returned when a standard cannot be found based on the provided filter options during control cloning
	ErrStandardNotFound = errors.New("standard not found, unable to clone controls")
	// ErrRefCodeSortKeyBackfill is returned when the ref code sort key backfill cannot read or update a row
	ErrRefCodeSortKeyBackfill = errors.New("ref code sort key backfill failed")
)
