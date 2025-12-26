package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/common/models"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// File defines the file schema.
type File struct {
	SchemaFuncs

	ent.Schema
}

const SchemaFile = "file"

func (File) Name() string {
	return SchemaFile
}

func (File) GetType() any {
	return File.Type
}

func (File) PluralName() string {
	return pluralize.NewClient().Plural(SchemaFile)
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
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the file").
			Optional(),
		field.String("storage_region").
			Comment("the region the file is stored in, if applicable").
			Optional(),
		field.String("storage_provider").
			Comment("the storage provider the file is stored in, if applicable").
			Optional(),
		field.Time("last_accessed_at").
			Optional().
			Annotations(
				entgql.OrderField("last_accessed_at"),
			).
			Nillable(),
	}
}

// Edges of the File
func (f File) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFrom(f, Organization{}),
		defaultEdgeFromWithPagination(f, Group{}),
		defaultEdgeFrom(f, Contact{}),
		defaultEdgeFrom(f, Entity{}),
		defaultEdgeFrom(f, OrganizationSetting{}),
		defaultEdgeFrom(f, Template{}),
		defaultEdgeFrom(f, DocumentData{}),
		defaultEdgeFrom(f, Program{}),
		defaultEdgeFrom(f, Evidence{}),
		defaultEdgeToWithPagination(f, Event{}),
		defaultEdgeFrom(f, TrustCenterSetting{}),
		defaultEdgeToWithPagination(f, Integration{}),
		defaultEdgeToWithPagination(f, Hush{}),
		defaultEdgeToWithPagination(f, TrustcenterEntity{}),
	}
}

// Mixin of the File
func (f File) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.File](f,
				withParents(
					Organization{}, Program{}, Control{}, Procedure{}, Template{}, Subcontrol{}, DocumentData{},
					Contact{}, InternalPolicy{}, Narrative{}, Evidence{}, TrustCenterSetting{}, Subprocessor{}, Export{},
					TrustCenterDoc{}, Standard{}, TrustcenterEntity{}), // used to create parent tuples for the file
				withHookFuncs(), // use an empty hook, file processing is handled in middleware
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(f)
}

func (File) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the File
func (f File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the File
func (f File) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorPresignedURL(),
		interceptors.InterceptorFile(), // filter on the organization id
	}
}

// Policy of the File
func (f File) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithOnMutationRules(
			// check permissions on delete and update operations, creation is handled by the parent object
			ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne,
			entfga.CheckEditAccess[*generated.FileMutation](),
		),
		policy.WithMutationRules(
			policy.AllowCreate(),
		),
	)
}

// Hooks of the File
func (File) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookFileDelete(),
	}
}
