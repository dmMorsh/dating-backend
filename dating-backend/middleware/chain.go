package middleware

import "net/http"

// Chain применяет несколько middleware к одному обработчику.
// Middleware вызываются в порядке их перечисления.
func Chain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}