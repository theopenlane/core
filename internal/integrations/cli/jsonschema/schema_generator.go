package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	invyaml "github.com/invopop/yaml"
	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v3"

	"github.com/theopenlane/core/internal/integrations/cli/config"
)

// const values used for the schema generator
const (
	tagName        = "koanf"
	jsonSchemaPath = "./config/integrations.config.json"
	yamlConfigPath = "./config/config.example.yaml"
	envConfigPath  = "./config/.env.example"
	ownerReadWrite = 0600

	// commentWrapWidth is the column width for wrapping YAML comments
	commentWrapWidth = 88
	// yamlIndentSpaces is the number of spaces for YAML indentation
	yamlIndentSpaces = 4
)

// includedPackages is a list of packages to include in the schema generation
// that contain Go comments to be added to the schema
// any external packages must use the jsonschema description tags to add comments
var includedPackages = []string{
	"./config",
}

// schemaConfig represents the configuration for the schema generator
type schemaConfig struct {
	// jsonSchemaPath represents the file path of the JSON schema to be generated
	jsonSchemaPath string
	// yamlConfigPath is the file path to the YAML configuration to be generated
	yamlConfigPath string
	// envConfigPath is the file path to the environment variable configuration to be generated
	envConfigPath string
}

func main() {
	c := schemaConfig{
		jsonSchemaPath: jsonSchemaPath,
		yamlConfigPath: yamlConfigPath,
		envConfigPath:  envConfigPath,
	}

	if err := generateSchema(c, &config.Config{}); err != nil {
		panic(err)
	}
}

// generateSchema generates a JSON schema and a YAML schema based on the provided schemaConfig and structure
func generateSchema(c schemaConfig, structure any) error {
	// override the default name to using the prefixed pkg name
	r := jsonschema.Reflector{Namer: namePkg}
	r.ExpandedStruct = true
	// set `jsonschema:required` tag to true to generate required fields
	r.RequiredFromJSONSchemaTags = true
	// set the tag name to `koanf` for the koanf struct tags
	r.FieldNameTag = tagName

	// add go comments to the schema
	for _, pkg := range includedPackages {
		if err := r.AddGoComments("github.com/theopenlane/core/internal/integrations/cli", pkg); err != nil {
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

	var schemaDoc map[string]any
	if err := json.Unmarshal(data, &schemaDoc); err != nil {
		return fmt.Errorf("unmarshal json schema: %w", err)
	}

	// generate yaml schema with default values and inline comments
	yamlConfig := &config.Config{}
	defaults.SetDefaults(yamlConfig)

	rawYAML, err := invyaml.Marshal(yamlConfig)
	if err != nil {
		panic(err.Error())
	}

	var node yaml.Node
	if err := yaml.Unmarshal(rawYAML, &node); err != nil {
		return fmt.Errorf("unmarshal yaml defaults: %w", err)
	}

	root := &node
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		root = node.Content[0]
	}

	descriptions := buildDescriptionMap(schemaDoc)
	annotateNode(root, nil, descriptions)
	root.HeadComment = wrapComment("integrations CLI configuration. Comments describe each key; edit values as needed.", commentWrapWidth)

	var buf bytes.Buffer

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(yamlIndentSpaces)

	if err := encoder.Encode(root); err != nil {
		return fmt.Errorf("encode yaml config: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return fmt.Errorf("close yaml encoder: %w", err)
	}

	if err = os.WriteFile(c.yamlConfigPath, buf.Bytes(), ownerReadWrite); err != nil {
		panic(err.Error())
	}

	return nil
}

func namePkg(r reflect.Type) string {
	return r.String()
}

func buildDescriptionMap(schema map[string]any) map[string]string {
	descriptions := make(map[string]string)
	defs, _ := schema["$defs"].(map[string]any)

	var walk func(node map[string]any, path []string)

	walk = func(node map[string]any, path []string) {
		if node == nil {
			return
		}

		if desc, ok := node["description"].(string); ok && strings.TrimSpace(desc) != "" && len(path) > 0 {
			key := strings.Join(path, ".")
			if _, exists := descriptions[key]; !exists {
				descriptions[key] = strings.TrimSpace(desc)
			}
		}

		if ref, ok := node["$ref"].(string); ok {
			if refNode := resolveRef(ref, defs); refNode != nil {
				walk(refNode, path)
			}
		}

		props, ok := node["properties"].(map[string]any)
		if !ok {
			return
		}

		for name, raw := range props {
			child, ok := raw.(map[string]any)
			if !ok {
				continue
			}

			walk(child, append(path, name))
		}
	}

	walk(schema, nil)

	return descriptions
}

func resolveRef(ref string, defs map[string]any) map[string]any {
	if !strings.HasPrefix(ref, "#/$defs/") {
		return nil
	}

	name := strings.TrimPrefix(ref, "#/$defs/")
	if def, ok := defs[name].(map[string]any); ok {
		return def
	}

	return nil
}

func annotateNode(node *yaml.Node, path []string, descriptions map[string]string) {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			key := keyNode.Value

			nextPath := make([]string, len(path)+1)
			copy(nextPath, path)
			nextPath[len(path)] = key

			if desc, ok := descriptions[strings.Join(nextPath, ".")]; ok && strings.TrimSpace(desc) != "" {
				keyNode.HeadComment = wrapComment(desc, commentWrapWidth)
			}

			annotateNode(valueNode, nextPath, descriptions)
		}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			annotateNode(item, path, descriptions)
		}
	default:
		return
	}
}

func wrapComment(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	if width <= 0 {
		return strings.Join(words, " ")
	}

	lines := []string{}

	line := words[0]
	for _, word := range words[1:] {
		if len(line)+1+len(word) > width {
			lines = append(lines, line)
			line = word

			continue
		}

		line += " " + word
	}

	lines = append(lines, line)

	return strings.Join(lines, "\n")
}
