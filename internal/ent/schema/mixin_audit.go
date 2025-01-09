package schema

import (
	"context"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/entx"
)

// NewAuditMixin creates a new AuditMixin
func NewAuditMixin() AuditMixin {
	return AuditMixin{}
}

// NewAuditMixinWithExcludedEdges creates a new AuditMixin with the edges excluded
func NewAuditMixinWithExcludedEdges() AuditMixin {
	return AuditMixin{
		ExcludeEdge: true,
	}
}

// AuditMixin provides auditing for all records where enabled. The created_at, created_by_i, updated_at, and updated_by_id records are automatically populated when this mixin is enabled.
type AuditMixin struct {
	mixin.Schema
	// ExcludeEdge is a boolean to indicate if the edges should be excluded
	ExcludeEdge bool
}

// Fields of the AuditMixin
func (AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Optional().
			Default(time.Now).
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.Time("updated_at").
			Default(time.Now).
			Optional().
			UpdateDefault(time.Now).
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("created_by_id").
			Immutable().
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("updated_by_id").
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("created_by_user_id").
			Immutable().
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("updated_by_user_id").
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("created_by_service_id").
			Immutable().
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
		field.String("updated_by_service_id").
			Optional().
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
				entx.FieldAdminSearchable(false),
			),
	}
}

// Edges of the AuditMixin
func (a AuditMixin) Edges() []ent.Edge {
	if a.ExcludeEdge {
		return nil
	}

	return []ent.Edge{
		edge.To("created_by_user", User.Type).
			Unique().
			Field("created_by_user_id").
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
			).
			Immutable(),
		edge.To("updated_by_user", User.Type).
			Unique().
			Field("updated_by_user_id").
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
			),
		edge.To("created_by_service", APIToken.Type).
			Unique().
			Field("created_by_service_id").
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
			).
			Immutable(),
		edge.To("updated_by_service", APIToken.Type).
			Unique().
			Field("updated_by_service_id").
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput,
				),
			),
	}
}

// Hooks of the AuditMixin
func (AuditMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		AuditHook,
	}
}

func (a AuditMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		AuditInterceptor,
	}
}

var AuditInterceptor ent.Interceptor = ent.InterceptFunc(func(next ent.Querier) ent.Querier {
	return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
		res, err := next.Query(ctx, query)
		if err != nil {
			return nil, err
		}

		// get the fields that were queried and check for the SubscriptionURL field
		fields := graphql.CollectFieldsCtx(ctx, []string{createdBy, updatedBy})

		// if the fields are empty, return the query as is
		if len(fields) == 0 {
			return res, nil
		}

		v, ok := res.(AuditFields)
		if !ok {
			return res, nil
		}

		out := res.(map[string]interface{})

		for _, f := range fields {
			switch f.Name {
			case createdBy:
				out = setField(ctx, v.CreatedByID, out, createdBy)
			case updatedBy:
				out = setField(ctx, v.UpdatedByID, out, updatedBy)
			}
		}

		return out, nil
	})
})

func setField(ctx context.Context, id string, out map[string]interface{}, fieldName string) map[string]interface{} {
	user, _ := hooks.TransactionFromContext(ctx).User.Get(ctx, id)
	service, _ := hooks.TransactionFromContext(ctx).APIToken.Get(ctx, id)

	if user != nil {
		out[fieldName] = getUserActor(user)
	} else if service != nil {
		out[fieldName] = getServiceActor(service)
	}

	return out
}

func getUserActor(user *generated.User) models.Actor {
	return models.Actor{
		ID:      user.ID,
		Name:    user.FirstName + " " + user.LastName,
		Type:    "user",
		Details: user,
	}
}

func getServiceActor(service *generated.APIToken) models.Actor {
	return models.Actor{
		ID:      service.ID,
		Name:    service.Name,
		Type:    "service",
		Details: service,
	}
}

const (
	createdBy = "createdBy"
	updatedBy = "updatedBy"
)

type AuditFields struct {
	CreatedByID string `json:"createdByID"`
	UpdatedByID string `json:"updatedByID"`
}

// AuditHook sets and returns the created_at, updated_at, etc., fields
func AuditHook(next ent.Mutator) ent.Mutator {
	type AuditLogger interface {
		SetCreatedAt(time.Time)
		CreatedAt() (v time.Time, exists bool) // exists if present before this hook
		SetUpdatedAt(time.Time)
		UpdatedAt() (v time.Time, exists bool)
		SetCreatedByID(string)
		CreatedByID() (id string, exists bool)
		SetUpdatedByID(string)
		UpdatedByID() (id string, exists bool)

		SetUpdatedByUserID(string)
		UpdatedByUserID() (id string, exists bool)

		SetCreatedByUserID(string)
		CreatedByUserID() (id string, exists bool)

		SetUpdatedByServiceID(string)
		UpdatedByServiceID() (id string, exists bool)

		SetCreatedByServiceID(string)
		CreatedByServiceID() (id string, exists bool)

		SetCreatedByActorID(string)
		CreatedByActorID() (id string, exists bool)
	}

	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		ml, ok := m.(AuditLogger)
		if !ok {
			return nil, newUnexpectedAuditError(m)
		}

		actor, _ := auth.GetUserIDFromContext(ctx) // ignore error, leave as null if not found
		actorType := auth.GetAuthzSubjectType(ctx)

		switch op := m.Op(); {
		case op.Is(ent.OpCreate):
			ml.SetCreatedAt(time.Now())

			if actor != "" {
				ml.SetCreatedByID(actor)
				ml.SetUpdatedByID(actor)
				ml.SetCreatedByActorID(actor)

				if actorType != string(auth.APITokenAuthentication) {
					ml.SetCreatedByUserID(actor)
					ml.SetCreatedByServiceID("")
				} else {
					ml.SetCreatedByServiceID(actor)
					ml.SetCreatedByUserID("")
				}
			}

		case op.Is(ent.OpUpdateOne | ent.OpUpdate):
			ml.SetUpdatedAt(time.Now())

			if actor != "" {
				ml.SetUpdatedByID(actor)

				if actorType != string(auth.APITokenAuthentication) {
					ml.SetUpdatedByUserID(actor)
					ml.SetUpdatedByServiceID("")
				} else {
					ml.SetUpdatedByServiceID(actor)
					ml.SetUpdatedByUserID("")
				}
			}
		}

		return next.Mutate(ctx, m)
	})
}
