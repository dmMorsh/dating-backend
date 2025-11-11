package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"dating-backend/internal/logging"
)

// RequestIDMiddleware is a handler-style middleware (func(http.HandlerFunc) http.HandlerFunc)
// that ensures each request has a unique request id. It sets the "X-Request-ID"
// response header and stores the id into the request context so downstream
// code can enrich logs using logging.FromContext(ctx).
func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // honor incoming header if provided
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            b := make([]byte, 16)
            _, _ = rand.Read(b)
            reqID = hex.EncodeToString(b)
        }

        // set response header for clients
        w.Header().Set("X-Request-ID", reqID)

        // attach to context for logging helpers
        ctx := logging.ContextWithRequestID(r.Context(), reqID)

        // call next with updated context
        next(w, r.WithContext(ctx))
    }
}
