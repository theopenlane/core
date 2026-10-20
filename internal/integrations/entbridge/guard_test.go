package entbridge

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// installationExcluded lists the integration columns deliberately left out of types.Installation.
// A new column fails TestInstallationCoversEntColumns until it is either added to the view (and its
// conversion) or listed here with the reason it stays internal; a removed column fails to compile
var installationExcluded = map[string]string{
	integration.FieldCreatedBy:                "audit metadata",
	integration.FieldUpdatedBy:                "audit metadata",
	integration.FieldUpdatedByImpersonator:    "audit metadata",
	integration.FieldDeletedAt:                "soft-delete bookkeeping",
	integration.FieldDeletedBy:                "soft-delete bookkeeping",
	integration.FieldTags:                     "user tagging, not definition-facing",
	integration.FieldInternalNotes:            "staff-only notes",
	integration.FieldSystemInternalID:         "platform-internal identifier",
	integration.FieldEnvironmentName:          "environment scoping handled by the runtime",
	integration.FieldEnvironmentID:            "environment scoping handled by the runtime",
	integration.FieldScopeName:                "scope handled by the runtime",
	integration.FieldScopeID:                  "scope handled by the runtime",
	integration.FieldKind:                     "legacy classification",
	integration.FieldIntegrationType:          "legacy classification",
	integration.FieldPlatformID:               "platform linkage handled by ingest",
	integration.FieldProviderMetadata:         "catalog metadata, available from the definition",
	integration.FieldHealth:                   "health is owned by the runtime",
	integration.FieldDefinitionSlug:           "derived from the definition",
	integration.FieldProviderMetadataSnapshot: "catalog snapshot, available from the definition",
}

// webhookExcluded lists the integration webhook columns deliberately left out of types.WebhookEndpoint
var webhookExcluded = map[string]string{
	integrationwebhook.FieldCreatedAt:             "audit metadata",
	integrationwebhook.FieldUpdatedAt:             "audit metadata",
	integrationwebhook.FieldCreatedBy:             "audit metadata",
	integrationwebhook.FieldUpdatedBy:             "audit metadata",
	integrationwebhook.FieldUpdatedByImpersonator: "audit metadata",
	integrationwebhook.FieldDeletedAt:             "soft-delete bookkeeping",
	integrationwebhook.FieldDeletedBy:             "soft-delete bookkeeping",
	integrationwebhook.FieldProvider:              "redundant with the installation's definition",
	integrationwebhook.FieldLastDeliveryAt:        "delivery tracking owned by the runtime",
	integrationwebhook.FieldLastDeliveryStatus:    "delivery tracking owned by the runtime",
	integrationwebhook.FieldLastDeliveryError:     "delivery tracking owned by the runtime",
}

func TestInstallationCoversEntColumns(t *testing.T) {
	t.Parallel()

	assertViewCovers(t, integration.Columns, reflect.TypeFor[ent.Integration](), reflect.TypeFor[types.Installation](), installationExcluded)
}

func TestWebhookCoversEntColumns(t *testing.T) {
	t.Parallel()

	assertViewCovers(t, integrationwebhook.Columns, reflect.TypeFor[ent.IntegrationWebhook](), reflect.TypeFor[types.WebhookEndpoint](), webhookExcluded)
}

func TestInstallationCopiesEveryMappedColumn(t *testing.T) {
	t.Parallel()

	record := &ent.Integration{}
	fillMapped(t, integration.Columns, reflect.ValueOf(record).Elem(), installationExcluded)

	assertCopied(t, reflect.ValueOf(record).Elem(), reflect.ValueOf(Installation(record)).Elem())
}

func TestWebhookCopiesEveryMappedColumn(t *testing.T) {
	t.Parallel()

	record := &ent.IntegrationWebhook{}
	fillMapped(t, integrationwebhook.Columns, reflect.ValueOf(record).Elem(), webhookExcluded)

	assertCopied(t, reflect.ValueOf(record).Elem(), reflect.ValueOf(Webhook(record)).Elem())
}

// assertViewCovers requires every column to be mapped into the view under its ent field name and type, or excluded,
// and every view field to correspond to a column
func assertViewCovers(t *testing.T, columns []string, entType, viewType reflect.Type, excluded map[string]string) {
	t.Helper()

	fields := entFieldsByColumn(t, columns, entType)

	for _, column := range columns {
		field := fields[column]
		_, isExcluded := excluded[column]
		viewField, inView := viewType.FieldByName(field.Name)

		switch {
		case isExcluded:
			require.Falsef(t, inView, "column %s is excluded but %s.%s exists", column, viewType.Name(), field.Name)
		case !inView:
			require.Failf(t, "unmapped column", "column %s (%s.%s) is not in %s: add it to the view and its conversion, or exclude it with a reason", column, entType.Name(), field.Name, viewType.Name())
		default:
			require.Equalf(t, field.Type, viewField.Type, "%s.%s has a different type than column %s", viewType.Name(), field.Name, column)
		}
	}

	byName := map[string]bool{}
	for _, field := range fields {
		byName[field.Name] = true
	}

	for i := range viewType.NumField() {
		name := viewType.Field(i).Name
		require.Truef(t, byName[name], "%s.%s does not correspond to a %s column", viewType.Name(), name, entType.Name())
	}
}

// entFieldsByColumn resolves each column to its ent struct field by json tag, falling back to the snake-cased
// field name for fields whose json tag is suppressed (such as secrets)
func entFieldsByColumn(t *testing.T, columns []string, entType reflect.Type) map[string]reflect.StructField {
	t.Helper()

	index := map[string]reflect.StructField{}

	for i := range entType.NumField() {
		field := entType.Field(i)
		if !field.IsExported() {
			continue
		}

		if tag, _, _ := strings.Cut(field.Tag.Get("json"), ","); tag != "" && tag != "-" {
			index[tag] = field
		}

		if _, ok := index[snake(field.Name)]; !ok {
			index[snake(field.Name)] = field
		}
	}

	out := make(map[string]reflect.StructField, len(columns))

	for _, column := range columns {
		field, ok := index[column]
		require.Truef(t, ok, "column %s has no field on %s", column, entType.Name())

		out[column] = field
	}

	return out
}

var (
	snakeBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snakeAcronym  = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
)

// snake converts an ent Go field name to its column name (EndpointURL -> endpoint_url)
func snake(name string) string {
	return strings.ToLower(snakeBoundary.ReplaceAllString(snakeAcronym.ReplaceAllString(name, "${1}_${2}"), "${1}_${2}"))
}

// assertCopied requires every view field to equal the same-named field on the ent record
func assertCopied(t *testing.T, record, view reflect.Value) {
	t.Helper()

	for i := range view.NumField() {
		name := view.Type().Field(i).Name

		require.Truef(t, reflect.DeepEqual(record.FieldByName(name).Interface(), view.Field(i).Interface()), "%s is not copied by the conversion", name)
	}
}

// fillMapped sets every mapped column's field to a distinct non-zero value so a missed copy shows up as a zero value
func fillMapped(t *testing.T, columns []string, record reflect.Value, excluded map[string]string) {
	t.Helper()

	for column, field := range entFieldsByColumn(t, columns, record.Type()) {
		if _, skip := excluded[column]; skip {
			continue
		}

		setNonZero(record.FieldByName(field.Name), column, 0)
	}
}

// setNonZero assigns a deterministic non-zero value derived from seed to v
func setNonZero(v reflect.Value, seed string, depth int) {
	if depth > 4 || !v.CanSet() {
		return
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString("v-" + seed)
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(len(seed) + 1))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(len(seed) + 1))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(len(seed)) + 0.5)
	case reflect.Pointer:
		elem := reflect.New(v.Type().Elem())
		setNonZero(elem.Elem(), seed, depth+1)
		v.Set(elem)
	case reflect.Slice:
		slice := reflect.MakeSlice(v.Type(), 1, 1)
		setNonZero(slice.Index(0), seed, depth+1)
		v.Set(slice)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		key := reflect.New(v.Type().Key()).Elem()
		val := reflect.New(v.Type().Elem()).Elem()
		setNonZero(key, seed, depth+1)
		setNonZero(val, seed, depth+1)
		m.SetMapIndex(key, val)
		v.Set(m)
	case reflect.Interface:
		v.Set(reflect.ValueOf("v-" + seed))
	case reflect.Struct:
		if v.Type() == reflect.TypeFor[time.Time]() {
			v.Set(reflect.ValueOf(time.Date(2026, 1, 2, 3, 4, 5, len(seed), time.UTC)))

			return
		}

		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				setNonZero(v.Field(i), seed+"."+v.Type().Field(i).Name, depth+1)
			}
		}
	}
}
