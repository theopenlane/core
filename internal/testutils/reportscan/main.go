//go:build examples

package main

import (
	"log"
	"net/http"
)

func main() {
	l := log.Default()

	// 5500 is already an allowed cors origin in the local config
	origin := "http://localhost:5500/reportscan/"

	l.Printf("[INFO] register routes")
	// serve the whole testutils dir so the shared bootstrap stylesheet resolves
	http.Handle("/", http.FileServer(http.Dir("./internal/testutils/")))

	l.Printf("[INFO] start server at %s", origin)
	if err := http.ListenAndServe(":5500", nil); err != nil {
		l.Println(err)
	}
}
