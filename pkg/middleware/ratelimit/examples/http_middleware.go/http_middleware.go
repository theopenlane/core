//go:build examples

package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

// GetRemoteIP returns the remote IP address of the request
func GetRemoteIP(ipLookups []string, forwardedForIndexFromBehind int, r *http.Request) string {
	realIP := r.Header.Get("True-Client-IP")
	forwardedFor := r.Header.Get("X-Forwarded-For")

	for _, lookup := range ipLookups {
		if lookup == "RemoteAddr" {
			// 1. Cover the basic use cases for both ipv4 and ipv6
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// 2. Upon error, just return the remote addr.
				return r.RemoteAddr
			}

			return ip
		}

		if lookup == "X-Forwarded-For" && forwardedFor != "" {
			// X-Forwarded-For is potentially a list of addresses separated with ","
			parts := strings.Split(forwardedFor, ",")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}

			partIndex := len(parts) - 1 - forwardedForIndexFromBehind
			if partIndex < 0 {
				partIndex = 0
			}

			return parts[partIndex]
		}

		if lookup == "True-Client-IP" && realIP != "" {
			return realIP
		}
	}

	return ""
}

// rateLimitMiddleware is a middleware that rate limits the requests
func rateLimitMiddleware(rateLimiter *ratelimit.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			remoteIP := GetRemoteIP([]string{"X-Forwarded-For", "RemoteAddr", "True-Client-IP"}, 0, r)
			key := fmt.Sprintf("%s_%s_%s", remoteIP, r.URL.String(), r.Method)

			limitStatus, err := rateLimiter.Check(key)
			if err != nil {
				// if rate limit error then pass the request
				next.ServeHTTP(w, r)
			}

			if limitStatus.IsLimited {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}

			if err := rateLimiter.Inc(key); err != nil {
				log.Printf("could not increment key: %s", key)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hello(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func main() {
	windowSize := 1 * time.Minute
	dataStore := ratelimit.NewMapLimitStore(context.Background(), 2*windowSize, 10*time.Second)
	rateLimiter := ratelimit.New(dataStore, 5, windowSize)
	rateLimiterHandler := rateLimitMiddleware(rateLimiter)
	helloHandler := http.HandlerFunc(hello)
	http.Handle("/", rateLimiterHandler(helloHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
