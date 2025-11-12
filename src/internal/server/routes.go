package server

import (
	"net/http"
	"time"

	handlers "dating-backend/internal/handlers"
	middleware "dating-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// NewRouter creates and returns a configured chi HTTP handler.
//
// It wires application routes and middleware. Public routes are registered
// without authentication; protected routes are grouped and require the
// `middleware.ChiAuthMiddleware` to set authenticated user id in the
// request context.
func NewRouter() http.Handler {
    r := chi.NewRouter()

	// Connecting common middleware as chi middlewares
	
	// r.Use(middleware.ChiRequestIDMiddleware)
	r.Use(middleware.ChiLoggingMiddleware)
	r.Use(middleware.ChiCORSMiddleware)

    // Public routes
    r.Group(func(r chi.Router) {
        r.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + " -=- pong"))
        }))
		
        r.Post("/register", http.HandlerFunc(handlers.RegisterHandler))
        r.Post("/login", 	http.HandlerFunc(handlers.LoginHandler))
        r.Post("/refresh", 	http.HandlerFunc(handlers.RefreshHandler))
		r.Get("/ws/chat", 	http.HandlerFunc(handlers.ChatWebSocketHandler))

    })

    // Protected routes - use chi-group with chi-compatible Auth middleware
    r.Group(func(r chi.Router) {
        r.Use(middleware.ChiAuthMiddleware)

		r.Delete("/clear/my/swipes",http.HandlerFunc(handlers.ClearMySwipesHandler))
		r.Post("/logout", 			http.HandlerFunc(handlers.LogoutHandler))
		
		r.Get("/me", 				http.HandlerFunc(handlers.GetMyProfileHandler))
		r.Put("/me", 				http.HandlerFunc(handlers.UpdateProfileHandler))
		r.Get("/user/{id}", 		http.HandlerFunc(handlers.GetUserHandler))
		r.Get("/followers", 		http.HandlerFunc(handlers.GetMyFollowersHandler))
		
		r.Post("/swipe", 			http.HandlerFunc(handlers.SwipeHandler))
		r.Get("/profiles/search", 	http.HandlerFunc(handlers.GetSwipeCandidatesHandler))

		r.Post("/ws/start", 		http.HandlerFunc(handlers.StartWebSocketSession))
		r.Post("/messages/send", 	http.HandlerFunc(handlers.SendMessageHandler))
		r.Post("/messages/read", 	http.HandlerFunc(handlers.MarkMessagesReadHandler))
		r.Get("/chats", 			http.HandlerFunc(handlers.GetChatsHandler))
		r.Post("/chat/read", 		http.HandlerFunc(handlers.MarkChatMessagesAsReadHandler))
		r.Get("/chat/messages/{chatId}", 	http.HandlerFunc(handlers.GetChatMessagesHandler))
    })

    return r
}
