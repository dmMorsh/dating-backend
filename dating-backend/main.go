package main

import (
	"log"
	"net/http"

	data_access "dating-backend/data-access"
)


func main() {
	data_access.InitDB()
	registerRoutes()

	log.Println("Server running on :8088")
	log.Fatal(http.ListenAndServe(":8088", nil))
}