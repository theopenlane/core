//go:build windmill

package windmill

import (
	"time"

	"github.com/theopenlane/core/pkg/enums"
)

// CreateFlowRequest represents the request structure for creating a new flow
type CreateFlowRequest struct {
	Path        string                `json:"path"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Value       []any                 `json:"value"`
	Schema      *map[string]any       `json:"schema,omitempty"`
	Language    enums.JobPlatformType `json:"language"`
}

// UpdateFlowRequest represents the request structure for updating an existing flow
type UpdateFlowRequest struct {
	Summary     string                  `json:"summary,omitempty"`
	Description string                  `json:"description,omitempty"`
	Value       []any                   `json:"value"`
	Schema      *map[string]interface{} `json:"schema,omitempty"`
	Language    enums.JobPlatformType   `json:"language"`
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
	Schedule string `json:"schedule"`
	FlowPath string `json:"script_path"`
	Args     any    `json:"args,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
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
