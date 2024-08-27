package sessions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/sessions"
)

func TestSet(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string

		userID  string
		session string
	}{
		{
			name:        "happy path",
			sessionName: "__Secure-SessionId",
			userID:      "01HMDBSNBGH4DTEP0SR8118Y96",
			session:     ulids.New().String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// with a string first
			cs := sessions.NewCookieStore[string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			session := cs.New(tc.sessionName)

			// Set sessions
			session.Set(tc.userID, tc.session)

			assert.Equal(t, tc.session, session.Get(tc.userID))

			// Again, with a string map
			csMap := sessions.NewCookieStore[map[string]string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			sessionMap := csMap.New(tc.sessionName)

			// Set sessions
			sessionMap.Set(tc.session, map[string]string{sessions.UserIDKey: tc.userID})

			assert.Equal(t, tc.userID, sessionMap.Get(tc.session)[sessions.UserIDKey])
		})
	}
}

func TestGetOk(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		userID      string
		session     string
	}{
		{
			name:        "happy path",
			sessionName: "__Secure-SessionId",
			userID:      "01HMDBSNBGH4DTEP0SR8118Y96",
			session:     ulids.New().String(),
		},
		{
			name:        "another session name",
			sessionName: "MeOWzErZ!",
			userID:      ulids.New().String(),
			session:     "01HMDBSNBGH4DTEP0SR8118Y96",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// test with a string
			cs := sessions.NewCookieStore[string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			s := cs.New(tc.sessionName)

			s.Set(sessions.UserIDKey, tc.userID)
			s.Set("session", tc.session)

			uID, ok := s.GetOk(sessions.UserIDKey)
			assert.True(t, ok)

			sess, ok := s.GetOk("session")
			assert.True(t, ok)

			assert.Equal(t, tc.userID, uID)
			assert.Equal(t, tc.session, sess)

			// Again, but with a string map this time
			csMap := sessions.NewCookieStore[map[string]string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			sMap := csMap.New(tc.sessionName)
			sMap.Set(tc.session, map[string]string{sessions.UserIDKey: tc.userID})

			sessMap, ok := sMap.GetOk(tc.session)
			assert.True(t, ok)
			assert.Equal(t, tc.userID, sessMap[sessions.UserIDKey])
		})
	}
}
