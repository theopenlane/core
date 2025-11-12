//go:build cli

package speccli

import (
	"encoding/json"
	"fmt"

	"github.com/TylerBrock/colorjson"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
)

var colorFormatter = func() *colorjson.Formatter {
	f := colorjson.NewFormatter()
	f.Indent = 2
	return f
}()

// PrintJSON renders the provided value as formatted JSON. Accepts structs,
// maps, slices, []byte, or json.RawMessage.
func PrintJSON(value any) error {
	switch v := value.(type) {
	case nil:
		fmt.Println("null")
		return nil
	case []byte:
		return printJSONBytes(v)
	case json.RawMessage:
		return printJSONBytes([]byte(v))
	default:
		payload, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return printJSONBytes(payload)
	}
}

// printJSONBytes pretty prints raw JSON bytes using the shared formatter.
func printJSONBytes(data []byte) error {
	if len(data) == 0 {
		fmt.Println("{}")
		return nil
	}

	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	formatted, err := colorFormatter.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(cmdpkg.RootCmd.OutOrStdout(), string(formatted))
	return err
}
