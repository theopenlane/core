package sessions

import (
	"net/http"
	"net/url"

	"github.com/gorilla/securecookie"
)

const (
	UserIDKey         = "userID"
	ExternalUserIDKey = "externalUserID"
	SessionNameKey    = "name"
	UserTypeKey       = "userType"
	UsernameKey       = "username"
	EmailKey          = "email"
	WebAuthnKey       = "webauthn"
)

type Store[T any] interface {
	// New returns a new named Session
	New(name string) *Session[T]
	// Get a named Session from the request
	Get(req *http.Request, name string) (*Session[T], error)
	// Save writes a Session to the ResponseWriter
	Save(w http.ResponseWriter, session *Session[T]) error
	// Destroy removes (expires) a named Session
	Destroy(w http.ResponseWriter, name string)
	// GetSessionIDFromCookie returns the key, which should be the sessionID, in the map
	GetSessionIDFromCookie(sess *Session[T]) string
	// GetSessionDataFromCookie returns the value stored map
	GetSessionDataFromCookie(sess *Session[T]) any
	// EncodeCookie encodes the cookie
	EncodeCookie(session *Session[T]) (string, error)
}

var _ Store[any] = &cookieStore[any]{}

// cookieStore stores Sessions in secure cookies (i.e. client-side)
type cookieStore[T any] struct {
	config *CookieConfig
	// encodes and decodes signed and optionally encrypted cookie values
	codecs []securecookie.Codec
}

// NewCookieStore returns a new Store that signs and optionally encrypts
// session state in http cookies.
func NewCookieStore[T any](config *CookieConfig, keyPairs ...[]byte) Store[T] {
	if config == nil {
		config = DefaultCookieConfig
	}

	return &cookieStore[T]{
		config: config,
		codecs: securecookie.CodecsFromPairs(keyPairs...),
	}
}

// New returns a new named Session
func (s *cookieStore[T]) New(name string) *Session[T] {
	return NewSession[T](s, name)
}

// Get returns the named Session from the Request. Returns an error if the
// session cookie cannot be found, the cookie verification fails, or an error
// occurs decoding the cookie value.
func (s *cookieStore[T]) Get(req *http.Request, name string) (session *Session[T], err error) {
	cookie, err := req.Cookie(name)
	if err == nil {
		// decode the session cookie, UIs like to encode the cookie value
		decodedSession, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			return nil, err
		}

		session = s.New(name)
		if err = securecookie.DecodeMulti(name, decodedSession, &session.values, s.codecs...); err != nil {
			return nil, err
		}
	}

	return session, err
}

// GetSessionIDFromCookie gets the cookies from the http.Request and gets the key (session ID) from the values
func (s *cookieStore[T]) GetSessionIDFromCookie(sess *Session[T]) string {
	for k := range sess.values {
		return k
	}

	return ""
}

// GetSessionDataFromCookie gets the cookies from the http.Request and gets session values
func (s *cookieStore[T]) GetSessionDataFromCookie(sess *Session[T]) any {
	for _, v := range sess.values {
		return v
	}

	return ""
}

// Save adds or updates the Session on the response via a signed and optionally
// encrypted session cookie. Session Values are encoded into the cookie value
// and the session Config sets cookie properties.
func (s *cookieStore[T]) Save(w http.ResponseWriter, session *Session[T]) error {
	cookieValue, err := securecookie.EncodeMulti(session.Name(), &session.values, s.codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, NewCookie(session.Name(), cookieValue, s.config))

	return nil
}

func (s *cookieStore[T]) EncodeCookie(session *Session[T]) (string, error) {
	return securecookie.EncodeMulti(session.Name(), &session.values, s.codecs...)
}

// Destroy deletes the Session with the given name by issuing an expired
// session cookie with the same name.
func (s *cookieStore[T]) Destroy(w http.ResponseWriter, name string) {
	http.SetCookie(w, NewCookie(name, "", &CookieConfig{MaxAge: -1, Path: s.config.Path}))
}

// NewSessionCookie creates a cookie from a session id
func NewSessionCookie(session string) *http.Cookie {
	return NewCookie(DefaultCookieName, session, DefaultCookieConfig)
}

// NewDevSessionCookie creates a cookie from a session id using the dev cookie name
func NewDevSessionCookie(session string) *http.Cookie {
	return NewCookie(DevCookieName, session, DefaultCookieConfig)
}
