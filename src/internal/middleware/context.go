package middleware

import (
	"context"
	"errors"
)

// UserIDFromContext extracts the authenticated user id from the provided
// context. It returns an error if no user id is present or the type is wrong.
func UserIDFromContext(ctx context.Context) (int64, error) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, errors.New("no user in context")
	}
	id, ok := v.(int64)
	if !ok {
		return 0, errors.New("invalid user id type")
	}
	return id, nil
}
