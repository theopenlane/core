package resolvers

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/logx"
)

// init registers all built-in resolver functions so they are available automatically
// without requiring explicit registration elsewhere in the codebase
func init() {
	// Control resolvers
	Register("CONTROL_OWNER", resolveObjectOwner)
	Register("CONTROL_AUDITOR", resolveControlAuditor)
	Register("RESPONSIBLE_PARTY", resolveResponsibleParty)

	// InternalPolicy resolvers
	Register("POLICY_OWNER", resolveObjectOwner)
	Register("POLICY_APPROVER", resolvePolicyApprover)
	Register("POLICY_DELEGATE", resolvePolicyDelegate)

	// Evidence resolvers
	Register("EVIDENCE_OWNER", resolveObjectOwner)

	// Universal resolvers
	Register("OBJECT_CREATOR", resolveObjectCreator)
}

// resolveControlAuditor returns the auditor reference ID for a control, if set.
func resolveControlAuditor(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error) {
	return resolveOptionalIDList(ctx, client, obj, "auditor_reference_id")
}

// resolveResponsibleParty returns the responsible party for a control, if set
func resolveResponsibleParty(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error) {
	return resolveOptionalIDList(ctx, client, obj, "responsible_party_id")
}

// resolvePolicyApprover returns the approvers for a policy from the approver group
func resolvePolicyApprover(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error) {
	approverID, err := resolveOptionalID(ctx, client, obj, "approver_id")
	if err != nil {
		return nil, err
	}
	if approverID == "" {
		return []string{}, nil
	}

	return ResolveGroupMembers(ctx, client, approverID)
}

// resolvePolicyDelegate returns the delegates for a policy from the delegate group
func resolvePolicyDelegate(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error) {
	delegateID, err := resolveOptionalID(ctx, client, obj, "delegate_id")
	if err != nil {
		return nil, err
	}
	if delegateID == "" {
		return []string{}, nil
	}

	return ResolveGroupMembers(ctx, client, delegateID)
}

// resolveObjectCreator returns the creator of any object, if available
func resolveObjectCreator(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error) {
	return resolveOptionalIDList(ctx, client, obj, "created_by")
}

// ownerQueryer is an interface for objects that can query their owning organization
type ownerQueryer interface {
	QueryOwner() *generated.OrganizationQuery
}

// loadWorkflowNode loads the workflow node for the given object, using the node if present or loading from the client otherwise
func loadWorkflowNode(ctx context.Context, client *generated.Client, obj *workflows.Object) (any, error) {
	if obj.Node != nil {
		return obj.Node, nil
	}

	return workflows.LoadWorkflowObject(ctx, client, obj.Type.String(), obj.ID)
}

// resolveObjectOwner returns the user IDs of the owners for the given object, if available
func resolveObjectOwner(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error) {
	node, err := loadWorkflowNode(ctx, client, obj)
	if err != nil {
		return nil, err
	}

	ownerID, err := workflows.StringField(node, "owner_id")
	if err != nil {
		return nil, err
	}
	if ownerID != "" {
		return workflows.OrganizationOwnerIDs(ctx, client, ownerID)
	}

	queryer, ok := node.(ownerQueryer)
	if !ok {
		return []string{}, nil
	}

	org, err := queryer.QueryOwner().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to resolve owner organization")
		return []string{}, nil
	}

	return workflows.OrganizationOwnerIDs(ctx, client, org.ID)
}

// resolveOptionalID loads a workflow node and extracts a string field
func resolveOptionalID(ctx context.Context, client *generated.Client, obj *workflows.Object, field string) (string, error) {
	node, err := loadWorkflowNode(ctx, client, obj)
	if err != nil {
		return "", err
	}

	return workflows.StringField(node, field)
}

// resolveOptionalIDList returns a single-element slice when the field is present
func resolveOptionalIDList(ctx context.Context, client *generated.Client, obj *workflows.Object, field string) ([]string, error) {
	value, err := resolveOptionalID(ctx, client, obj, field)
	if err != nil {
		return nil, err
	}

	if value == "" {
		return []string{}, nil
	}

	return []string{value}, nil
}

// ResolveGroupMembers returns the user IDs of all users in the specified group.
func ResolveGroupMembers(ctx context.Context, client *generated.Client, groupID string) ([]string, error) {
	users, err := client.Group.
		Query().
		Where(group.IDEQ(groupID)).
		QueryUsers().
		IDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve group members: %w", err)
	}
	return users, nil
}
