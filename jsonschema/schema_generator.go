package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/invopop/yaml"
	"github.com/mcuadros/go-defaults"

	"github.com/theopenlane/utils/envparse"

	"github.com/theopenlane/core/config"
)

// const values used for the schema generator
const (
	tagName        = "koanf"
	skipper        = "-"
	defaultTag     = "default"
	jsonSchemaPath = "./jsonschema/core.config.json"
	yamlConfigPath = "./config/config.example.yaml"
	envConfigPath  = "./config/.env.example"
	configMapPath  = "./config/configmap.yaml"
	varPrefix      = "CORE"
	ownerReadWrite = 0600
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
}

func main() {
	c := schemaConfig{
		jsonSchemaPath: jsonSchemaPath,
		yamlConfigPath: yamlConfigPath,
		envConfigPath:  envConfigPath,
		configMapPath:  configMapPath,
	}

	if err := generateSchema(c, &config.Config{}); err != nil {
		panic(err)
	}
}

// generateSchema generates a JSON schema and a YAML schema based on the provided schemaConfig and structure
func generateSchema(c schemaConfig, structure interface{}) error {
	// override the default name to using the prefixed pkg name
	r := jsonschema.Reflector{Namer: namePkg}
	r.ExpandedStruct = true
	// set `jsonschema:required` tag to true to generate required fields
	r.RequiredFromJSONSchemaTags = true
	// set the tag name to `koanf` for the koanf struct tags
	r.FieldNameTag = tagName

	// add go comments to the schema
	for _, pkg := range includedPackages {
		if err := r.AddGoComments("github.com/theopenlane/core/", pkg); err != nil {
			panic(err.Error())
		}
	}

	s := r.Reflect(structure)

	// generate the json schema
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		panic(err.Error())
	}

	if err = os.WriteFile(c.jsonSchemaPath, data, ownerReadWrite); err != nil {
		panic(err.Error())
	}

	// generate yaml schema with default
	yamlConfig := &config.Config{}
	defaults.SetDefaults(yamlConfig)

	// this uses the `json` tag to generate the yaml schema
	yamlSchema, err := yaml.Marshal(yamlConfig)
	if err != nil {
		panic(err.Error())
	}

	if err = os.WriteFile(c.yamlConfigPath, yamlSchema, ownerReadWrite); err != nil {
		panic(err.Error())
	}

	cp := envparse.Config{
		FieldTagName: tagName,
		Skipper:      skipper,
	}

	out, err := cp.GatherEnvInfo(varPrefix, &config.Config{})
	if err != nil {
		panic(err.Error())
	}

	// generate the environment variables from the config
	envSchema := ""
	configMapSchema := "\n"

	for _, k := range out {
		defaultVal := k.Tags.Get(defaultTag)

		envSchema += fmt.Sprintf("%s=\"%s\"\n", k.Key, defaultVal)

		// if the default value is empty, use the value from the values.yaml
		if defaultVal == "" {
			configMapSchema += fmt.Sprintf("  %s: {{ .Values.%s }}\n", k.Key, k.FullPath)
		} else {
			switch k.Type.Kind() {
			case reflect.String, reflect.Int64:
				defaultVal = "\"" + defaultVal + "\"" // add quotes to the string
			case reflect.Slice:
				defaultVal = strings.Replace(defaultVal, "[", "", 1)
				defaultVal = strings.Replace(defaultVal, "]", "", 1)
				defaultVal = "\"" + defaultVal + "\"" // add quotes to the string
			}

			configMapSchema += fmt.Sprintf("  %s: {{ .Values.%s | default %s }}\n", k.Key, k.FullPath, defaultVal)
		}
	}

	// write the environment variables to a file
	if err = os.WriteFile(c.envConfigPath, []byte(envSchema), ownerReadWrite); err != nil {
		panic(err.Error())
	}

	// Get the configmap header
	cm, err := os.ReadFile("./jsonschema/templates/configmap.tmpl")
	if err != nil {
		panic(err.Error())
	}

	// append the configmap schema to the header
	cm = append(cm, []byte(configMapSchema)...)

	// write the configmap to a file
	if err = os.WriteFile(c.configMapPath, cm, ownerReadWrite); err != nil {
		panic(err.Error())
	}

	return nil
}

func namePkg(r reflect.Type) string {
	return r.String()
}
