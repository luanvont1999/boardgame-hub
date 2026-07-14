package handler

import (
	"net/http"

	"boardgame-hub/backend/pkg/server"
)

var mux = server.NewRouter()

func Handler(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}
