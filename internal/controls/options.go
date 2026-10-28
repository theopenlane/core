package controls

import "github.com/theopenlane/core/v2/pkg/gala"

// CloneOptions holds the options for cloning controls
type CloneOptions struct {
	programID *string
	orgID     string
	pool      *gala.Pool
}

// CloneOption is a function type that modifies the CloneOptions
type CloneOption func(*CloneOptions)

// WithProgramID sets the ProgramID in the CloneOptions
func WithProgramID(programID string) CloneOption {
	return func(co *CloneOptions) {
		co.programID = &programID
	}
}

// WithOrgID sets the OrgID in the CloneOptions
func WithOrgID(orgID string) CloneOption {
	return func(co *CloneOptions) {
		co.orgID = orgID
	}
}

// WithPool clones the controls in parallel on the pool, a nil pool clones them one at a time for callers inside a transaction
func WithPool(pool *gala.Pool) CloneOption {
	return func(co *CloneOptions) {
		co.pool = pool
	}
}

// runTasks runs the tasks on the pool, or in order when the pool is nil
func runTasks(pool *gala.Pool, tasks []func()) error {
	if pool != nil {
		return pool.SubmitMultipleAndWait(tasks)
	}

	for _, task := range tasks {
		task()
	}

	return nil
}
