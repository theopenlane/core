package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
)

// File defines the file schema.
type File struct {
	ent.Schema
}

// Fields returns file fields.
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("provided_file_name").
			Comment("the name of the file provided in the payload key without the extension"),
		field.String("provided_file_extension").
			Comment("the extension of the file provided"),
		field.Int64("provided_file_size").
			Comment("the computed size of the file in the original http request").
			NonNegative().
			Optional(),
		field.Int64("persisted_file_size").
			NonNegative().
			Optional(),
		field.String("detected_mime_type").
			Comment("the mime type detected by the system").
			Optional(),
		field.String("md5_hash").
			Comment("the computed md5 hash of the file calculated after we received the contents of the file, but before the file was written to permanent storage").
			Optional(),
		field.String("detected_content_type").
			Comment("the content type of the HTTP request - may be different than MIME type as multipart-form can transmit multiple files and different types"),
		field.String("store_key").
			Comment("the key parsed out of a multipart-form request; if we allow multiple files to be uploaded we may want our API specifications to require the use of different keys allowing us to perform easier conditional evaluation on the key and what to do with the file based on key").
			Optional(),
		field.String("category_type").
			Comment("the category type of the file, if any (e.g. evidence, invoice, etc.)").
			Optional(),
		field.String("uri").
			Comment("the full URI of the file").
			Optional(),
		field.String("storage_scheme").
			Comment("the storage scheme of the file, e.g. file://, s3://, etc.").
			Optional(),
		field.String("storage_volume").
			Comment("the storage volume of the file which typically will be the organization ID the file belongs to - this is not a literal volume but the overlay file system mapping").
			Optional(),
		field.String("storage_path").
			Comment("the storage path is the second-level directory of the file path, typically the correlating logical object ID the file is associated with; files can be stand alone objects and not always correlated to a logical one, so this path of the tree may be empty").
			Optional(),
		field.Bytes("file_contents").
			Comment("the contents of the file").
			Optional().
			Annotations(
				entgql.Skip(), // Don't return file content in GraphQL queries
			),
	}
}

// Edges of the File
func (File) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("files"),
		edge.From("organization", Organization.Type).
			Ref("files"),
		edge.From("group", Group.Type).
			Ref("files"),
		edge.From("contact", Contact.Type).
			Ref("files"),
		edge.From("entity", Entity.Type).
			Ref("files"),
		edge.From("usersetting", UserSetting.Type).
			Ref("files"),
		edge.From("organizationsetting", OrganizationSetting.Type).
			Ref("files"),
		edge.From("template", Template.Type).
			Ref("files"),
		edge.From("documentdata", DocumentData.Type).
			Ref("files"),
		edge.To("events", Event.Type),
		edge.From("program", Program.Type).
			Ref("files"),
	}
}

// Mixin of the File
func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewObjectOwnedMixin(ObjectOwnedMixin{
			HookFuncs: []HookFunc{}, // use an empty hook, file processing is handled in middleware
		}),
	}
}

// Annotations of the File
func (File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "file",
			IncludeHooks: false,
		},
	}
}

// Policy of the File
func (File) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.OnMutationOperation(
				privacy.FileMutationRuleFunc(func(ctx context.Context, m *generated.FileMutation) error {
					return m.CheckAccessForEdit(ctx)
				}),
				// check permissions on delete and update operations, creation is handled by the parent object
				ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne,
			),
			privacy.AlwaysAllowRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.FileQueryRuleFunc(func(ctx context.Context, q *generated.FileQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
