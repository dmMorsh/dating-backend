package server

import (
	"net/http"

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

    // Подключаем общие middleware как chi middlewares
    r.Use(middleware.ChiLoggingMiddleware)
    r.Use(middleware.ChiCORSMiddleware)

    // Публичные маршруты
    r.Group(func(r chi.Router) {
        r.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("pong"))
        }))
		
        r.Post("/register", http.HandlerFunc(handlers.RegisterHandler))
        r.Post("/login", 	http.HandlerFunc(handlers.LoginHandler))
        r.Post("/refresh", 	http.HandlerFunc(handlers.RefreshHandler))
		r.Get("/ws/chat", 	http.HandlerFunc(handlers.ChatWebSocketHandler))

    })

    // Защищенные маршруты — используем chi-group с chi-совместимым Auth middleware
    r.Group(func(r chi.Router) {
        r.Use(middleware.ChiAuthMiddleware)

		r.Handle("/logout", 		http.HandlerFunc(handlers.LogoutHandler))
		
        r.Get("/profiles", 			handlers.ProfilesHandler)
		r.Get("/me", 				http.HandlerFunc(handlers.GetMyProfileHandler))
		r.Put("/me", 				http.HandlerFunc(handlers.UpdateProfileHandler))
		r.Get("/user/{id}", 		http.HandlerFunc(handlers.GetUserHandler))
		r.Get("/followers", 		http.HandlerFunc(handlers.MyFollowersHandler))
		
		r.Post("/swipe", 			http.HandlerFunc(handlers.SwipeHandler))
		r.Get("/matches", 			http.HandlerFunc(handlers.MatchesHandler))
		r.Get("/recommendations", 	http.HandlerFunc(handlers.RecommendationsHandler))
		r.Get("/profiles/search", 	http.HandlerFunc(handlers.GetSwipeCandidatesHandler))
		r.Get("/clear/my/swipes", 	http.HandlerFunc(handlers.ClearMySwipesHandler))

		r.Get("/ws/start", 			http.HandlerFunc(handlers.StartWebSocketSession))
		r.Post("/messages/send", 	http.HandlerFunc(handlers.SendMessageHandler))
		r.Get("/chats", 			http.HandlerFunc(handlers.GetChatsHandler))
		r.Get("/chat/messages/{chatId}", 	http.HandlerFunc(handlers.GetChatMessagesHandler))
		r.Get("/chat/read/{chatId}", 	http.HandlerFunc(handlers.MarkChatMessagesAsReadHandler))
		r.Post("/messages/read", 	http.HandlerFunc(handlers.MarkMessagesReadHandler))
    })

    return r
}
