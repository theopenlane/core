//go:build generate

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/invopop/yaml"
	"github.com/mcuadros/go-defaults"

	"github.com/theopenlane/utils/envparse"

	"github.com/theopenlane/core/config"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

// const values used for the schema generator
const (
	tagName                 = "koanf"
	skipper                 = "-"
	defaultTag              = "default"
	jsonSchemaPath          = "./jsonschema/core.config.json"
	yamlConfigPath          = "./config/config.example.yaml"
	envConfigPath           = "./config/.env.example"
	configFileConfigMapPath = "./config/configmap-config-file.yaml"
	externalSecretsDir      = "./config/external-secrets" // #nosec G101 - this is a directory path, not a secret
	helmValuesPath          = "./config/helm-values.yaml"
	sensitiveTag            = "sensitive"
	varPrefix               = "CORE"
	ownerReadWrite          = 0600
	dirPermission           = 0755
)

// includedPackages is a list of packages to include in the schema generation
// that contain Go comments to be added to the schema
// any external packages must use the jsonschema description tags to add comments
var includedPackages = []string{
	"./config",
	"./ent",
	"./ent/entdb",
	"./internal/httpserve/handlers",
	"./pkg/middleware",
	"./pkg/objects",
	"./ent/entconfig",
	"./pkg/entitlements",
	"./pkg/summarizer",
}

// sensitiveFields lists configuration paths that are sensitive but reside in external packages
// SensitiveField represents a sensitive configuration field
type SensitiveField struct {
	Key        string // Environment variable key (e.g., CORE_AUTH_PROVIDERS_GITHUB_CLIENTSECRET)
	Path       string // Config path (e.g., auth.providers.github.clientSecret)
	SecretName string // Secret name for external secrets (e.g., github-client-secret)
}

// externalSensitiveFields maps configuration paths for sensitive fields from external packages
// that the envparse library cannot automatically detect due to struct tag traversal limitations
// Only includes fields where envparse fails to detect the sensitive:"true" tag
var externalSensitiveFields = map[string]struct{}{
	"core.auth.providers.github.clientSecret":          {},
	"core.auth.providers.google.clientSecret":          {},
	"core.authz.credentials.apiToken":                  {},
	"core.authz.credentials.clientSecret":              {},
	"core.totp.secret":                                 {},
	"core.entConfig.summarizer.llm.openai.apiKey":      {},
	"core.entConfig.summarizer.llm.anthropic.apiKey":   {},
	"core.entConfig.summarizer.llm.mistral.apiKey":     {},
	"core.entConfig.summarizer.llm.huggingface.apiKey": {},
	"core.entConfig.summarizer.llm.cloudflare.apiKey":  {},
}

// schemaConfig represents the configuration for the schema generator
type schemaConfig struct {
	// jsonSchemaPath represents the file path of the JSON schema to be generated
	jsonSchemaPath string
	// yamlConfigPath is the file path to the YAML configuration to be generated
	yamlConfigPath string
	// envConfigPath is the file path to the environment variable configuration to be generated
	envConfigPath string
	// configFileConfigMapPath is the file path to the kubernetes config map that embeds config.yaml
	configFileConfigMapPath string
	// externalSecretsDir is the directory for ExternalSecret files for Helm chart
	externalSecretsDir string
	// helmValuesPath is the file path to the helm values.yaml to be generated
	helmValuesPath string
}

// schemaOption defines a functional option for schemaConfig
type schemaOption func(*schemaConfig)

// newSchemaConfig creates a schemaConfig with provided options
func newSchemaConfig(opts ...schemaOption) schemaConfig {
	c := schemaConfig{
		jsonSchemaPath:          jsonSchemaPath,
		yamlConfigPath:          yamlConfigPath,
		envConfigPath:           envConfigPath,
		configFileConfigMapPath: configFileConfigMapPath,
		externalSecretsDir:      externalSecretsDir,
		helmValuesPath:          helmValuesPath,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// main is the entry point for the schema generator. It creates a new schemaConfig and generates all configuration files.
func main() {
	c := newSchemaConfig()

	// Generate all schema/config files from the config structure
	if err := generateSchema(c, &config.Config{}); err != nil {
		panic(err)
	}
}

// generateSchema generates all configuration files (JSON schema, YAML config, env file, config map, secrets, Helm values)
// from the provided structure and schemaConfig paths.
func generateSchema(c schemaConfig, structure interface{}) error {
	// Generate JSON schema file
	if err := generateJSONSchema(c.jsonSchemaPath, structure); err != nil {
		return err
	}

	// Generate YAML config file with defaults
	if err := generateYAMLConfig(c.yamlConfigPath); err != nil {
		return err
	}

	// Process environment variables and sensitive fields
	envVars, sensitiveFields, err := processEnvironmentVariables()
	if err != nil {
		return err
	}

	// Generate .env.example file
	if err := generateEnvironmentFile(c.envConfigPath, envVars); err != nil {
		return err
	}

	// Generate ConfigMap containing the rendered config.yaml content
	if err := generateConfigFileConfigMap(c.configFileConfigMapPath); err != nil {
		return err
	}

	// Generate secret files if any sensitive fields are present
	if len(sensitiveFields) > 0 {
		if err := generateSecretFiles(c, sensitiveFields); err != nil {
			return err
		}
	}

	// Generate Helm values.yaml file
	if err := generateAndWriteHelmValues(c.helmValuesPath, structure); err != nil {
		return err
	}

	return nil
}

// generateJSONSchema creates the JSON schema file from the config structure using the invopop/jsonschema package.
// It also attaches Go comments from included packages for documentation.
func generateJSONSchema(jsonSchemaPath string, structure interface{}) error {
	r := jsonschema.Reflector{Namer: namePkg}
	r.ExpandedStruct = true
	r.RequiredFromJSONSchemaTags = true
	r.FieldNameTag = tagName

	// Attach Go comments from included packages for schema documentation
	for _, pkg := range includedPackages {
		if err := r.AddGoComments("github.com/theopenlane/core/", pkg); err != nil {
			return fmt.Errorf("failed to add go comments for package %s: %w", pkg, err)
		}
	}

	// Reflect the structure to generate the schema
	s := r.Reflect(structure)

	// Marshal the schema to JSON
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON schema: %w", err)
	}

	// Write the JSON schema to file
	if err := os.WriteFile(jsonSchemaPath, data, ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write JSON schema file: %w", err)
	}

	return nil
}

// generateYAMLConfig creates the YAML configuration file with default values populated.
// It also ensures example map entries and initializes special config sections.
func generateYAMLConfig(yamlConfigPath string) error {
	yamlConfig := buildDefaultConfig()

	// Marshal the config struct to YAML
	yamlSchema, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML config: %w", err)
	}

	// Write the YAML config to file
	if err := os.WriteFile(yamlConfigPath, yamlSchema, ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write YAML config file: %w", err)
	}

	return nil
}

// buildDefaultConfig returns a config struct populated with default and example values.
func buildDefaultConfig() *config.Config {
	cfg := &config.Config{}
	defaults.SetDefaults(cfg)
	initializeStripeWebhookSecrets(cfg)
	initializeRateLimitOptions(cfg)
	return cfg
}

// initializeStripeWebhookSecrets ensures the stripe webhook secrets map includes entries for the current and discard API versions
func initializeStripeWebhookSecrets(cfg *config.Config) {
	if cfg == nil {
		return
	}

	ent := &cfg.Entitlements

	if ent.StripeWebhookSecrets == nil {
		ent.StripeWebhookSecrets = make(map[string]string)
	}

	versionKeys := []string{
		ent.StripeWebhookAPIVersion,
		ent.StripeWebhookDiscardAPIVersion,
	}

	for _, version := range versionKeys {
		if version == "" {
			continue
		}

		if _, exists := ent.StripeWebhookSecrets[version]; !exists {
			ent.StripeWebhookSecrets[version] = ""
		}
	}
}

func initializeRateLimitOptions(cfg *config.Config) {
	if cfg == nil {
		return
	}

	rl := &cfg.Ratelimit

	if len(rl.Options) == 0 {
		opt := ratelimit.RateOption{}
		defaults.SetDefaults(&opt)
		rl.Options = []ratelimit.RateOption{opt}
	} else {
		for i := range rl.Options {
			defaults.SetDefaults(&rl.Options[i])
		}
	}
}

// mapEntry represents a key/value pair discovered for map traversal
type mapEntry struct {
	Key   string
	Value reflect.Value
}

// getMapEntriesForPath returns sorted key/value pairs for the map identified by the provided config path
func getMapEntriesForPath(cfg *config.Config, fullPath string) ([]mapEntry, error) {
	if cfg == nil {
		return nil, nil
	}

	trimmed := strings.TrimPrefix(fullPath, "core.")
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ".")
	current := reflect.ValueOf(cfg).Elem()

	for _, part := range parts {
		if !current.IsValid() {
			return nil, nil
		}

		switch current.Kind() {
		case reflect.Struct:
			found := false
			for i := 0; i < current.NumField(); i++ {
				field := current.Type().Field(i)
				if field.Tag.Get(tagName) == part {
					current = current.Field(i)
					if current.Kind() == reflect.Ptr && !current.IsNil() {
						current = current.Elem()
					}
					found = true
					break
				}
			}
			if !found {
				return nil, nil
			}
		default:
			// Unsupported path traversal beyond structs for map discovery
			return nil, nil
		}
	}

	if current.Kind() != reflect.Map || current.IsNil() {
		return nil, nil
	}

	iter := current.MapRange()
	entries := make([]mapEntry, 0, current.Len())

	for iter.Next() {
		keyVal := iter.Key()
		valVal := iter.Value()

		if keyVal.Kind() != reflect.String {
			continue
		}

		entries = append(entries, mapEntry{
			Key:   keyVal.String(),
			Value: valVal,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	return entries, nil
}

// getSliceValuesForPath returns the slice of reflect.Values for the slice identified by the provided config path
func getSliceValuesForPath(cfg *config.Config, fullPath string) ([]reflect.Value, error) {
	if cfg == nil {
		return nil, nil
	}

	trimmed := strings.TrimPrefix(fullPath, "core.")
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ".")
	current := reflect.ValueOf(cfg).Elem()

	for _, part := range parts {
		if !current.IsValid() {
			return nil, nil
		}

		switch current.Kind() {
		case reflect.Struct:
			found := false
			for i := 0; i < current.NumField(); i++ {
				field := current.Type().Field(i)
				if field.Tag.Get(tagName) == part {
					current = current.Field(i)
					if current.Kind() == reflect.Ptr && !current.IsNil() {
						current = current.Elem()
					}
					found = true
					break
				}
			}
			if !found {
				return nil, nil
			}
		case reflect.Slice:
			idx, err := strconv.Atoi(part)
			if err != nil {
				return nil, nil
			}
			if idx < 0 || idx >= current.Len() {
				return nil, nil
			}
			current = current.Index(idx)
			if current.Kind() == reflect.Ptr && !current.IsNil() {
				current = current.Elem()
			}
		default:
			return nil, nil
		}
	}

	if current.Kind() != reflect.Slice {
		return nil, nil
	}

	values := make([]reflect.Value, 0, current.Len())
	for i := 0; i < current.Len(); i++ {
		values = append(values, current.Index(i))
	}

	return values, nil
}

// sanitizeMapKeyForEnv produces a stable suffix for environment variables from a map key
func sanitizeMapKeyForEnv(mapKey string) string {
	sanitized := strings.ToUpper(mapKey)
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return sanitized
}

// appendMapFieldEnvVars processes a map field and appends its entries to the environment variables
func appendMapFieldEnvVars(envVars *strings.Builder, cfg *config.Config, field envparse.VarInfo, isSecret bool, sensitiveFields *[]SensitiveField) error {
	mapEntries, err := getMapEntriesForPath(cfg, field.FullPath)
	if err != nil {
		return fmt.Errorf("failed to derive map entries for %s: %w", field.FullPath, err)
	}

	for _, entry := range mapEntries {
		if entry.Key == "" {
			continue
		}

		baseKey := fmt.Sprintf("%s_%s", field.Key, sanitizeMapKeyForEnv(entry.Key))
		basePath := fmt.Sprintf("%s.%s", field.FullPath, entry.Key)
		appendMapValueEnvVars(envVars, baseKey, basePath, entry.Value, isSecret, sensitiveFields)
	}

	return nil
}

// appendSliceStructFieldEnvVars processes a slice of structs field and appends its entries to the environment variables
func appendSliceStructFieldEnvVars(envVars *strings.Builder, cfg *config.Config, field envparse.VarInfo, isSecret bool, sensitiveFields *[]SensitiveField) error {
	values, err := getSliceValuesForPath(cfg, field.FullPath)
	if err != nil {
		return fmt.Errorf("failed to derive slice entries for %s: %w", field.FullPath, err)
	}

	for idx, val := range values {
		baseKey := fmt.Sprintf("%s_%d", field.Key, idx)
		basePath := fmt.Sprintf("%s.%d", field.FullPath, idx)
		appendMapValueEnvVars(envVars, baseKey, basePath, val, isSecret, sensitiveFields)
	}

	return nil
}

// appendMapValueEnvVars recursively appends environment variable entries for a given reflect.Value
func appendMapValueEnvVars(envVars *strings.Builder, baseKey, basePath string, value reflect.Value, parentSecret bool, sensitiveFields *[]SensitiveField) {
	if !value.IsValid() {
		return
	}

	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		appendStructEnvVars(envVars, baseKey, basePath, value, parentSecret, sensitiveFields)
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			keyVal := iter.Key()
			if keyVal.Kind() != reflect.String {
				continue
			}

			subKey := fmt.Sprintf("%s_%s", baseKey, sanitizeMapKeyForEnv(keyVal.String()))
			subPath := fmt.Sprintf("%s.%s", basePath, keyVal.String())
			appendMapValueEnvVars(envVars, subKey, subPath, iter.Value(), parentSecret, sensitiveFields)
		}
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Struct {
			for i := 0; i < value.Len(); i++ {
				sub := value.Index(i)
				subKey := fmt.Sprintf("%s_%d", baseKey, i)
				subPath := fmt.Sprintf("%s.%d", basePath, i)
				appendMapValueEnvVars(envVars, subKey, subPath, sub, parentSecret, sensitiveFields)
			}
		} else {
			envVars.WriteString(fmt.Sprintf("%s=\"%s\"\n", baseKey, sliceToEnvString(value)))
			if parentSecret {
				appendSensitiveField(sensitiveFields, baseKey, basePath)
			}
		}
	default:
		envVars.WriteString(fmt.Sprintf("%s=\"%v\"\n", baseKey, value.Interface()))
		if parentSecret {
			appendSensitiveField(sensitiveFields, baseKey, basePath)
		}
	}
}

// appendStructEnvVars processes a struct value and appends its fields to the environment variables
func appendStructEnvVars(envVars *strings.Builder, baseKey, basePath string, val reflect.Value, parentSecret bool, sensitiveFields *[]SensitiveField) {
	typeInfo := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typeInfo.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get(tagName)
		if tag == "" || tag == skipper {
			continue
		}

		subValue := val.Field(i)
		subKey := fmt.Sprintf("%s_%s", baseKey, sanitizeMapKeyForEnv(tag))
		subPath := fmt.Sprintf("%s.%s", basePath, tag)
		fieldSecret := parentSecret || field.Tag.Get(sensitiveTag) == "true" || isExternalSensitiveField(subPath)

		appendMapValueEnvVars(envVars, subKey, subPath, subValue, fieldSecret, sensitiveFields)
	}
}

// appendSensitiveField adds a sensitive field to the list of sensitive fields
func appendSensitiveField(target *[]SensitiveField, key, path string) {
	*target = append(*target, SensitiveField{
		Key:        key,
		Path:       path,
		SecretName: generateSecretName(path),
	})
}

// sliceToEnvString converts a slice reflect.Value to a comma-separated string for environment variables
func sliceToEnvString(v reflect.Value) string {
	if v.Len() == 0 {
		return ""
	}

	parts := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		parts[i] = fmt.Sprintf("%v", v.Index(i).Interface())
	}

	return strings.Join(parts, ",")
}

// buildHelmValueReference constructs a Helm values reference string from a full config path
func buildHelmValueReference(fullPath string) string {
	trimmed := strings.TrimPrefix(fullPath, "core.")
	parts := strings.Split(trimmed, ".")
	expr := ".Values.openlane.coreConfiguration"
	for _, part := range parts {
		if part == "" {
			continue
		}
		if idx, err := strconv.Atoi(part); err == nil {
			expr = fmt.Sprintf("(index %s %d)", expr, idx)
		} else {
			expr = fmt.Sprintf("%s.%s", expr, part)
		}
	}
	return expr
}

func structValueToMap(val reflect.Value) map[string]any {
	val = reflect.Indirect(val)
	result := make(map[string]any)
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			name = strings.ToLower(field.Name)
		}

		fieldValue := val.Field(i)
		result[name] = convertValueForYAML(fieldValue)
	}
	return result
}

// convertValueForYAML converts a reflect.Value to a juicy YAML-compatible representation
func convertValueForYAML(val reflect.Value) any {
	if !val.IsValid() {
		return nil
	}

	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	typ := val.Type()
	if typ.PkgPath() == "time" && typ.Name() == "Duration" {
		return time.Duration(val.Int()).String()
	}

	switch val.Kind() {
	case reflect.Struct:
		return structValueToMap(val)
	case reflect.Slice, reflect.Array:
		length := val.Len()
		items := make([]any, 0, length)
		for i := 0; i < length; i++ {
			items = append(items, convertValueForYAML(val.Index(i)))
		}
		return items
	case reflect.Map:
		iter := val.MapRange()
		out := make(map[string]any)
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				continue
			}
			out[key.String()] = convertValueForYAML(iter.Value())
		}
		return out
	default:
		return val.Interface()
	}
}

// processEnvironmentVariables extracts and processes all environment variables from the config
func processEnvironmentVariables() (string, []SensitiveField, error) {
	defaultConfig := buildDefaultConfig()

	cp := envparse.Config{
		FieldTagName: tagName,
		Skipper:      skipper,
	}

	out, err := cp.GatherEnvInfo(varPrefix, defaultConfig)
	if err != nil {
		return "", nil, fmt.Errorf("failed to gather environment info: %w", err)
	}

	var envVars strings.Builder
	var sensitiveFields []SensitiveField

	for _, field := range out {
		if strings.EqualFold(field.FieldName, "domain") {
			// skip the domain field, this is used for inheritance, not alone
			continue
		}

		defaultVal := field.Tags.Get(defaultTag)
		isSecret := field.Tags.Get(sensitiveTag) == "true" || isExternalSensitiveField(field.FullPath)

		isComplexCollection := field.Type.Kind() == reflect.Map || (field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct)

		switch {
		case !isComplexCollection:
			// Always add to environment variables
			envVars.WriteString(fmt.Sprintf("%s=\"%s\"\n", field.Key, defaultVal))

			if isSecret {
				appendSensitiveField(&sensitiveFields, field.Key, field.FullPath)
			}
		case isSecret:
			appendSensitiveField(&sensitiveFields, field.Key, field.FullPath)
		}

		if field.Type.Kind() == reflect.Map {
			if err := appendMapFieldEnvVars(&envVars, defaultConfig, field, isSecret, &sensitiveFields); err != nil {
				return "", nil, err
			}
		}

		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			if err := appendSliceStructFieldEnvVars(&envVars, defaultConfig, field, isSecret, &sensitiveFields); err != nil {
				return "", nil, err
			}
		}
	}

	return envVars.String(), sensitiveFields, nil
}

// generateEnvironmentFile writes the environment variables to a file
func generateEnvironmentFile(envConfigPath, envVars string) error {
	if err := os.WriteFile(envConfigPath, []byte(envVars), ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}

	return nil
}

// generateConfigFileConfigMap creates a ConfigMap that embeds the rendered config.yaml file.
func generateConfigFileConfigMap(configMapPath string) error {
	header, err := os.ReadFile("./jsonschema/templates/configmap-configfile.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read config file ConfigMap template: %w", err)
	}

	cfg := buildDefaultConfig()

	var body strings.Builder
	body.WriteString("\n")
	if err := writeConfigYAML(&body, reflect.ValueOf(cfg).Elem(), nil, 4); err != nil {
		return fmt.Errorf("failed to render config.yaml content: %w", err)
	}

	configMapData := body.String() + "{{- end }}\n"
	content := append(header, []byte(configMapData)...)

	if err := os.WriteFile(configMapPath, content, ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write config file ConfigMap: %w", err)
	}

	return nil
}

// writeConfigYAML walks the config struct and emits Helm-templated YAML content.
func writeConfigYAML(builder *strings.Builder, val reflect.Value, path []string, indent int) error {
	val = reflect.Indirect(val)
	if !val.IsValid() || val.Kind() != reflect.Struct {
		return nil
	}

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		key := field.Tag.Get(tagName)
		if key == "" || key == skipper {
			continue
		}

		fullPath := strings.Join(append(path, key), ".")
		prefixedPath := fmt.Sprintf("core.%s", fullPath)
		if isFieldSensitive(field, prefixedPath) {
			continue
		}

		fieldValue := unwrapValue(val.Field(i))
		helmRef := buildHelmValueReference(prefixedPath)

		if field.Tag.Get("domain") == "inherit" {
			builder.WriteString(generateDomainConfigYAML(key, helmRef, field.Tag.Get("domainPrefix"), field.Tag.Get("domainSuffix"), fieldValue, indent))
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			if err := writeStructField(builder, key, helmRef, fieldValue, appendPath(path, key), indent); err != nil {
				return err
			}
		case reflect.Map:
			writeMapField(builder, key, helmRef, indent)
		case reflect.Slice, reflect.Array:
			writeSliceField(builder, key, helmRef, indent)
		default:
			builder.WriteString(formatScalarFieldLine(key, helmRef, fieldValue, indent))
		}
	}

	return nil
}

// writeStructField renders a nested struct block guarded by Helm conditionals.
func writeStructField(builder *strings.Builder, key, helmRef string, value reflect.Value, path []string, indent int) error {
	var inner strings.Builder
	if err := writeConfigYAML(&inner, value, path, indent+2); err != nil {
		return err
	}

	indentStr := strings.Repeat(" ", indent)
	if inner.Len() == 0 {
		return nil
	}

	builder.WriteString(fmt.Sprintf("%s{{- if %s }}\n", indentStr, helmRef))
	builder.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
	builder.WriteString(inner.String())
	builder.WriteString(fmt.Sprintf("%s{{- end }}\n", indentStr))
	return nil
}

// writeSliceField renders slice values using Helm helpers while preserving YAML formatting.
func writeSliceField(builder *strings.Builder, key, helmRef string, indent int) {
	indentStr := strings.Repeat(" ", indent)
	builder.WriteString(fmt.Sprintf("%s{{- $sliceValue := (%s | default (list)) }}\n", indentStr, helmRef))
	builder.WriteString(fmt.Sprintf("%s{{- if gt (len $sliceValue) 0 }}\n", indentStr))
	builder.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
	builder.WriteString(fmt.Sprintf("%s{{- toYaml $sliceValue | nindent %d }}\n", indentStr, indent+2))
	builder.WriteString(fmt.Sprintf("%s{{- end }}\n", indentStr))
}

// writeMapField renders map values using Helm helpers while preserving YAML formatting.
func writeMapField(builder *strings.Builder, key, helmRef string, indent int) {
	indentStr := strings.Repeat(" ", indent)
	builder.WriteString(fmt.Sprintf("%s{{- $mapValue := (%s | default (dict)) }}\n", indentStr, helmRef))
	builder.WriteString(fmt.Sprintf("%s{{- if gt (len $mapValue) 0 }}\n", indentStr))
	builder.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
	builder.WriteString(fmt.Sprintf("%s{{- toYaml $mapValue | nindent %d }}\n", indentStr, indent+2))
	builder.WriteString(fmt.Sprintf("%s{{- end }}\n", indentStr))
}

// formatScalarFieldLine renders a single scalar field with Helm defaults.
func formatScalarFieldLine(key, helmRef string, value reflect.Value, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s{{- if %s }}\n", indentStr, helmRef))
	if needsQuotes(value) {
		builder.WriteString(fmt.Sprintf("%s%s: {{ %s | quote }}\n", indentStr, key, helmRef))
	} else {
		builder.WriteString(fmt.Sprintf("%s%s: {{ %s }}\n", indentStr, key, helmRef))
	}
	builder.WriteString(fmt.Sprintf("%s{{- end }}\n", indentStr))
	return builder.String()
}

// formatScalarDefaultLiteral converts a scalar value into a Helm-friendly literal.
func formatScalarDefaultLiteral(value reflect.Value) string {
	if !value.IsValid() {
		return "\"\""
	}

	switch value.Kind() {
	case reflect.String:
		return strconv.Quote(value.String())
	case reflect.Bool:
		if value.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if isDurationType(value.Type()) {
			return strconv.Quote(time.Duration(value.Int()).String())
		}
		return fmt.Sprintf("%d", value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", value.Uint())
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'f', -1, 64)
	default:
		return strconv.Quote(fmt.Sprintf("%v", value.Interface()))
	}
}

// needsQuotes determines if a scalar field should be quoted in YAML output.
func needsQuotes(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	if isDurationType(value.Type()) {
		return true
	}

	return value.Kind() == reflect.String
}

// isDurationType checks if the provided type is a time.Duration.
func isDurationType(t reflect.Type) bool {
	return t == reflect.TypeOf(time.Duration(0))
}

// unwrapValue dereferences pointer values and returns the underlying value.
func unwrapValue(v reflect.Value) reflect.Value {
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Zero(v.Type().Elem())
		}
		v = v.Elem()
	}
	return v
}

// appendPath returns a copy of the path with the additional key appended.
func appendPath(path []string, key string) []string {
	newPath := make([]string, len(path)+1)
	copy(newPath, path)
	newPath[len(path)] = key
	return newPath
}

// isFieldSensitive determines whether a field should be omitted due to sensitivity.
func isFieldSensitive(field reflect.StructField, fullPath string) bool {
	return field.Tag.Get(sensitiveTag) == "true" || isExternalSensitiveField(fullPath)
}

// generateDomainConfigYAML builds conditional YAML for domain-inherited fields.
func generateDomainConfigYAML(key, helmRef, domainPrefix, domainSuffix string, value reflect.Value, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	var template strings.Builder

	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		template.WriteString(fmt.Sprintf("%s{{- if %s }}\n", indentStr, helmRef))
		template.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
		template.WriteString(fmt.Sprintf("%s{{- toYaml %s | nindent %d }}\n", indentStr, helmRef, indent+2))
		template.WriteString(fmt.Sprintf("%s{{- else if .Values.domain }}\n", indentStr))
		template.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
		template.WriteString(buildDomainSliceValues(domainPrefix, domainSuffix, indent+2))
		template.WriteString(fmt.Sprintf("%s{{- else }}\n", indentStr))
		template.WriteString(fmt.Sprintf("%s%s:\n", indentStr, key))
		template.WriteString(renderStaticYAMLValue(value.Interface(), indent+2, "[]"))
		template.WriteString(fmt.Sprintf("%s{{- end }}\n", indentStr))
		return template.String()
	}

	template.WriteString(fmt.Sprintf("%s{{- if %s }}\n", indentStr, helmRef))
	template.WriteString(fmt.Sprintf("%s%s: {{ %s | quote }}\n", indentStr, key, helmRef))
	template.WriteString(fmt.Sprintf("%s{{- else if .Values.domain }}\n", indentStr))
	template.WriteString(fmt.Sprintf("%s%s: %s\n", indentStr, key, buildDomainStringValue(domainPrefix, domainSuffix)))
	template.WriteString(fmt.Sprintf("%s{{- else }}\n", indentStr))
	template.WriteString(fmt.Sprintf("%s%s: %s\n", indentStr, key, formatScalarDefaultLiteral(value)))
	template.WriteString(fmt.Sprintf("%s{{- end }}\n", indentStr))
	return template.String()
}

// buildDomainSliceValues renders domain-derived slice entries.
func buildDomainSliceValues(domainPrefix, domainSuffix string, indent int) string {
	indentStr := strings.Repeat(" ", indent)
	prefixes := normalizeDomainPrefixes(domainPrefix)
	if len(prefixes) == 0 {
		prefixes = []string{""}
	}

	var builder strings.Builder
	for _, prefix := range prefixes {
		builder.WriteString(indentStr)
		builder.WriteString("- ")
		builder.WriteString(fmt.Sprintf("\"%s\"\n", composeDomainValue(prefix, domainSuffix)))
	}

	return builder.String()
}

// buildDomainStringValue renders the string value for domain overrides.
func buildDomainStringValue(domainPrefix, domainSuffix string) string {
	prefixes := normalizeDomainPrefixes(domainPrefix)
	prefix := ""
	if len(prefixes) > 0 {
		prefix = prefixes[0]
	}

	return fmt.Sprintf("\"%s\"", composeDomainValue(prefix, domainSuffix))
}

// normalizeDomainPrefixes splits and trims prefix definitions.
func normalizeDomainPrefixes(prefix string) []string {
	if prefix == "" {
		return nil
	}

	parts := strings.Split(prefix, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// composeDomainValue composes the domain-based value using prefix/suffix hints.
func composeDomainValue(prefix, suffix string) string {
	var builder strings.Builder
	if prefix != "" {
		builder.WriteString(formatDomainPrefix(prefix))
	} else {
		builder.WriteString("{{ .Values.domain }}")
	}

	if suffix != "" {
		builder.WriteString(suffix)
	}

	return builder.String()
}

// formatDomainPrefix formats the domain prefix into a Helm template expression.
func formatDomainPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}

	if strings.HasSuffix(prefix, "@") {
		return prefix + "{{ .Values.domain }}"
	}

	return prefix + ".{{ .Values.domain }}"
}

// renderStaticYAMLValue renders a static YAML representation for default data.
func renderStaticYAMLValue(value any, indent int, emptyLiteral string) string {
	data, err := yaml.Marshal(value)
	if err != nil {
		return strings.Repeat(" ", indent) + emptyLiteral + "\n"
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return strings.Repeat(" ", indent) + emptyLiteral + "\n"
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	indentStr := strings.Repeat(" ", indent)
	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(indentStr)
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return builder.String()
}

// generateAndWriteHelmValues generates Helm values and writes them to a file
func generateAndWriteHelmValues(helmValuesPath string, structure interface{}) error {
	helmValues, err := generateHelmValues(structure)
	if err != nil {
		return fmt.Errorf("failed to generate Helm values: %w", err)
	}

	if err := os.WriteFile(helmValuesPath, []byte(helmValues), ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write Helm values file: %w", err)
	}

	return nil
}

func namePkg(r reflect.Type) string {
	return r.String()
}

// generateHelmValues creates a Helm-compatible values.yaml from the config structure
func generateHelmValues(structure interface{}) (string, error) {
	// Create a reflector to extract comments and schema information
	r := new(jsonschema.Reflector)
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true

	// Add go comments to the reflector
	for _, pkg := range includedPackages {
		if err := r.AddGoComments("github.com/theopenlane/core/", pkg); err != nil {
			return "", err
		}
	}

	// Get the defaults for the structure
	defaults.SetDefaults(structure)
	if cfg, ok := structure.(*config.Config); ok {
		initializeStripeWebhookSecrets(cfg)
		initializeRateLimitOptions(cfg)
	}

	// Generate schema to extract field information
	schema := r.Reflect(structure)

	// Generate values with comments
	var regularResult strings.Builder

	// Add helm values header comment for regular values
	regularResult.WriteString(`# Helm values.yaml for Openlane
# This file is auto-generated from the core config structure
# Manual changes may be overwritten when regenerated
#
# Domain Inheritance:
# Set 'domain' to enable automatic domain inheritance for fields tagged with domain:"inherit"
# Fields with domainPrefix will be prefixed (e.g., "https://api" becomes "https://api.yourdomain.com")
# Fields with domainSuffix will be suffixed (e.g., "/.well-known/jwks.json" becomes "yourdomain.com/.well-known/jwks.json")
# Individual fields can still be overridden by setting them explicitly

`)

	// Generate YAML with comments recursively, excluding sensitive values
	// Nest under coreConfiguration key to be merged into openlane.coreConfiguration
	regularResult.WriteString("coreConfiguration:\n")
	err := generateYAMLWithComments(&regularResult, "", reflect.ValueOf(structure).Elem(), schema, 1)
	if err != nil {
		return "", err
	}

	// Add external secrets configuration to regular values
	sensitiveFields := findSensitiveFields(reflect.ValueOf(structure).Elem(), "")
	if len(sensitiveFields) > 0 {
		regularResult.WriteString("\n# -- External Secrets configuration\n")
		regularResult.WriteString("externalSecrets:\n")
		regularResult.WriteString("  # -- Enable external secrets integration\n")
		regularResult.WriteString("  enabled: true  # @schema type:boolean; default:true\n")
		regularResult.WriteString("  # -- List of external secrets to create\n")
		regularResult.WriteString("  secrets:\n")

		for _, field := range sensitiveFields {
			regularResult.WriteString(fmt.Sprintf("    # -- %s secret configuration\n", field.SecretName))
			regularResult.WriteString(fmt.Sprintf("    %s:\n", field.SecretName))
			regularResult.WriteString("      # -- Enable this external secret\n")
			regularResult.WriteString("      enabled: true  # @schema type:boolean; default:true\n")
			regularResult.WriteString(fmt.Sprintf("      # -- Environment variable key for %s\n", field.Path))
			regularResult.WriteString(fmt.Sprintf("      secretKey: \"%s\"  # @schema type:string\n", field.Key))
			regularResult.WriteString("      # -- Remote key in GCP Secret Manager\n")
			regularResult.WriteString(fmt.Sprintf("      remoteKey: \"%s\"  # @schema type:string\n", field.SecretName))
		}
	}

	return regularResult.String(), nil
}

// WriteFieldDescription writes multi-line field descriptions as comments
func WriteFieldDescription(result *strings.Builder, description, indentStr string) {
	if description == "" {
		return
	}
	lines := strings.Split(description, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if i == 0 {
				fmt.Fprintf(result, "%s# -- %s\n", indentStr, line)
			} else {
				fmt.Fprintf(result, "%s# %s\n", indentStr, line)
			}
		}
	}
}

// handleSliceField processes slice fields for YAML generation
func handleSliceField(result *strings.Builder, fieldName, description, indentStr string, fieldValue reflect.Value) {
	WriteFieldDescription(result, description, indentStr)

	fmt.Fprintf(result, "%s%s:", indentStr, fieldName)
	if fieldValue.IsNil() || fieldValue.Len() == 0 {
		result.WriteString(" []\n")
	} else {
		result.WriteString("\n")

		if fieldValue.Type().Elem().Kind() == reflect.Struct {
			items := make([]any, 0, fieldValue.Len())

			for j := 0; j < fieldValue.Len(); j++ {
				items = append(items, structValueToMap(fieldValue.Index(j)))
			}

			if data, err := yaml.Marshal(items); err == nil {
				lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")

				for _, line := range lines {
					if strings.TrimSpace(line) == "" {
						continue
					}

					result.WriteString(indentStr)
					result.WriteString("  ")
					result.WriteString(line)
					result.WriteString("\n")
				}
				return
			}
		}

		for j := 0; j < fieldValue.Len(); j++ {
			elem := fieldValue.Index(j)
			fmt.Fprintf(result, "%s    - %v\n", indentStr, elem.Interface())
		}
	}
}

// handleMapField processes map fields for YAML generation
func handleMapField(result *strings.Builder, fieldName, description, indentStr string) {
	WriteFieldDescription(result, description, indentStr)
	fmt.Fprintf(result, "%s%s: {}\n", indentStr, fieldName)
}

// handlePrimitiveField processes primitive fields for YAML generation
func handlePrimitiveField(result *strings.Builder, fieldName, description, indentStr, defaultTag string, field reflect.StructField, fieldValue reflect.Value) {
	WriteFieldDescription(result, description, indentStr)

	typeInfo := getTypeSchemaInfo(field, defaultTag)
	value := formatValue(fieldValue.Interface())
	fmt.Fprintf(result, "%s%s: %s%s\n", indentStr, fieldName, value, typeInfo)
}

// generateYAMLWithComments recursively generates YAML with comments from struct fields (legacy function for compatibility)
func generateYAMLWithComments(result *strings.Builder, prefix string, v reflect.Value, schema *jsonschema.Schema, indent int) error {
	if !v.IsValid() {
		return nil
	}

	t := v.Type()
	indentStr := strings.Repeat("  ", indent)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		// Build full path for sensitive field checking
		var fullPath string
		if prefix != "" {
			fullPath = prefix + fieldName
		} else {
			fullPath = fieldName
		}

		// Skip sensitive fields - they are handled by External Secrets
		if field.Tag.Get("sensitive") == "true" || isExternalSensitiveField(fullPath) {
			continue
		}

		// Get field description from schema
		var description string
		if schema != nil && schema.Properties != nil {
			if propSchema, exists := schema.Properties.Get(fieldName); exists {
				description = propSchema.Description
			}
		}

		// Get default value from struct tag
		defaultTag := field.Tag.Get("default")

		// Generate comment
		WriteFieldDescription(result, description, indentStr)

		// Handle different field types
		switch fieldValue.Kind() {
		case reflect.Struct:
			// Write field name for struct
			fmt.Fprintf(result, "%s%s:\n", indentStr, fieldName)

			// Get nested schema
			var nestedSchema *jsonschema.Schema
			if schema != nil && schema.Properties != nil {
				if propSchema, exists := schema.Properties.Get(fieldName); exists {
					nestedSchema = propSchema
				}
			}

			// Recurse into struct
			err := generateYAMLWithComments(result, fullPath+".", fieldValue, nestedSchema, indent+1)
			if err != nil {
				return err
			}

		case reflect.Slice:
			handleSliceField(result, fieldName, "", indentStr, fieldValue)

		case reflect.Map:
			handleMapField(result, fieldName, "", indentStr)

		default:
			handlePrimitiveField(result, fieldName, "", indentStr, defaultTag, field, fieldValue)
		}
	}

	return nil
}

// getTypeSchemaInfo generates schema annotation for field types
func getTypeSchemaInfo(field reflect.StructField, defaultTag string) string {
	var parts []string

	// Add type information
	fieldType := field.Type
	switch fieldType.Kind() {
	case reflect.String:
		parts = append(parts, "type:string")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parts = append(parts, "type:integer")
	case reflect.Bool:
		parts = append(parts, "type:boolean")
	case reflect.Float32, reflect.Float64:
		parts = append(parts, "type:number")
	}

	// Add default if specified
	if defaultTag != "" {
		parts = append(parts, fmt.Sprintf("default:%s", defaultTag))
	}

	if len(parts) > 0 {
		return fmt.Sprintf("  # @schema %s", strings.Join(parts, "; "))
	}

	return ""
}

// formatValue formats a value for YAML output (compatible with Helm templating)
func formatValue(v any) string {
	// Check if v is a nil pointer
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
		return "\"\""
	}

	switch val := v.(type) {
	case string:
		// Always quote strings for Helm/YAML compatibility
		return fmt.Sprintf(`"%s"`, val)
	case time.Duration:
		// Quote durations so Helm parses them as literal strings
		return fmt.Sprintf(`"%s"`, val.String())
	case bool:
		// Helm expects true/false as unquoted
		return fmt.Sprintf("%t", val)
	default:
		// For numbers and other types, output as string if needed
		formatted := fmt.Sprintf("%v", val)
		if formatted == "<nil>" {
			return "\"\""
		}

		return formatted
	}
}

// formatHelmDefaultLiteral formats a value as a Helm template-friendly default literal.
func formatHelmDefaultLiteral(v any) string {
	if v == nil {
		return "\"\""
	}

	value := reflect.ValueOf(v)
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return "\"\""
		}
		value = value.Elem()
	}

	unwrapped := value.Interface()

	if b, ok := unwrapped.(bool); ok {
		return fmt.Sprintf("%t", b)
	}

	if d, ok := unwrapped.(time.Duration); ok {
		return strconv.Quote(d.String())
	}

	return strconv.Quote(fmt.Sprintf("%v", unwrapped))
}

// hasSecretChildren checks if a struct has any sensitive child fields
func hasSecretChildren(v reflect.Value, prefix string) bool {
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		fullPath := prefix + fieldName

		// Check if this field is sensitive
		if field.Tag.Get(sensitiveTag) == "true" || isExternalSensitiveField(fullPath) {
			return true
		}

		// Recurse into nested structs
		if fieldValue.Kind() == reflect.Struct {
			if hasSecretChildren(fieldValue, fullPath+".") {
				return true
			}
		}
	}
	return false
}

// hasNonSecretChildren checks if a struct has any non-sensitive child fields
//
//nolint:unused // actually used but linter doesn't see it
func hasNonSecretChildren(v reflect.Value, prefix string) bool {
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return false
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		fullPath := prefix + fieldName

		// Check if this field is NOT sensitive
		if field.Tag.Get(sensitiveTag) != "true" && !isExternalSensitiveField(fullPath) {
			// For non-struct fields, this counts as a non-secret child
			if fieldValue.Kind() != reflect.Struct {
				return true
			}
			// For struct fields, recurse to check their children
			if hasNonSecretChildren(fieldValue, fullPath+".") {
				return true
			}
		}
	}
	return false
}

// generateSecretName creates a Kubernetes-friendly secret name from a config path
func generateSecretName(path string) string {
	// Convert path like "auth.providers.github.clientSecret" to "auth-providers-github-client-secret"
	name := strings.ToLower(path)
	name = strings.ReplaceAll(name, ".", "-")
	// Convert camelCase to kebab-case
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// generateSecretFiles ensures the ExternalSecret template exists when sensitive fields are present
func generateSecretFiles(c schemaConfig, fields []SensitiveField) error {
	if len(fields) == 0 {
		return nil
	}

	if err := os.MkdirAll(c.externalSecretsDir, dirPermission); err != nil {
		return err
	}

	return generateExternalSecretsTemplate(c.externalSecretsDir)
}

// generateExternalSecretsTemplate creates a single dynamic ExternalSecret template for Helm chart
func generateExternalSecretsTemplate(dir string) error {
	templateFile := fmt.Sprintf("%s/external-secrets.yaml", dir)

	content := `{{- if and .Values.externalSecrets .Values.externalSecrets.enabled }}
{{- range $secretName, $config := .Values.externalSecrets.secrets | default dict }}
{{- if $config.enabled }}
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: {{ $secretName }}-ext
  namespace: {{ $.Release.Namespace }}
spec:
  secretStoreRef:
    name: gcp-secret-store
    kind: ClusterSecretStore
  target:
    name: {{ $secretName }}
    creationPolicy: Owner
  data:
  - secretKey: {{ $config.secretKey }}
    remoteRef:
      key: {{ $config.remoteKey }}
{{- end }}
{{- end }}
{{- end }}
`
	return os.WriteFile(templateFile, []byte(content), ownerReadWrite)
}

// isExternalSensitiveField checks if a field path corresponds to a sensitive field from external packages
func isExternalSensitiveField(path string) bool {
	_, ok := externalSensitiveFields[path]
	return ok
}

// findSensitiveFields recursively finds all sensitive fields in a struct
func findSensitiveFields(v reflect.Value, prefix string) []SensitiveField {
	var fields []SensitiveField

	if !v.IsValid() {
		return fields
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		if fieldName == "" {
			continue
		}

		// Build current path
		var currentPath string
		if prefix != "" {
			currentPath = fmt.Sprintf("%s.%s", prefix, fieldName)
		} else {
			currentPath = fieldName
		}

		envKeyName := strings.ToUpper(strings.ReplaceAll(currentPath, ".", "_"))
		secretKeyName := strings.ToLower(strings.ReplaceAll(currentPath, ".", "-"))
		baseEnvKey := fmt.Sprintf("CORE_%s", envKeyName)
		baseSecretName := fmt.Sprintf("core-%s", secretKeyName)

		// Check if this field is sensitive
		if sensitiveTag := field.Tag.Get("sensitive"); sensitiveTag == "true" {
			fields = append(fields, SensitiveField{
				Key:        baseEnvKey,
				Path:       currentPath,
				SecretName: baseSecretName,
			})

			if fieldValue.Kind() == reflect.Map && currentPath == "subscription.stripeWebhookSecrets" && fieldValue.Len() > 0 {
				keys := fieldValue.MapKeys()
				keyStrings := make([]string, 0, len(keys))
				for _, k := range keys {
					if k.Kind() == reflect.String {
						keyStrings = append(keyStrings, k.String())
					}
				}

				sort.Strings(keyStrings)

				for _, mapKey := range keyStrings {
					if mapKey == "" {
						continue
					}
					envKey := fmt.Sprintf("%s_%s", baseEnvKey, sanitizeMapKeyForEnv(mapKey))
					mapPath := fmt.Sprintf("%s.%s", currentPath, mapKey)
					fields = append(fields, SensitiveField{
						Key:        envKey,
						Path:       mapPath,
						SecretName: generateSecretName(fmt.Sprintf("core.%s", mapPath)),
					})
				}
			}
		}

		// Recursively check nested structs
		if fieldValue.Kind() == reflect.Struct && fieldValue.Type() != reflect.TypeOf(time.Time{}) {
			nestedFields := findSensitiveFields(fieldValue, currentPath)
			fields = append(fields, nestedFields...)
		} else if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() && fieldValue.Elem().Kind() == reflect.Struct {
			nestedFields := findSensitiveFields(fieldValue.Elem(), currentPath)
			fields = append(fields, nestedFields...)
		}
	}

	return fields
}
