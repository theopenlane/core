package windmill

import "time"

// CreateFlowRequest represents the request structure for creating a new flow
type CreateFlowRequest struct {
	Path    string `json:"path"`
	Summary string `json:"summary,omitempty"`
	Value   []any  `json:"value"` // Array of modules that make up the flow
	Schema  any    `json:"schema,omitempty"`
}

// UpdateFlowRequest represents the request structure for updating an existing flow
type UpdateFlowRequest struct {
	Summary string `json:"summary,omitempty"`
	Value   []any  `json:"value"` // Array of modules that make up the flow
	Schema  any    `json:"schema,omitempty"`
}

// CreateFlowResponse represents the response after creating a flow
type CreateFlowResponse struct {
	Path string `json:"path"`
}

// Flow represents a complete flow definition
type Flow struct {
	Path      string    `json:"path"`
	Summary   string    `json:"summary,omitempty"`
	Value     any       `json:"value"`
	Schema    any       `json:"schema,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"edited_at"`
}

// CreateScheduledJobRequest represents the request structure for creating a scheduled job
type CreateScheduledJobRequest struct {
	Path     string `json:"path"`
	Schedule string `json:"schedule"`          // Cron expression
	FlowPath string `json:"script_path"`       // Path to the flow to execute
	Args     any    `json:"args,omitempty"`    // Arguments to pass to the flow
	Summary  string `json:"summary,omitempty"` // Description of the scheduled job
	Enabled  *bool  `json:"enabled,omitempty"` // Whether the scheduled job is enabled
}

// CreateScheduledJobResponse represents the response after creating a scheduled job
type CreateScheduledJobResponse struct {
	Path string `json:"path"`
}

// ScheduledJob represents a scheduled job in Windmill
type ScheduledJob struct {
	Path      string    `json:"path"`
	Schedule  string    `json:"schedule"`
	FlowPath  string    `json:"script_path"`
	Args      any       `json:"args,omitempty"`
	Summary   string    `json:"summary,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"edited_at"`
}
