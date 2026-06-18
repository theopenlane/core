package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookSSOExemptionAttribution stamps the grantor and timestamp when a membership's SSO exemption is
// set, and clears the attribution and reason when the exemption is removed. The grantor defaults to
// the acting caller when it is not explicitly provided by the mutation, which lets API driven grants
// record who performed the change while server driven flows can attribute it to a specific user
func HookSSOExemptionAttribution() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OrgMembershipFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) (generated.Value, error) {
			exempt, ok := m.SSOExempt()
			if !ok {
				// the exemption field is not part of this mutation
				return next.Mutate(ctx, m)
			}

			switch {
			case exempt:
				if _, set := m.SSOExemptGrantedAt(); !set {
					m.SetSSOExemptGrantedAt(models.DateTime(time.Now()))
				}

				if _, set := m.SSOExemptGrantedBy(); !set {
					if uid := callerSubjectID(ctx); uid != "" {
						m.SetSSOExemptGrantedBy(uid)
					}
				}
			default:
				m.ClearSSOExemptReason()
				m.ClearSSOExemptGrantedBy()
				m.ClearSSOExemptGrantedAt()
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// callerSubjectID returns the acting caller's subject id, or empty when no caller is in context
func callerSubjectID(ctx context.Context) string {
	uid, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return ""
	}

	return uid
}
