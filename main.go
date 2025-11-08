// package main is the entry point
package main

import ( 
	"github.com/theopenlane/core/cmd"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
)

func main() {
	cmd.Execute()
}
