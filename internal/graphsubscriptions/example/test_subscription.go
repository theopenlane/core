//go:build example

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const (
	serverURL = "ws://localhost:17608/query"
)

var token = os.Getenv("API_TOKEN")

// GraphQL WebSocket protocol messages
type ConnectionInit struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type Subscribe struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func main() {
	// Set up interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create WebSocket connection with auth header
	header := http.Header{}
	header.Add("Authorization", "Bearer "+token)
	header.Add("Sec-WebSocket-Protocol", "graphql-transport-ws")

	log.Println("Connecting to", serverURL)

	conn, resp, err := websocket.DefaultDialer.Dial(serverURL, header)
	if err != nil {
		log.Printf("Dial error: %v", err)
		if resp != nil {
			log.Printf("Response status: %s", resp.Status)
		}
		return
	}
	defer conn.Close()

	log.Println("Connected! Sending connection_init...")

	// Send connection_init (graphql-ws protocol)
	initMsg := ConnectionInit{
		Type: "connection_init",
		Payload: map[string]interface{}{
			"Authorization": "Bearer " + token,
		},
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		log.Fatal("Failed to send connection_init:", err)
	}

	// Wait for connection_ack
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Failed to read connection_ack:", err)
	}
	log.Printf("Received: %s", string(msg))

	// Send subscription
	log.Println("Sending subscription request...")
	subMsg := Subscribe{
		ID:   "1",
		Type: "subscribe",
		Payload: map[string]interface{}{
			"query": `subscription NotificationCreated {
				notificationCreated {
					id
					title
					body
					userID
					createdAt
				}
			}`,
		},
	}
	if err := conn.WriteJSON(subMsg); err != nil {
		log.Fatal("Failed to send subscription:", err)
	}

	log.Println("Subscription sent! Waiting for notifications...")
	log.Println("(Create a task with assigneeID to trigger a notification)")

	// Read messages in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			var response map[string]interface{}
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Raw message: %s", string(message))
				continue
			}

			prettyJSON, _ := json.MarshalIndent(response, "", "  ")
			log.Printf("Received:\n%s", string(prettyJSON))

			if response["type"] == "next" {
				log.Println("NOTIFICATION RECEIVED!")
			}
		}
	}()

	// Keep alive ping
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteJSON(map[string]string{"type": "ping"}); err != nil {
				log.Println("Ping error:", err)
				return
			}
		case <-interrupt:
			log.Println("Interrupted, closing connection...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Close error:", err)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
