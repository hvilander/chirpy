package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type postUserBody struct {
	Email string `json:"email"`
}

type userCreatedRes struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	platform := os.Getenv("PLATFORM")

	if platform != "dev" {
		respondWithError(w, 403, "DOH")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := postUserBody{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("error decoding params:", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error creating user")
	}

	responseBody := userCreatedRes{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt.Time.String(),
		UpdatedAt: user.UpdatedAt.Time.String(),
		Email:     user.Email,
	}

	respondWithJson(w, 201, responseBody)
}
