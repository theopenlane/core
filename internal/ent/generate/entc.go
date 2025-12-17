//go:build ignore

// See Upstream docs for more details: https://entgo.io/docs/code-gen/#use-entc-as-a-package

package main

import (
	"embed"
	"flag"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/vektah/gqlparser/v2/ast"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	_ "github.com/jackc/pgx/v5"
	"gocloud.dev/secrets"

	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/internal/ent/historygenerated"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/internal/entitlements/genfeatures"
	"github.com/theopenlane/core/internal/genhelpers"
	"github.com/theopenlane/core/internal/graphapi/directives"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/enums/exportenums"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/summarizer"
	"github.com/theopenlane/core/pkg/windmill"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/genhooks"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/iam/totp"
)

var (
	//go:embed templates/entgql/*.tmpl
	_entqlTemplates embed.FS

	// add flags for skipping parts of generation for simple dev runs
	skipHistory = flag.Bool("skip-history", false, "skip history schema generation")
	skipModules = flag.Bool("skip-modules", false, "skip module per schema generation")

	onlySchemas = flag.Bool("only-schemas", false, "only generate base schema, skip history, modules, and exportable validation")
)

const (
	graphDir       = "./internal/graphapi/"
	graphSchemaDir = graphDir + "schema/"

	graphHistorySchemaDir = graphDir + "schemahistory/"
	graphQueryDir         = graphDir + "query/"
	graphHistoryQueryDir  = graphDir + "query/history/"

	schemaPath        = "./internal/ent/schema"
	historySchemaPath = "./internal/ent/historyschema"

	templateDir   = "./internal/ent/generate/templates/ent"
	featureMapDir = "./internal/entitlements/features/"

	entGeneratedPath        = "internal/ent/generated"
	entGeneratedHistoryPath = "internal/ent/historygenerated"
	entGeneratedAuthzPath   = "internal/ent/authzgenerated"
)

var enabledFeatures = []gen.Feature{
	gen.FeatureVersionedMigration,
	gen.FeaturePrivacy,
	gen.FeatureEntQL,
	gen.FeatureNamedEdges,
	gen.FeatureSchemaConfig,
	gen.FeatureIntercept,
	gen.FeatureModifier,
	// this is disabled because it is not compatible with the entcache driver
	// gen.FeatureExecQuery,
}

func main() {
	// setup logging
	genhelpers.SetupLogging()

	// change to the root of the repo so that the config hierarchy is correct
	genhelpers.ChangeToRootDir("../../../")

	if err := os.Mkdir("schema", 0755); err != nil && !os.IsExist(err) {
		log.Fatal().Err(err).Msg("creating schema directory")
	}

	log.Info().Msg("running ent codegen with extensions")

	// add the history hook on the main schema because the auditing template has be
	// based off of the main schema
	schemaGenerate(getEntfgaExtension(), getEntGqlExtension(), getHistoryExtension())
	if *onlySchemas {
		// if only generating schemas, return now
		log.Info().Msg("only-schemas flag set, skipping history and module generation")
		return
	}

	log.Info().Msg("running ent history codegen")

	if !*skipHistory {
		historySchemaGenerate(getEntHistoryGqlExtension())
	} else {
		log.Info().Msg("skipping history schema generation")
	}

	log.Info().Msg("generating module per schema for entitlements")

	if !*skipModules {
		if err := genfeatures.GenerateModulePerSchema(schemaPath, featureMapDir); err != nil {
			log.Fatal().Err(err).Msg("generating module per schema")
		}
	} else {
		log.Info().Msg("skipping module per schema generation")
	}
}

// WithGqlWithTemplates is a schema hook to replace entgql default template used for pagination
// The only change to the template is the function used to get the totalCount field uses
// CountIDs(ctx) instead of `Count(ctx)`. The rest is a direct copy of the default template from:
// https://github.com/ent/contrib/tree/master/entgql/template
func WithGqlWithTemplates() entgql.ExtensionOption {
	paginationTmpl := gen.MustParse(gen.NewTemplate("node").
		Funcs(entgql.TemplateFuncs).ParseFS(_entqlTemplates, "templates/entgql/gql_where.tmpl", "templates/entgql/pagination.tmpl"))
	return entgql.WithTemplates(append(entgql.AllTemplates, paginationTmpl)...)
}

func getEntfgaExtension() *entfga.AuthzExtension {
	entfgaExt := entfga.New(
		entfga.WithSoftDeletes(),
		entfga.WithSchemaPath(schemaPath),
	)

	// generate authz checks, this needs to happen before we regenerate the schema
	// because the authz checks are generated based on the schema
	if err := entfgaExt.GenerateAuthzChecks(); err != nil {
		log.Fatal().Err(err).Msg("generating authz checks")
	}

	return entfgaExt
}

func getEntGqlExtension() *entgql.Extension {
	// initialize schema hooks for entgql
	schemaHooks := []entgql.SchemaHook{
		removeExtraScalars,
	}

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

	return gqlExt
}

func getEntHistoryGqlExtension() *entgql.Extension {
	schemaHooks := []entgql.SchemaHook{
		removeExtraScalars,
	}

	gqlExt, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath(graphHistorySchemaDir+"ent.graphql"),
		entgql.WithConfigPath(graphDir+"/generate/.gqlgen_history.yml"),
		entgql.WithWhereInputs(true),
		entgql.WithSchemaHook(schemaHooks...),
		WithGqlWithTemplates(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("creating entgql extension")
	}

	return gqlExt
}

// getHistoryExtension generates the history schemas and returns the history extension to be used in the ent codegen
func getHistoryExtension() *history.Extension {
	// generate the history schemas
	log.Info().Msg("generating the history schemas")

	historyExt := history.New(
		history.WithImmutableFields(),
		history.WithHistoryTimeIndex(),
		history.WithNillableFields(),
		history.WithGQLQuery(),
		history.WithAuthzPolicy(),
		history.WithInputSchemaPath(schemaPath),
		history.WithOutputSchemaPath(historySchemaPath),
		history.WithPackageName("historyschema"),
		history.WithFirstRun(false),
		history.WithAllowedRelation("audit_log_viewer"),
		history.WithUpdatedByFromSchema(history.ValueTypeString, false),
	)

	if err := historyExt.GenerateSchemas(); err != nil {
		log.Fatal().Err(err).Msg("generating history schema")
	}

	return historyExt
}

func exportableSchema() {
	// generate exportable schemas validation using existing entx method
	exportableGen := entx.NewExportableGenerator(schemaPath, "internal/ent/hooks").
		WithPackage("hooks")

	if err := exportableGen.Generate(); err != nil {
		log.Fatal().Err(err).Msg("generating exportable validation")
	}
}

func schemaGenerate(extensions ...entc.Extension) {
	accessMapExt := accessmap.New(
		accessmap.WithSchemaPath(schemaPath),
		accessmap.WithGeneratedDir(entGeneratedAuthzPath),
		accessmap.WithPackageName("authzgenerated"),
	)

	if err := entc.Generate(schemaPath, &gen.Config{
		Target: "./" + entGeneratedPath,
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
		Package:  "github.com/theopenlane/core/" + entGeneratedPath,
		Features: enabledFeatures,
	},
		entc.Dependency(
			entc.DependencyName("EntConfig"),
			entc.DependencyType(&entconfig.Config{}),
		),
		entc.Dependency(
			entc.DependencyName("HistoryClient"),
			entc.DependencyType(&historygenerated.Client{}),
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
			entc.DependencyType(&objects.Service{}),
		),
		entc.Dependency(
			entc.DependencyName("Summarizer"),
			entc.DependencyType(&summarizer.Client{}),
		),
		entc.Dependency(
			entc.DependencyName("Windmill"),
			entc.DependencyType(&windmill.Client{}),
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
			extensions...,
		)); err != nil {
		log.Fatal().Err(err).Msg("running ent codegen")
	}

	// generate exportable schema validation
	log.Info().Msg("generating exportable schema validation")
	exportableSchema()
}

func historySchemaGenerate(extensions ...entc.Extension) {
	log.Info().Msg("running ent history codegen")
	if err := entc.Generate(historySchemaPath, &gen.Config{
		Target: "./" + entGeneratedHistoryPath,
		Hooks: []gen.Hook{
			genhooks.GenSchema(graphHistorySchemaDir),
			genhooks.GenQuery(graphHistoryQueryDir),
		},
		Package:  "github.com/theopenlane/core/" + entGeneratedHistoryPath,
		Features: enabledFeatures,
	},
		entc.Dependency(
			entc.DependencyName("EntConfig"),
			entc.DependencyType(&entconfig.Config{}),
		),
		entc.TemplateFiles(templateDir+"/client.tmpl", templateDir+"/config.tmpl", templateDir+"/count.tmpl"),
		entc.Extensions(
			extensions...,
		)); err != nil {
		log.Fatal().Err(err).Msg("running ent codegen")
	}
}

// TODO: see if this does anything?
var removeExtraScalars entgql.SchemaHook = func(_ *gen.Graph, s *ast.Schema) error {
	// Remove scalar *definitions* from the generated schema file.
	// these scalars are defined already in the common graphql schema file
	for _, name := range []string{"JSON", "Map", "Cursor", "Time"} {
		def, ok := s.Types[name]
		if ok && def.Kind == ast.Scalar {
			delete(s.Types, name)
		}
	}
	return nil
}
