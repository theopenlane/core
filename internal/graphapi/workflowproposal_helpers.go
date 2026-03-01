package graphapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/mutations"
	"github.com/theopenlane/core/pkg/mapx"
)

// requireWorkflowObjectEditAccess checks that the user in the context has edit access to the given workflow object
func (r *Resolver) requireWorkflowObjectEditAccess(ctx context.Context, objectType enums.WorkflowObjectType, objectID string) error {
	if objectID == "" || objectType == "" {
		return fmt.Errorf("%w: missing workflow object context", rout.ErrBadRequest)
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return rout.ErrPermissionDenied
	}

	if caller.Has(auth.CapSystemAdmin) {
		return nil
	}

	if caller.SubjectID == "" {
		return rout.ErrPermissionDenied
	}

	allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		ObjectType:  fgax.Kind(strcase.SnakeCase(objectType.String())),
		ObjectID:    objectID,
		Relation:    fgax.CanEdit,
		SubjectID:   caller.SubjectID,
		SubjectType: caller.SubjectType(),
	})
	if err != nil {
		return err
	}
	if !allow {
		return rout.ErrPermissionDenied
	}

	return nil
}

// workflowProposalDomainFields parses the domain key into object type and fields
func workflowProposalDomainFields(domainKey string) (string, []string, error) {
	trimmed := strings.TrimSpace(domainKey)
	if trimmed == "" {
		return "", nil, fmt.Errorf("%w: proposal domain key missing", rout.ErrBadRequest)
	}

	parts := strings.SplitN(trimmed, ":", 2) //nolint:mnd
	if len(parts) != 2 {                     //nolint:mnd
		return "", nil, fmt.Errorf("%w: invalid proposal domain key", rout.ErrBadRequest)
	}

	objectType := strings.TrimSpace(parts[0])
	fieldsPart := strings.TrimSpace(parts[1])
	if fieldsPart == "" {
		return objectType, nil, fmt.Errorf("%w: proposal domain key missing fields", rout.ErrBadRequest)
	}

	fields := mutations.NormalizeStrings(strings.Split(fieldsPart, ","))
	if len(fields) == 0 {
		return objectType, nil, fmt.Errorf("%w: proposal domain key missing fields", rout.ErrBadRequest)
	}

	return objectType, fields, nil
}

// validateWorkflowProposalChanges ensures that the proposed changes are valid for the given domain
func validateWorkflowProposalChanges(domainKey string, objectType enums.WorkflowObjectType, changes map[string]any) error {
	domainObjectType, fields, err := workflowProposalDomainFields(domainKey)
	if err != nil {
		return err
	}

	if domainObjectType != "" && !strings.EqualFold(domainObjectType, objectType.String()) {
		return fmt.Errorf("%w: proposal domain does not match object type", rout.ErrBadRequest)
	}

	if len(changes) == 0 {
		return nil
	}

	allowed := mapx.MapSetFromSlice(fields)

	for field := range changes {
		if _, ok := allowed[field]; !ok {
			return fmt.Errorf("%w: field %q is not part of the proposal domain", rout.ErrBadRequest, field)
		}
	}

	return nil
}
