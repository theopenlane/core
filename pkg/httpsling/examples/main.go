package main

import (
	"context"
	"log"
	"time"

	"github.com/theopenlane/core/pkg/httpsling"
)

// Post represents a simple structure to map JSON Placeholder posts
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

func main() {
	// Initialize the client with a base URL and a default timeout
	client := httpsling.Create(&httpsling.Config{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Timeout: 30 * time.Second,
	})

	// Perform a GET request to the /posts endpoint
	resp, err := client.Get("/posts/{post_id}").PathParam("post_id", "1").Send(context.Background())
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}

	// Decode the JSON response into our Post struct
	var post Post
	if err := resp.ScanJSON(&post); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Output the result
	log.Printf("Post Received: %+v\n", post)
}
