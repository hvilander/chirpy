package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
)

const PORT_STR = ":8080"

type apiConfig struct {
	hitCount atomic.Int32
}

type jsonRes struct {
	Error       string `json:"error"`
	Valid       bool   `json:"valid"`
	Body        string `json:"body"`
	CleanedBody string `json:"cleaned_body"`
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

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(" <html> <body> <h1>Welcome, Chirpy Admin</h1> <p>Chirpy has been visited %d times!</p> </body> </html> ", apiConfig.hitCount.Load())))
	})
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		apiConfig.hitCount.Store(0)
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := jsonRes{}
		err := decoder.Decode(&params)
		if err != nil {
			fmt.Println("error decoding params:", err)
			w.WriteHeader(500)
			return
		}

		chirp := params.Body

		if len(chirp) > 140 {
			respondWithError(w, 400, "Chirp is too long")
		}

		profane := []string{"kerfuffle", "sharbert", "fornax"}

		words := strings.Split(chirp, " ")
		fmt.Println(words)

		for i, w := range words {
			if slices.Contains(profane, strings.ToLower(w)) {
				words[i] = "****"
			}
		}

		cleaned := strings.Join(words, " ")
		respondWithJson(w, 200, jsonRes{CleanedBody: cleaned})

	})

	fmt.Println("Server starting on", "http://localhost"+PORT_STR)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	dat, err := json.Marshal(jsonRes{Error: msg})
	if err != nil {
		fmt.Println("error marshaling:", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error marshaling:", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
