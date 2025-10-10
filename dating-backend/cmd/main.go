package main

import (
	"log"
	"net/http"

	data_access "dating-backend/internal/data-access"
	server "dating-backend/internal/server"
)


func main() {
	data_access.InitDB()
	mux := server.NewRouter()

	log.Println("Server running on :8088")
	log.Fatal(http.ListenAndServe(":8088", mux))
}