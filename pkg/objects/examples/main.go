package main

import (
	"context"
	"log"
	"os"

	"github.com/theopenlane/core/pkg/objects/examples/app"
)

func main() {
	cmd := app.NewCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
