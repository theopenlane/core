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

	buildFlags = "-tags=codegen"
)

const (
	graphDir       = "./internal/graphapi/"
	graphSchemaDir = graphDir + "schema/"

	graphHistorySchemaDir = graphDir + "schemahistory/"
	graphQueryDir         = graphDir + "query/"
	graphHistoryQueryDir  = graphDir + "query/history/"

	schemaPath           = "./internal/ent/schema"
	mixinPath            = "./internal/ent/mixin"
	historySchemaPath    = "./internal/ent/historyschema"
	entTemplatesPath     = "./internal/ent/generate/templates"
	entGenerateConfigDir = "./internal/ent/generate"

	templateDir   = "./internal/ent/generate/templates/ent"
	featureMapDir = "./internal/entitlements/features/"

	entGeneratedPath        = "internal/ent/generated"
	entGeneratedHistoryPath = "internal/ent/historygenerated"
	entGeneratedAuthzPath   = "internal/ent/authzgenerated"

	schemaInputChecksumFile  = "./internal/ent/checksum/.schema_checksum"
	historyInputChecksumFile = "./internal/ent/checksum/.history_schema_checksum"
)

var (
	// changes to these paths should trigger full schema generation
	mainSchemaInputPaths = []string{
		schemaPath,
		mixinPath,
		entTemplatesPath,
		entGenerateConfigDir,
	}

	// changes to these paths should trigger history schema generation
	historySchemaInputPaths = []string{
		schemaPath,
		entGeneratedPath,
		entGenerateConfigDir,
	}
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

	log.Info().Msg("running ent codegen")

	// check if there were schema changes before running full codegen
	hasChanges, err := genhelpers.HasSchemaChanges(schemaInputChecksumFile, mainSchemaInputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for schema changes, running history generation anyway")
		hasChanges = true
	}

	if hasChanges {
		schemaGenerate(getEntfgaExtension(hasChanges), getEntGqlExtension(), getHistoryExtension(hasChanges))
	} else {
		log.Info().Msg("no schema changes detected, skipping main schema codegen")
	}

	// only run if there were changes to the internal/ent/generated or internal/ent/schema directories
	hasChangesForHistory, err := genhelpers.HasSchemaChanges(historyInputChecksumFile, historySchemaInputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for schema changes, running history generation anyway")
		hasChangesForHistory = true
	}

	if hasChangesForHistory {
		historySchemaGenerate(getEntHistoryGqlExtension())
	} else {
		log.Info().Msg("no schema changes detected, skipping history generation")
	}

	log.Info().Msg("generating module per schema for entitlements")
	if hasChanges {
		if err := genfeatures.GenerateModulePerSchema(schemaPath, featureMapDir); err != nil {
			log.Fatal().Err(err).Msg("generating module per schema")
		}
	}

	log.Info().Msg("ent codegen completed successfully, setting new schema checksums")

	// set final schema checksum
	if hasChanges {
		err := genhelpers.SetSchemaChecksum(schemaInputChecksumFile, mainSchemaInputPaths...)
		if err != nil {
			log.Warn().Err(err).Msg("error setting schema checksum")
		}
	}

	if hasChangesForHistory {
		err := genhelpers.SetSchemaChecksum(historyInputChecksumFile, historySchemaInputPaths...)
		if err != nil {
			log.Warn().Err(err).Msg("error setting history schema checksum")
		}
	}
}

// WithGqlWithTemplates is a schema hook to replace entgql default template used for pagination
// The only change to the template is the function used to get the totalCount field uses
// CountIDs(ctx) instead of `Count(ctx)`. The rest is a direct copy of the default template from:
// https://github.com/ent/contrib/tree/master/entgql/template
// 12/18/2025 MKA - This was modified to remove the use of prepareQuery and withInterceptors to prevent duplicate query execution
func WithGqlWithTemplates() entgql.ExtensionOption {
	paginationTmpl := gen.MustParse(gen.NewTemplate("node").
		Funcs(entgql.TemplateFuncs).ParseFS(_entqlTemplates, "templates/entgql/gql_where.tmpl", "templates/entgql/pagination.tmpl"))
	return entgql.WithTemplates(append(entgql.AllTemplates, paginationTmpl)...)
}

func getEntfgaExtension(hasChanges bool) *entfga.AuthzExtension {
	entfgaExt := entfga.New(
		entfga.WithSoftDeletes(),
		entfga.WithSchemaPath(schemaPath),
	)

	if hasChanges {
		// generate authz checks, this needs to happen before we regenerate the schema
		// because the authz checks are generated based on the schema
		if err := entfgaExt.GenerateAuthzChecks(); err != nil {
			log.Fatal().Err(err).Msg("generating authz checks")
		}
	} else {
		log.Info().Msg("no schema changes detected, skipping authz check generation")
	}

	return entfgaExt
}

func getEntGqlExtension() *entgql.Extension {
	// initialize schema hooks for entgql
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

	return gqlExt
}

func getEntHistoryGqlExtension() *entgql.Extension {
	gqlExt, err := entgql.NewExtension(
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath(graphHistorySchemaDir+"ent.graphql"),
		entgql.WithConfigPath(graphDir+"/generate/.gqlgen_history.yml"),
		entgql.WithWhereInputs(true),
		WithGqlWithTemplates(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("creating entgql extension")
	}

	return gqlExt
}

// getHistoryExtension generates the history schemas and returns the history extension to be used in the ent codegen
func getHistoryExtension(hasChanges bool) *history.Extension {
	// generate the history schemas
	log.Info().Msg("creating history extension")

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

	if hasChanges {
		log.Info().Msg("main schema changes detected, updating history schemas")
		if err := historyExt.GenerateSchemas(buildFlags); err != nil {
			log.Fatal().Err(err).Msg("generating history schemas")
		}
	} else {
		log.Info().Msg("no schema changes detected, skipping history schema generation")
	}

	return historyExt
}

func exportableSchema() {
	// generate exportable schemas validation using existing entx method
	exportableGen := entx.NewExportableGenerator(schemaPath, "internal/ent/hooks").
		WithPackage("hooks")

	if err := exportableGen.Generate(buildFlags); err != nil {
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
		Header: "// Code generated by ent, DO NOT EDIT.\n",
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
		Package:    "github.com/theopenlane/core/" + entGeneratedPath,
		Features:   enabledFeatures,
		BuildFlags: []string{buildFlags},
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
	log.Info().Msg("generating history schema codegen")
	if err := entc.Generate(historySchemaPath, &gen.Config{
		Target: "./" + entGeneratedHistoryPath,
		Header: "//go:build !enthistorycodegen\n// Code generated by ent, DO NOT EDIT.\n",
		Hooks: []gen.Hook{
			genhooks.GenSchema(graphHistorySchemaDir),
			genhooks.GenQuery(graphHistoryQueryDir),
		},
		Package:    "github.com/theopenlane/core/" + entGeneratedHistoryPath,
		Features:   enabledFeatures,
		BuildFlags: []string{buildFlags},
	},
		entc.Dependency(
			entc.DependencyName("EntConfig"),
			entc.DependencyType(&entconfig.Config{}),
		),
		entc.Dependency(
			entc.DependencyName("Authz"),
			entc.DependencyType(fgax.Client{}),
		),
		entc.TemplateFiles(templateDir+"/client.tmpl", templateDir+"/config.tmpl", templateDir+"/count.tmpl"),
		entc.Extensions(
			extensions...,
		)); err != nil {
		log.Fatal().Err(err).Msg("running ent codegen")
	}
}
