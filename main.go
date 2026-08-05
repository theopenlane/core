// package main is the entry point
package main

import (
	_ "time/tzdata"

	"github.com/theopenlane/core/cmd"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	_ "github.com/theopenlane/core/internal/ent/historygenerated/runtime"
)

func main() {
	cmd.Execute()
}
