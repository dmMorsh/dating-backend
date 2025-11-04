package middleware

import (
	"net/http"
)

// Adapter converts existing handler-style middleware (func(http.HandlerFunc) http.HandlerFunc)
// into chi-compatible middleware (func(http.Handler) http.Handler).
// Adapter converts an existing handler-style middleware (func(http.HandlerFunc)http.HandlerFunc)
// into a chi-compatible middleware (func(http.Handler) http.Handler).
//
// This lets us reuse existing middleware implementations without rewriting
// them for chi. The returned middleware wraps the provided next handler and
// calls the original middleware with an http.HandlerFunc adapter.
func Adapter(hm func(http.HandlerFunc) http.HandlerFunc) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // wrap next.ServeHTTP into http.HandlerFunc and pass into existing middleware
            wrapped := func(w http.ResponseWriter, r *http.Request) {
                next.ServeHTTP(w, r)
            }
            hm(wrapped).ServeHTTP(w, r)
        })
    }
}

// Chi* variables are convenience chi-compatible middleware wrappers for the
// existing handler-style middleware defined in this package.
var (
    // ChiLoggingMiddleware logs requests using the existing LoggingMiddleware.
    ChiLoggingMiddleware = Adapter(LoggingMiddleware)

    // ChiCORSMiddleware applies CORS headers and handling using CORSMiddleware.
    ChiCORSMiddleware = Adapter(CORSMiddleware)

    // ChiAuthMiddleware performs authorization checks and injects user id into
    // the request context using the existing AuthMiddleware implementation.
    ChiAuthMiddleware = Adapter(AuthMiddleware)
)
