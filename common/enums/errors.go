package enums

import (
	"errors"
)

var (
	// ErrWrongTypeWorkflowObjectType indicates the value type for WorkflowActionType is incorrect.
	ErrWrongTypeWorkflowObjectType = errors.New("wrong type for WorkflowActionType")
)
