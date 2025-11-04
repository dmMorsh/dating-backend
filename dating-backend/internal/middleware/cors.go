package middleware

import (
	"net/http"
)

// CORSMiddleware applies permissive CORS headers and handles OPTIONS
// preflight requests. It's intentionally permissive (Access-Control-Allow-Origin: *).
// Adjust for production to restrict origins if necessary.
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // If this is an OPTIONS preflight request, respond immediately.
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next(w, r)
    }
}