package middleware

import (
	"net/http"
)

// Adapter converts existing handler-style middleware (func(http.HandlerFunc) http.HandlerFunc)
// into chi-compatible middleware (func(http.Handler) http.Handler).
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

// Add a small convenience to expose common adapters
var (
    ChiLoggingMiddleware = Adapter(LoggingMiddleware)
    ChiCORSMiddleware     = Adapter(CORSMiddleware)
    // Expose chi-compatible auth middleware as well.
    ChiAuthMiddleware = Adapter(AuthMiddleware)
)
