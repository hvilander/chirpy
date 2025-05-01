package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	cfg.hitCount.Store(0)
}
