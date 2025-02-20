package main

import (
	"github.com/theopenlane/core/pkg/sleuth/static"
)

func main() {
	items, err := static.NewSecListFromEmbeddedFile("mitb", "well-known.txt")
	if err != nil {
		panic(err)
	}

	for _, item := range items.Items {
		println(item)
	}
}
