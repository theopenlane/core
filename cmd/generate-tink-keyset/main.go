package main

import (
	"fmt"
	"log"
	"os"

	"github.com/theopenlane/core/internal/ent/hooks"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println("Usage: generate-tink-keyset")
		fmt.Println("Generates a new Tink keyset for field encryption")
		fmt.Println("")
		fmt.Println("Set the output as OPENLANE_TINK_KEYSET environment variable")
		return
	}

	// Generate a new keyset using the hooks package function
	keyset, err := hooks.GenerateTinkKeyset()
	if err != nil {
		log.Fatalf("Failed to generate keyset: %v", err)
	}

	fmt.Printf("Generated Tink keyset:\n")
	fmt.Printf("OPENLANE_TINK_KEYSET=%s\n", keyset)
	fmt.Printf("\nTo use this keyset:\n")
	fmt.Printf("export OPENLANE_TINK_KEYSET=%s\n", keyset)
}

