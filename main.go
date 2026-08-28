// package main is the entry point
package main

import (
	_ "time/tzdata"

	"github.com/theopenlane/core/v2/cmd"
	_ "github.com/theopenlane/core/v2/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/v2/internal/ent/historygenerated/runtime"
)

func main() {
	cmd.Execute()
}
