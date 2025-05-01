package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	cfg.hitCount.Store(0)

	cfg.db.ResetUsers(req.Context())

	w.WriteHeader(200)
}
