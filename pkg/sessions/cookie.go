package sessions

import (
	"encoding/base64"
	"net/http"
	"time"
)

const (
	defaultMaxAgeSeconds = 60 * 60 // 1 hour (in seconds)
)

var (
	DefaultCookieName = "__Secure-SessionId"
	DevCookieName     = "temporary-cookie"
)

// DefaultCookieConfig configures http.Cookie creation for production (AKA default secure values are set)
var DefaultCookieConfig = &CookieConfig{
	Path:     "/",
	Domain:   "",
	MaxAge:   defaultMaxAgeSeconds,
	HTTPOnly: true,
	Secure:   true,
	SameSite: http.SameSiteStrictMode,
}

// DebugCookieConfig configures http.Cookie creation for debugging
var DebugCookieConfig = &CookieConfig{
	Path:     "/",
	MaxAge:   defaultMaxAgeSeconds,
	HTTPOnly: true,
	Secure:   false, // allow to work over http
	SameSite: http.SameSiteLaxMode,
}

// DebugOnlyCookieConfig is different in that it's not a receiver and the name is set, so it can be called directly
var DebugOnlyCookieConfig = CookieConfig{
	Name:     DevCookieName,
	Path:     "/",
	MaxAge:   defaultMaxAgeSeconds,
	HTTPOnly: true,
	Secure:   false, // allow to work over http
	SameSite: http.SameSiteLaxMode,
}

// CookieConfig configures http.Cookie creation
type CookieConfig struct {
	Name string
	// Cookie domain/path scope (leave zeroed for requested resource scope)
	// Defaults to the domain name of the responding server when unset
	Domain string
	// Defaults to the path of the responding URL when unset
	Path string
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge int
	// cookie may only be transferred over HTTPS. Recommend true
	Secure bool
	// browser should prohibit non-HTTP (i.e. javascript) cookie access. Recommend true
	HTTPOnly bool
	// prohibit sending in cross-site requests with SameSiteLaxMode or SameSiteStrictMode
	SameSite http.SameSite
}

// NewCookie returns a new chocolate chip http.Cookie with the given name, value, and properties from config
func NewCookie(name, value string, config *CookieConfig) *http.Cookie {
	if name != "" {
		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Path:     config.Path,
			Domain:   config.Domain,
			MaxAge:   config.MaxAge,
			HttpOnly: config.HTTPOnly,
			Secure:   config.Secure,
			SameSite: config.SameSite,
		}

		if expires, ok := expiresTime(config.MaxAge); ok {
			cookie.Expires = expires
		}

		return cookie
	}

	if name == "" {
		cookie := &http.Cookie{
			Name:     config.Name,
			Value:    value,
			Path:     config.Path,
			Domain:   config.Domain,
			MaxAge:   config.MaxAge,
			HttpOnly: config.HTTPOnly,
			Secure:   config.Secure,
			SameSite: config.SameSite,
		}

		if expires, ok := expiresTime(config.MaxAge); ok {
			cookie.Expires = expires
		}

		return cookie
	}

	return nil
}

// expiresTime converts a maxAge time in seconds to a time.Time in the future
// ref http://golang.org/src/net/http/cookie.go?s=618:801#L23
func expiresTime(maxAge int) (time.Time, bool) {
	if maxAge > 0 {
		d := time.Duration(maxAge) * time.Second
		return time.Now().Add(d), true
	} else if maxAge < 0 {
		return time.Unix(1, 0), true
	}

	return time.Time{}, false
}

// GetCookie function retrieves a specific cookie from an HTTP request
func GetCookie(r *http.Request, cookieName string) (*http.Cookie, error) {
	return r.Cookie(cookieName)
}

// CookieExpired checks to see if a cookie is expired
func CookieExpired(cookie *http.Cookie) bool {
	if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
		return true
	}

	if cookie.MaxAge < 0 {
		return true
	}

	return false
}

// SetCookieB64 function sets a base64-encoded cookie with the given name and value in the HTTP response
func SetCookieB64(w http.ResponseWriter, body []byte, cookieName string, v CookieConfig) string {
	cookieValue := base64.StdEncoding.EncodeToString(body)
	// set the cookie
	SetCookie(w, cookieValue, cookieName, v)

	return cookieValue
}

// SetCookie function sets a cookie with the given value and name
func SetCookie(w http.ResponseWriter, value string, cookieName string, v CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Domain:   v.Domain,
		Path:     v.Path,
		MaxAge:   v.MaxAge,
		Secure:   v.Secure,
		SameSite: v.SameSite,
		HttpOnly: v.HTTPOnly,
	})
}

// RemoveCookie function removes a cookie from the HTTP response
func RemoveCookie(w http.ResponseWriter, cookieName string, v CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Domain:   v.Domain,
		Path:     v.Path,
		MaxAge:   -1,
		Secure:   v.Secure,
		SameSite: v.SameSite,
		HttpOnly: v.HTTPOnly,
	})
}
