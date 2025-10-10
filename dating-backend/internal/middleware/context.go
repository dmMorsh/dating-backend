package middleware

import (
	"context"
	"errors"
)

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
