package enums

import "io"

type ReviewStatus string

var (
	ReviewStatusOpen       ReviewStatus = "OPEN"
	ReviewStatusInProgress ReviewStatus = "IN_PROGRESS"
	ReviewStatusInReview   ReviewStatus = "IN_REVIEW"
	ReviewStatusCompleted  ReviewStatus = "COMPLETED"
	ReviewStatusWontDo     ReviewStatus = "WONT_DO"
	ReviewStatusInvalid    ReviewStatus = "INVALID"
)

var reviewStatusValues = []ReviewStatus{ReviewStatusOpen, ReviewStatusInProgress, ReviewStatusInReview, ReviewStatusCompleted, ReviewStatusWontDo}

// Values returns a slice of strings that represents all the possible values of the ReviewStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS", "IN_REVIEW", "COMPLETED", and "WONT_DO".
func (ReviewStatus) Values() []string { return stringValues(reviewStatusValues) }

// String returns the ReviewStatus as a string
func (r ReviewStatus) String() string { return string(r) }

// ToReviewStatus returns the review status enum based on string input
func ToReviewStatus(r string) *ReviewStatus {
	return parse(r, reviewStatusValues, &ReviewStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r ReviewStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *ReviewStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
