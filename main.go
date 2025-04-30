package main

import (
	"encoding/json"
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
		fmt.Println("hi")
		type jsonBody struct {
			Body string `json:"body"`
		}

		type jsonRes struct {
			Error string `json:"error"`
			Valid bool   `json:"valid"`
		}

		decoder := json.NewDecoder(req.Body)
		params := jsonBody{}
		err := decoder.Decode(&params)
		if err != nil {
			fmt.Println("error decoding params:", err)
			w.WriteHeader(500)
			return
		}

		if len(params.Body) > 140 {
			dat, err := json.Marshal(jsonRes{Error: "Chirp is too long"})
			if err != nil {
				fmt.Println("error marshaling:", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			w.Write(dat)
		} else {
			dat, err := json.Marshal(jsonRes{Valid: true})
			if err != nil {
				fmt.Println("error marshaling:", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
			w.Write(dat)
		}

	})

	fmt.Println("Server starting on", "http://localhost"+PORT_STR)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}
