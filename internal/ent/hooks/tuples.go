package hooks

import (
	"strings"

	"github.com/datumforge/fgax"

	"github.com/theopenlane/core/pkg/enums"
)

// getTupleKeyFromRole creates a Tuple key with the provided subject, object, and role
func getTupleKeyFromRole(req fgax.TupleRequest, role enums.Role) (fgax.TupleKey, error) {
	fgaRelation, err := roleToRelation(role)
	if err != nil {
		return fgax.NewTupleKey(), err
	}

	req.Relation = fgaRelation

	return fgax.GetTupleKey(req), nil
}

func roleToRelation(r enums.Role) (string, error) {
	switch r {
	case enums.RoleOwner, enums.RoleAdmin, enums.RoleMember:
		return strings.ToLower(r.String()), nil
	case fgax.ParentRelation:
		return r.String(), nil
	default:
		return "", ErrUnsupportedFGARole
	}
}
