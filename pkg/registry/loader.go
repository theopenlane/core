package registry

import (
	"bytes"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// memoryLoader is a custom loader that loads the predefined JSON schema in memory
// instead of from a file
type memoryLoader struct {
	schema []byte
}

// Load loads the predefined JSON schema from memory
// it satisfies the jsonschema.Loader interface, which requires a string argument,
// this is unused in this case because it changes the name to the full "filepath" e.g.
// "file:///Users/sarahfunkhouser/go/src/github.com/theopenlane/policytemplates/standards/jsonschema/standards.json"
// which is not what we want here instead we just use the schema bytes from the struct
func (s memoryLoader) Load(_ string) (any, error) {
	r := bytes.NewBuffer(s.schema)

	return jsonschema.UnmarshalJSON(r)
}

// loadMemorySchema loads the predefined JSON schema from memory
// use this helper function over the `Load` method to load the schema unless you know what you're doing
func loadMemorySchema(s []byte) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()

	// always set the default draft to prevent unknown schema errors
	// with newer JSON schema drafts
	compiler.DefaultDraft(jsonschema.Draft2020)

	// use a custom loader to load the predefined JSON schema from go:embed
	compiler.UseLoader(memoryLoader{
		schema: s,
	})

	// compile the schema
	return compiler.Compile("")
}
