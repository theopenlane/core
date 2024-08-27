package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/sessions"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		sessionName  string
		expectedName string
	}{
		{
			name:         "happy path",
			sessionName:  "huddle",
			expectedName: "huddle",
		},
		{
			name:         "empty name, use default",
			sessionName:  "",
			expectedName: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// test with a string first
			cs := sessions.NewCookieStore[string](sessions.DefaultCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			session := cs.New(tc.sessionName)

			assert.Equal(t, tc.expectedName, session.Name())

			// Again, with a string map
			csMap := sessions.NewCookieStore[map[string]string](sessions.DefaultCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			sessionMap := csMap.New(tc.sessionName)

			assert.Equal(t, tc.expectedName, sessionMap.Name())
		})
	}
}

func TestNewSessionCookie(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		session     string
		userID      string
	}{
		{
			name:        "happy path",
			sessionName: "__Host-meow",
			session:     ulids.New().String(),
			userID:      ulids.New().String(),
		},
		{
			name:        "empty string still results in session",
			sessionName: "__Host-woof",
			session:     "",
			userID:      "",
		},
		{
			name:        "empty session name",
			sessionName: "__Host-meow",
			session:     "",
			userID:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// test with a string first
			cs := sessions.NewCookieStore[string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			cooky := cs.New(tc.name)

			cooky.Set(sessions.SessionNameKey, tc.name)
			cooky.Set("session", tc.session)

			name := cooky.Get(sessions.SessionNameKey)
			session := cooky.Get("session")

			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.session, session)

			// Again, with a string map
			csMap := sessions.NewCookieStore[map[string]string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			cookyMap := csMap.New(tc.sessionName)

			cookyMap.Set(tc.session,
				map[string]string{
					sessions.SessionNameKey: tc.sessionName,
					sessions.UserIDKey:      tc.userID,
				},
			)

			assert.Equal(t, tc.sessionName, cookyMap.Get(tc.session)[sessions.SessionNameKey])
			assert.Equal(t, tc.userID, cookyMap.Get(tc.session)[sessions.UserIDKey])
		})
	}
}

func TestSaveGet(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		session     string
		userID      string
	}{
		{
			name:        "happy path",
			sessionName: "__Host-meow",
			session:     ulids.New().String(),
			userID:      "mitb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs := sessions.NewCookieStore[map[string]string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			session := sessions.NewSession(cs, tc.sessionName)

			setSessionMap := map[string]string{}
			setSessionMap[sessions.UserIDKey] = tc.userID
			setSessionMap[sessions.SessionNameKey] = tc.sessionName

			session.Set(tc.session, setSessionMap)

			err := cs.Save(recorder, session)
			require.NoError(t, err)

			// Copy the Cookie over to a new Request
			res := recorder.Result()
			defer res.Body.Close()

			cooky := res.Header["Set-Cookie"]
			request := &http.Request{Header: http.Header{"Cookie": cooky}}

			sess, err := cs.Get(request, tc.sessionName)
			require.NoError(t, err)
			assert.Equal(t, tc.session, sess.GetKey())
			assert.Equal(t, tc.sessionName, sess.Get(sess.GetKey())[sessions.SessionNameKey])
		})
	}
}

func TestGetSessionIDFromCookie(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		userID      string
	}{
		{
			name:        "happy path",
			sessionName: "__Host-meow",
			userID:      "mitb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs := sessions.NewCookieStore[map[string]string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			session := cs.New(tc.sessionName)
			sessionID := sessions.GenerateSessionID()

			setSessionMap := map[string]string{}
			setSessionMap[sessions.UserIDKey] = tc.userID
			setSessionMap[sessions.SessionNameKey] = tc.sessionName

			session.Set(sessionID, setSessionMap)

			err := cs.Save(recorder, session)
			require.NoError(t, err)

			// Copy the Cookie over to a new Request
			res := recorder.Result()
			defer res.Body.Close()

			cooky := res.Header["Set-Cookie"]
			request := &http.Request{Header: http.Header{"Cookie": cooky}}

			session, err = cs.Get(request, tc.sessionName)
			require.NoError(t, err)

			id := cs.GetSessionIDFromCookie(session)

			require.Equal(t, sessionID, id)
		})
	}
}

func TestGetSessionDataFromCookie(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		userID      string
	}{
		{
			name:        "happy path",
			sessionName: "__Host-meow",
			userID:      "mitb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs := sessions.NewCookieStore[map[string]string](sessions.DebugCookieConfig,
				[]byte("my-signing-secret"), []byte("encryptionsecret"))

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			session := cs.New(tc.sessionName)
			sessionID := sessions.GenerateSessionID()

			setSessionMap := map[string]string{}
			setSessionMap[sessions.UserIDKey] = tc.userID
			setSessionMap[sessions.SessionNameKey] = tc.sessionName

			session.Set(sessionID, setSessionMap)

			err := cs.Save(recorder, session)
			require.NoError(t, err)

			// Copy the Cookie over to a new Request
			res := recorder.Result()
			defer res.Body.Close()

			cooky := res.Header["Set-Cookie"]
			request := &http.Request{Header: http.Header{"Cookie": cooky}}

			session, err = cs.Get(request, tc.sessionName)
			require.NoError(t, err)

			sd := cs.GetSessionDataFromCookie(session)

			require.Equal(t, setSessionMap, sd)
		})
	}
}
