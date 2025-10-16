package server

import (
	"net/http"

	handlers "dating-backend/internal/handlers"
	middleware "dating-backend/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// NewRouter создает и возвращает настроенный chi.Router
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
		
		r.Post("/swipe", 			http.HandlerFunc(handlers.SwipeHandler))
		r.Get("/matches", 			http.HandlerFunc(handlers.MatchesHandler))
		r.Get("/recommendations", 	http.HandlerFunc(handlers.RecommendationsHandler))

		r.Post("/messages/send", 	http.HandlerFunc(handlers.SendMessageHandler))
		r.Get("/messages", 			http.HandlerFunc(handlers.GetMessagesHandler))
		r.Get("/chats", 			http.HandlerFunc(handlers.GetChatsHandler))
		r.Get("/ws/start", 			http.HandlerFunc(handlers.StartWebSocketSession))
		// r.Get("/ws/chat", 			http.HandlerFunc(handlers.ChatWebSocketHandler))
		r.Get("/chat/messages/{chatId}", 	http.HandlerFunc(handlers.GetChatMessagesHandler))
    })

    return r
}
