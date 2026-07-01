package graphapi

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// supportUserEdgeExtension prevents list-edge model methods on a synthetic support
type supportUserEdgeExtension struct{}

// ExtensionName satisfies the extension interface
func (supportUserEdgeExtension) ExtensionName() string { return "SupportUserEdges" }

// Validate satisfies the extension interface
func (supportUserEdgeExtension) Validate(graphql.ExecutableSchema) error { return nil }

// InterceptField set named edges for the support user when retrieving the user data
func (supportUserEdgeExtension) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || (fc.Object != "User") {
		return next(ctx)
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || !caller.Has(auth.CapOrgSupport) {
		return next(ctx)
	}

	switch fc.Field.Name {
	case "tfaSettings":
		return &generated.TFASettingConnection{Edges: []*generated.TFASettingEdge{}}, nil
	case "organizations":
		return &generated.OrganizationConnection{Edges: []*generated.OrganizationEdge{}}, nil
	case "orgMemberships":
		return &generated.OrgMembershipConnection{Edges: []*generated.OrgMembershipEdge{}}, nil
	case "groups":
		return &generated.GroupConnection{Edges: []*generated.GroupEdge{}}, nil
	case "groupMemberships":
		return &generated.GroupMembershipConnection{Edges: []*generated.GroupMembershipEdge{}}, nil
	case "personalAccessTokens":
		return &generated.PersonalAccessTokenConnection{Edges: []*generated.PersonalAccessTokenEdge{}}, nil
	}

	return next(ctx)
}

// WithSupportUserEdges registers the support-user edge extension
func WithSupportUserEdges(srv *handler.Server) {
	srv.Use(supportUserEdgeExtension{})
}

// supportUserFromCaller defines the support user for calls from the UI such as GetUserProfile
func supportUserFromCaller(caller *auth.Caller) *generated.User {
	orgID := caller.OrganizationID

	setting := &generated.UserSetting{
		UserID:         caller.SubjectID,
		Status:         enums.UserStatusActive,
		EmailConfirmed: true,
		Edges: generated.UserSettingEdges{
			DefaultOrg: &generated.Organization{
				ID: orgID,
			},
		},
	}

	return &generated.User{
		ID:          caller.SubjectID,
		DisplayName: caller.SubjectName,
		Email:       caller.SubjectEmail,
		Edges: generated.UserEdges{
			Setting:    setting,
			AvatarFile: &generated.File{},
		},
	}
}
