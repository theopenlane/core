package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/invopop/yaml"
	"github.com/mcuadros/go-defaults"

	"github.com/theopenlane/utils/envparse"

	"github.com/theopenlane/core/config"
)

// const values used for the schema generator
// const values used for the schema generator
const (
	tagName            = "koanf"
	skipper            = "-"
	defaultTag         = "default"
	jsonSchemaPath     = "./jsonschema/core.config.json"
	yamlConfigPath     = "./config/config.example.yaml"
	envConfigPath      = "./config/.env.example"
	configMapPath      = "./config/configmap.yaml"
	pushSecretsDir     = "./config/pushsecrets"      // #nosec G101 - this is a directory path, not a secret
	externalSecretsDir = "./config/external-secrets" // #nosec G101 - this is a directory path, not a secret
	helmValuesPath     = "./config/helm-values.yaml"
	sensitiveTag       = "sensitive"
	varPrefix          = "CORE"
	ownerReadWrite     = 0600
	dirPermission      = 0755
)

// includedPackages is a list of packages to include in the schema generation
// that contain Go comments to be added to the schema
// any external packages must use the jsonschema description tags to add comments
var includedPackages = []string{
	"./config",
	"./internal/ent",
	"./internal/entdb",
	"./internal/httpserve/handlers",
	"./pkg/middleware",
	"./pkg/objects",
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
	// configMapPath is the file path to the kubernetes config map configuration to be generated
	configMapPath string
	// pushSecretsDir is the directory for individual PushSecret files
	pushSecretsDir string
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
		jsonSchemaPath:     jsonSchemaPath,
		yamlConfigPath:     yamlConfigPath,
		envConfigPath:      envConfigPath,
		configMapPath:      configMapPath,
		pushSecretsDir:     pushSecretsDir,
		externalSecretsDir: externalSecretsDir,
		helmValuesPath:     helmValuesPath,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

func main() {
	c := newSchemaConfig()

	if err := generateSchema(c, &config.Config{}); err != nil {
		panic(err)
	}
}

// generateSchema generates all configuration files from the provided structure
func generateSchema(c schemaConfig, structure interface{}) error {
	if err := generateJSONSchema(c.jsonSchemaPath, structure); err != nil {
		return err
	}

	if err := generateYAMLConfig(c.yamlConfigPath); err != nil {
		return err
	}

	envFields, sensitiveFields, err := processEnvironmentVariables()
	if err != nil {
		return err
	}

	if err := generateEnvironmentFile(c.envConfigPath, envFields.EnvVars); err != nil {
		return err
	}

	if err := generateConfigMap(c.configMapPath, envFields.ConfigMap); err != nil {
		return err
	}

	if len(sensitiveFields) > 0 {
		if err := generateSecretFiles(c, sensitiveFields); err != nil {
			return err
		}
	}

	if err := generateAndWriteHelmValues(c.helmValuesPath, structure); err != nil {
		return err
	}

	return nil
}

// generateJSONSchema creates the JSON schema file from the config structure
func generateJSONSchema(jsonSchemaPath string, structure interface{}) error {
	r := jsonschema.Reflector{Namer: namePkg}
	r.ExpandedStruct = true
	r.RequiredFromJSONSchemaTags = true
	r.FieldNameTag = tagName

	// Add go comments to the schema
	for _, pkg := range includedPackages {
		if err := r.AddGoComments("github.com/theopenlane/core/", pkg); err != nil {
			return fmt.Errorf("failed to add go comments for package %s: %w", pkg, err)
		}
	}

	s := r.Reflect(structure)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON schema: %w", err)
	}

	if err := os.WriteFile(jsonSchemaPath, data, ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write JSON schema file: %w", err)
	}

	return nil
}

// generateYAMLConfig creates the YAML configuration file with defaults
func generateYAMLConfig(yamlConfigPath string) error {
	yamlConfig := &config.Config{}
	defaults.SetDefaults(yamlConfig)

	yamlSchema, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML config: %w", err)
	}

	if err := os.WriteFile(yamlConfigPath, yamlSchema, ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write YAML config file: %w", err)
	}

	return nil
}

// EnvironmentFields represents the processed environment variable data
type EnvironmentFields struct {
	EnvVars   string
	ConfigMap string
}

// processEnvironmentVariables extracts and processes all environment variables from the config
func processEnvironmentVariables() (*EnvironmentFields, []SensitiveField, error) {
	cp := envparse.Config{
		FieldTagName: tagName,
		Skipper:      skipper,
	}

	out, err := cp.GatherEnvInfo(varPrefix, &config.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to gather environment info: %w", err)
	}

	envVars := ""
	configMapData := "\n"
	var sensitiveFields []SensitiveField

	for _, field := range out {
		defaultVal := field.Tags.Get(defaultTag)
		isSecret := field.Tags.Get(sensitiveTag) == "true" || isExternalSensitiveField(field.FullPath)

		if !isSecret {
			// Add to environment variables
			envVars += fmt.Sprintf("%s=\"%s\"\n", field.Key, defaultVal)

			// Add to ConfigMap
			configMapEntry := generateConfigMapEntry(field, defaultVal)
			configMapData += configMapEntry
		} else {
			// Track sensitive fields for secret generation
			secretName := generateSecretName(field.FullPath)
			sensitiveFields = append(sensitiveFields, SensitiveField{
				Key:        field.Key,
				Path:       field.FullPath,
				SecretName: secretName,
			})
		}
	}

	return &EnvironmentFields{
		EnvVars:   envVars,
		ConfigMap: configMapData,
	}, sensitiveFields, nil
}

// generateConfigMapEntry creates a ConfigMap entry for a single field
func generateConfigMapEntry(field envparse.VarInfo, defaultVal string) string {
	// Check if this field has domain inheritance
	domainTag := field.Tags.Get("domain")
	domainPrefix := field.Tags.Get("domainPrefix")
	domainSuffix := field.Tags.Get("domainSuffix")

	if domainTag == "inherit" {
		// Generate Helm template logic for domain inheritance
		return generateDomainHelmTemplate(field.Key, field.FullPath, domainPrefix, domainSuffix, defaultVal)
	}

	// Standard non-domain field processing
	if defaultVal == "" {
		return fmt.Sprintf("  %s: {{ .Values.%s }}\n", field.Key, field.FullPath)
	}

	// Format default value based on type
	formattedDefault := formatDefaultValue(defaultVal, field.Type.Kind())
	return fmt.Sprintf("  %s: {{ .Values.%s | default %s }}\n", field.Key, field.FullPath, formattedDefault)
}

// formatDefaultValue formats a default value based on its type
func formatDefaultValue(defaultVal string, kind reflect.Kind) string {
	switch kind {
	case reflect.String, reflect.Int64:
		return "\"" + defaultVal + "\""
	case reflect.Slice:
		// Remove brackets and add quotes
		formatted := strings.Replace(defaultVal, "[", "", 1)
		formatted = strings.Replace(formatted, "]", "", 1)
		return "\"" + formatted + "\""
	default:
		return "\"" + defaultVal + "\""
	}
}

// generateEnvironmentFile writes the environment variables to a file
func generateEnvironmentFile(envConfigPath, envVars string) error {
	if err := os.WriteFile(envConfigPath, []byte(envVars), ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}
	return nil
}

// generateConfigMap creates the ConfigMap file with header template and data
func generateConfigMap(configMapPath, configMapData string) error {
	header, err := os.ReadFile("./jsonschema/templates/configmap.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read ConfigMap template: %w", err)
	}

	header = append(header, []byte(configMapData)...)
	configMapContent := header

	if err := os.WriteFile(configMapPath, configMapContent, ownerReadWrite); err != nil {
		return fmt.Errorf("failed to write ConfigMap file: %w", err)
	}

	return nil
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
	err := generateYAMLWithComments(&regularResult, "", reflect.ValueOf(structure).Elem(), schema, 0)
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
	indentStr := strings.Repeat("    ", indent)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag or use field name
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

// formatValue formats a value for YAML output
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if val == "" {
			return `""`
		}
		return fmt.Sprintf(`"%s"`, val)
	case bool:
		return fmt.Sprintf("%t", val)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// hasSecretChildren checks if a struct has any sensitive child fields
//
//nolint:unused // actually used but linter doesn't see it
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

// generateSecretFiles creates individual PushSecret and ExternalSecret files
func generateSecretFiles(c schemaConfig, fields []SensitiveField) error {
	// Create directories
	if err := os.MkdirAll(c.pushSecretsDir, dirPermission); err != nil {
		return err
	}
	if err := os.MkdirAll(c.externalSecretsDir, dirPermission); err != nil {
		return err
	}

	for _, field := range fields {
		// Generate PushSecret file
		if err := generatePushSecretFile(c.pushSecretsDir, field); err != nil {
			return err
		}
	}

	// Generate single dynamic ExternalSecret template (only if we have sensitive fields)
	if len(fields) > 0 {
		if err := generateExternalSecretsTemplate(c.externalSecretsDir); err != nil {
			return err
		}
	}

	return nil
}

// generatePushSecretFile creates an individual PushSecret file
func generatePushSecretFile(dir string, field SensitiveField) error {
	content := fmt.Sprintf(`---
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: openlane-secret-push
type: Opaque
data:
  # Base64 encode your secret value and paste it here
  # Example: echo -n "your-secret-value" | base64
  %s: ""
---
apiVersion: external-secrets.io/v1alpha1
kind: PushSecret
metadata:
  name: %s-push
  namespace: openlane-secret-push
spec:
  refreshInterval: 1h
  secretStoreRefs:
    - name: gcp-secretstore
      kind: ClusterSecretStore
  updatePolicy: Replace
  deletionPolicy: Delete
  selector:
    secret:
      name: %s  # References the Secret above
  data:
    - match:
        secretKey: %s  # The key from the Secret above
        remoteRef:
          remoteKey: %s  # The destination secret name in GCP Secret Manager
          property: %s  # The key within the destination secret object (matches secretKey)
`, field.SecretName, field.Key, field.SecretName, field.SecretName, field.Key, field.SecretName, field.Key)

	filename := fmt.Sprintf("%s.yaml", field.SecretName)
	filepath := fmt.Sprintf("%s/%s", dir, filename)
	return os.WriteFile(filepath, []byte(content), ownerReadWrite)
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

// generateDomainHelmTemplate creates Helm template logic for domain inheritance fields
func generateDomainHelmTemplate(envKey, fieldPath, domainPrefix, domainSuffix, defaultVal string) string {
	var template strings.Builder

	// Generate block-level conditional template for proper Helm formatting
	template.WriteString("{{- if .Values.")
	template.WriteString(fieldPath)
	template.WriteString(" }}\n")
	template.WriteString("  ")
	template.WriteString(envKey)
	template.WriteString(": {{ .Values.")
	template.WriteString(fieldPath)
	template.WriteString(" }}\n")
	template.WriteString("{{- else if .Values.domain }}\n")
	template.WriteString("  ")
	template.WriteString(envKey)
	template.WriteString(": ")

	// Handle domain transformation logic
	switch {
	case domainPrefix != "" && domainSuffix != "":
		template.WriteString("\"")
		template.WriteString(domainPrefix)
		template.WriteString(".{{ .Values.domain }}")
		template.WriteString(domainSuffix)
		template.WriteString("\"")
	case domainPrefix != "":
		// Handle multiple prefixes for slice fields
		if strings.Contains(domainPrefix, ",") {
			prefixes := strings.Split(domainPrefix, ",")
			template.WriteString("\"")
			for i, prefix := range prefixes {
				if i > 0 {
					template.WriteString(",")
				}
				template.WriteString(strings.TrimSpace(prefix))
				template.WriteString(".{{ .Values.domain }}")
			}
			template.WriteString("\"")
		} else {
			template.WriteString("\"")
			template.WriteString(domainPrefix)
			template.WriteString(".{{ .Values.domain }}")
			template.WriteString("\"")
		}
	case domainSuffix != "":
		template.WriteString("\"{{ .Values.domain }}")
		template.WriteString(domainSuffix)
		template.WriteString("\"")
	default:
		template.WriteString("\"{{ .Values.domain }}\"")
	}

	template.WriteString("\n{{- else }}\n")
	template.WriteString("  ")
	template.WriteString(envKey)
	template.WriteString(": ")

	// Fallback to default value
	if defaultVal != "" {
		template.WriteString("\"")
		template.WriteString(defaultVal)
		template.WriteString("\"")
	} else {
		template.WriteString("\"\"")
	}

	template.WriteString("\n{{- end }}\n")

	return template.String()
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

		// Get JSON tag for field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		// Build current path
		var currentPath string
		if prefix != "" {
			currentPath = fmt.Sprintf("%s.%s", prefix, fieldName)
		} else {
			currentPath = fieldName
		}

		// Check if this field is sensitive
		if sensitiveTag := field.Tag.Get("sensitive"); sensitiveTag == "true" {
			envKey := strings.ToUpper(strings.ReplaceAll(currentPath, ".", "_"))
			secretName := strings.ToLower(strings.ReplaceAll(currentPath, ".", "-"))

			fields = append(fields, SensitiveField{
				Key:        fmt.Sprintf("CORE_%s", envKey),
				Path:       currentPath,
				SecretName: fmt.Sprintf("core-%s", secretName),
			})
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
