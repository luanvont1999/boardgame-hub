package main

import (
	"log"
	"net/http"

	"boardgame-hub/backend/pkg/server"
)

func main() {
	mux := server.NewRouter()

	log.Println("Server Go local đang khởi chạy tại http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
