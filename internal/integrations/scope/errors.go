package scope

import "errors"

var (
	// ErrScopeExpressionRequired indicates an expression is required for validation
	ErrScopeExpressionRequired = errors.New("scope: expression required")
	// ErrScopeCompilationFailed indicates expression compilation failed
	ErrScopeCompilationFailed = errors.New("scope: compilation failed")
	// ErrScopeProgramCreationFailed indicates expression program creation failed
	ErrScopeProgramCreationFailed = errors.New("scope: program creation failed")
	// ErrScopeEvaluationFailed indicates runtime evaluation failed
	ErrScopeEvaluationFailed = errors.New("scope: evaluation failed")
	// ErrScopeEvaluationTimeout indicates evaluation exceeded the timeout budget
	ErrScopeEvaluationTimeout = errors.New("scope: evaluation timeout")
	// ErrScopeEvaluationOutputNil indicates evaluation completed without output
	ErrScopeEvaluationOutputNil = errors.New("scope: evaluation output nil")
	// ErrScopeConditionType indicates condition expressions must evaluate to bool
	ErrScopeConditionType = errors.New("scope: condition type mismatch")
)
