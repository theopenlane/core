package oauthserver

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
	fgamodel "github.com/theopenlane/core/v2/fga/model"
)

var (
	// ErrScopesNotRequested is returned when a user grants scopes the application didn't ask for
	ErrScopesNotRequested = errors.New("scopes were not requested by the application")
	// ErrNoScopes is returned when an application is approved with nothing granted
	ErrNoScopes = errors.New("grant at least one scope")
	// ErrInvalidPersonalScope is returned for a scope a personal connection can't hold
	ErrInvalidPersonalScope = errors.New("personal connections take openlane:read, openlane:write, or object:read, object:write, object:delete scopes")
	// ErrInvalidObjectScope is returned for an object scope that doesn't map to a service relation in the FGA model
	ErrInvalidObjectScope = errors.New("unknown scopes")
)

// Personal connection scopes clients request, write implies read
const (
	// ScopeUser asks for a personal connection and grants nothing by itself, so clients can name exact object scopes
	ScopeUser = "openlane:user"
	// ScopeRead reads every object the user can see
	ScopeRead = "openlane:read"
	// ScopeWrite reads and changes every object the user can
	ScopeWrite = "openlane:write"
)

// SupportedScopes are the personal connection scopes advertised in the authorization server metadata
var SupportedScopes = []string{ScopeUser, ScopeRead, ScopeWrite}

// impliedBy lists, for each object relation, the relations that also grant it, so delete implies edit and view
var impliedBy = map[string][]string{
	fgax.CanView:   {fgax.CanView, fgax.CanEdit, fgax.CanDelete},
	fgax.CanEdit:   {fgax.CanEdit, fgax.CanDelete},
	fgax.CanDelete: {fgax.CanDelete},
}

// callerGrantKey stores the grant of a request made by an installed application
var callerGrantKey = contextx.NewKey[CallerGrant]()

// AccessScopes returns the personal scopes an access level grants
func AccessScopes(level enums.OAuthAccessLevel) []string {
	if level == enums.OAuthAccessLevelWrite {
		return []string{ScopeRead, ScopeWrite}
	}

	return []string{ScopeRead}
}

// MaxAccess returns the highest access level requested scopes allow, a client that requests nothing may get write
func MaxAccess(requested []string) enums.OAuthAccessLevel {
	if len(requested) == 0 || grantsWrite(requested) {
		return enums.OAuthAccessLevelWrite
	}

	return enums.OAuthAccessLevelRead
}

// grantsWrite reports whether scopes include any write or delete
func grantsWrite(scopes []string) bool {
	for _, s := range scopes {
		if s == ScopeWrite {
			return true
		}

		if relation, _, ok := fgamodel.ParseScopeRelation(s); ok && relation != fgax.CanView {
			return true
		}
	}

	return false
}

// ObjectScopes returns the object:action scopes in a list, without the personal scopes
func ObjectScopes(scopes []string) []string {
	return lo.Without(scopes, SupportedScopes...)
}

// InvalidObjectScopes returns the scopes that don't map to a service relation in the FGA model
func InvalidObjectScopes(scopes []string) []string {
	set, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return scopes
	}

	return lo.Filter(scopes, func(s string, _ int) bool {
		_, ok := set[fgamodel.NormalizeScope(s)]

		return !ok
	})
}

// validPersonalScope reports whether a personal connection can hold scope
func validPersonalScope(scope string) bool {
	if slices.Contains(SupportedScopes, scope) {
		return true
	}

	_, _, ok := fgamodel.ParseScopeRelation(scope)

	return ok && len(InvalidObjectScopes([]string{scope})) == 0
}

// ScopeSet answers what a list of personal scopes allows, keyed by FGA relation such as can_edit_control
type ScopeSet map[string]bool

// NewScopeSet returns the set for a list of scopes
func NewScopeSet(scopes []string) ScopeSet {
	set := ScopeSet{}

	for _, s := range scopes {
		if slices.Contains(SupportedScopes, s) {
			set[s] = true

			continue
		}

		set[fgamodel.NormalizeScope(s)] = true
	}

	return set
}

// ReadsAll reports whether every object is readable
func (s ScopeSet) ReadsAll() bool { return s[ScopeRead] || s[ScopeWrite] }

// WritesAll reports whether every object is writable
func (s ScopeSet) WritesAll() bool { return s[ScopeWrite] }

// Allows reports whether relation (fgax.CanView, CanEdit, or CanDelete) is allowed on object
func (s ScopeSet) Allows(object, relation string) bool {
	switch {
	case relation == fgax.CanView && s.ReadsAll():
		return true
	case relation != fgax.CanView && s.WritesAll():
		return true
	}

	for _, implied := range impliedBy[relation] {
		if s[fgamodel.ScopeRelation(implied, object)] {
			return true
		}
	}

	return false
}

// ValidatePersonalGrant checks the scopes granted to a personal connection against what the application
// requested and returns the resulting access level, an application that requested nothing may get anything
func ValidatePersonalGrant(requested, granted []string) (enums.OAuthAccessLevel, error) {
	if !slices.ContainsFunc(granted, func(g string) bool { return g != ScopeUser }) {
		return "", ErrNoScopes
	}

	req := NewScopeSet(requested)

	var extra []string

	for _, g := range granted {
		if g == ScopeUser {
			continue
		}

		if !validPersonalScope(g) {
			return "", fmt.Errorf("%w: %s", ErrInvalidPersonalScope, g)
		}

		if len(requested) > 0 && !requestCovers(req, g) {
			extra = append(extra, g)
		}
	}

	if len(extra) > 0 {
		return "", fmt.Errorf("%w: %s", ErrScopesNotRequested, strings.Join(extra, " "))
	}

	if grantsWrite(granted) {
		return enums.OAuthAccessLevelWrite, nil
	}

	return enums.OAuthAccessLevelRead, nil
}

// requestCovers reports whether the requested scope set includes a granted personal scope
func requestCovers(req ScopeSet, granted string) bool {
	switch granted {
	case ScopeRead:
		return req.ReadsAll()
	case ScopeWrite:
		return req.WritesAll()
	}

	relation, object, _ := fgamodel.ParseScopeRelation(granted)

	return req.Allows(object, relation)
}

// ValidateOrganizationGrant checks the object scopes granted to an organization install against what the
// application requested, an application that requested nothing may get any known scope
func ValidateOrganizationGrant(requested, granted []string) error {
	if len(granted) == 0 {
		return ErrNoScopes
	}

	if bad := InvalidObjectScopes(granted); len(bad) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidObjectScope, strings.Join(bad, " "))
	}

	if len(requested) == 0 {
		return nil
	}

	if extra := lo.Without(granted, requested...); len(extra) > 0 {
		return fmt.Errorf("%w: %s", ErrScopesNotRequested, strings.Join(extra, " "))
	}

	return nil
}

// ClassifyScopes returns the grant kind requested scopes ask for, any personal scope means a user connection
// and only object scopes mean an organization install, no scopes leaves the choice to the user
func ClassifyScopes(requested []string) (kind enums.OAuthGrantKind, decided bool, err error) {
	if len(requested) == 0 {
		return "", false, nil
	}

	if len(ObjectScopes(requested)) == len(requested) {
		if bad := InvalidObjectScopes(requested); len(bad) > 0 {
			return "", false, fmt.Errorf("%w: %s", ErrInvalidObjectScope, strings.Join(bad, " "))
		}

		return enums.OAuthGrantKindOrganization, true, nil
	}

	for _, s := range requested {
		if !validPersonalScope(s) {
			return "", false, fmt.Errorf("%w: %s", ErrInvalidPersonalScope, s)
		}
	}

	return enums.OAuthGrantKindUser, true, nil
}

// CallerGrant identifies a request made by an installed application
type CallerGrant struct {
	// ID is the OauthGrant id
	ID string
	// Kind is whether the application acts as a user or as the organization
	Kind enums.OAuthGrantKind
	// ClientID is the application's client_id
	ClientID string
	// Access is the grant's access level
	Access enums.OAuthAccessLevel
	// Scopes are what a personal connection was granted
	Scopes []string
}

// scopeSet returns the grant's scopes as a ScopeSet
func (g CallerGrant) scopeSet() ScopeSet { return NewScopeSet(g.Scopes) }

// ReadLimited reports whether a personal connection may read only some objects
func (g CallerGrant) ReadLimited() bool {
	return g.Kind == enums.OAuthGrantKindUser && !g.scopeSet().ReadsAll()
}

// WriteLimited reports whether a personal connection with write access may change only some objects
func (g CallerGrant) WriteLimited() bool {
	return g.Kind == enums.OAuthGrantKindUser && g.Access == enums.OAuthAccessLevelWrite && !g.scopeSet().WritesAll()
}

// WithCallerGrant marks a request context as made by an installed application
func WithCallerGrant(ctx context.Context, g CallerGrant) context.Context {
	return callerGrantKey.Set(ctx, g)
}

// CallerGrantFromContext returns the application grant of a request made by an installed application
func CallerGrantFromContext(ctx context.Context) (CallerGrant, bool) {
	return callerGrantKey.Get(ctx)
}
