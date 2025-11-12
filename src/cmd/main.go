package main

import (
	"net/http"
	"os"

	data_access "dating-backend/internal/data-access"
	"dating-backend/internal/logging"
	"dating-backend/internal/realtime"
	server "dating-backend/internal/server"

	"github.com/redis/go-redis/v9"
)

func main() {
	// initialize structured logging
	if err := logging.Init(); err != nil {
		// fallback: panic so the operator notices
		panic(err)
	}
	defer logging.Sync()

	data_access.InitDB()
	mux := server.NewRouter()

	// Optionally use Redis for session tokens. Set REDIS_ADDR env var
	// (for example: "localhost:6379") to enable Redis-backed session store.
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		realtime.DefaultSessionStore = realtime.NewRedisSessionStore(&redis.Options{Addr: addr})
		logging.Log.Infof("using Redis session store at %s", addr)
	}
	realtime.StartPingLoop()

	logging.Log.Infow("server starting", "addr", ":8088")
	if err := http.ListenAndServe(":8088", mux); err != nil {
		logging.Log.Fatalw("server exited", "err", err)
	}
}