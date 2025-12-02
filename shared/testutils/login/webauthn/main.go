//go:build examples

package main

import (
	"log"
	"net/http"
)

func main() {
	l := log.Default()

	origin := "http://localhost:5500"

	l.Printf("[INFO] register routes")
	// Serve the web files
	http.Handle("/", http.FileServer(http.Dir("./pkg/testutils/login/webauthn/")))

	l.Printf("[INFO] start server at %s", origin)
	if err := http.ListenAndServe(":5500", nil); err != nil {
		l.Println(err)
	}
}
