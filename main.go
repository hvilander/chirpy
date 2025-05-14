package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	dbmod "github.com/hvilander/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const PORT_STR = ":8080"

type apiConfig struct {
	hitCount atomic.Int32
	db       *dbmod.Queries
	secret   string
	polkaKey string
}

type jsonRes struct {
	Error       string `json:"error"`
	Valid       bool   `json:"valid"`
	Body        string `json:"body"`
	CleanedBody string `json:"cleaned_body"`
}

func main() {
	fmt.Println("hello world")
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	SECRET := os.Getenv("SECRET")
	POLKA := os.Getenv("POLKA_KEY")
	database, err := sql.Open("postgres", dbURL)
	apiConfig := apiConfig{}

	apiConfig.db = dbmod.New(database)
	apiConfig.secret = SECRET
	apiConfig.polkaKey = POLKA

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /api/users", apiConfig.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiConfig.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiConfig.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handlerGetChirpByID)
	mux.HandleFunc("POST /api/login", apiConfig.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiConfig.handleRefresh)
	mux.HandleFunc("POST /api/revoke", apiConfig.handleRevoke)
	mux.HandleFunc("PUT /api/users", apiConfig.handlePutUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiConfig.handleChripyRedWebHook)

	fmt.Println("Server starting on", "http://localhost"+PORT_STR)

	err = server.ListenAndServe()
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
