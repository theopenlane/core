package gala

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

const (
	// durableContextCodecKey is the context snapshot key for durable context values
	durableContextCodecKey ContextKey = "durable"
)

// DurableContextSnapshot captures context values that should persist across durable event processing
type DurableContextSnapshot struct {
	// Auth contains authenticated user context when present
	Auth *AuthSnapshot `json:"auth,omitempty"`
	// LogFields contains logger context fields for correlation and tracing
	LogFields map[string]any `json:"log_fields,omitempty"`
}

// AuthSnapshot is a JSON-safe snapshot of authenticated user context values
type AuthSnapshot struct {
	// SubjectID is the authenticated principal identifier
	SubjectID string `json:"subject_id,omitempty"`
	// SubjectName is the authenticated principal display name
	SubjectName string `json:"subject_name,omitempty"`
	// SubjectEmail is the authenticated principal email
	SubjectEmail string `json:"subject_email,omitempty"`
	// OrganizationID is the active organization scope
	OrganizationID string `json:"organization_id,omitempty"`
	// OrganizationName is the active organization display name
	OrganizationName string `json:"organization_name,omitempty"`
	// OrganizationIDs contains organizations available in caller scope
	OrganizationIDs []string `json:"organization_ids,omitempty"`
	// AuthenticationType identifies the authentication method used by the caller
	AuthenticationType string `json:"authentication_type,omitempty"`
	// OrganizationRole captures the caller role within the active organization
	OrganizationRole string `json:"organization_role,omitempty"`
	// Capabilities is the full capability bitset carried by the caller.
	Capabilities uint64 `json:"capabilities,omitempty"`
	// IsSystemAdmin reports whether the caller has system-admin privileges
	// and is kept for backward compatibility with older snapshots.
	IsSystemAdmin bool `json:"is_system_admin,omitempty"`
}

// toCaller converts a snapshot into an auth.Caller
func (s AuthSnapshot) toCaller() *auth.Caller {
	caps := auth.Capability(s.Capabilities)
	if s.IsSystemAdmin {
		caps |= auth.CapSystemAdmin
	}

	return &auth.Caller{
		SubjectID:          s.SubjectID,
		SubjectName:        s.SubjectName,
		SubjectEmail:       s.SubjectEmail,
		OrganizationID:     s.OrganizationID,
		OrganizationName:   s.OrganizationName,
		OrganizationIDs:    append([]string(nil), s.OrganizationIDs...),
		AuthenticationType: auth.AuthenticationType(s.AuthenticationType),
		OrganizationRole:   auth.OrganizationRoleType(s.OrganizationRole),
		Capabilities:       caps,
	}
}

// authSnapshotFromCaller converts an auth.Caller into a JSON-safe snapshot
func authSnapshotFromCaller(caller *auth.Caller) *AuthSnapshot {
	if caller == nil {
		return nil
	}

	return &AuthSnapshot{
		SubjectID:          caller.SubjectID,
		SubjectName:        caller.SubjectName,
		SubjectEmail:       caller.SubjectEmail,
		OrganizationID:     caller.OrganizationID,
		OrganizationName:   caller.OrganizationName,
		OrganizationIDs:    append([]string(nil), caller.OrgIDs()...),
		AuthenticationType: string(caller.AuthenticationType),
		OrganizationRole:   string(caller.OrganizationRole),
		Capabilities:       uint64(caller.Capabilities),
		IsSystemAdmin:      caller.Has(auth.CapSystemAdmin),
	}
}

// DurableContextCodec captures and restores durable context values including auth and logger fields
type DurableContextCodec struct{}

// NewContextCodec creates a context codec for durable context capture
func NewContextCodec() DurableContextCodec {
	return DurableContextCodec{}
}

// Key returns the stable snapshot key used by the context codec
func (DurableContextCodec) Key() ContextKey {
	return durableContextCodecKey
}

// Capture extracts durable context values and encodes them as JSON
func (DurableContextCodec) Capture(ctx context.Context) (json.RawMessage, bool, error) {
	snapshot := DurableContextSnapshot{}
	hasData := false

	// Capture auth context
	if caller, callerOk := auth.CallerFromContext(ctx); callerOk && caller != nil {
		snapshot.Auth = authSnapshotFromCaller(caller)
		hasData = true
	}

	// Capture logger fields if available. Clone to avoid aliasing the live context map,
	// which may be written concurrently by pool workers restoring context snapshots.
	if fields := logx.FieldsFromContext(ctx); len(fields) > 0 {
		snapshot.LogFields = maps.Clone(fields)
		hasData = true
	}

	if !hasData {
		return nil, false, nil
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, false, ErrContextSnapshotCaptureFailed
	}

	return append(json.RawMessage(nil), encoded...), true, nil
}

// Restore decodes durable context values and restores them on the supplied context
func (DurableContextCodec) Restore(ctx context.Context, raw json.RawMessage) (context.Context, error) {
	var snapshot DurableContextSnapshot
	if err := jsonx.RoundTrip(raw, &snapshot); err != nil {
		return ctx, ErrContextSnapshotRestoreFailed
	}

	// Restore auth context
	if snapshot.Auth != nil {
		ctx = auth.WithCaller(ctx, snapshot.Auth.toCaller())
	}

	// Restore logger with captured fields
	if len(snapshot.LogFields) > 0 {
		ctx = logx.WithFields(ctx, snapshot.LogFields)
	}

	return ctx, nil
}
