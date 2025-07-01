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
