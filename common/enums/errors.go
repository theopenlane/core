package enums

import (
	"errors"
)

var (
	// ErrInvalidType indicates an unexpected type was provided during enum unmarshalling
	ErrInvalidType = errors.New("unexpected type for enum unmarshal")

	// ErrWrongTypeWorkflowObjectType indicates the value type for WorkflowActionType is incorrect
	ErrWrongTypeWorkflowObjectType = errors.New("wrong type for WorkflowActionType")
)
