package graphapi_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
)

// TestNotificationCreated tests the notificationCreated subscription
// to ensure that we handle the websocket connection correctly
// this does not test the actual notification sending logic
// as that is covered in unit tests for the notification service
func TestNotificationCreated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		testUser    testUserDetails
		expectError bool
	}{
		{
			name:        "happy path",
			testUser:    testUser1,
			expectError: false,
		},
		{
			name: "invalid auth",
			testUser: testUserDetails{
				ID: "nonexistent-user-id",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(newTestGraphServer(t))
			defer srv.Close()

			u, _ := url.Parse(srv.URL)
			u.Scheme = "ws"
			u.Path = "/query"

			ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			assert.NilError(t, err)

			claims := &tokens.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: tc.testUser.ID,
				},
				UserID: tc.testUser.ID,
				OrgID:  tc.testUser.OrganizationID,
			}

			access, _, err := suite.client.db.TokenManager.CreateTokenPair(claims)
			assert.NilError(t, err)

			// Send connection_init
			initMsg := map[string]interface{}{
				"type": "connection_init",
				"payload": map[string]interface{}{
					"Authorization": fmt.Sprintf("Bearer %s", access),
				},
			}
			err = ws.WriteJSON(initMsg)
			assert.NilError(t, err)

			_, ackMsg, err := ws.ReadMessage()
			assert.NilError(t, err)
			assert.Assert(t, string(ackMsg) != "", "Expected connection_ack message")

			// Send the subscription start message
			startMsg := map[string]interface{}{
				"id":   "1",
				"type": "start",
				"payload": map[string]interface{}{
					"query": "subscription { notificationCreated { id } }",
				},
			}
			err = ws.WriteJSON(startMsg)
			assert.NilError(t, err)

			// Read messages until we get a non-keep-alive message or an error
			for {
				msgType, msg, err := ws.ReadMessage()
				// we close early for error cases
				// so we expect an error here if expectError is true
				if tc.expectError {
					assert.ErrorContains(t, err, "close")
					return
				}

				assert.NilError(t, err)
				if string(msg) == "" {
					continue
				}

				// check if it's a keep-alive message
				// so we can continue reading
				if msgType == websocket.PingMessage {
					continue // skip keep-alive
				}

				// We received a non-keep-alive message, exit the loop
				break
			}

			// initiate close handshake
			err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "test complete"))
			assert.NilError(t, err)

			// Wait for server to close the connection or context to expire
			done := make(chan struct{})
			var readErr error
			go func() {
				defer close(done)
				_, _, readErr = ws.ReadMessage() // expecting a close error
			}()

			select {
			case <-TestAfterCancel:
				t.Fatalf("Non nil-response after context cancelled, this will cause an infinite loop if not fixed")
			case <-done:
				assert.ErrorContains(t, readErr, "close")

				// ensure no late messages are received after close
				// from test extension to signal bad behavior
				select {
				case <-TestAfterCancel:
					t.Fatalf("Non nil-response after context cancelled (late), this will cause an infinite loop if not fixed")
				case <-time.After(1 * time.Second):
					// no late messages, test passes
				}
			}
		})
	}
}
