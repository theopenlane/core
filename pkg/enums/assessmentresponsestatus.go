package enums

import (
	"fmt"
	"io"
	"strconv"
)

type AssessmentResponseStatus string

const (
	// AssessmentResponseStatusNotStarted indicates the assessment has been assigned but not started
	AssessmentResponseStatusNotStarted AssessmentResponseStatus = "NOT_STARTED"
	// AssessmentResponseStatusInProgress indicates the assessment has been started but not completed
	AssessmentResponseStatusInProgress AssessmentResponseStatus = "IN_PROGRESS"
	// AssessmentResponseStatusCompleted indicates the assessment has been completed
	AssessmentResponseStatusCompleted AssessmentResponseStatus = "COMPLETED"
	// AssessmentResponseStatusOverdue indicates the assessment is past its due date
	AssessmentResponseStatusOverdue AssessmentResponseStatus = "OVERDUE"
	// AssessmentResponseStatusReviewRequired indicates the assessment needs review
	AssessmentResponseStatusReviewRequired AssessmentResponseStatus = "REVIEW_REQUIRED"
)

// String returns the string representation of AssessmentResponseStatus
func (ars AssessmentResponseStatus) String() string {
	return string(ars)
}

// Values returns all possible values for AssessmentResponseStatus enum
func (AssessmentResponseStatus) Values() (kinds []string) {
	for _, s := range []AssessmentResponseStatus{
		AssessmentResponseStatusNotStarted,
		AssessmentResponseStatusInProgress,
		AssessmentResponseStatusCompleted,
		AssessmentResponseStatusOverdue,
		AssessmentResponseStatusReviewRequired,
	} {
		kinds = append(kinds, string(s))
	}
	return
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (ars *AssessmentResponseStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*ars = AssessmentResponseStatus(str)
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (ars AssessmentResponseStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(ars.String()))
}
