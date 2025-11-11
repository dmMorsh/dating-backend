package logging

import (
	"context"
	"os"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

// internal context key type to avoid collisions
type ctxKey string

const requestIDKey ctxKey = "request_id"

// Init initializes the global sugared logger. Call once at startup.
// level is ignored currently but kept for future extension.
func Init() error {
    cfg := zap.NewProductionConfig()
    // respect environment override
    if os.Getenv("DEBUG") != "" {
        cfg = zap.NewDevelopmentConfig()
    }
    logger, err := cfg.Build()
    if err != nil {
        return err
    }
    Log = logger.Sugar()
    return nil
}

// Sync flushes buffered logs. Call before exit.
func Sync() {
    if Log == nil {
        return
    }
    _ = Log.Sync()
}

// ContextWithRequestID returns a new context carrying the provided request id.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext extracts request id from the context or empty string.
func RequestIDFromContext(ctx context.Context) string {
    if ctx == nil {
        return ""
    }
    v := ctx.Value(requestIDKey)
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return ""
}

// FromContext returns a sugared logger with request_id field when present in ctx.
// If no request id is present the global logger is returned.
func FromContext(ctx context.Context) *zap.SugaredLogger {
    if Log == nil {
        return nil
    }
    rid := RequestIDFromContext(ctx)
    if rid == "" {
        return Log
    }
    return Log.With("request_id", rid)
}
