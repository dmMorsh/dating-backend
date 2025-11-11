package middleware

import (
	"net/http"

	"dating-backend/internal/logging"
)

// LoggingMiddleware logs each incoming HTTP request and then calls the next
// handler. Uses structured logging via internal/logging.
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Use context-aware logger so request_id (if present) is included.
        logger := logging.FromContext(r.Context())
        if logger == nil {
            logger = logging.Log
        }
        logger.Infow("http request", "remote", r.RemoteAddr, "method", r.Method, "path", r.URL.Path)
        next(w, r)
    }
}
