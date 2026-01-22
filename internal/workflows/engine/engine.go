package engine

import (
	"sync"

	"github.com/google/cel-go/cel"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/observability"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// WorkflowEngine orchestrates workflow execution via event emission
type WorkflowEngine struct {
	client       *generated.Client
	emitter      soiree.Emitter
	observer     *observability.Observer
	config       *workflows.Config
	env          *cel.Env
	programCache sync.Map
}

// NewWorkflowEngine creates a new workflow engine using the provided configuration options
func NewWorkflowEngine(client *generated.Client, emitter soiree.Emitter, opts ...workflows.ConfigOpts) (*WorkflowEngine, error) {
	config := workflows.NewDefaultConfig(opts...)

	return NewWorkflowEngineWithConfig(client, emitter, config)
}

// NewWorkflowEngineWithConfig creates a new workflow engine using the provided configuration
func NewWorkflowEngineWithConfig(client *generated.Client, emitter soiree.Emitter, config *workflows.Config) (*WorkflowEngine, error) {
	if client == nil {
		return nil, ErrNilClient
	}

	if config == nil {
		config = workflows.NewDefaultConfig()
	}

	// we set default opts within this
	env, err := workflows.NewCELEnvWithConfig(config)
	if err != nil {
		return nil, err
	}

	return &WorkflowEngine{
		client:   client,
		emitter:  emitter,
		observer: observability.New(),
		config:   config,
		env:      env,
	}, nil
}
