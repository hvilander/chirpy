package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

const PORT_STR = ":8080"

type apiConfig struct {
	hitCount atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.hitCount.Add(1)
		next.ServeHTTP(w, req)
	})

}

func main() {
	fmt.Println("hello world")

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	apiConfig := apiConfig{}

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("Hits: %d", apiConfig.hitCount.Load())))
	})
	mux.HandleFunc("POST /reset", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		apiConfig.hitCount.Store(0)
	})

	fmt.Println("Server starting on", "http://localhost"+PORT_STR)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}
