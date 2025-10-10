package main

import (
	"net/http"

	handlers "dating-backend/handlers"
	middleware "dating-backend/middleware"
)

func registerRoutes() {
	// Общие middleware (например логирование и CORS)
	common := []func(http.HandlerFunc) http.HandlerFunc{
		middleware.LoggingMiddleware,
		middleware.CORSMiddleware,
	}

	// Middleware для защищённых маршрутов
	protected := append(common, middleware.AuthMiddleware)

	// Публичные маршруты
	http.HandleFunc("/register", middleware.Chain(handlers.RegisterHandler, common...))
	http.HandleFunc("/login", middleware.Chain(handlers.LoginHandler, common...))
	http.HandleFunc("/refresh", middleware.Chain(handlers.RefreshHandler, common...))

	// Защищённые маршруты
	http.HandleFunc("/logout", middleware.Chain(handlers.LogoutHandler, protected...))
	http.HandleFunc("/profiles", middleware.Chain(handlers.ProfilesHandler, protected...))
	http.HandleFunc("/like/", middleware.Chain(handlers.LikeHandler, protected...))
	http.HandleFunc("/swipe", middleware.Chain(handlers.SwipeHandler, protected...))
	http.HandleFunc("/matches", middleware.Chain(handlers.MatchesHandler, protected...))
	http.HandleFunc("/recommendations", middleware.Chain(handlers.RecommendationsHandler, protected...))

	http.HandleFunc("/myprofile", middleware.Chain(handlers.GetMyProfileHandler, protected...))
	http.HandleFunc("PUT /me", middleware.Chain(handlers.UpdateProfileHandler, protected...))
	http.HandleFunc("/user/", middleware.Chain(handlers.GetUserHandler, protected...))

	http.HandleFunc("/messages/send", middleware.Chain(handlers.SendMessageHandler, protected...))
	http.HandleFunc("/messages", middleware.Chain(handlers.GetMessagesHandler, protected...))
	http.HandleFunc("/ws/chat", middleware.Chain(handlers.ChatWebSocketHandler, protected...))

}
