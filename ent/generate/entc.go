//go:build ignore

// See Upstream docs for more details: https://entgo.io/docs/code-gen/#use-entc-as-a-package
package main

import (
	"embed"
	"os"

	"github.com/rs/zerolog/log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	_ "github.com/jackc/pgx/v5"
	"gocloud.dev/secrets"

	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/ent/entconfig"
	"github.com/theopenlane/ent/entitlements/genfeatures"
	"github.com/theopenlane/ent/validator"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/genhooks"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/shared/directives"
	"github.com/theopenlane/shared/entitlements"
	"github.com/theopenlane/shared/enums/exportenums"
	"github.com/theopenlane/shared/genhelpers"
	"github.com/theopenlane/shared/objects/objstore"
	"github.com/theopenlane/shared/soiree"
	"github.com/theopenlane/shared/summarizer"
)

const (
	entGeneratedDir = "./ent/generated"
	graphDir        = "./internal/graphapi/"
	graphSchemaDir  = graphDir + "schema/"
	graphQueryDir   = graphDir + "query/"
	schemaPath      = "./ent/schema"
	templateDir     = "./ent/generate/templates/ent"
	featureMapDir   = "./ent/entitlements/features/"
)

func main() {
	// setup logging
	genhelpers.SetupLogging()

	// change to the root of the repo so that the config hierarchy is correct
	genhelpers.ChangeToRootDir("../../")

	if err := os.Mkdir("schema", 0755); err != nil && !os.IsExist(err) {
		log.Fatal().Err(err).Msg("creating schema directory")
	}

	// generate the history schemas
	historyExt, entfgaExt := preRun()
	schemaHooks := []entgql.SchemaHook{}

	xExt, err := entx.NewExtension(entx.WithJSONScalar())
	if err != nil {
		log.Fatal().Err(err).Msg("creating entx extension")
	}

	schemaHooks = append(schemaHooks, xExt.GQLSchemaHooks()...)

	dExt, err := directives.NewExtension()
	if err != nil {
		log.Fatal().Err(err).Msg("creating directives extension")
	}

	schemaHooks = append(schemaHooks, dExt.SchemaHooks()...)

	gqlExt, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath(graphSchemaDir+"ent.graphql"),
		entgql.WithConfigPath(graphDir+"/generate/.gqlgen.yml"),
		entgql.WithWhereInputs(true),
		entgql.WithSchemaHook(schemaHooks...),
		WithGqlWithTemplates(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("creating entgql extension")
	}

	accessMapExt := accessmap.New(
		accessmap.WithSchemaPath(schemaPath),
		accessmap.WithGeneratedDir(entGeneratedDir),
	)

	log.Info().Msg("running ent codegen with extensions")

	if err := entc.Generate(schemaPath, &gen.Config{
		Target: entGeneratedDir,
		Hooks: []gen.Hook{
			genhooks.GenSchema(graphSchemaDir),
			genhooks.GenQuery(graphQueryDir),
			genhooks.GenSearchSchema(
				genhooks.WithGraphQueryDir(graphQueryDir),
				genhooks.WithGraphSchemaDir(graphSchemaDir),
				genhooks.WithIncludeAdminSearch(false)),
			accessMapExt.Hook(),
			exportenums.New().Hook(),
		},
		Package: "github.com/theopenlane/ent/generated",
		Features: []gen.Feature{
			gen.FeatureVersionedMigration,
			gen.FeaturePrivacy,
			gen.FeatureEntQL,
			gen.FeatureNamedEdges,
			gen.FeatureSchemaConfig,
			gen.FeatureIntercept,
			gen.FeatureModifier,
			// this is disabled because it is not compatible with the entcache driver
			// gen.FeatureExecQuery,
		},
	},
		entc.Dependency(
			entc.DependencyName("EntConfig"),
			entc.DependencyType(&entconfig.Config{}),
		),
		entc.Dependency(
			entc.DependencyName("Secrets"),
			entc.DependencyType(&secrets.Keeper{}),
		),
		entc.Dependency(
			entc.DependencyName("Authz"),
			entc.DependencyType(fgax.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("TokenManager"),
			entc.DependencyType(&tokens.TokenManager{}),
		),
		entc.Dependency(
			entc.DependencyName("SessionConfig"),
			entc.DependencyType(&sessions.SessionConfig{}),
		),
		entc.Dependency(
			entc.DependencyName("Emailer"),
			entc.DependencyType(&emailtemplates.Config{}),
		),
		entc.Dependency(
			entc.DependencyName("TOTP"),
			entc.DependencyType(&totp.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("EntitlementManager"),
			entc.DependencyType(&entitlements.StripeClient{}),
		),
		entc.Dependency(
			entc.DependencyName("ObjectManager"),
			entc.DependencyType(&objstore.Service{}),
		),
		entc.Dependency(
			entc.DependencyName("Summarizer"),
			entc.DependencyType(&summarizer.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("PondPool"),
			entc.DependencyType(&soiree.PondPool{}),
		),
		entc.Dependency(
			entc.DependencyName("EmailVerifier"),
			entc.DependencyType(&validator.EmailVerifier{}),
		),
		entc.TemplateDir(templateDir),
		entc.Extensions(
			gqlExt,
			historyExt,
			entfgaExt,
		)); err != nil {
		log.Fatal().Err(err).Msg("running ent codegen")
	}

	if err := genfeatures.GenerateModulePerSchema(schemaPath, featureMapDir); err != nil {
		log.Fatal().Err(err).Msg("generating module per schema")
	}
}

// preRun runs before the ent codegen to generate the history schemas and authz checks
// and returns the history and fga extensions to be used in the ent codegen
func preRun() (*history.Extension, *entfga.AuthzExtension) {
	// generate the history schemas
	log.Info().Msg("pre-run: generating the history schemas")

	historyExt := history.New(
		history.WithAuditing(),
		history.WithImmutableFields(),
		history.WithHistoryTimeIndex(),
		history.WithNillableFields(),
		history.WithGQLQuery(),
		history.WithAuthzPolicy(),
		history.WithSchemaPath(schemaPath),
		history.WithFirstRun(true),
		history.WithAllowedRelation("audit_log_viewer"),
		history.WithUpdatedByFromSchema(history.ValueTypeString, false),
	)

	if err := historyExt.GenerateSchemas(); err != nil {
		log.Fatal().Err(err).Msg("generating history schema")
	}

	log.Info().Msg("pre-run: generating the authz checks")

	// initialize the entfga extension
	entfgaExt := entfga.New(
		entfga.WithSoftDeletes(),
		entfga.WithSchemaPath(schemaPath),
		entfga.WithGeneratedPath(entGeneratedDir),
	)

	// generate authz checks, this needs to happen before we regenerate the schema
	// because the authz checks are generated based on the schema
	if err := entfgaExt.GenerateAuthzChecks(); err != nil {
		log.Fatal().Err(err).Msg("generating authz checks")
	}

	log.Info().Msg("pre-run: generating the history schemas with authz checks")

	// run again with policy
	historyExt.SetFirstRun(false)

	// generate the updated history schemas with authz checks
	if err := historyExt.GenerateSchemas(); err != nil {
		log.Fatal().Err(err).Msg("generating history schema")
	}

	log.Info().Msg("pre-run: generating exportable validation and enum")

	// generate exportable schemas validation using existing entx method
	exportableGen := entx.NewExportableGenerator(schemaPath, "ent/hooks").
		WithPackage("hooks")

	if err := exportableGen.Generate(); err != nil {
		log.Fatal().Err(err).Msg("generating exportable validation")
	}

	return historyExt, entfgaExt
}

var (
	//go:embed templates/entgql/*.tmpl
	_entqlTemplates embed.FS
)

// WithGqlWithTemplates is a schema hook to replace entgql default template used for pagination
// The only change to the template is the function used to get the totalCount field uses
// CountIDs(ctx) instead of `Count(ctx)`. The rest is a direct copy of the default template from:
// https://github.com/ent/contrib/tree/master/entgql/template
func WithGqlWithTemplates() entgql.ExtensionOption {
	paginationTmpl := gen.MustParse(gen.NewTemplate("node").
		Funcs(entgql.TemplateFuncs).ParseFS(_entqlTemplates, "templates/entgql/gql_where.tmpl", "templates/entgql/pagination.tmpl"))
	return entgql.WithTemplates(append(entgql.AllTemplates, paginationTmpl)...)
}
