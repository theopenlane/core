package gala

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/iam/auth"
)

const (
	// authContextCodecKey is the default context snapshot key for authenticated user context.
	authContextCodecKey ContextKey = "auth_user"
)

// AuthContextSnapshot is a JSON-safe snapshot of authenticated user context values.
type AuthContextSnapshot struct {
	// SubjectID is the authenticated principal identifier.
	SubjectID string `json:"subject_id,omitempty"`
	// SubjectName is the authenticated principal display name.
	SubjectName string `json:"subject_name,omitempty"`
	// SubjectEmail is the authenticated principal email.
	SubjectEmail string `json:"subject_email,omitempty"`
	// OrganizationID is the active organization scope.
	OrganizationID string `json:"organization_id,omitempty"`
	// OrganizationName is the active organization display name.
	OrganizationName string `json:"organization_name,omitempty"`
	// OrganizationIDs contains organizations available in caller scope.
	OrganizationIDs []string `json:"organization_ids,omitempty"`
	// AuthenticationType identifies the authentication method used by the caller.
	AuthenticationType string `json:"authentication_type,omitempty"`
	// OrganizationRole captures the caller role within the active organization.
	OrganizationRole string `json:"organization_role,omitempty"`
	// IsSystemAdmin reports whether the caller has system-admin privileges.
	IsSystemAdmin bool `json:"is_system_admin,omitempty"`
}

// ToAuthenticatedUser converts a snapshot into an auth.AuthenticatedUser payload.
func (s AuthContextSnapshot) ToAuthenticatedUser() *auth.AuthenticatedUser {
	return &auth.AuthenticatedUser{
		SubjectID:          s.SubjectID,
		SubjectName:        s.SubjectName,
		SubjectEmail:       s.SubjectEmail,
		OrganizationID:     s.OrganizationID,
		OrganizationName:   s.OrganizationName,
		OrganizationIDs:    append([]string(nil), s.OrganizationIDs...),
		AuthenticationType: auth.AuthenticationType(s.AuthenticationType),
		OrganizationRole:   auth.OrganizationRoleType(s.OrganizationRole),
		IsSystemAdmin:      s.IsSystemAdmin,
	}
}

// authContextSnapshotFromUser converts an auth.AuthenticatedUser into a JSON-safe snapshot.
func authContextSnapshotFromUser(user *auth.AuthenticatedUser) AuthContextSnapshot {
	if user == nil {
		return AuthContextSnapshot{}
	}

	return AuthContextSnapshot{
		SubjectID:          user.SubjectID,
		SubjectName:        user.SubjectName,
		SubjectEmail:       user.SubjectEmail,
		OrganizationID:     user.OrganizationID,
		OrganizationName:   user.OrganizationName,
		OrganizationIDs:    append([]string(nil), user.OrganizationIDs...),
		AuthenticationType: string(user.AuthenticationType),
		OrganizationRole:   string(user.OrganizationRole),
		IsSystemAdmin:      user.IsSystemAdmin,
	}
}

// AuthContextCodec captures and restores auth.AuthenticatedUser context values.
type AuthContextCodec struct{}

// NewAuthContextCodec creates a context codec for authenticated user context.
func NewAuthContextCodec() AuthContextCodec {
	return AuthContextCodec{}
}

// Key returns the stable snapshot key used by the auth context codec.
func (AuthContextCodec) Key() ContextKey {
	return authContextCodecKey
}

// Capture extracts authenticated user context and encodes it as JSON.
func (AuthContextCodec) Capture(ctx context.Context) (json.RawMessage, bool, error) {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || au == nil {
		return nil, false, nil
	}

	encoded, err := json.Marshal(authContextSnapshotFromUser(au))
	if err != nil {
		return nil, false, ErrAuthContextEncodeFailed
	}

	return append(json.RawMessage(nil), encoded...), true, nil
}

// Restore decodes authenticated user context and restores it on the supplied context.
func (AuthContextCodec) Restore(ctx context.Context, raw json.RawMessage) (context.Context, error) {
	var snapshot AuthContextSnapshot
	if err := jsonx.RoundTrip(raw, &snapshot); err != nil {
		return ctx, ErrAuthContextDecodeFailed
	}

	return auth.WithAuthenticatedUser(ctx, snapshot.ToAuthenticatedUser()), nil
}
